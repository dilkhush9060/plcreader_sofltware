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
	ProcessStartTime   string `json:"processStartTime"` // Changed to string for HH:MM:SS
	TimeOfReaction     string `json:"timeOfReaction"`
	ProcessEndTime     string `json:"processEndTime"`
	CoolingEndTime     string `json:"coolingEndTime"`
	NitrogenPurging    int    `json:"nitrogenPurging"`    // Reverted to int
	CarbonDoorStatus   int    `json:"carbonDoorStatus"`   // Reverted to int
	CoCh4Leakage       int    `json:"coCh4Leakage"`       // Reverted to int
	JaaliBlockage      int    `json:"jaaliBlockage"`      // Reverted to int
	MachineMaintenance int    `json:"machineMaintenance"` // Reverted to int
	AutoShutDown       int    `json:"autoShutDown"`       // Reverted to int
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
	if !a.isModbusConnected {
		return nil, fmt.Errorf("Modbus client not connected")
	}

	// Adjust startAddress and quantity based on the UI data (e.g., D368 to D409 covers 42 registers)
	startAddress := uint16(4466) // Starting from D368 as per UI
	quantity := uint16(42)      // Total registers for 3 boilers (14 registers per boiler x 3)

	results, err := a.client.ReadHoldingRegisters(startAddress, quantity)
	if err != nil {
		runtime.LogError(a.ctx, fmt.Sprintf("Error reading holding registers: %v", err))
		return nil, err
	}

	if len(results) != int(quantity)*2 {
		runtime.LogError(a.ctx, "Invalid register data length")
		return nil, fmt.Errorf("invalid register data length, expected %d bytes, got %d", quantity*2, len(results))
	}

	var boilerDataArray []BoilerData
	registersPerBoiler := 14 // Each boiler has 14 data points (e.g., temps, pressures, indicators)

	for boilerID := 0; boilerID < 3; boilerID++ {
		startIdx := boilerID * registersPerBoiler
		endIdx := startIdx + registersPerBoiler

		if endIdx*2 > len(results) {
			break
		}

		data := results[startIdx*2 : endIdx*2]
		boiler := BoilerData{
			ID:                 boilerID + 1, // Boiler 1, 2, 3
			ReactorTemp:        int(binary.BigEndian.Uint16(data[0:2])),
			SeparatorTemp:      int(binary.BigEndian.Uint16(data[2:4])),
			FurnaceTemp:        int(binary.BigEndian.Uint16(data[4:6])),
			CondenserTemp:      int(binary.BigEndian.Uint16(data[6:8])),
			AtmTemp:            int(binary.BigEndian.Uint16(data[8:10])),
			ReactorPressure:    int(binary.BigEndian.Uint16(data[10:12])),
			GasTankPressure:    int(binary.BigEndian.Uint16(data[12:14])),
			ProcessStartTime:   fmt.Sprintf("%02d:%02d:%02d", int(binary.BigEndian.Uint16(data[14:16])/3600), (int(binary.BigEndian.Uint16(data[14:16])/60)%60), int(binary.BigEndian.Uint16(data[14:16])%60)), // Example conversion
			TimeOfReaction:     fmt.Sprintf("%02d:%02d:%02d", int(binary.BigEndian.Uint16(data[16:18])/3600), (int(binary.BigEndian.Uint16(data[16:18])/60)%60), int(binary.BigEndian.Uint16(data[16:18])%60)),
			ProcessEndTime:     fmt.Sprintf("%02d:%02d:%02d", int(binary.BigEndian.Uint16(data[18:20])/3600), (int(binary.BigEndian.Uint16(data[18:20])/60)%60), int(binary.BigEndian.Uint16(data[18:20])%60)),
			CoolingEndTime:     fmt.Sprintf("%02d:%02d:%02d", int(binary.BigEndian.Uint16(data[20:22])/3600), (int(binary.BigEndian.Uint16(data[20:22])/60)%60), int(binary.BigEndian.Uint16(data[20:22])%60)),
			NitrogenPurging:    int(binary.BigEndian.Uint16(data[22:24])),    // Kept as int
			CarbonDoorStatus:   int(binary.BigEndian.Uint16(data[24:26])),    // Kept as int
			CoCh4Leakage:       int(binary.BigEndian.Uint16(data[26:28])),    // Kept as int
			JaaliBlockage:      int(binary.BigEndian.Uint16(data[28:30])),    // Kept as int
			MachineMaintenance: int(binary.BigEndian.Uint16(data[30:32])),    // Kept as int
			AutoShutDown:       int(binary.BigEndian.Uint16(data[32:34])),    // Kept as int
		}
		boilerDataArray = append(boilerDataArray, boiler)
	}

	runtime.LogInfo(a.ctx, fmt.Sprintf("Successfully read %d holding registers starting at address %d", quantity, startAddress))
	return boilerDataArray, nil
}