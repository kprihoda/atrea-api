package main

import (
	"crypto/md5"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// WebClient provides access to the Atrea RD5 web API
type WebClient struct {
	baseURL    string
	auth       string
	httpClient *http.Client
}

// NewWebClient creates a new web client for the Atrea RD5
func NewWebClient(ip string) *WebClient {
	return &WebClient{
		baseURL:    "http://" + ip,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// Login authenticates with the device using the password
//
// AUTHENTICATION FLOW:
// 1. Create MD5 hash of literal string "\r\n" + password (e.g., "\r\n6378")
// 2. GET /config/login.cgi?magic=<HASH>&rnd=<RANDOM_NUMBER>
// 3. Device returns XML: <?xml version="1.0"?><root lng="0">XXXXX</root>
// 4. Extract 5-digit session ID from between <root> and </root>
// 5. Use session ID in all subsequent requests via auth parameter
//
// Example:
//
//	password: "6378"
//	hash of "\r\n6378": 993278d1925c378ab94a6fe664ea6c60
//	request: GET /config/login.cgi?magic=993278d1925c378ab94a6fe664ea6c60&rnd=123
//	response: <?xml version="1.0" encoding="UTF-8"?><root lng="0">15736</root>
//	sessionID: "15736"
func (wc *WebClient) Login(password string) (string, error) {
	// STEP 1: Create MD5 hash of "\r\n" + password
	// CRITICAL: The hash input is the literal string with actual carriage return and newline
	hash := md5.New()
	io.WriteString(hash, "\r\n"+password)
	magic := fmt.Sprintf("%x", hash.Sum(nil))

	// STEP 2: Generate random number for nonce (prevents replay attacks, any random digits work)
	randStr := generateRandomString(3)

	// STEP 3: Call login endpoint with magic hash and random nonce
	params := url.Values{}
	params.Set("magic", magic)
	params.Set("rnd", randStr)

	resp, err := wc.httpClient.Get(wc.baseURL + "/config/login.cgi?" + params.Encode())
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	responseStr := strings.TrimSpace(string(body))

	// STEP 4: Extract session ID from XML response
	// Response format expected:
	//   <?xml version="1.0" encoding="UTF-8"?><root lng="0">XXXXX</root>
	// Robustly locate the content inside the <root> element.
	if rootStart := strings.Index(responseStr, "<root"); rootStart != -1 {
		// find the '>' that closes the opening <root ...> tag
		if gt := strings.Index(responseStr[rootStart:], ">"); gt != -1 {
			start := rootStart + gt + 1
			if endTag := strings.Index(responseStr, "</root>"); endTag != -1 && start < endTag {
				sessionID := strings.TrimSpace(responseStr[start:endTag])
				// Validate: must not be empty, "0", or "denied"; must be numeric
				if sessionID != "" && sessionID != "0" && sessionID != "denied" {
					if _, err := strconv.Atoi(sessionID); err == nil {
						wc.auth = sessionID
						return sessionID, nil
					}
				}
			}
		}
	}

	// If we got here, either parsing failed or response was "denied"
	return "", fmt.Errorf("authentication failed: invalid response from device")

} // GetData retrieves the XML configuration data from the device
func (wc *WebClient) GetData() (string, error) {
	params := url.Values{}
	if wc.auth != "" {
		params.Set("auth", wc.auth)
	}
	params.Set("rnd", generateRandomString(2))

	resp, err := wc.httpClient.Get(wc.baseURL + "/config/xml.xml?" + params.Encode())
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	return string(body), err
}

// SetValue sends a parameter update to the device
// Parameter should be in format like "H12345=1000"
func (wc *WebClient) SetValue(parameter string) error {
	params := url.Values{}
	params.Set("auth", wc.auth)
	params.Set(strings.Split(parameter, "=")[0], strings.Split(parameter, "=")[1])

	resp, err := wc.httpClient.Get(wc.baseURL + "/config/xml.cgi?" + params.Encode())
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to set value: status %d", resp.StatusCode)
	}

	return nil
}

// SetMultipleValues sends multiple parameter updates to the device
// Parameters should be in format like []string{"H12345=1000", "H12346=2000"}
func (wc *WebClient) SetMultipleValues(parameters []string) error {
	params := url.Values{}
	params.Set("auth", wc.auth)

	for _, param := range parameters {
		parts := strings.Split(param, "=")
		if len(parts) == 2 {
			params.Set(parts[0], parts[1])
		}
	}

	resp, err := wc.httpClient.Get(wc.baseURL + "/config/xml.cgi?" + params.Encode())
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to set values: status %d", resp.StatusCode)
	}

	return nil
}

// GetAlarms retrieves alarm information from the device
func (wc *WebClient) GetAlarms() (string, error) {
	params := url.Values{}
	if wc.auth != "" {
		params.Set("auth", wc.auth)
	}
	params.Set("rnd", generateRandomString(2))

	resp, err := wc.httpClient.Get(wc.baseURL + "/config/alarms.xml?" + params.Encode())
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	return string(body), err
}

// GetWeeklyProgram retrieves weekly program settings
// deviceType can be "RTS" or "RNS"
// programType can be "vzt" or "izt"
func (wc *WebClient) GetWeeklyProgram(deviceType, programType string) (string, error) {
	var endpoint string
	if deviceType == "RTS" {
		if programType == "vzt" {
			endpoint = "/config/rtssetup.xml"
		} else {
			endpoint = "/config/rgtssetup.xml"
		}
	} else if deviceType == "RNS" {
		if programType == "vzt" {
			endpoint = "/config/rnssetup.xml"
		} else {
			endpoint = "/config/rgnssetup.xml"
		}
	} else {
		return "", fmt.Errorf("invalid device type: %s", deviceType)
	}

	params := url.Values{}
	if wc.auth != "" {
		params.Set("auth", wc.auth)
	}
	params.Set("rnd", generateRandomString(2))

	resp, err := wc.httpClient.Get(wc.baseURL + endpoint + "?" + params.Encode())
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	return string(body), err
}

// SetWeeklyProgram updates weekly program settings
// deviceType can be "RTS" or "RNS"
// programType can be "vzt" or "izt"
func (wc *WebClient) SetWeeklyProgram(deviceType, programType, data string) error {
	var endpoint string
	if deviceType == "RTS" {
		if programType == "vzt" {
			endpoint = "/config/rtssetup.cgi"
		} else {
			endpoint = "/config/rgtssetup.cgi"
		}
	} else if deviceType == "RNS" {
		if programType == "vzt" {
			endpoint = "/config/rnssetup.cgi"
		} else {
			endpoint = "/config/rgnssetup.cgi"
		}
	} else {
		return fmt.Errorf("invalid device type: %s", deviceType)
	}

	params := url.Values{}
	params.Set("auth", wc.auth)
	params.Set("rnd", generateRandomString(2))

	// Append data to query string
	fullURL := wc.baseURL + endpoint + "?" + params.Encode() + "&" + data

	resp, err := wc.httpClient.Get(fullURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to set weekly program: status %d", resp.StatusCode)
	}

	return nil
}

// GetNetworkSettings retrieves network configuration
func (wc *WebClient) GetNetworkSettings() (string, error) {
	params := url.Values{}
	if wc.auth != "" {
		params.Set("auth", wc.auth)
	}
	params.Set("rnd", generateRandomString(2))

	resp, err := wc.httpClient.Get(wc.baseURL + "/config/ip.cgi?" + params.Encode())
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	return string(body), err
}

// SetNetworkSettings updates network configuration
// Example: "dhcp=1" or "dhcp=0&ip=192168068106&ip4mask=255255255000..."
func (wc *WebClient) SetNetworkSettings(settings string) error {
	params := url.Values{}
	params.Set("auth", wc.auth)
	params.Set("rnd", generateRandomString(2))

	fullURL := wc.baseURL + "/config/ip.cgi?" + params.Encode() + "&" + settings

	resp, err := wc.httpClient.Get(fullURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to set network settings: status %d", resp.StatusCode)
	}

	return nil
}

// IsAuthenticated returns whether the client has an active session
func (wc *WebClient) IsAuthenticated() bool {
	return wc.auth != ""
}

// GetSessionID returns the current session ID (auth token)
func (wc *WebClient) GetSessionID() string {
	return wc.auth
}

// SetSessionID sets the session ID manually (useful for restoring sessions)
func (wc *WebClient) SetSessionID(sessionID string) {
	wc.auth = sessionID
}

// Helper function to generate random string (like the JS randStr function)
func generateRandomString(length int) string {
	rand.Seed(time.Now().UnixNano())
	const charset = "0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// Helper function to format parameter for SetValue
// Example: FormatParam("H12345", 1000) -> "H12345=1000"
func FormatParam(key string, value interface{}) string {
	return fmt.Sprintf("%s=%v", key, value)
}

// Helper function to convert two 16-bit values to IP address parts
// Used for parsing network settings
func ValuesToIPArray(low, high int32) [4]int {
	if high < 0 {
		high += 65536
	}
	if low < 0 {
		low += 65536
	}

	lowHex := fmt.Sprintf("%04x", low)
	highHex := fmt.Sprintf("%04x", high)

	var result [4]int
	result[0], _ = strconv.Atoi(lowHex[2:4])
	result[1], _ = strconv.Atoi(lowHex[0:2])
	result[2], _ = strconv.Atoi(highHex[2:4])
	result[3], _ = strconv.Atoi(highHex[0:2])

	return result
}
