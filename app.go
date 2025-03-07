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
	ctx               context.Context
	client            modbus.Client
	handler           *modbus.ASCIIClientHandler // Added to allow proper cleanup
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

// shutdown is called when the app is closing
func (a *App) shutdown(ctx context.Context) {
	if a.isModbusConnected && a.handler != nil {
		a.handler.Close()
		runtime.LogInfo(a.ctx, "Modbus connection closed")
	}
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
	runtime.LogInfo(a.ctx, "Configuration saved successfully")
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
	if err := json.Unmarshal(data, &data); err != nil {
		return Config{}, fmt.Errorf("failed to unmarshal config: %v", err)
	}
	return config, nil
}

// Connect establishes a Modbus connection to the specified COM port
func (a *App) Connect(comPort string) bool {
	if a.isModbusConnected {
		runtime.LogWarning(a.ctx, "Already connected, disconnecting first")
		if a.handler != nil {
			a.handler.Close()
		}
	}

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
	a.handler = handler // Store handler for later cleanup
	a.isModbusConnected = true
	runtime.LogInfo(a.ctx, "Connected with ASCII 7E1 to COM port: "+comPort)
	return true
}

// Disconnect closes the Modbus connection
func (a *App) Disconnect() bool {
	if !a.isModbusConnected || a.handler == nil {
		runtime.LogWarning(a.ctx, "No active connection to disconnect")
		return false
	}
	err := a.handler.Close()
	if err != nil {
		runtime.LogError(a.ctx, fmt.Sprintf("Failed to disconnect: %v", err))
		return false
	}
	a.isModbusConnected = false
	a.handler = nil
	a.client = nil
	runtime.LogInfo(a.ctx, "Disconnected from Modbus")
	return true
}

// PLC_DATA reads data from Modbus holding registers and returns structured data
func (a *App) PLC_DATA() ([]uint16, error) {
	if !a.isModbusConnected {
		runtime.LogError(a.ctx, "Modbus client not connected")
		return nil, fmt.Errorf("modbus client not connected")
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
	const maxRetries = 3
	const retryDelay = 500 * time.Millisecond

	for _, reg := range registerRanges {
		var results []byte
		var err error

		// Retry mechanism for timeout errors
		for attempt := 0; attempt < maxRetries; attempt++ {
			// Check context cancellation
			select {
			case <-a.ctx.Done():
				return nil, fmt.Errorf("operation cancelled: %v", a.ctx.Err())
			default:
			}

			results, err = a.client.ReadHoldingRegisters(reg.start, reg.length)
			if err == nil {
				break
			}

			if err.Error() == "serial: timeout" {
				runtime.LogWarning(a.ctx, fmt.Sprintf("Timeout reading registers %d-%d (attempt %d/%d)", 
					reg.start, reg.start+reg.length-1, attempt+1, maxRetries))
				if attempt < maxRetries-1 {
					time.Sleep(retryDelay)
					continue
				}
			}

			runtime.LogError(a.ctx, fmt.Sprintf("Error reading registers %d-%d: %v", 
				reg.start, reg.start+reg.length-1, err))
			return nil, fmt.Errorf("failed to read registers %d-%d: %v", 
				reg.start, reg.start+reg.length-1, err)
		}

		if err != nil {
			return nil, fmt.Errorf("exhausted retries reading registers %d-%d", 
				reg.start, reg.start+reg.length-1)
		}

		// Convert []byte to []uint16
		for i := 0; i < len(results); i += 2 {
			value := binary.BigEndian.Uint16(results[i : i+2])
			allData = append(allData, value)
			runtime.LogDebug(a.ctx, fmt.Sprintf("Register %d (Address %d): %d", 
				i/2, reg.start+uint16(i/2), value))
		}
	}

	runtime.LogInfo(a.ctx, "Successfully read all registers")
	runtime.LogDebug(a.ctx, fmt.Sprintf("PLC Data: %v", allData))
	return allData, nil
}