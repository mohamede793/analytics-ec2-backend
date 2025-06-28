// internal/auth/api_key.go
package auth

import (
	"net/http"
	"strings"
	"time"
)

type APIKeyAuth struct {
	expectedKey string
}

func NewAPIKeyAuth(apiKey string) *APIKeyAuth {
	return &APIKeyAuth{
		expectedKey: apiKey,
	}
}

// ValidateBearerToken checks if the request has a valid Bearer token
func (a *APIKeyAuth) ValidateBearerToken(r *http.Request) bool {
	authHeader := r.Header.Get("Authorization")
	
	// Check if Authorization header exists and starts with "Bearer "
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return false
	}
	
	// Extract the token
	token := strings.TrimPrefix(authHeader, "Bearer ")
	
	// Validate the token
	return token == a.expectedKey
}

// RequireAuth is a helper function that can be called at the start of protected endpoints
func (a *APIKeyAuth) RequireAuth(w http.ResponseWriter, r *http.Request) bool {
	if !a.ValidateBearerToken(r) {
		// Return standardized unauthorized response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{
			"success": false,
			"error": {
				"code": "UNAUTHORIZED",
				"message": "Valid Bearer token required",
				"details": "Authorization header must contain 'Bearer <token>'"
			},
			"server": "microcontroller-api",
			"timestamp": "` + getCurrentTimestamp() + `"
		}`))
		return false
	}
	return true
}

func getCurrentTimestamp() string {
	return time.Now().UTC().Format(time.RFC3339)
}