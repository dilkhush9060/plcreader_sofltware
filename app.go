package main

import (
	"context"
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

// Config struct for storing application configuration
type Config struct {
	PlantID string `json:"plantId"`
	COMPort string `json:"comPort"`
}

// PLCDataResponse represents the structured response for PLC_DATA
type PLCDataResponse struct {
	Success bool      `json:"success"`
	Data    []uint16  `json:"data"`
	Error   string    `json:"error,omitempty"`
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

// SaveConfig saves the configuration to a JSON file
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

// LoadConfig loads the configuration from a JSON file
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

// Connect establishes a Modbus connection to the specified COM port
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

// PLC_DATA reads data from Modbus holding registers and returns it as JSON
func (a *App) PLC_DATA() (string, error) {
	response := PLCDataResponse{
		Success: false,
		Data:    nil,
		Error:   "",
	}

	if !a.isModbusConnected {
		response.Error = "Modbus client not connected"
		return a.marshalResponse(response)
	}

	startAddress := uint16(4466)
	quantity := uint16(2) // Start with a smaller quantity for debugging

	results, err := a.client.ReadHoldingRegisters(startAddress, quantity)
	if err != nil {
		response.Error = fmt.Sprintf("Error reading holding registers: %v", err)
		runtime.LogError(a.ctx, response.Error)
		return a.marshalResponse(response)
	}

	// Convert raw byte data to []uint16
	registerValues := make([]uint16, 0, quantity)
	for i := 0; i < len(results); i += 2 {
		if i+1 >= len(results) {
			break
		}
		value := uint16(results[i])<<8 | uint16(results[i+1])
		registerValues = append(registerValues, value)
	}

	response.Success = true
	response.Data = registerValues
	runtime.LogInfo(a.ctx, fmt.Sprintf("Successfully read %d holding registers starting at address %d", quantity, startAddress))

	return a.marshalResponse(response)
}

// marshalResponse converts the PLCDataResponse struct to JSON
func (a *App) marshalResponse(response PLCDataResponse) (string, error) {
	jsonData, err := json.Marshal(response)
	if err != nil {
		runtime.LogError(a.ctx, fmt.Sprintf("Failed to marshal response: %v", err))
		return "", err
	}
	return string(jsonData), nil
}
