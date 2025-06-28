package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)
//dummy commit

// Standard API response structure
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Server  string      `json:"server"`
}

// Data structure for the name endpoint
type NameData struct {
	Name    string `json:"name"`
	Greeting string `json:"greeting"`
}

// Helper function to get server hostname
func getServerName() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return hostname
}

// Helper function to send JSON responses
func sendResponse(w http.ResponseWriter, statusCode int, success bool, message string, data interface{}, errorMsg string) {
	response := APIResponse{
		Success: success,
		Message: message,
		Data:    data,
		Error:   errorMsg,
		Server:  getServerName(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// Health check endpoint for ALB
func healthHandler(w http.ResponseWriter, r *http.Request) {
	sendResponse(w, http.StatusOK, true, "This is the final setup testing server", nil, "")
}

// Name endpoint - accepts name parameter and returns greeting
func nameHandler(w http.ResponseWriter, r *http.Request) {
	// Get API key from Authorization header only
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		sendResponse(w, http.StatusUnauthorized, false, "Unauthorized", nil, "Authorization header with Bearer token is required")
		return
	}
	
	// Extract the token
	apiKey := strings.TrimPrefix(authHeader, "Bearer ")
	
	// Validate API key is provided
	if apiKey == "" {
		sendResponse(w, http.StatusUnauthorized, false, "Unauthorized", nil, "API key is required")
		return
	}
	
	// Get expected API key from environment variable
	expectedAPIKey := os.Getenv("API_KEY")
	if expectedAPIKey == "" {
		sendResponse(w, http.StatusInternalServerError, false, "Internal Server Error", nil, "Server configuration error")
		return
	}
	
	// Validate API key matches
	if apiKey != expectedAPIKey {
		sendResponse(w, http.StatusUnauthorized, false, "Unauthorized", nil, "Invalid API key")
		return
	}

	// Get name from query parameter
	name := r.URL.Query().Get("name")
	
	// Validate input
	if name == "" {
		sendResponse(w, http.StatusBadRequest, false, "Bad Request", nil, "Name parameter is required")
		return
	}

	// Create response data
	data := NameData{
		Name:     name,
		Greeting: fmt.Sprintf("Hello, %s! Welcome to our NEW API.", name),
	}

	// Send success response
	sendResponse(w, http.StatusOK, true, "Name processed successfully", data, "")
}

func main() {
	r := mux.NewRouter()

	// Define routes
	r.HandleFunc("/health", healthHandler).Methods("GET")
	r.HandleFunc("/api/name", nameHandler).Methods("GET")

	// Add logging middleware
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Printf("%s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
			next.ServeHTTP(w, r)
		})
	})

	fmt.Printf("Go API server starting on port 3000...\n")
	fmt.Printf("Server: %s\n", getServerName())
	fmt.Println("Available endpoints:")
	fmt.Println("  GET /health - Health check")
	fmt.Println("  GET /api/name?name=YourName - Name greeting endpoint")
	
	log.Fatal(http.ListenAndServe(":3000", r))
}// Deployment test
