package main

import (
	"fmt"
	"log"
	"time"

	"github.com/goburrow/modbus"
)

// readHoldingRegisters reads specified number of holding registers from the Modbus client
func readHoldingRegisters(client modbus.Client, startAddress, quantity uint16) error {
    results, err := client.ReadHoldingRegisters(startAddress, quantity)
    if err != nil {
        return fmt.Errorf("error reading holding registers: %v", err)
    }

    // Print the results
    fmt.Printf("Successfully read %d holding registers starting at address %d:\n", quantity, startAddress)
    for i, value := range results {
        fmt.Printf("Register %d (address %d): %d\n", i, startAddress+uint16(i), value)
    }
    return nil
}

func main() {
    // Modbus ASCII configuration
    handler := modbus.NewASCIIClientHandler("COM9") // Replace with your serial port
    handler.BaudRate = 9600
    handler.DataBits = 7
    handler.Parity = "E" // Even parity
    handler.StopBits = 1
    handler.SlaveId = 1
    handler.Timeout = 5 * time.Second

    // Connect to the Modbus server
    err := handler.Connect()
    if err != nil {
        log.Fatalf("Error connecting to Modbus server: %v", err)
    }
    defer handler.Close()

    // Create a new Modbus client
    client := modbus.NewClient(handler)

    // Define reading parameters
    startAddress := uint16(4466)
    quantity := uint16(50)

    // Create a ticker that triggers every 5 seconds
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()

    // Initial read before starting the loop
    if err := readHoldingRegisters(client, startAddress, quantity); err != nil {
        log.Printf("Initial read failed: %v", err)
    }

    // Infinite loop to read registers every 5 seconds
    for range ticker.C {
        fmt.Println("\nReading registers at", time.Now().Format(time.RFC1123))
        if err := readHoldingRegisters(client, startAddress, quantity); err != nil {
            log.Printf("Read failed: %v", err)
        }
    }
}