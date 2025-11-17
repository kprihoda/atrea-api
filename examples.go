package main

import (
	"fmt"
	"log"
)

// ExampleUsage demonstrates how to use the web interface
func ExampleUsage() {
	// Create a web client
	webClient := NewWebClient("192.168.68.106")

	// Authenticate with password
	sessionID, err := webClient.Login("6378")
	if err != nil {
		log.Fatalf("Authentication failed: %v", err)
	}
	fmt.Printf("✓ Authenticated with session: %s\n", sessionID)

	// Get current device data
	dataXML, err := webClient.GetData()
	if err != nil {
		log.Fatalf("Failed to get data: %v", err)
	}

	// Parse the XML data
	deviceData, err := ParseXMLData(dataXML)
	if err != nil {
		log.Fatalf("Failed to parse data: %v", err)
	}
	fmt.Printf("✓ Device data retrieved: %d parameters\n", len(deviceData.Items))

	// Extract common parameters
	commonParams := deviceData.ExtractCommonParameters()
	fmt.Printf("  - Desired Temperature: %.1f°C\n", commonParams.DesiredTemperature)
	fmt.Printf("  - Operating Mode: %s\n", commonParams.OperatingMode)
	fmt.Printf("  - Date: %d.%d.%d\n", commonParams.Day, commonParams.Month, commonParams.Year)

	// Get alarms
	alarmsXML, err := webClient.GetAlarms()
	if err != nil {
		log.Fatalf("Failed to get alarms: %v", err)
	}
	fmt.Printf("✓ Alarms retrieved: %d bytes\n", len(alarmsXML))

	// ========== TEMPERATURE CONTROL ==========

	tempControl := NewTemperatureControl(webClient)

	// Set desired temperature to 21°C in heating mode
	err = tempControl.SetDesiredTemperature(21, 1)
	if err != nil {
		log.Fatalf("Failed to set temperature: %v", err)
	}
	fmt.Println("✓ Temperature set to 21°C")

	// ========== SYSTEM CONTROL ==========

	sysControl := NewSystemControl(webClient)

	// Set timezone to UTC+1 (CET)
	err = sysControl.SetTimezone(1)
	if err != nil {
		log.Fatalf("Failed to set timezone: %v", err)
	}
	fmt.Println("✓ Timezone set to UTC+1")

	// ========== NETWORK SETTINGS ==========

	// Get network settings (if available)
	netSettings, err := webClient.GetNetworkSettings()
	if err != nil {
		log.Fatalf("Failed to get network settings: %v", err)
	}
	fmt.Printf("✓ Network settings: %d bytes\n", len(netSettings))

	// ========== WEEKLY PROGRAMS ==========

	// Get weekly program for RTS ventilation
	rtsProgram, err := webClient.GetWeeklyProgram("RTS", "vzt")
	if err != nil {
		log.Fatalf("Failed to get RTS program: %v", err)
	}
	fmt.Printf("✓ RTS weekly program: %d bytes\n", len(rtsProgram))

	fmt.Println("\n✓ All examples completed successfully!")
}

// ExampleMultipleCommands demonstrates sending multiple commands in sequence
func ExampleMultipleCommands() {
	webClient := NewWebClient("192.168.68.106")

	// Login
	_, err := webClient.Login("6378")
	if err != nil {
		log.Fatal(err)
	}

	// Set multiple parameters at once
	err = webClient.SetMultipleValues([]string{
		FormatParam("H11021", 22), // Temperature
		FormatParam("H11017", 1),  // Mode
		FormatParam("H11400", 1),  // Timezone
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("✓ Multiple parameters set")
}

// ExampleIPAddressHandling demonstrates IP address encoding/decoding
func ExampleIPAddressHandling() {
	// Encode IP address to device parameters
	ip := "192.168.1.100"
	params, _ := IPParameterEncoder(ip)
	fmt.Printf("IP %s encoded as: low=%s, high=%s\n", ip, params["low"], params["high"])

	// Decode device parameters back to IP address
	// Parse params as integers first
	lowVal := 0
	highVal := 0
	_, err := fmt.Sscanf(params["low"], "%d", &lowVal)
	_, err = fmt.Sscanf(params["high"], "%d", &highVal)
	if err != nil {
		log.Fatal(err)
	}
	decoded := IPParameterDecoder(lowVal, highVal)
	fmt.Printf("Decoded IP address: %s\n", decoded)
}

// ExampleDataParsing demonstrates parsing device data
func ExampleDataParsing() {
	// Example XML response from device
	xmlData := `<?xml version="1.0" encoding="utf-8"?>
<root>
  <RD5>
    <Item id="H10715" val="1"/>
    <Item id="H11021" val="21"/>
    <Item id="H11017" val="1"/>
  </RD5>
</root>`

	// Parse the data
	data, err := ParseXMLData(xmlData)
	if err != nil {
		log.Fatal(err)
	}

	// Access specific values
	if temp, ok := data.GetValue("H11021"); ok {
		fmt.Printf("Desired temperature: %s°C\n", temp)
	}

	if mode, ok := data.GetValue("H10715"); ok {
		fmt.Printf("Operating mode: %s\n", mode)
	}

	// Get numeric values
	tempInt, _ := data.GetIntValue("H11021")
	fmt.Printf("Temperature (as int): %d°C\n", tempInt)
}
