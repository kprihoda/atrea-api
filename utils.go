package main

import (
	"encoding/xml"
	"strconv"
	"strings"
	"time"
)

// DeviceData represents the parsed device configuration
type DeviceData struct {
	Items map[string]string
}

// AlarmData represents parsed alarm information
type AlarmData struct {
	Alarms map[string]string
}

// ParseXMLData parses the XML response from GetData()
func ParseXMLData(xmlStr string) (*DeviceData, error) {
	var root struct {
		XMLName xml.Name `xml:"RD5WEB"`
		RD5     struct {
			IntegerR struct {
				Items []struct {
					ID    string `xml:"I,attr"`
					Value string `xml:"V,attr"`
				} `xml:"O"`
			} `xml:"INTEGER_R"`
			StringR struct {
				Items []struct {
					ID    string `xml:"I,attr"`
					Value string `xml:"V,attr"`
				} `xml:"O"`
			} `xml:"STRING_R"`
			FloatR struct {
				Items []struct {
					ID    string `xml:"I,attr"`
					Value string `xml:"V,attr"`
				} `xml:"O"`
			} `xml:"FLOAT_R"`
			EnumR struct {
				Items []struct {
					ID    string `xml:"I,attr"`
					Value string `xml:"V,attr"`
				} `xml:"O"`
			} `xml:"ENUM_R"`
		} `xml:"RD5"`
	}

	err := xml.Unmarshal([]byte(xmlStr), &root)
	if err != nil {
		return nil, err
	}

	data := &DeviceData{
		Items: make(map[string]string),
	}

	// Collect all items from different sections
	for _, item := range root.RD5.IntegerR.Items {
		data.Items[item.ID] = item.Value
	}
	for _, item := range root.RD5.StringR.Items {
		data.Items[item.ID] = item.Value
	}
	for _, item := range root.RD5.FloatR.Items {
		data.Items[item.ID] = item.Value
	}
	for _, item := range root.RD5.EnumR.Items {
		data.Items[item.ID] = item.Value
	}

	return data, nil
}

// GetValue retrieves a specific parameter value
func (d *DeviceData) GetValue(key string) (string, bool) {
	val, ok := d.Items[key]
	return val, ok
}

// GetIntValue retrieves and converts a parameter to int
func (d *DeviceData) GetIntValue(key string) (int, error) {
	val, ok := d.Items[key]
	if !ok {
		return 0, nil
	}
	return strconv.Atoi(val)
}

// GetFloatValue retrieves and converts a parameter to float64
func (d *DeviceData) GetFloatValue(key string) (float64, error) {
	val, ok := d.Items[key]
	if !ok {
		return 0, nil
	}
	return strconv.ParseFloat(val, 64)
}

// ParameterNames maps device parameter IDs to human-readable names
// Based on Atrea RD5 official parameter documentation
var ParameterNames = map[string]string{
	// System Status & Mode
	"I00000": "System Status",
	"I00001": "Mode",
	"I00002": "Temperature",
	"I00004": "Year",
	
	// Temperature Readings (I1xxxx series)
	"I10222": "Indoor Air Temperature",
	"I10224": "Extract Air Temperature",
	"I10225": "Extract Air Temperature",
	"I10249": "Supply Air Temperature",
	"I10275": "Outdoor Air Temperature",
	"I10281": "Outdoor Air Temperature",
	"I10282": "Outdoor Air Temperature",
	
	// Fan Control
	"I10215": "Fan Speed",
	"I10230": "Supply Fan Speed",
	"I10244": "Extract Fan Speed",
	"I10251": "Supply Air Pressure",
	"I10262": "Extract Air Pressure",
	"I10265": "Fan Status",
	
	// Filter Status
	"I12015": "Filter Status",
	"I12020": "Filter Hours",
	
	// Control Parameters (H10xxx, H11xxx, H12xxx series)
	"H10715": "Operating Mode",
	"H11010": "Temperature Setpoint Mode 1",
	"H11017": "Temperature Control Mode",
	"H11021": "Desired Temperature",
	"H11400": "Timezone Offset",
	"H11406": "System Uptime",
	
	// Date/Time
	"H10905": "Year",
	"H10906": "Month",
	"H10907": "Day",
	
	// Network & System
	"H12200": "Network DHCP",
	"H12201": "IP Address",
	"H12202": "Subnet Mask",
	"H12203": "Gateway",
	"H12204": "DNS Server",
	
	// System Commands
	"C10005": "System Reset",
	"C10007": "Clear Mode",
}

// GetParameterName returns the human-readable name for a parameter ID
func GetParameterName(id string) string {
	if name, ok := ParameterNames[id]; ok {
		return name
	}
	return id
}

// GetCurrentTemperature reads the current room/indoor temperature from the device
// Primary parameter: I10222 (Indoor Air Temperature) from official RD5 documentation
func (d *DeviceData) GetCurrentTemperature() (float64, error) {
	// Try indoor temperature parameter IDs in priority order
	tempIDs := []string{"I10222", "I10224", "I10225", "I10249"}

	for _, id := range tempIDs {
		if val, ok := d.Items[id]; ok {
			temp, err := strconv.ParseFloat(val, 64)
			if err == nil && temp > -500 && temp < 10000 { // Raw values are typically 500-3500 (5-35°C)
				return temp / 100, nil // Device stores temps as integers in hundredths: 2500 = 25.00°C
			}
		}
	}

	return 0, nil // Return 0 if no valid temperature found
}

// GetOutdoorTemperature reads the outdoor air temperature from the device
// Primary parameter: I10275 (Outdoor Air Temperature) from official RD5 documentation
func (d *DeviceData) GetOutdoorTemperature() (float64, error) {
	tempIDs := []string{"I10275", "I10282", "I10281"}

	for _, id := range tempIDs {
		if val, ok := d.Items[id]; ok {
			temp, err := strconv.ParseFloat(val, 64)
			if err == nil && temp > -500 && temp < 10000 {
				return temp / 100, nil
			}
		}
	}

	return 0, nil
}

// GetAllTemperatures returns a map of all temperature-like parameters
func (d *DeviceData) GetAllTemperatures() map[string]float64 {
	temps := make(map[string]float64)

	for id, val := range d.Items {
		// Temperature parameters typically start with I1 and are 5 digits
		if strings.HasPrefix(id, "I1") && len(id) == 6 {
			if temp, err := strconv.ParseFloat(val, 64); err == nil {
				// Convert from device format (raw value / 100) to Celsius
				tempCelsius := temp / 100
				// Only include reasonable temperatures
				if tempCelsius > -50 && tempCelsius < 100 {
					name := GetParameterName(id)
					temps[name] = tempCelsius
				}
			}
		}
	}

	return temps
}

// CommonParameters defines common device parameters
type CommonParameters struct {
	// Operating mode
	OperatingMode string // H10715
	// Temperature settings
	DesiredTemperature float64 // H11021
	TemperatureMode    int     // H11017
	// Date/Time
	Year  int // H10905
	Month int // H10906
	Day   int // H10907
	// Network
	DHCPEnabled bool
	IPAddress   string
	SubnetMask  string
	Gateway     string
	DNS         string
}

// ExtractCommonParameters extracts commonly used parameters
func (d *DeviceData) ExtractCommonParameters() *CommonParameters {
	params := &CommonParameters{}

	if val, ok := d.GetValue("H10715"); ok {
		params.OperatingMode = val
	}

	if val, err := d.GetFloatValue("H11021"); err == nil {
		params.DesiredTemperature = val
	}

	if val, err := d.GetIntValue("H11017"); err == nil {
		params.TemperatureMode = val
	}

	if val, err := d.GetIntValue("H10905"); err == nil {
		params.Year = val
	}

	if val, err := d.GetIntValue("H10906"); err == nil {
		params.Month = val
	}

	if val, err := d.GetIntValue("H10907"); err == nil {
		params.Day = val
	}

	return params
}

// IPParameterEncoder encodes IP address octets into device parameter values
// The device expects IP parts as: low=octet1+(octet2<<8), high=octet3+(octet4<<8)
func IPParameterEncoder(ip string) (map[string]string, error) {
	parts := strings.Split(ip, ".")
	if len(parts) != 4 {
		return nil, nil
	}

	octets := make([]int, 4)
	for i, p := range parts {
		val, _ := strconv.Atoi(p)
		octets[i] = val
	}

	result := make(map[string]string)
	// Low 16-bit: octet1 + (octet2 << 8)
	low := octets[0] + (octets[1] << 8)
	result["low"] = strconv.Itoa(low)
	// High 16-bit: octet3 + (octet4 << 8)
	high := octets[2] + (octets[3] << 8)
	result["high"] = strconv.Itoa(high)

	return result, nil
}

// IPParameterDecoder decodes device parameter values back to IP address
func IPParameterDecoder(low, high int) string {
	octet1 := low & 0xFF
	octet2 := (low >> 8) & 0xFF
	octet3 := high & 0xFF
	octet4 := (high >> 8) & 0xFF

	return strconv.Itoa(octet1) + "." + strconv.Itoa(octet2) + "." +
		strconv.Itoa(octet3) + "." + strconv.Itoa(octet4)
}

// TemperatureControl provides convenience methods for temperature settings
type TemperatureControl struct {
	client *WebClient
}

// NewTemperatureControl creates a temperature control helper
func NewTemperatureControl(client *WebClient) *TemperatureControl {
	return &TemperatureControl{client: client}
}

// SetDesiredTemperature sets the target temperature
// mode can be: 0 (off), 1 (heating), 2 (cooling), etc.
func (tc *TemperatureControl) SetDesiredTemperature(temperature float64, mode int) error {
	params := []string{
		FormatParam("H11021", int(temperature)),
		FormatParam("H11017", mode),
	}
	return tc.client.SetMultipleValues(params)
}

// SystemControl provides convenience methods for system control
type SystemControl struct {
	client *WebClient
}

// NewSystemControl creates a system control helper
func NewSystemControl(client *WebClient) *SystemControl {
	return &SystemControl{client: client}
}

// Reset performs a system reset
func (sc *SystemControl) Reset() error {
	return sc.client.SetValue(FormatParam("C10005", 1))
}

// ClearMode clears the current mode
func (sc *SystemControl) ClearMode() error {
	return sc.client.SetValue(FormatParam("C10007", 1))
}

// SetTimezone sets the timezone offset (in hours from UTC)
func (sc *SystemControl) SetTimezone(offsetHours int) error {
	return sc.client.SetValue(FormatParam("H11400", offsetHours))
}

// SetSystemTime sets the current system date/time
func (sc *SystemControl) SetSystemTime(t time.Time) error {
	params := []string{
		FormatParam("H10905", t.Year()),
		FormatParam("H10906", int(t.Month())),
		FormatParam("H10907", t.Day()),
	}
	return sc.client.SetMultipleValues(params)
}

// SessionManager helps manage authenticated sessions
type SessionManager struct {
	client      *WebClient
	password    string
	sessionFile string
	lastLogin   time.Time
}

// NewSessionManager creates a session manager
func NewSessionManager(client *WebClient, password string) *SessionManager {
	return &SessionManager{
		client:   client,
		password: password,
	}
}

// EnsureAuthenticated ensures the client is authenticated, logging in if necessary
func (sm *SessionManager) EnsureAuthenticated() error {
	if sm.client.IsAuthenticated() {
		return nil
	}

	_, err := sm.client.Login(sm.password)
	if err == nil {
		sm.lastLogin = time.Now()
	}
	return err
}

// Logout clears the current session
func (sm *SessionManager) Logout() {
	sm.client.SetSessionID("")
}

// GetSessionAge returns how long the current session has been active
func (sm *SessionManager) GetSessionAge() time.Duration {
	if !sm.client.IsAuthenticated() {
		return 0
	}
	return time.Since(sm.lastLogin)
}
