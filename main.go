package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

var (
	atreaIP       = "192.168.68.106"
	atreaPassword = "6378"
)

func loadConfig() error {
	file, err := os.Open("config.env")
	if err != nil {
		// If config.env doesn't exist, use defaults
		if os.IsNotExist(err) {
			fmt.Println("Note: config.env not found, using default configuration")
			return nil
		}
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "DEVICE_IP":
			atreaIP = value
		case "DEVICE_PASSWORD":
			atreaPassword = value
		}
	}

	return scanner.Err()
}

func main() {
	// Check for --capture flag
	captureFlag := flag.Bool("capture", false, "Capture real device responses and save to testdata/")
	flag.Parse()

	if *captureFlag {
		if err := CaptureTestData(); err != nil {
			log.Fatalf("Error capturing test data: %v", err)
		}
		os.Exit(0)
	}
	// Load configuration from config.env
	if err := loadConfig(); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	fmt.Println("=== Atrea RD5 Web API Client ===")

	// Create web client
	webClient := NewWebClient(atreaIP)
	fmt.Printf("Created client for: %s\n", atreaIP)

	// Authenticate with the device
	fmt.Printf("\nAttempting authentication with password...\n")
	sessionID, err := webClient.Login(atreaPassword)
	if err != nil {
		fmt.Printf("❌ Authentication failed: %v\n", err)
		fmt.Println("\nVerify:")
		fmt.Println("  - Device IP is correct: 192.168.68.106")
		fmt.Println("  - Device is accessible on network")
		fmt.Println("  - Password is correct: 6378")
		os.Exit(1)
	}
	fmt.Printf("✓ Session ID obtained: %s\n", sessionID)

	// Get device data
	fmt.Println("\nRetrieving device configuration...")
	data, err := webClient.GetData()
	if err != nil {
		log.Fatalf("Failed to get data: %v", err)
	}
	fmt.Printf("✓ Data retrieved: %d bytes\n", len(data))

	// Parse device data
	deviceData, err := ParseXMLData(data)
	if err != nil {
		log.Fatalf("Failed to parse data: %v", err)
	}

	fmt.Printf("✓ Parsed %d parameters\n", len(deviceData.Items))

	// Show sample parameters with names
	if len(deviceData.Items) > 0 {
		fmt.Println("\nSample parameters:")
		count := 0
		for key, val := range deviceData.Items {
			if count >= 10 {
				break
			}
			name := GetParameterName(key)
			fmt.Printf("  %s (%s) = %s\n", key, name, val)
			count++
		}
	}

	// Display current temperatures
	fmt.Println("\nCurrent Temperatures:")
	if indoor, err := deviceData.GetCurrentTemperature(); err == nil && indoor > 0 {
		fmt.Printf("  Indoor: %.1f°C\n", indoor)
	}
	if outdoor, err := deviceData.GetOutdoorTemperature(); err == nil && outdoor > -50 {
		fmt.Printf("  Outdoor: %.1f°C\n", outdoor)
	}

	// Try to get alarms
	fmt.Println("\nRetrieving alarms...")
	alarms, err := webClient.GetAlarms()
	if err != nil {
		log.Fatalf("Failed to get alarms: %v", err)
	}
	fmt.Printf("✓ Alarms retrieved: %d bytes\n", len(alarms))

	fmt.Println("\n=== All operations completed successfully ===")
}
