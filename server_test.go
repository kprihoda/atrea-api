package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestHealthEndpoint tests the health check endpoint
func TestHealthEndpoint(t *testing.T) {
	// Create a test server with mock device data
	server := &Server{
		deviceIP:       "192.168.68.106",
		devicePassword: "6378",
		client:         NewWebClient("192.168.68.106"),
	}

	// Create a mock HTTP server
	ts := httptest.NewServer(http.HandlerFunc(server.handleHealth))
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	var result APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !result.Success {
		t.Error("expected success=true")
	}
}

// TestStatusEndpoint tests the status endpoint without real device
func TestStatusEndpointWithoutDevice(t *testing.T) {
	server := &Server{
		deviceIP:       "192.168.68.106",
		devicePassword: "6378",
		client:         NewWebClient("192.168.68.106"),
		deviceData:     nil, // Not initialized
	}

	ts := httptest.NewServer(http.HandlerFunc(server.handleStatus))
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("expected status 503, got %d", resp.StatusCode)
	}

	var result APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if result.Success {
		t.Error("expected success=false for uninitialized device")
	}
}

// TestParameterEndpoint tests getting a single parameter
func TestParameterEndpoint(t *testing.T) {
	// Load test data
	configPath := "testdata/response_config.xml"
	configData, err := ioutil.ReadFile(configPath)
	if err != nil {
		t.Skipf("skipping: cannot load test data (%v)", err)
	}

	deviceData, err := ParseXMLData(string(configData))
	if err != nil {
		t.Fatalf("failed to parse test data: %v", err)
	}

	server := &Server{
		deviceIP:       "192.168.68.106",
		devicePassword: "6378",
		client:         NewWebClient("192.168.68.106"),
		deviceData:     deviceData,
	}

	// Test getting existing parameter
	req := httptest.NewRequest("GET", "/parameter/I10215", nil)
	w := httptest.NewRecorder()
	server.handleParameter(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var result APIResponse
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !result.Success {
		t.Error("expected success=true")
	}
}

// TestParameterEndpointNotFound tests getting non-existent parameter
func TestParameterEndpointNotFound(t *testing.T) {
	configPath := "testdata/response_config.xml"
	configData, err := ioutil.ReadFile(configPath)
	if err != nil {
		t.Skipf("skipping: cannot load test data (%v)", err)
	}

	deviceData, err := ParseXMLData(string(configData))
	if err != nil {
		t.Fatalf("failed to parse test data: %v", err)
	}

	server := &Server{
		deviceIP:       "192.168.68.106",
		devicePassword: "6378",
		client:         NewWebClient("192.168.68.106"),
		deviceData:     deviceData,
	}

	// Test getting non-existent parameter
	req := httptest.NewRequest("GET", "/parameter/UNKNOWN", nil)
	w := httptest.NewRecorder()
	server.handleParameter(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}

	var result APIResponse
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if result.Success {
		t.Error("expected success=false for missing parameter")
	}
}

// TestTemperatureEndpoint tests temperature endpoint
func TestTemperatureEndpoint(t *testing.T) {
	configPath := "testdata/response_config.xml"
	configData, err := ioutil.ReadFile(configPath)
	if err != nil {
		t.Skipf("skipping: cannot load test data (%v)", err)
	}

	deviceData, err := ParseXMLData(string(configData))
	if err != nil {
		t.Fatalf("failed to parse test data: %v", err)
	}

	server := &Server{
		deviceIP:       "192.168.68.106",
		devicePassword: "6378",
		client:         NewWebClient("192.168.68.106"),
		deviceData:     deviceData,
	}

	req := httptest.NewRequest("GET", "/temperature", nil)
	w := httptest.NewRecorder()
	server.handleTemperature(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var result APIResponse
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !result.Success {
		t.Error("expected success=true")
	}

	tempResp, ok := result.Data.(map[string]interface{})
	if !ok {
		t.Fatal("expected temperature response data")
	}

	if _, hasIndoor := tempResp["indoor_celsius"]; !hasIndoor {
		t.Error("missing indoor_celsius in response")
	}
	if _, hasOutdoor := tempResp["outdoor_celsius"]; !hasOutdoor {
		t.Error("missing outdoor_celsius in response")
	}
}

// TestParametersEndpoint tests listing parameters
func TestParametersEndpoint(t *testing.T) {
	configPath := "testdata/response_config.xml"
	configData, err := ioutil.ReadFile(configPath)
	if err != nil {
		t.Skipf("skipping: cannot load test data (%v)", err)
	}

	deviceData, err := ParseXMLData(string(configData))
	if err != nil {
		t.Fatalf("failed to parse test data: %v", err)
	}

	server := &Server{
		deviceIP:       "192.168.68.106",
		devicePassword: "6378",
		client:         NewWebClient("192.168.68.106"),
		deviceData:     deviceData,
	}

	req := httptest.NewRequest("GET", "/parameters", nil)
	w := httptest.NewRecorder()
	server.handleParameters(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var result APIResponse
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !result.Success {
		t.Error("expected success=true")
	}
}

// TestCORSMiddleware tests CORS headers are set
func TestCORSMiddleware(t *testing.T) {
	handler := corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	corsOrigin := w.Header().Get("Access-Control-Allow-Origin")
	if corsOrigin != "*" {
		t.Errorf("expected CORS origin *, got %s", corsOrigin)
	}

	corsMethods := w.Header().Get("Access-Control-Allow-Methods")
	if corsMethods == "" {
		t.Error("missing CORS methods header")
	}
}
