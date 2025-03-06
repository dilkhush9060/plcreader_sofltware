package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/goburrow/modbus"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx              context.Context
	client           modbus.Client
	isModbusConnected bool
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

func (a *App) GetBoilerData() []BoilerData {
	if !a.isModbusConnected || a.client == nil {
		runtime.LogInfo(a.ctx, "Modbus not connected, returning empty data")
		return []BoilerData{}
	}

	startAddress := uint16(4466)
	quantity := uint16(100)
	results, err := a.client.ReadHoldingRegisters(startAddress, quantity)

	if err != nil {
		runtime.LogError(a.ctx, "Modbus read error: "+err.Error())
		runtime.LogError(a.ctx, "Raw data received: "+string(results))
		return []BoilerData{}
	}

	// Filter out non-hexadecimal characters
	filteredResults := make([]byte, 0)
	for _, b := range results {
		if (b >= '0' && b <= '9') || (b >= 'A' && b <= 'F') || (b >= 'a' && b <= 'f') {
			filteredResults = append(filteredResults, b)
		}
	}

	// Convert filtered results to integers
	var intResults []int
	for i := 0; i < len(filteredResults); i += 2 {
		if i+1 >= len(filteredResults) {
			break // Skip incomplete pairs
		}
		hexStr := string(filteredResults[i]) + string(filteredResults[i+1])
		intVal, err := strconv.ParseInt(hexStr, 16, 16)
		if err != nil {
			runtime.LogError(a.ctx, "Failed to parse hex: "+err.Error())
			return []BoilerData{}
		}
		intResults = append(intResults, int(intVal))
	}

	// Ensure we have enough data to populate the BoilerData struct
	if len(intResults) < 4 {
		runtime.LogError(a.ctx, "Insufficient data received from Modbus")
		return []BoilerData{}
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
	if a.client != nil {
		if handler, ok := a.client.(interface{ Close() error }); ok {
			handler.Close()
		}
		a.client = nil
		a.isModbusConnected = false
	}

	handler := modbus.NewASCIIClientHandler(comPort)
	handler.BaudRate = 9600
	handler.DataBits = 7
	handler.Parity = "E"
	handler.StopBits = 1
	handler.SlaveId = 1
	handler.Timeout = 10 * time.Second // Increased timeout

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

	return "Connected to " + comPort
}

func (a *App) IsModbusConnected() bool {
	return a.isModbusConnected
}

func (a *App) saveConfig(config Config) error {
	configPath := filepath.Join(".", "config.json")
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, data, 0644)
}

func (a *App) loadConfig() (Config, error) {
	configPath := filepath.Join(".", "config.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return Config{}, err
	}
	var config Config
	err = json.Unmarshal(data, &config)
	return config, err
}