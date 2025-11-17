# Atrea Alerts - Project Structure

## Overview

This project implements a Go client library for the Atrea RD5 heat recovery ventilation unit, providing an HTTP web API interface.

## File Structure

### Core Files

1. **web.go** - Reverse-engineered web API client
   - HTTP-based communication
   - MD5 password authentication
   - Parameter reading and writing
   - Network configuration
   - Weekly program management
   - Alarm retrieval

2. **utils.go** - Utility functions and helpers
   - XML parsing utilities
   - Temperature and system control helpers
   - IP address encoding/decoding
   - Session management
   - Common parameter extraction

3. **main.go** - Example usage
   - Demonstrates web interface
   - Shows authentication and basic operations

4. **examples.go** - Extended examples
   - Multiple commands
   - IP address handling
   - Data parsing
   - Session management

5. **WEB_API.md** - API documentation
   - Complete endpoint reference
   - Parameter IDs
   - Authentication details
   - Response formats
   - Security notes

## Key Features

### Web Interface
- HTTP client with timeout handling
- MD5-based password authentication
- XML data parsing
- Persistent session management
- Support for:
  - Device configuration
  - Parameter setting
  - Alarm retrieval
  - Network settings
  - Weekly programs

### Helper Utilities
- XML parsing into Go structs
- Temperature control convenience methods
- System control functions
- IP address encoding/decoding
- Session management with age tracking

## Usage Example

```go
// Web interface
webClient := NewWebClient("192.168.68.106")
sessionID, err := webClient.Login("6378")

// Get and parse data
data, _ := webClient.GetData()
deviceData, _ := ParseXMLData(data)

// Set parameters
tempControl := NewTemperatureControl(webClient)
tempControl.SetDesiredTemperature(21, 1)
```

## Building and Running

```bash
# Build the project
go build

# Run the application
./atrea_alerts

# Run specific examples
go run *.go
```

## API Endpoints Discovered

During reverse-engineering, the following endpoints were discovered:

- `/config/login.cgi` - Authentication
- `/config/xml.xml` - Read device configuration
- `/config/xml.cgi` - Write parameters
- `/config/alarms.xml` - Read alarms
- `/config/ip.cgi` - Network settings
- `/config/rtssetup.xml` / `.cgi` - RTS setup
- `/config/rgtssetup.xml` / `.cgi` - RTS intelligent setup
- `/config/rnssetup.xml` / `.cgi` - RNS setup
- `/config/rgnssetup.xml` / `.cgi` - RNS intelligent setup

## Authentication Method

The device uses a custom authentication scheme:
1. Client sends: `magic=MD5("\r\n" + password)`
2. Server responds with: 5-character session ID
3. All subsequent requests include: `auth=<SESSION_ID>`

## Next Steps

Potential improvements:
1. Implement more parameter parsing
2. Add support for scheduling
3. Implement data logging
4. Add alert notifications
5. Create REST API wrapper
6. Add web dashboard
