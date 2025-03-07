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

// PLC_DATA reads registers 4466-4507 in chunks and returns uint16 values
func (a *App) PLC_DATA() ([]uint16, error) {
    if !a.isModbusConnected {
        err := fmt.Errorf("Modbus client not connected")
        runtime.LogError(a.ctx, err.Error())
        return nil, err
    }

    var allData []uint16

    // Read first 10 registers (4466-4475)
    results1, err1 := a.client.ReadHoldingRegisters(uint16(4466), uint16(10))
    if err1 != nil {
        err := fmt.Errorf("error reading registers 4466-4475: %v", err1)
        runtime.LogError(a.ctx, err.Error())
        runtime.LogDebug(a.ctx, fmt.Sprintf("Raw response (if any): %v", results1))
        return nil, err
    }
    for i := 0; i < len(results1); i += 2 {
        allData = append(allData, uint16(results1[i])<<8|uint16(results1[i+1]))
    }


    

    runtime.LogInfo(a.ctx, fmt.Sprintf("Successfully read %d uint16 values from PLC (registers 4466-4507)", len(allData)))
    runtime.LogDebug(a.ctx, fmt.Sprintf("Values: %v", allData))

    return allData, nil
}