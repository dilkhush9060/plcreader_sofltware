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

// App struct remains the same
type App struct {
    ctx              context.Context
    client           modbus.Client
    isModbusConnected bool
}

// BoilerData and Config structs remain the same
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
    NitrogenPurging    string `json:"nitrogenPurging"`
    CarbonDoorStatus   string `json:"carbonDoorStatus"`
    CoCh4Leakage       string `json:"coCh4Leakage"`
    JaaliBlockage      string `json:"jaaliBlockage"`
    MachineMaintenance string `json:"machineMaintenance"`
    AutoShutDown       string `json:"autoShutDown"`
}

type Config struct {
    PlantID string `json:"plantId"`
    COMPort string `json:"comPort"`
}

func NewApp() *App {
    return &App{
        isModbusConnected: false,
    }
}

func (a *App) Startup(ctx context.Context) {
    a.ctx = ctx
    runtime.LogInfo(ctx, "App started")
    config, err := a.loadConfig()
    if err == nil && config.COMPort != "" {
        a.Connect(config.PlantID, config.COMPort)
    }
}

func (a *App) readHoldingRegisters(startAddress, quantity uint16) ([]int, error) {
    if !a.isModbusConnected || a.client == nil {
        return nil, fmt.Errorf("Modbus not connected")
    }

    var results []byte
    var err error
    for retry := 0; retry < 3; retry++ {
        results, err = a.client.ReadHoldingRegisters(startAddress, quantity)
        if err == nil {
            break
        }
        runtime.LogWarning(a.ctx, fmt.Sprintf("Retry %d: Modbus read failed: %v", retry+1, err))
        time.Sleep(1 * time.Second)
    }
    if err != nil {
        return nil, fmt.Errorf("failed to read holding registers after retries: %v", err)
    }

    // Log raw bytes for debugging
    runtime.LogDebug(a.ctx, fmt.Sprintf("Raw Modbus data: %x", results))

    // Convert 16-bit registers to integers (2 bytes per register)
    if len(results) < int(quantity)*2 {
        return nil, fmt.Errorf("insufficient data: got %d bytes, expected %d", len(results), quantity*2)
    }

    intResults := make([]int, quantity)
    for i := 0; i < int(quantity); i++ {
        // Combine two bytes into one 16-bit value
        highByte := uint16(results[i*2])
        lowByte := uint16(results[i*2+1])
        value := int(highByte<<8 | lowByte)
        intResults[i] = value
    }

    return intResults, nil
}

func (a *App) GetBoilerData() []BoilerData {
    startAddress := uint16(4466)
    quantity := uint16(4)

    intResults, err := a.readHoldingRegisters(startAddress, quantity)
    if err != nil {
        runtime.LogError(a.ctx, "Failed to read Modbus registers: "+err.Error())
        return []BoilerData{
            {
                ID:            1,
                ReactorTemp:   -1, // Indicate error with invalid value
                SeparatorTemp: -1,
                FurnaceTemp:   -1,
                CondenserTemp: -1,
            },
        }
    }

    now := time.Now()
    return []BoilerData{
        {
            ID:                 1,
            ReactorTemp:        intResults[0],
            SeparatorTemp:      intResults[1],
            FurnaceTemp:        intResults[2],
            CondenserTemp:      intResults[3],
            AtmTemp:            25,
            ReactorPressure:    10,
            GasTankPressure:    5,
            ProcessStartTime:   now.Add(-time.Hour).Format("15:04:05"),
            TimeOfReaction:     "01:00:00",
            ProcessEndTime:     "00:00:00",
            CoolingEndTime:     "00:00:00",
            NitrogenPurging:    "red",
            CarbonDoorStatus:   "green",
            CoCh4Leakage:       "red",
            JaaliBlockage:      "green",
            MachineMaintenance: "red",
            AutoShutDown:       "green",
        },
    }
}

func (a *App) Connect(plantID, comPort string) string {
    // Cleanup existing connection
    a.Disconnect()

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
        runtime.LogError(a.ctx, "Failed to connect to COM port: "+err.Error())
        return "Connection failed: " + err.Error()
    }

    a.client = modbus.NewClient(handler)
    a.isModbusConnected = true
    runtime.LogInfo(a.ctx, "Connected to COM port: "+comPort)

    config := Config{PlantID: plantID, COMPort: comPort}
    if err := a.saveConfig(config); err != nil {
        runtime.LogError(a.ctx, "Failed to save config: "+err.Error())
    }

    go a.startPeriodicReading()
    return "Connected to " + comPort
}

func (a *App) Disconnect() {
    if a.client != nil {
        if handler, ok := a.client.(interface{ Close() error }); ok {
            if err := handler.Close(); err != nil {
                runtime.LogError(a.ctx, "Failed to close Modbus connection: "+err.Error())
            }
        }
        a.client = nil
        a.isModbusConnected = false
    }
}

func (a *App) startPeriodicReading() {
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            if !a.isModbusConnected {
                runtime.LogWarning(a.ctx, "Skipping periodic read: not connected")
                continue
            }
            boilerData := a.GetBoilerData()
            if len(boilerData) > 0 {
                runtime.LogInfo(a.ctx, fmt.Sprintf("Boiler Data: %+v", boilerData[0]))
            }
        case <-a.ctx.Done():
            runtime.LogInfo(a.ctx, "Stopping periodic reading")
            return
        }
    }
}

func (a *App) IsModbusConnected() bool {
    return a.isModbusConnected
}

func (a *App) saveConfig(config Config) error {
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

func (a *App) loadConfig() (Config, error) {
    configPath := filepath.Join(".", "config.json")
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