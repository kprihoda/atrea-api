package main

import (
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestLoginSuccess tests successful authentication with valid credentials
func TestLoginSuccess(t *testing.T) {
	// Calculate expected MD5 hash
	hash := md5.New()
	io.WriteString(hash, "\r\n6378")
	expectedHash := fmt.Sprintf("%x", hash.Sum(nil))

	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/config/login.cgi" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		// Check query parameters
		magic := r.URL.Query().Get("magic")
		if magic != expectedHash {
			t.Errorf("invalid magic hash: got %s, want %s", magic, expectedHash)
		}

		if r.URL.Query().Get("rnd") == "" {
			t.Error("missing rnd parameter")
		}

		// Return valid login response with session ID
		w.Header().Set("Content-Type", "application/xml")
		fmt.Fprint(w, `<?xml version="1.0" encoding="UTF-8"?><root lng="0">12345</root>`)
	}))
	defer server.Close()

	// Test
	client := NewWebClient(server.Listener.Addr().String())
	client.baseURL = server.URL

	sessionID, err := client.Login("6378")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if sessionID != "12345" {
		t.Errorf("got session ID %s, want 12345", sessionID)
	}
}

// TestLoginFailure tests authentication failure when device returns "denied"
func TestLoginFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		fmt.Fprint(w, `<?xml version="1.0" encoding="UTF-8"?><root lng="0">denied</root>`)
	}))
	defer server.Close()

	client := NewWebClient(server.Listener.Addr().String())
	client.baseURL = server.URL

	_, err := client.Login("wrongpassword")
	if err == nil {
		t.Error("expected error for denied response, got nil")
	}

	if !strings.Contains(err.Error(), "authentication failed") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// TestLoginInvalidResponse tests handling of malformed responses
func TestLoginInvalidResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		// No root tag in response
		fmt.Fprint(w, `<?xml version="1.0" encoding="UTF-8"?>`)
	}))
	defer server.Close()

	client := NewWebClient(server.Listener.Addr().String())
	client.baseURL = server.URL

	_, err := client.Login("6378")
	if err == nil {
		t.Error("expected error for invalid response, got nil")
	}
}

// TestGetData tests retrieving configuration data
func TestGetData(t *testing.T) {
	expectedData := `<?xml version="1.0"?><RD5WEB><RD5><INTEGER_R><O I="I10215" V="201"/></INTEGER_R></RD5></RD5WEB>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/config/xml.xml" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		if r.URL.Query().Get("auth") == "" {
			t.Error("missing auth parameter")
		}

		w.Header().Set("Content-Type", "application/xml")
		fmt.Fprint(w, expectedData)
	}))
	defer server.Close()

	client := NewWebClient(server.Listener.Addr().String())
	client.baseURL = server.URL
	client.auth = "12345"

	data, err := client.GetData()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if data != expectedData {
		t.Errorf("got data %s, want %s", data, expectedData)
	}
}

// TestGetAlarms tests retrieving alarm data
func TestGetAlarms(t *testing.T) {
	expectedData := `<?xml version="1.0"?><RD5WEB><ALARMS><ALARM>No alarms</ALARM></ALARMS></RD5WEB>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/config/alarms.xml" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/xml")
		fmt.Fprint(w, expectedData)
	}))
	defer server.Close()

	client := NewWebClient(server.Listener.Addr().String())
	client.baseURL = server.URL
	client.auth = "12345"

	data, err := client.GetAlarms()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if data != expectedData {
		t.Errorf("got data %s, want %s", data, expectedData)
	}
}

// TestSetValue tests setting a single parameter
func TestSetValue(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/config/xml.cgi" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		// Verify parameters
		if r.URL.Query().Get("auth") != "12345" {
			t.Errorf("invalid auth: got %s, want 12345", r.URL.Query().Get("auth"))
		}

		if r.URL.Query().Get("H11021") != "21" {
			t.Errorf("invalid parameter: got %s, want 21", r.URL.Query().Get("H11021"))
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewWebClient(server.Listener.Addr().String())
	client.baseURL = server.URL
	client.auth = "12345"

	err := client.SetValue("H11021=21")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestIsAuthenticated tests session ID validation
func TestIsAuthenticated(t *testing.T) {
	client := NewWebClient("192.168.68.106")

	if client.IsAuthenticated() {
		t.Error("should not be authenticated without session ID")
	}

	client.SetSessionID("12345")
	if !client.IsAuthenticated() {
		t.Error("should be authenticated with session ID")
	}
}
