// internal/handlers/health.go
package handlers

import (
	"fmt"
	"net/http"
	"os"
	"runtime"
	"time"

	"microcontroller-api/pkg/logger"
	"microcontroller-api/pkg/response"
)

var startTime = time.Now()

type HealthHandler struct {
	logger logger.Logger
}

func NewHealthHandler(logger logger.Logger) *HealthHandler {
	return &HealthHandler{
		logger: logger,
	}
}

type HealthData struct {
	Status      string            `json:"status"`
	Version     string            `json:"version"`
	Uptime      string            `json:"uptime"`
	Environment string            `json:"environment"`
	System      map[string]string `json:"system"`
}

// Health handles GET /health - No authentication required
func (hh *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	healthData := &HealthData{
		Status:      "healthy",
		Version:     "1.0.0",
		Uptime:      time.Since(startTime).String(),
		Environment: getEnv("ENVIRONMENT", "development"),
		System: map[string]string{
			"go_version": runtime.Version(),
			"arch":       runtime.GOARCH,
			"os":         runtime.GOOS,
			"goroutines": fmt.Sprintf("%d", runtime.NumGoroutine()),
		},
	}

	hh.logger.Info("Health check requested", "remote_addr", r.RemoteAddr)
	response.Success(w, http.StatusOK, "Microcontroller API is healthy", healthData)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}