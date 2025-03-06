package main

import (
	"context"
	"encoding/json"
	"fmt"
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
		time.Sleep(1 * time.Second) // Wait before retrying
	}
	if err != nil {
		runtime.LogError(a.ctx, "Raw data received: "+string(results))
		return nil, fmt.Errorf("error reading holding registers: %v", err)
	}

	runtime.LogInfo(a.ctx, "Raw data received: "+string(results))

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
			return nil, fmt.Errorf("failed to parse hex: %v", err)
		}
		intResults = append(intResults, int(intVal))
	}

	// Handle insufficient data
	if len(intResults) < 4 {
		runtime.LogWarning(a.ctx, "Insufficient data received from Modbus")
		for len(intResults) < 4 {
			intResults = append(intResults, 0)
		}
	}

	return intResults, nil
}

func (a *App) GetBoilerData() []BoilerData {
	startAddress := uint16(4466)
	quantity := uint16(4) // Read 4 registers for ReactorTemp, SeparatorTemp, FurnaceTemp, CondenserTemp

	intResults, err := a.readHoldingRegisters(startAddress, quantity)
	if err != nil {
		runtime.LogError(a.ctx, "Failed to read Modbus registers: "+err.Error())
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

	// Start periodic reading of Modbus registers
	go a.startPeriodicReading()

	return "Connected to " + comPort
}

func (a *App) startPeriodicReading() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		boilerData := a.GetBoilerData()
		if len(boilerData) > 0 {
			runtime.LogInfo(a.ctx, fmt.Sprintf("Boiler Data: %+v", boilerData[0]))
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

