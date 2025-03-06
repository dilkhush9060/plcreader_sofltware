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
	ID                 int `json:"id"`
	ReactorTemp        int `json:"reactorTemp"`
	SeparatorTemp      int `json:"separatorTemp"`
	FurnaceTemp        int `json:"furnaceTemp"`
	CondenserTemp      int `json:"condenserTemp"`
	AtmTemp            int `json:"atmTemp"`
	ReactorPressure    int `json:"reactorPressure"`
	GasTankPressure    int `json:"gasTankPressure"`
	ProcessStartTime   int `json:"processStartTime"`
	TimeOfReaction     int `json:"timeOfReaction"`
	ProcessEndTime     int `json:"processEndTime"`
	CoolingEndTime     int `json:"coolingEndTime"`
	NitrogenPurging    int `json:"nitrogenPurging"`
	CarbonDoorStatus   int `json:"carbonDoorStatus"`
	CoCh4Leakage       int `json:"coCh4Leakage"`
	JaaliBlockage      int `json:"jaaliBlockage"`
	MachineMaintenance int `json:"machineMaintenance"`
	AutoShutDown       int `json:"autoShutDown"`
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
		runtime.LogError(a.ctx, "Failed to connect to "+comPort+": "+err.Error())
		return false
	}

	a.client = modbus.NewClient(handler)
	a.isModbusConnected = true
	runtime.LogInfo(a.ctx, "Connected to COM port: "+comPort)
	return true
}

func (a *App) PLC_DATA() ([]BoilerData, error) {
	startAddress := uint16(4466)
	quantity := uint16(20)

	results, err := a.client.ReadHoldingRegisters(startAddress, quantity)
	if err != nil {
		runtime.LogError(a.ctx, fmt.Sprintf("Error reading holding registers: %v", err))
		return nil, err
	}

	if len(results)%2 != 0 {
		runtime.LogError(a.ctx, "Invalid register data length")
		return nil, fmt.Errorf("invalid register data length")
	}

	var boilerDataArray []BoilerData

	for i := 0; i < len(results); i += 2 {
		value := binary.BigEndian.Uint16(results[i : i+2])
		boilerData := BoilerData{
			ID:          i / 2,
			ReactorTemp: int(value),
		}
		boilerDataArray = append(boilerDataArray, boilerData)
	}

	runtime.LogInfo(a.ctx, fmt.Sprintf("Successfully read %d holding registers starting at address %d", quantity, startAddress))
	return boilerDataArray, nil
}
