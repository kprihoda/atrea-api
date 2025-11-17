# Atrea RD5 Web API

REST API server for controlling and monitoring Atrea RD5 heat recovery ventilation unit.

## Starting the Server

```bash
go build -o server.exe .
./server.exe
```

The server will authenticate with the device and start listening on port 8080 (configurable in `config.env`).

## Configuration

Create `config.env` file:

```
DEVICE_IP=192.168.68.106
DEVICE_PASSWORD=6378
SERVER_PORT=8080
```

## API Endpoints

### Health Check

```
GET /health
```

Returns server health status.

**Response:**
```json
{
  "success": true,
  "message": "Server is running",
  "data": {
    "status": "ok",
    "time": "2025-11-17T11:40:55Z"
  }
}
```

### Device Status

```
GET /status
```

Returns current device status including temperatures and parameter count.

**Response:**
```json
{
  "success": true,
  "data": {
    "device": "Atrea RD5",
    "ip": "192.168.68.106",
    "is_authenticated": true,
    "session_id": "25263",
    "parameter_count": 315,
    "last_update": "2025-11-17T11:40:55Z",
    "indoor_temp_celsius": 20.1,
    "outdoor_temp_celsius": 3.6
  }
}
```

### Get Temperatures

```
GET /temperature
```

Returns current indoor and outdoor temperatures.

**Response:**
```json
{
  "success": true,
  "data": {
    "indoor_celsius": 20.1,
    "outdoor_celsius": 3.6,
    "timestamp": "2025-11-17T11:40:55Z"
  }
}
```

### List All Parameters

```
GET /parameters
```

Lists all device parameters with optional limit.

**Query Parameters:**
- `limit` (optional): Maximum number of parameters to return

**Response:**
```json
{
  "success": true,
  "data": {
    "count": 10,
    "parameters": [
      {
        "id": "I10215",
        "name": "Indoor Air Temperature (T-IDA)",
        "value": "201"
      },
      {
        "id": "I10211",
        "name": "Outdoor Air Temperature (T-ODA)",
        "value": "36"
      }
    ]
  }
}
```

**Example:**
```bash
# Get first 10 parameters
curl "http://localhost:8080/parameters?limit=10"

# Get all parameters
curl "http://localhost:8080/parameters"
```

### Get Specific Parameter

```
GET /parameter/:id
```

Returns a single parameter by ID.

**Path Parameters:**
- `id`: Parameter ID (e.g., I10215, H11021)

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "I10215",
    "name": "Indoor Air Temperature (T-IDA)",
    "value": "201"
  }
}
```

**Example:**
```bash
# Get indoor temperature parameter
curl "http://localhost:8080/parameter/I10215"

# Get outdoor temperature parameter
curl "http://localhost:8080/parameter/I10211"

# Get desired temperature setpoint
curl "http://localhost:8080/parameter/H11021"
```

### Refresh Device Data

```
POST /refresh
```

Refreshes all parameter data from the device. Useful when you want fresh data without restarting the server.

**Response:**
```json
{
  "success": true,
  "message": "Device data refreshed",
  "data": {
    "timestamp": "2025-11-17T11:40:55Z"
  }
}
```

**Example:**
```bash
curl -X POST "http://localhost:8080/refresh"
```

## Common Parameters

| Parameter ID | Description | Type |
|---|---|---|
| I10215 | Indoor Air Temperature (T-IDA) | Read-only |
| I10211 | Outdoor Air Temperature (T-ODA) | Read-only |
| I10212 | Supply Air Temperature (T-SUP) | Read-only |
| I10213 | Extract Air Temperature (T-ETA) | Read-only |
| I10214 | Exhaust Air Temperature (T-EHA) | Read-only |
| H11021 | Desired Temperature Setpoint | Read/Write |
| H10715 | Operating Mode | Read/Write |
| C10005 | System Reset Command | Write-only |

## Error Responses

### Service Unavailable (503)
Device not initialized or data not available:
```json
{
  "success": false,
  "error": "Device not initialized"
}
```

### Not Found (404)
Parameter ID not found:
```json
{
  "success": false,
  "error": "Parameter UNKNOWN not found"
}
```

### Bad Request (400)
Invalid request parameters:
```json
{
  "success": false,
  "error": "Missing parameter ID"
}
```

## Temperature Value Encoding

Temperature values are encoded in the device as follows:

**Positive Temperatures (0.1°C to 130.0°C):**
- Raw value divided by 10
- Example: raw value 201 = 20.1°C

**Negative Temperatures (-50.0°C to -0.1°C):**
- Two's complement encoding
- Range: 65036-65535 maps to -50.0 to -0.1°C
- Example: raw value 65526 = -1.0°C

## CORS

The API supports CORS for cross-origin requests:
- **Access-Control-Allow-Origin**: `*`
- **Access-Control-Allow-Methods**: `GET, POST, PUT, DELETE, OPTIONS`
- **Access-Control-Allow-Headers**: `Content-Type, Authorization`

## Examples

### Get current indoor and outdoor temperatures
```bash
curl "http://localhost:8080/temperature"
```

### Monitor device status
```bash
curl "http://localhost:8080/status" | jq '.data | {ip, is_authenticated, indoor_temp_celsius, outdoor_temp_celsius}'
```

### Refresh device data and get status
```bash
curl -X POST "http://localhost:8080/refresh"
curl "http://localhost:8080/status"
```

### Get all temperature-related parameters
```bash
curl "http://localhost:8080/parameters?limit=100" | jq '.data.parameters[] | select(.name | contains("Temperature"))'
```

## Testing

Run the test suite:
```bash
go test -v
```

Run only server tests:
```bash
go test -v server_test.go server.go web.go utils.go testdata_capture.go
```

Run with coverage:
```bash
go test -cover
```
