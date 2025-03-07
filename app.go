package main

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
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

// PLC_DATA reads data from Modbus holding registers and returns structured data
func (a *App) PLC_DATA() []uint16 {
	if !a.isModbusConnected {
		log.Println("Modbus client not connected")
		return nil
	}

	registerRanges := []struct {
		start  uint16
		length uint16
	}{
		{4466, 8},
		{4474, 8},
		{4482, 8},
	}

	var allData []uint16

	for _, reg := range registerRanges {
		results, err := a.client.ReadHoldingRegisters(reg.start, reg.length)
		if err != nil {
			log.Printf("Error reading registers %d-%d: %v", reg.start, reg.start+reg.length-1, err)
			return nil
		}

		// Convert []byte to []uint16
		for i := 0; i < len(results); i += 2 {
			value := binary.BigEndian.Uint16(results[i : i+2])
			allData = append(allData, value)
			log.Printf("Register %d (Address %d): %d", i/2, reg.start+uint16(i/2), value)
		}
	}

	log.Println("Successfully read all registers")
	log.Println(allData)
	return allData
}

