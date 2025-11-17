# Atrea RD5 Go Client Library

This Go library provides access to the Atrea RD5 heat recovery ventilation unit via an HTTP-based web API (reverse-engineered from the official web client).

**For detailed authentication information, see `AUTHENTICATION.md`**

## Overview

The Web interface communicates via HTTP. The API uses MD5-hashed password authentication.

```go
// Create client
webClient := NewWebClient("192.168.68.106")

// Login with password
sessionID, err := webClient.Login("6378")

// Get device configuration data
data, err := webClient.GetData()

// Set parameter values
err := webClient.SetValue("H12345=1000")

// Set multiple values
err := webClient.SetMultipleValues([]string{"H12345=1000", "H12346=2000"})
```

## Web API Endpoints

Based on reverse-engineering the web client, the following endpoints are available:

### Authentication
- **GET `/config/login.cgi?magic=<MD5_HASH>&rnd=<RANDOM>`**
  - MD5 hash format: MD5("\r\n" + password)
  - Returns: Session ID (5 characters)

### Data Access
- **GET `/config/xml.xml?auth=<SESSION_ID>&rnd=<RANDOM>`**
  - Returns: XML configuration data
  
- **GET `/config/alarms.xml?auth=<SESSION_ID>&rnd=<RANDOM>`**
  - Returns: XML alarm data

- **GET `/config/xml.cgi?auth=<SESSION_ID>&<PARAM>=<VALUE>`**
  - Sets configuration parameters
  - Format: Parameter IDs like "H12345", "C10005", etc.

### Network Configuration
- **GET `/config/ip.cgi?auth=<SESSION_ID>&<SETTINGS>`**
  - Get/Set network settings
  - Parameters:
    - `dhcp=1` or `dhcp=0` (enable/disable DHCP)
    - `ip=<VALUE>` - IP address (as concatenated bytes)
    - `ip4mask=<VALUE>` - Subnet mask
    - `ip4gw=<VALUE>` - Gateway
    - `ip4dns1=<VALUE>` - DNS server

### Weekly Programs
- **GET `/config/rtssetup.xml`** - RTS Ventilation setup
- **GET `/config/rgtssetup.xml`** - RTS Intelligent setup
- **GET `/config/rnssetup.xml`** - RNS Ventilation setup
- **GET `/config/rgnssetup.xml`** - RNS Intelligent setup

- **GET `/config/rtssetup.cgi`** - Set RTS Ventilation
- **GET `/config/rgtssetup.cgi`** - Set RTS Intelligent
- **GET `/config/rnssetup.cgi`** - Set RNS Ventilation
- **GET `/config/rgnssetup.cgi`** - Set RNS Intelligent

## Parameter IDs

Common parameter IDs observed in the web interface:

- `H10715` - Operating mode
- `H11010` - Temperature setpoint (mode 1)
- `H11017` - Temperature control mode
- `H11021` - Desired temperature
- `H11400` - Timezone offset
- `H10905-H10907` - Date (year, month, day)
- `H12200-H12209` - Network settings (DHCP, IP, mask, gateway, DNS)
- `C10005` - System reset
- `C10007` - Clear mode
- `C11400-C11415` - Restore settings

## Security Notes

1. The device uses MD5 hashing for password authentication (not cryptographically secure)
2. Session IDs appear to be persistent
3. No HTTPS support (HTTP only)
4. The device is intended for local network use only

## Example Usage

```go
package main

import (
	"fmt"
	"log"
)

func main() {
	// Create web client
	client := NewWebClient("192.168.68.106")

	// Authenticate
	sessionID, err := client.Login("6378")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Authenticated with session: %s\n", sessionID)

	// Get current configuration
	data, err := client.GetData()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Config data: %s\n", data)

	// Set temperature setpoint to 21°C
	err = client.SetValue(FormatParam("H11021", 21))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Temperature set successfully")
}
```

## Response Format

Most responses are XML documents. Example:
```xml
<?xml version="1.0" encoding="utf-8"?>
<root>
  <RD5>
    <Item id="H10715" val="0"/>
    <Item id="H11021" val="21"/>
    ...
  </RD5>
</root>
```

## Limitations & TODO

- [ ] Parse XML responses into Go structs
- [ ] Implement periodic data refresh
- [ ] Add parameter validation
- [ ] Implement HTTPS support (if available)
- [ ] Add alarm filtering and notifications
- [ ] Implement weekly program management
