# Quick Start Guide - Atrea RD5 Go Client

## Installation

The library is already built and ready to use. All source files are in the project directory.

## Quick Examples

### 1. Basic Web Authentication

```go
package main

import (
	"fmt"
	"log"
)

func main() {
	client := NewWebClient("192.168.68.106")
	sessionID, err := client.Login("6378")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Session ID:", sessionID)
}
```

### 2. Read Device Data

```go
client := NewWebClient("192.168.68.106")
client.Login("6378")

data, err := client.GetData()
if err != nil {
	log.Fatal(err)
}

// Parse the XML
deviceData, _ := ParseXMLData(data)

// Get temperature
if temp, ok := deviceData.GetValue("H11021"); ok {
	fmt.Println("Current temperature:", temp)
}
```

### 3. Set Parameters

```go
client := NewWebClient("192.168.68.106")
client.Login("6378")

// Set single parameter
client.SetValue(FormatParam("H11021", 21))

// Set multiple parameters
client.SetMultipleValues([]string{
	FormatParam("H11021", 21),
	FormatParam("H11017", 1),
})
```

### 4. Use Temperature Control Helper

```go
client := NewWebClient("192.168.68.106")
client.Login("6378")

tempControl := NewTemperatureControl(client)
tempControl.SetDesiredTemperature(21, 1)
```

### 5. Use System Control Helper

```go
client := NewWebClient("192.168.68.106")
client.Login("6378")

sysControl := NewSystemControl(client)
sysControl.SetTimezone(1)      // UTC+1
sysControl.Reset()             // Reset system
sysControl.ClearMode()         // Clear mode
```

### 6. Use Modbus Interface

```go
modbusClient, err := NewModbusClient("192.168.68.106", "502")
if err != nil {
	log.Fatal(err)
}
defer modbusClient.Close()

// Read single register
value, err := modbusClient.ReadInputRegister(0)
fmt.Println("Register 0:", value)

// Read multiple registers
values, err := modbusClient.ReadInputRegisters(0, 10)
fmt.Println("Registers 0-9:", values)
```

## Common Parameter IDs

| ID | Description | Type |
|----|-------------|------|
| H10715 | Operating Mode | int |
| H11021 | Desired Temperature | float |
| H11017 | Temperature Mode | int |
| H11400 | Timezone Offset | int |
| H10905 | Year | int |
| H10906 | Month | int |
| H10907 | Day | int |
| C10005 | System Reset | command |
| C10007 | Clear Mode | command |

## Troubleshooting

### Authentication Failed
- Verify the IP address (default: 192.168.68.106)
- Verify the password (default: 6378)
- Check network connectivity

### No Data Returned
- Ensure you're authenticated before calling GetData()
- Verify the device is online
- Check for firewall/network issues

### Modbus Connection Issues
- Ensure Modbus is enabled on the device
- Check TCP port 502 is accessible
- Verify IP address is correct

## Architecture Overview

```
┌─────────────────────────────────────┐
│       Atrea RD5 Device              │
│   (IP: 192.168.68.106)              │
└──────────┬──────────────────────────┘
           │
     ┌─────┴──────┐
     │            │
     ▼            ▼
┌─────────┐  ┌────────────┐
│ Modbus  │  │ Web API    │
│ TCP:502 │  │ HTTP:80    │
└─────────┘  └────────────┘
     │            │
     │            │
┌────▼────────────▼────────┐
│  Go Client Library       │
├──────────────────────────┤
│ • modbus.go              │
│ • web.go                 │
│ • utils.go               │
│ • examples.go            │
└──────────────────────────┘
```

## File Reference

| File | Purpose | Lines |
|------|---------|-------|
| modbus.go | Modbus TCP client | ~79 |
| web.go | Web API client (reverse-engineered) | ~280 |
| utils.go | Helpers and utilities | ~350 |
| examples.go | Usage examples | ~190 |
| main.go | Basic demo | ~40 |
| WEB_API.md | API documentation | ~150 |
| ARCHITECTURE.md | Project overview | ~130 |

## Building

```bash
# Build executable
go build -o atrea_alerts.exe

# Run (with custom main.go showing both interfaces)
./atrea_alerts.exe
```

## Key Classes/Structs

- **ModbusClient** - Direct Modbus communication
- **WebClient** - HTTP-based web API
- **DeviceData** - Parsed device configuration
- **TemperatureControl** - Temperature management helper
- **SystemControl** - System control helper
- **SessionManager** - Session lifecycle management

## Notes

- All web API endpoints discovered through reverse-engineering
- Password is hashed with MD5("\r\n" + password)
- Session IDs are 5 characters long
- Both Modbus and Web interfaces can be used simultaneously
- Most operations return XML which is automatically parsed

## Support

For detailed API documentation, see WEB_API.md
For architecture details, see ARCHITECTURE.md
For code examples, see examples.go
