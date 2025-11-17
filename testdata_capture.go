package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// CaptureTestData runs the actual device integration and saves responses to testdata files
// This is used ONCE to capture real device responses for testing
func CaptureTestData() error {
	// Create testdata directory
	testdataDir := "testdata"
	if err := os.MkdirAll(testdataDir, 0755); err != nil {
		return err
	}

	// Connect to device
	client := NewWebClient("192.168.68.106")

	// STEP 1: Capture login response
	fmt.Println("Capturing login response...")
	sessionID, err := client.Login("6378")
	if err != nil {
		return fmt.Errorf("login failed: %w", err)
	}
	fmt.Printf("✓ Session ID obtained: %s\n", sessionID)

	// STEP 2: Capture config data response
	fmt.Println("Capturing config data...")
	configData, err := client.GetData()
	if err != nil {
		return fmt.Errorf("get data failed: %w", err)
	}
	fmt.Printf("✓ Config data captured: %d bytes\n", len(configData))

	if err := os.WriteFile(filepath.Join(testdataDir, "response_config.xml"), []byte(configData), 0644); err != nil {
		return err
	}

	// STEP 3: Capture alarms response
	fmt.Println("Capturing alarms data...")
	alarmsData, err := client.GetAlarms()
	if err != nil {
		return fmt.Errorf("get alarms failed: %w", err)
	}
	fmt.Printf("✓ Alarms data captured: %d bytes\n", len(alarmsData))

	if err := os.WriteFile(filepath.Join(testdataDir, "response_alarms.xml"), []byte(alarmsData), 0644); err != nil {
		return err
	}

	fmt.Println("\n✓ All test data captured successfully!")
	fmt.Printf("Test data saved to %s/\n", testdataDir)

	return nil
}
