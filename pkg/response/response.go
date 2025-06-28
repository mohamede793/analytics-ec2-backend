// pkg/response/response.go
package response

import (
	"encoding/json"
	"net/http"
	"os"
	"time"
)

// StandardResponse is the consistent structure for all API responses
type StandardResponse struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	Error     *ErrorInfo  `json:"error,omitempty"`
	Meta      *Meta       `json:"meta,omitempty"`
	Server    string      `json:"server"`
	Timestamp string      `json:"timestamp"`
}

// ErrorInfo provides detailed error information
type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// Meta provides additional response metadata
type Meta struct {
	Count       int `json:"count,omitempty"`
	Total       int `json:"total,omitempty"`
	ProcessedAt string `json:"processed_at,omitempty"`
}

// Success sends a successful response
func Success(w http.ResponseWriter, statusCode int, message string, data interface{}) {
	respond(w, statusCode, &StandardResponse{
		Success:   true,
		Message:   message,
		Data:      data,
		Server:    getServerName(),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
}

// SuccessWithMeta sends a successful response with metadata
func SuccessWithMeta(w http.ResponseWriter, statusCode int, message string, data interface{}, meta *Meta) {
	respond(w, statusCode, &StandardResponse{
		Success:   true,
		Message:   message,
		Data:      data,
		Meta:      meta,
		Server:    getServerName(),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
}

// Error sends an error response
func Error(w http.ResponseWriter, statusCode int, code, message, details string) {
	respond(w, statusCode, &StandardResponse{
		Success: false,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
			Details: details,
		},
		Server:    getServerName(),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
}

// Predefined error responses for common cases
func BadRequest(w http.ResponseWriter, details string) {
	Error(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid request data", details)
}

func Unauthorized(w http.ResponseWriter, details string) {
	Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required", details)
}

func NotFound(w http.ResponseWriter, resource string) {
	Error(w, http.StatusNotFound, "NOT_FOUND", "Resource not found", resource)
}

func ValidationError(w http.ResponseWriter, details string) {
	Error(w, http.StatusBadRequest, "VALIDATION_ERROR", "Validation failed", details)
}

func InternalError(w http.ResponseWriter, details string) {
	Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error", details)
}

func TooManyRequests(w http.ResponseWriter, details string) {
	Error(w, http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED", "Too many requests", details)
}

// Helper functions
func respond(w http.ResponseWriter, statusCode int, response *StandardResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

func getServerName() string {
	if hostname, err := os.Hostname(); err == nil {
		return hostname
	}
	return "github.com/mohamede793/analytics-ec2-backend"
}