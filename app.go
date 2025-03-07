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

// Config struct for storing application configuration
type Config struct {
    PlantID string `json:"plantId"`
    COMPort string `json:"comPort"`
}

// PLCDataResponse represents the structured response for PLC_DATA
type PLCDataResponse struct {
    Success bool     `json:"success"`
    Data    []uint16 `json:"data"`
    Error   string   `json:"error,omitempty"`
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

// bytesToUint16 converts a byte slice to uint16 slice assuming big-endian encoding
func bytesToUint16(byteSlice []byte) []uint16 {
    if len(byteSlice)%2 != 0 {
        return nil // Invalid byte slice length
    }
    result := make([]uint16, len(byteSlice)/2)
    for i := 0; i < len(byteSlice); i += 2 {
        result[i/2] = binary.BigEndian.Uint16(byteSlice[i : i+2])
    }
    return result
}

// PLC_DATA reads data from Modbus holding registers and returns structured data
func (a *App) PLC_DATA() PLCDataResponse {
    response := PLCDataResponse{
        Success: false,
        Data:    nil,
        Error:   "",
    }

    if !a.isModbusConnected {
        response.Error = "Modbus client not connected"
        return response
    }

    // Slice to collect all register values
    var allData []uint16

    // Read first set of registers (4466-4479)
    results1, err1 := a.client.ReadHoldingRegisters(4466, 14)
    if err1 != nil {
        response.Error = fmt.Sprintf("error reading registers 4466-4479: %v", err1)
        runtime.LogError(a.ctx, response.Error)
        return response
    }
    values1 := bytesToUint16(results1)
    if values1 == nil {
        response.Error = "error converting bytes to uint16 for registers 4466-4479"
        runtime.LogError(a.ctx, response.Error)
        return response
    }
    allData = append(allData, values1...)

    // Read second set of registers (4480-4493)
    results2, err2 := a.client.ReadHoldingRegisters(4480, 14)
    if err2 != nil {
        response.Error = fmt.Sprintf("error reading registers 4480-4493: %v", err2)
        runtime.LogError(a.ctx, response.Error)
        return response
    }
    values2 := bytesToUint16(results2)
    if values2 == nil {
        response.Error = "error converting bytes to uint16 for registers 4480-4493"
        runtime.LogError(a.ctx, response.Error)
        return response
    }
    allData = append(allData, values2...)

    // Read third set of registers (4494-4507)
    results3, err3 := a.client.ReadHoldingRegisters(4494, 14)
    if err3 != nil {
        response.Error = fmt.Sprintf("error reading registers 4494-4507: %v", err3)
        runtime.LogError(a.ctx, response.Error)
        return response
    }
    values3 := bytesToUint16(results3)
    if values3 == nil {
        response.Error = "error converting bytes to uint16 for registers 4494-4507"
        runtime.LogError(a.ctx, response.Error)
        return response
    }
    allData = append(allData, values3...)

    // Log the results
    for i, value := range values1 {
        runtime.LogInfo(a.ctx, fmt.Sprintf("Register %d (address %d): %d", i, 4466+i, value))
    }
    for i, value := range values2 {
        runtime.LogInfo(a.ctx, fmt.Sprintf("Register %d (address %d): %d", i, 4480+i, value))
    }
    for i, value := range values3 {
        runtime.LogInfo(a.ctx, fmt.Sprintf("Register %d (address %d): %d", i, 4494+i, value))
    }

    // Set successful response
    response.Success = true
    response.Data = allData
		println(allData)

    return response
}