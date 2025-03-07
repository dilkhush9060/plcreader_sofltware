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
    Success bool   `json:"success"`
    Data    []byte `json:"data"` // Changed from []uint16 to []byte
    Error   string `json:"error,omitempty"`
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
    handler.Timeout = 5 * time.Second

    err := handler.Connect()
    if err != nil {
        a.isModbusConnected = false
        runtime.LogError(a.ctx, fmt.Sprintf("Failed to connect with ASCII 7E1 to %s: %v", err))
        return false
    }

    a.client = modbus.NewClient(handler)
    a.isModbusConnected = true
    runtime.LogInfo(a.ctx, "Connected with ASCII 7E1 to COM port: "+comPort)
    return true
}



// PLC_DATA reads registers 4466-4507 in chunks and returns raw bytes
func (a *App) PLC_DATA() PLCDataResponse {
    response := PLCDataResponse{
        Success: false,
        Data:    nil,
        Error:   "",
    }

    if !a.isModbusConnected {
        response.Error = "Modbus client not connected"
        runtime.LogError(a.ctx, response.Error)
        return response
    }

    var allData []byte

    // Read first 10 registers (4466-4475)
    results1, err1 := a.client.ReadHoldingRegisters(4466, 10)
    if err1 != nil {
        response.Error = fmt.Sprintf("error reading registers 4466-4475: %v", err1)
        runtime.LogError(a.ctx, response.Error)
        runtime.LogDebug(a.ctx, fmt.Sprintf("Raw response (if any): %v", results1))
        return response
    }
    allData = append(allData, results1...)

    // Read second 10 registers (4476-4485)
    results2, err2 := a.client.ReadHoldingRegisters(4476, 10)
    if err2 != nil {
        response.Error = fmt.Sprintf("error reading registers 4476-4485: %v", err2)
        runtime.LogError(a.ctx, response.Error)
        runtime.LogDebug(a.ctx, fmt.Sprintf("Raw response (if any): %v", results2))
        return response
    }
    allData = append(allData, results2...)

    // Read third 10 registers (4486-4495)
    results3, err3 := a.client.ReadHoldingRegisters(4486, 10)
    if err3 != nil {
        response.Error = fmt.Sprintf("error reading registers 4486-4495: %v", err3)
        runtime.LogError(a.ctx, response.Error)
        runtime.LogDebug(a.ctx, fmt.Sprintf("Raw response (if any): %v", results3))
        return response
    }
    allData = append(allData, results3...)

    // Read remaining 12 registers (4496-4507)
    results4, err4 := a.client.ReadHoldingRegisters(4496, 12)
    if err4 != nil {
        response.Error = fmt.Sprintf("error reading registers 4496-4507: %v", err4)
        runtime.LogError(a.ctx, response.Error)
        runtime.LogDebug(a.ctx, fmt.Sprintf("Raw response (if any): %v", results4))
        return response
    }
    allData = append(allData, results4...)

    // Log the raw bytes for debugging
    runtime.LogInfo(a.ctx, fmt.Sprintf("Raw data length: %d bytes", len(allData)))
    runtime.LogDebug(a.ctx, fmt.Sprintf("Raw bytes: %v", allData))

    // Set successful response
    response.Success = true
    response.Data = allData
    runtime.LogInfo(a.ctx, "Successfully read PLC data (42 registers from 4466-4507 as raw bytes)")

		println(allData)

    return response
}