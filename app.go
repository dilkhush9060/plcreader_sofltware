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
	ReactorTemp        uint16 `json:"reactorTemp"`
	SeparatorTemp      uint16 `json:"separatorTemp"`
	FurnaceTemp        uint16 `json:"furnaceTemp"`
	CondenserTemp      uint16 `json:"condenserTemp"`
	AtmTemp            uint16 `json:"atmTemp"`
	ReactorPressure    uint16 `json:"reactorPressure"`
	GasTankPressure    uint16 `json:"gasTankPressure"`
	NitrogenPurging    uint16 `json:"nitrogenPurging"`
	CarbonDoorStatus   uint16 `json:"carbonDoorStatus"`
	CoCh4Leakage       uint16 `json:"coCh4Leakage"`
	JaaliBlockage      uint16 `json:"jaaliBlockage"`
	MachineMaintenance uint16 `json:"machineMaintenance"`
	AutoShutDown       uint16 `json:"autoShutDown"`
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
		runtime.LogError(a.ctx, "Invalid register data length")
		return nil, fmt.Errorf("invalid register data length, expected %d bytes, got %d", quantity*2, len(results))
	}

	data := make([]BoilerData, 1)
	data[0] = BoilerData{
		ID:                 1,
		ReactorTemp:        binary.BigEndian.Uint16(results[0:2]),
		SeparatorTemp:      binary.BigEndian.Uint16(results[2:4]),
		FurnaceTemp:        binary.BigEndian.Uint16(results[4:6]),
		CondenserTemp:      binary.BigEndian.Uint16(results[6:8]),
		AtmTemp:            binary.BigEndian.Uint16(results[8:10]),
		ReactorPressure:    binary.BigEndian.Uint16(results[10:12]),
		GasTankPressure:    binary.BigEndian.Uint16(results[12:14]),
		NitrogenPurging:    binary.BigEndian.Uint16(results[14:16]),
		CarbonDoorStatus:   binary.BigEndian.Uint16(results[16:18]),
		CoCh4Leakage:       binary.BigEndian.Uint16(results[18:20]),
		JaaliBlockage:      binary.BigEndian.Uint16(results[20:22]),
		MachineMaintenance: binary.BigEndian.Uint16(results[22:24]),
		AutoShutDown:       binary.BigEndian.Uint16(results[24:26]),
	}

	runtime.LogInfo(a.ctx, fmt.Sprintf("Successfully read %d holding registers starting at address %d", quantity, startAddress))
	return data, nil
}
