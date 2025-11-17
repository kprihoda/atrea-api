# Atrea RD5 Authentication Documentation

## Overview

The Atrea RD5 web API uses MD5-based authentication. This document explains the exact authentication flow to prevent future issues.

## Authentication Flow

### Step 1: Calculate MD5 Hash

Create an MD5 hash of the literal string `"\r\n"` (carriage return + newline) concatenated with the device password.

**CRITICAL:** The hash input must be the literal `\r\n` characters (0x0D 0x0A in hex), not escaped text.

Example with password `"6378"`:
- Input bytes: `0x0D 0x0A 0x36 0x33 0x37 0x38` 
- Input string: `"\r\n6378"`
- MD5 hash: `993278d1925c378ab94a6fe664ea6c60`

PowerShell verification:
```powershell
$password = "6378"
$hash = [System.Security.Cryptography.MD5]::Create().ComputeHash(
    [System.Text.Encoding]::UTF8.GetBytes("`r`n" + $password)
)
[System.BitConverter]::ToString($hash).Replace("-","").ToLower()
# Output: 993278d1925c378ab94a6fe664ea6c60
```

Go code:
```go
hash := md5.New()
io.WriteString(hash, "\r\n"+password)
magic := fmt.Sprintf("%x", hash.Sum(nil))
```

### Step 2: Generate Random Nonce

Generate a random number to include as the `rnd` parameter. This prevents replay attacks and can be any random sequence of digits.

Example: `123`, `456`, or any random 3+ digit number

### Step 3: Call Login Endpoint

Send a GET request to `/config/login.cgi` with the hash and nonce:

```
GET /config/login.cgi?magic=<HASH>&rnd=<RANDOM>
```

**Example:**
```
GET http://192.168.68.106/config/login.cgi?magic=993278d1925c378ab94a6fe664ea6c60&rnd=123
```

### Step 4: Extract Session ID

The device responds with XML containing a 5-digit session ID:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<root lng="0">15736</root>
```

Extract the number between `<root ...>` and `</root>` tags. This is the session ID.

**Valid session ID:** A 5-digit number (e.g., `15736`, `60588`, `46239`)

**Invalid responses:**
- `0` - Usually indicates an error
- `denied` - Authentication failed, wrong password
- Empty or missing - Malformed response

### Step 5: Use Session ID in Subsequent Requests

All subsequent API calls must include the session ID as the `auth` parameter:

```
GET /config/xml.xml?auth=<SESSION_ID>&rnd=<RANDOM>
GET /config/alarms.xml?auth=<SESSION_ID>&rnd=<RANDOM>
GET /config/xml.cgi?auth=<SESSION_ID>&param=value
```

## Complete Authentication Example

```go
package main

import (
    "crypto/md5"
    "fmt"
    "io"
    "math/rand"
    "net/http"
    "net/url"
    "strings"
    "time"
)

func authenticate(ip, password string) (string, error) {
    // Step 1: Calculate MD5 hash
    hash := md5.New()
    io.WriteString(hash, "\r\n"+password)
    magic := fmt.Sprintf("%x", hash.Sum(nil))

    // Step 2: Generate random nonce
    rand.Seed(time.Now().UnixNano())
    rnd := fmt.Sprintf("%03d", rand.Intn(1000))

    // Step 3: Call login endpoint
    client := &http.Client{Timeout: 10 * time.Second}
    params := url.Values{}
    params.Set("magic", magic)
    params.Set("rnd", rnd)
    
    loginURL := "http://" + ip + "/config/login.cgi?" + params.Encode()
    resp, err := client.Get(loginURL)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return "", err
    }

    // Step 4: Extract session ID
    responseStr := strings.TrimSpace(string(body))
    startIdx := strings.Index(responseStr, ">")
    endIdx := strings.LastIndex(responseStr, "<")

    if startIdx != -1 && endIdx != -1 && startIdx < endIdx {
        sessionID := strings.TrimSpace(responseStr[startIdx+1 : endIdx])
        if sessionID != "" && sessionID != "0" && sessionID != "denied" {
            return sessionID, nil
        }
    }

    return "", fmt.Errorf("authentication failed")
}

func main() {
    sessionID, err := authenticate("192.168.68.106", "6378")
    if err != nil {
        fmt.Printf("Authentication failed: %v\n", err)
        return
    }
    fmt.Printf("Session ID: %s\n", sessionID)
}
```

## Common Issues and Solutions

### Issue: "Authentication failed" or "denied" response

**Cause:** Incorrect password or hash calculation

**Solution:**
1. Verify the password is correct (check the device physical label or settings)
2. Verify the MD5 hash input includes literal `\r\n` (not escaped text)
3. Test the hash calculation manually:
   ```powershell
   $password = "6378"
   $hash = [System.Security.Cryptography.MD5]::Create().ComputeHash(
       [System.Text.Encoding]::UTF8.GetBytes("`r`n" + $password)
   )
   [System.BitConverter]::ToString($hash).Replace("-","").ToLower()
   ```

### Issue: Timeout connecting to device

**Cause:** Device IP is incorrect or device is offline

**Solution:**
1. Ping the device: `ping 192.168.68.106`
2. Check network connectivity
3. Verify the IP address matches the device's network configuration

### Issue: Session ID is `0` or empty

**Cause:** Malformed request or device returning error

**Solution:**
1. Check the response XML format
2. Verify the magic hash is correctly calculated
3. Try with a different random nonce

## Device Information

- **IP Address:** 192.168.68.106 (default, may vary)
- **Port:** 80 (HTTP only, no HTTPS)
- **Authentication Method:** MD5 hash
- **Session ID Format:** 5-digit number
- **Session Persistence:** Sessions appear to persist until device restart

## Reference

See `web.go` `Login()` method for the working implementation in this project.
