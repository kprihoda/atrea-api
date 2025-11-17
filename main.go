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
	serverPort    = 8080
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
		case "SERVER_PORT":
			fmt.Sscanf(value, "%d", &serverPort)
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

	fmt.Println("=== Atrea RD5 Web API Server ===")
	fmt.Printf("Device IP: %s\n", atreaIP)
	fmt.Printf("Server Port: %d\n", serverPort)

	// Create and start server
	server := NewServer(atreaIP, atreaPassword)
	if err := server.StartServer(serverPort); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
