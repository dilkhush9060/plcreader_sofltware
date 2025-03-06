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
	ProcessStartTime   int `json:"processStartTime"`
	TimeOfReaction     int `json:"timeOfReaction"`
	ProcessEndTime     int `json:"processEndTime"`
	CoolingEndTime     int `json:"coolingEndTime"`
	NitrogenPurging    int `json:"nitrogenPurging"`
	CarbonDoorStatus   int `json:"carbonDoorStatus"`
	CoCh4Leakage       int `json:"coCh4Leakage"`
	JaaliBlockage      int `json:"jaaliBlockage"`
	MachineMaintenance int `json:"machineMaintenance"`
	AutoShutDown       int `json:"autoShutDown"`
}


// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		isModbusConnected: false,
	}
}

// startup is called when the app starts. The context is saved
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

//save config
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

// load config
func (a *App) LoadConfig() (Config, error) {
	configPath := filepath.Join(".", "config.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		defaultConfig := Config{} // Initialize with default values if needed
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

// connect function
func (a *App) Connect(plantId string,comPort string) bool {
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
			return false
	}

	a.client = modbus.NewClient(handler)
	a.isModbusConnected = true
	runtime.LogInfo(a.ctx, "Connected to COM port: "+comPort)

	return true
}



// get plc data
func (a *App) PLC_DATA() bool{
	// data
	startAddress := uint16(4097)
  quantity := uint16(2)

		// result
	results, err := a.client.ReadHoldingRegisters(startAddress, quantity)
	if err != nil {
		runtime.LogInfo(a.ctx,"error reading holding registers: %v")
		return false
	}

	// Print the results
	fmt.Printf("Successfully read %d holding registers starting at address %d:\n", quantity, startAddress)
	for i, value := range results {
			fmt.Printf("Register %d (address %d): %d\n", i, startAddress+uint16(i), value)
	}
	 return true
}