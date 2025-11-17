package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// API Response types
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type StatusResponse struct {
	Device          string    `json:"device"`
	IP              string    `json:"ip"`
	IsAuthenticated bool      `json:"is_authenticated"`
	SessionID       string    `json:"session_id,omitempty"`
	ParameterCount  int       `json:"parameter_count"`
	LastUpdate      time.Time `json:"last_update"`
	IndoorTemp      float64   `json:"indoor_temp_celsius"`
	OutdoorTemp     float64   `json:"outdoor_temp_celsius"`
}

type TemperatureResponse struct {
	Indoor    float64   `json:"indoor_celsius"`
	Outdoor   float64   `json:"outdoor_celsius"`
	Timestamp time.Time `json:"timestamp"`
}

type ParameterResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

type ParametersResponse struct {
	Count      int                 `json:"count"`
	Parameters []ParameterResponse `json:"parameters"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// Server state
type Server struct {
	deviceIP       string
	devicePassword string
	client         *WebClient
	deviceData     *DeviceData
	lastUpdate     time.Time
	mutex          sync.RWMutex
}

// NewServer creates a new HTTP server
func NewServer(ip string, password string) *Server {
	return &Server{
		deviceIP:       ip,
		devicePassword: password,
		client:         NewWebClient(ip),
	}
}

// Authenticate with the device
func (s *Server) authenticate() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	sessionID, err := s.client.Login(s.devicePassword)
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Get initial data
	data, err := s.client.GetData()
	if err != nil {
		return fmt.Errorf("failed to get initial data: %w", err)
	}

	deviceData, err := ParseXMLData(data)
	if err != nil {
		return fmt.Errorf("failed to parse data: %w", err)
	}

	s.deviceData = deviceData
	s.lastUpdate = time.Now()

	log.Printf("✓ Authenticated with device (session: %s, %d parameters)", sessionID, len(deviceData.Items))
	return nil
}

// RefreshData updates device data
func (s *Server) refreshData() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	data, err := s.client.GetData()
	if err != nil {
		return err
	}

	deviceData, err := ParseXMLData(data)
	if err != nil {
		return err
	}

	s.deviceData = deviceData
	s.lastUpdate = time.Now()
	return nil
}

// HTTP Handlers

// GET /health - Health check
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := APIResponse{
		Success: true,
		Message: "Server is running",
		Data: map[string]string{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GET /status - Get device status
func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if s.deviceData == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(APIResponse{
			Success: false,
			Error:   "Device not initialized",
		})
		return
	}

	indoorTemp, _ := s.deviceData.GetCurrentTemperature()
	outdoorTemp, _ := s.deviceData.GetOutdoorTemperature()

	status := StatusResponse{
		Device:          "Atrea RD5",
		IP:              s.deviceIP,
		IsAuthenticated: s.client.IsAuthenticated(),
		SessionID:       s.client.GetSessionID(),
		ParameterCount:  len(s.deviceData.Items),
		LastUpdate:      s.lastUpdate,
		IndoorTemp:      indoorTemp,
		OutdoorTemp:     outdoorTemp,
	}

	response := APIResponse{
		Success: true,
		Data:    status,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GET /temperature - Get current temperatures
func (s *Server) handleTemperature(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if s.deviceData == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(APIResponse{
			Success: false,
			Error:   "Device not initialized",
		})
		return
	}

	indoor, errIn := s.deviceData.GetCurrentTemperature()
	outdoor, errOut := s.deviceData.GetOutdoorTemperature()

	if errIn != nil || errOut != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(APIResponse{
			Success: false,
			Error:   "Failed to read temperatures",
		})
		return
	}

	temps := TemperatureResponse{
		Indoor:    indoor,
		Outdoor:   outdoor,
		Timestamp: s.lastUpdate,
	}

	response := APIResponse{
		Success: true,
		Data:    temps,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GET /parameters - List all parameters
func (s *Server) handleParameters(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if s.deviceData == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(APIResponse{
			Success: false,
			Error:   "Device not initialized",
		})
		return
	}

	// Parse query parameters for filtering
	limit := r.URL.Query().Get("limit")
	limitInt := 0
	if limit != "" {
		limitInt, _ = strconv.Atoi(limit)
	}

	var params []ParameterResponse
	count := 0
	for id, value := range s.deviceData.Items {
		name := GetParameterName(id)
		params = append(params, ParameterResponse{
			ID:    id,
			Name:  name,
			Value: value,
		})
		count++
		if limitInt > 0 && count >= limitInt {
			break
		}
	}

	result := ParametersResponse{
		Count:      len(params),
		Parameters: params,
	}

	response := APIResponse{
		Success: true,
		Data:    result,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GET /parameter/:id - Get single parameter
func (s *Server) handleParameter(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract parameter ID from path
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/parameter/"), "/")
	paramID := parts[0]

	if paramID == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(APIResponse{
			Success: false,
			Error:   "Missing parameter ID",
		})
		return
	}

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if s.deviceData == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(APIResponse{
			Success: false,
			Error:   "Device not initialized",
		})
		return
	}

	value, ok := s.deviceData.Items[paramID]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Parameter %s not found", paramID),
		})
		return
	}

	param := ParameterResponse{
		ID:    paramID,
		Name:  GetParameterName(paramID),
		Value: value,
	}

	response := APIResponse{
		Success: true,
		Data:    param,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// POST /refresh - Refresh device data
func (s *Server) handleRefresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := s.refreshData(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to refresh data: %v", err),
		})
		return
	}

	response := APIResponse{
		Success: true,
		Message: "Device data refreshed",
		Data: map[string]interface{}{
			"timestamp": time.Now(),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Middleware for CORS
func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

// Middleware for logging
func loggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.Method, r.URL.Path, r.RemoteAddr)
		next(w, r)
	}
}

// Combined middleware
func (s *Server) withMiddleware(handler http.HandlerFunc) http.HandlerFunc {
	return loggingMiddleware(corsMiddleware(handler))
}

// StartServer starts the HTTP server
func (s *Server) StartServer(port int) error {
	// Authenticate first
	if err := s.authenticate(); err != nil {
		return err
	}

	// Setup routes
	http.HandleFunc("/health", s.withMiddleware(s.handleHealth))
	http.HandleFunc("/status", s.withMiddleware(s.handleStatus))
	http.HandleFunc("/temperature", s.withMiddleware(s.handleTemperature))
	http.HandleFunc("/parameters", s.withMiddleware(s.handleParameters))
	http.HandleFunc("/parameter/", s.withMiddleware(s.handleParameter))
	http.HandleFunc("/refresh", s.withMiddleware(s.handleRefresh))

	addr := fmt.Sprintf(":%d", port)
	log.Printf("🚀 Starting web server on %s", addr)
	log.Printf("Available endpoints:")
	log.Printf("  GET  /health             - Health check")
	log.Printf("  GET  /status             - Device status and temperatures")
	log.Printf("  GET  /temperature        - Current temperatures (indoor/outdoor)")
	log.Printf("  GET  /parameters         - List all parameters (?limit=10 to limit)")
	log.Printf("  GET  /parameter/:id      - Get specific parameter (e.g. /parameter/I10215)")
	log.Printf("  POST /refresh            - Refresh device data")

	return http.ListenAndServe(addr, nil)
}
