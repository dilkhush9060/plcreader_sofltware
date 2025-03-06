package main

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/goburrow/modbus"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx              context.Context
	client           modbus.Client
	isModbusConnected bool
}

type Config struct {
	PlantID string `json:"plantId"`
	COMPort string `json:"comPort"`
}

type BoilerData struct {
	ID                 int    `json:"id"`
	ReactorTemp        int    `json:"reactorTemp"`
	SeparatorTemp      int    `json:"separatorTemp"`
	FurnaceTemp        int    `json:"furnaceTemp"`
	CondenserTemp      int    `json:"condenserTemp"`
	AtmTemp            int    `json:"atmTemp"`
	ReactorPressure    int    `json:"reactorPressure"`
	GasTankPressure    int    `json:"gasTankPressure"`
	ProcessStartTime   string `json:"processStartTime"`
	TimeOfReaction     string `json:"timeOfReaction"`
	ProcessEndTime     string `json:"processEndTime"`
	CoolingEndTime     string `json:"coolingEndTime"`
	NitrogenPurging    int    `json:"nitrogenPurging"`
	CarbonDoorStatus   int    `json:"carbonDoorStatus"`
	CoCh4Leakage       int    `json:"coCh4Leakage"`
	JaaliBlockage      int    `json:"jaaliBlockage"`
	MachineMaintenance int    `json:"machineMaintenance"`
	AutoShutDown       int    `json:"autoShutDown"`
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		isModbusConnected: false,
	}
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) SaveConfig(config Config) error {
	configPath := filepath.Join(".", "config.json")
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}
	return nil
}

func (a *App) LoadConfig() (Config, error) {
	configPath := filepath.Join(".", "config.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		defaultConfig := Config{}
		if err := a.SaveConfig(defaultConfig); err != nil {
			return Config{}, fmt.Errorf("failed to create default config: %v", err)
		}
	}
	data, err := os.ReadFile(configPath)
	if err != nil {
		return Config{}, fmt.Errorf("failed to read config file: %v", err)
	}
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return Config{}, fmt.Errorf("failed to unmarshal config: %v", err)
	}
	return config, nil
}

func (a *App) Connect(comPort string) bool {
	handler := modbus.NewASCIIClientHandler(comPort)
	handler.BaudRate = 9600
	handler.DataBits = 7
	handler.Parity = "E"
	handler.StopBits = 1
	handler.SlaveId = 1
	handler.Timeout = 10 * time.Second

	err := handler.Connect()
	if err != nil {
		a.isModbusConnected = false
		runtime.LogError(a.ctx, fmt.Sprintf("Failed to connect with ASCII 7E1 to %s: %v", comPort, err))
		return false
	}

	a.client = modbus.NewClient(handler)
	a.isModbusConnected = true
	runtime.LogInfo(a.ctx, "Connected with ASCII 7E1 to COM port: "+comPort)
	return true
}

func (a *App) PLC_DATA() ([]BoilerData, error) {
	if !a.isModbusConnected {
		return nil, fmt.Errorf("Modbus client not connected")
	}

	startAddress := uint16(368)
	quantity := uint16(42)

	results, err := a.client.ReadHoldingRegisters(startAddress, quantity)
	if err != nil {
		runtime.LogError(a.ctx, fmt.Sprintf("Error reading holding registers: %v", err))
		return nil, err
	}

	if len(results) != int(quantity)*2 {
		return nil, fmt.Errorf("invalid register data length")
	}

	var boilerDataArray []BoilerData
	registersPerBoiler := 14

	for boilerID := 0; boilerID < 3; boilerID++ {
		startIdx := boilerID * registersPerBoiler
		endIdx := startIdx + registersPerBoiler

		if endIdx*2 > len(results) {
			break
		}

		data := results[startIdx*2 : endIdx*2]
		boiler := BoilerData{
			ID:              boilerID + 1,
			ReactorTemp:     int(safeUint16(data[0:2])),
			SeparatorTemp:   int(safeUint16(data[2:4])),
			FurnaceTemp:     int(safeUint16(data[4:6])),
			CondenserTemp:   int(safeUint16(data[6:8])),
			AtmTemp:         int(safeUint16(data[8:10])),
			ReactorPressure: int(safeUint16(data[10:12])),
			GasTankPressure: int(safeUint16(data[12:14])),
		}
		boilerDataArray = append(boilerDataArray, boiler)
	}

	runtime.LogInfo(a.ctx, fmt.Sprintf("Successfully read %d holding registers starting at address %d", quantity, startAddress))
	return boilerDataArray, nil
}

func safeUint16(data []byte) uint16 {
	if len(data) < 2 {
		runtime.LogError(context.Background(), "Invalid data length for uint16 conversion")
		return 0
	}
	return binary.BigEndian.Uint16(data)
}

