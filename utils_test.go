package main

import (
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"
)

// TestParseXMLDataWithRealResponse tests parsing real device response
func TestParseXMLDataWithRealResponse(t *testing.T) {
	// Load captured test data
	configPath := filepath.Join("testdata", "response_config.xml")
	configData, err := ioutil.ReadFile(configPath)
	if err != nil {
		t.Skipf("skipping test: cannot load test data (%v)", err)
	}

	deviceData, err := ParseXMLData(string(configData))
	if err != nil {
		t.Fatalf("failed to parse XML: %v", err)
	}

	// Verify basic parsing
	if len(deviceData.Items) == 0 {
		t.Error("no parameters parsed from XML")
	}

	// Check for expected temperature parameters
	expectedParams := []string{"I10215", "I10211"}
	for _, param := range expectedParams {
		if _, ok := deviceData.Items[param]; !ok {
			t.Errorf("missing expected parameter: %s", param)
		}
	}
}

// TestParseXMLDataBasic tests basic XML parsing with small sample
func TestParseXMLDataBasic(t *testing.T) {
	xml := `<?xml version="1.0"?>
<RD5WEB>
  <RD5>
    <INTEGER_R>
      <O I="I10215" V="201"/>
      <O I="I10211" V="36"/>
    </INTEGER_R>
    <FLOAT_R>
      <O I="I10230" V="50.5"/>
    </FLOAT_R>
  </RD5>
</RD5WEB>`

	deviceData, err := ParseXMLData(xml)
	if err != nil {
		t.Fatalf("failed to parse XML: %v", err)
	}

	tests := []struct {
		paramID      string
		expectedVal  string
		description  string
	}{
		{"I10215", "201", "indoor temperature"},
		{"I10211", "36", "outdoor temperature"},
		{"I10230", "50.5", "fan speed"},
	}

	for _, tt := range tests {
		val, ok := deviceData.Items[tt.paramID]
		if !ok {
			t.Errorf("missing parameter %s (%s)", tt.paramID, tt.description)
			continue
		}

		if val != tt.expectedVal {
			t.Errorf("%s: got %s, want %s", tt.description, val, tt.expectedVal)
		}
	}
}

// TestDecodeTemperaturePositive tests positive temperature decoding
func TestDecodeTemperaturePositive(t *testing.T) {
	tests := []struct {
		rawValue    float64
		expectedDeg float64
		description string
	}{
		{10, 1.0, "1°C"},
		{100, 10.0, "10°C"},
		{201, 20.1, "20.1°C (indoor)"},
		{215, 21.5, "21.5°C"},
		{360, 36.0, "36°C"},
		{1, 0.1, "0.1°C minimum"},
		{1300, 130.0, "130°C maximum"},
	}

	for _, tt := range tests {
		got := decodeTemperature(tt.rawValue)
		if got != tt.expectedDeg {
			t.Errorf("%s: got %.1f°C, want %.1f°C", tt.description, got, tt.expectedDeg)
		}
	}
}

// TestDecodeTemperatureNegative tests negative temperature decoding (two's complement)
func TestDecodeTemperatureNegative(t *testing.T) {
	tests := []struct {
		rawValue    float64
		expectedDeg float64
		description string
	}{
		{65526, -1.0, "-1°C"},
		{65536, 0.0, "zero (boundary)"},  // 65536 mod 65536 = 0
		{65036, -50.0, "-50°C minimum"},
		{65535, -0.1, "-0.1°C near zero"},
		{65496, -4.0, "-4°C"},
	}

	for _, tt := range tests {
		got := decodeTemperature(tt.rawValue)
		if got != tt.expectedDeg {
			t.Errorf("%s: got %.1f°C, want %.1f°C", tt.description, got, tt.expectedDeg)
		}
	}
}

// TestDecodeTemperatureEdgeCases tests edge cases and boundary conditions
func TestDecodeTemperatureEdgeCases(t *testing.T) {
	tests := []struct {
		rawValue    float64
		shouldBeValid bool
		description string
	}{
		{0, true, "zero (out of range, returns as-is)"},
		{1300, true, "1300 (positive max)"},
		{1301, true, "1301 (just above positive max, out of range)"},
		{65035, true, "65035 (just below negative min)"},
		{65036, true, "65036 (negative min)"},
	}

	for _, tt := range tests {
		// Just verify it doesn't panic
		_ = decodeTemperature(tt.rawValue)
	}
}

// TestGetCurrentTemperatureWithRealData tests getting temperature with real device response
func TestGetCurrentTemperatureWithRealData(t *testing.T) {
	configPath := filepath.Join("testdata", "response_config.xml")
	configData, err := ioutil.ReadFile(configPath)
	if err != nil {
		t.Skipf("skipping test: cannot load test data (%v)", err)
	}

	deviceData, err := ParseXMLData(string(configData))
	if err != nil {
		t.Fatalf("failed to parse XML: %v", err)
	}

	temp, err := deviceData.GetCurrentTemperature()
	if err != nil {
		t.Fatalf("failed to get temperature: %v", err)
	}

	// Sanity check - room temperature should be between 0-30°C
	if temp < 0 || temp > 40 {
		t.Errorf("unreasonable temperature: %.1f°C", temp)
	}
}

// TestGetOutdoorTemperatureWithRealData tests getting outdoor temperature with real device response
func TestGetOutdoorTemperatureWithRealData(t *testing.T) {
	configPath := filepath.Join("testdata", "response_config.xml")
	configData, err := ioutil.ReadFile(configPath)
	if err != nil {
		t.Skipf("skipping test: cannot load test data (%v)", err)
	}

	deviceData, err := ParseXMLData(string(configData))
	if err != nil {
		t.Fatalf("failed to parse XML: %v", err)
	}

	temp, err := deviceData.GetOutdoorTemperature()
	if err != nil {
		t.Fatalf("failed to get temperature: %v", err)
	}

	// Sanity check - outdoor temperature should be between -50 to 50°C
	if temp < -50 || temp > 50 {
		t.Errorf("unreasonable outdoor temperature: %.1f°C", temp)
	}
}

// TestGetParameterName tests parameter name lookup
func TestGetParameterName(t *testing.T) {
	tests := []struct {
		paramID      string
		expectedName string
	}{
		{"I10215", "Indoor Air Temperature (T-IDA)"},
		{"I10211", "Outdoor Air Temperature (T-ODA)"},
		{"I10212", "Supply Air Temperature (T-SUP)"},
		{"H11021", "Desired Temperature"},
		{"UNKNOWN", "UNKNOWN"},
	}

	for _, tt := range tests {
		name := GetParameterName(tt.paramID)
		if name != tt.expectedName {
			t.Errorf("param %s: got name %q, want %q", tt.paramID, name, tt.expectedName)
		}
	}
}

// TestAlarmsParsingWithRealData tests parsing alarm data with real response
func TestAlarmsParsingWithRealData(t *testing.T) {
	alarmsPath := filepath.Join("testdata", "response_alarms.xml")
	alarmsData, err := ioutil.ReadFile(alarmsPath)
	if err != nil {
		t.Skipf("skipping test: cannot load test data (%v)", err)
	}

	// Just verify it's valid XML and contains expected root
	xmlStr := string(alarmsData)
	if !strings.Contains(xmlStr, "<root>") && !strings.Contains(xmlStr, "<errors>") {
		t.Error("alarms XML missing expected root element")
	}
}
