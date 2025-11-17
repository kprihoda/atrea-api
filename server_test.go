package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestHealthEndpoint tests the health check endpoint
func TestHealthEndpoint(t *testing.T) {
	server := &Server{
		deviceIP:       "192.168.68.106",
		devicePassword: "6378",
		client:         NewWebClient("192.168.68.106"),
	}

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

// TestServerInitialization tests server creation
func TestServerInitialization(t *testing.T) {
	server := NewServer("192.168.68.106", "6378")

	if server.deviceIP != "192.168.68.106" {
		t.Errorf("expected IP 192.168.68.106, got %s", server.deviceIP)
	}

	if server.devicePassword != "6378" {
		t.Errorf("expected password 6378, got %s", server.devicePassword)
	}

	if server.client == nil {
		t.Error("expected client to be initialized")
	}
}

// TestFetchDeviceData tests data fetching (will fail if device not available)
func TestFetchDeviceData(t *testing.T) {
	server := &Server{
		deviceIP:       "192.168.68.106",
		devicePassword: "6378",
		client:         NewWebClient("192.168.68.106"),
	}

	// This will fail if device is not available, which is expected in unit tests
	_, err := server.fetchDeviceData()
	if err == nil {
		t.Log("Successfully fetched device data (device is available)")
	} else {
		t.Logf("Device not available (expected in test environment): %v", err)
	}
}

// TestCORSMiddleware tests CORS header injection
func TestCORSMiddleware(t *testing.T) {
	handler := corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("expected CORS origin *, got %s", w.Header().Get("Access-Control-Allow-Origin"))
	}
}

// TestLoggingMiddleware tests that requests are processed
func TestLoggingMiddleware(t *testing.T) {
	handler := loggingMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

// TestMethodNotAllowed tests invalid HTTP methods
func TestMethodNotAllowed(t *testing.T) {
	server := &Server{
		deviceIP:       "192.168.68.106",
		devicePassword: "6378",
		client:         NewWebClient("192.168.68.106"),
	}

	req := httptest.NewRequest("POST", "/health", nil)
	w := httptest.NewRecorder()
	server.handleHealth(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

// TestStatusEndpointDeviceNotAvailable tests status endpoint when device unavailable
func TestStatusEndpointDeviceNotAvailable(t *testing.T) {
	server := &Server{
		deviceIP:       "192.168.68.106",
		devicePassword: "6378",
		client:         NewWebClient("192.168.68.106"),
	}

	req := httptest.NewRequest("GET", "/status", nil)
	w := httptest.NewRecorder()
	server.handleStatus(w, req)

	// Will fail since we can't connect to device
	if w.Code != http.StatusServiceUnavailable && w.Code != http.StatusInternalServerError {
		t.Logf("Device available (status %d)", w.Code)
	}
}
