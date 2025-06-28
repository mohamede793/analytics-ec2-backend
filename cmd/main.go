// cmd/main.go
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	
	"microcontroller-api/internal/auth"
	"microcontroller-api/internal/handlers"
	"microcontroller-api/internal/storage"
	"microcontroller-api/pkg/logger"
)

func main() {
	// Initialize logger
	logger := logger.New(getEnv("LOG_LEVEL", "info"))
	
	// Get API key from environment
	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		logger.Fatal("API_KEY environment variable is required")
	}
	
	// Initialize storage
	store := storage.NewMemoryStorage()
	
	// Initialize auth
	authService := auth.NewAPIKeyAuth(apiKey)
	
	// Initialize handlers
	sensorHandler := handlers.NewSensorHandler(store, logger, authService)
	healthHandler := handlers.NewHealthHandler(logger)
	
	// Setup router
	router := mux.NewRouter()
	
	// Public endpoints (no auth)
	router.HandleFunc("/health", healthHandler.Health).Methods("GET")
	
	// Protected API endpoints (bearer token required)
	api := router.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc("/sensor-data", sensorHandler.ReceiveSensorData).Methods("POST")
	api.HandleFunc("/sensor-data/batch", sensorHandler.ReceiveBatchData).Methods("POST")
	api.HandleFunc("/devices/{device_id}/latest", sensorHandler.GetLatestReading).Methods("GET")
	api.HandleFunc("/devices/{device_id}/status", sensorHandler.GetDeviceStatus).Methods("GET")
	
	// Server configuration
	port := getEnv("PORT", ":3000")
	srv := &http.Server{
		Addr:         port,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server
	go func() {
		logger.Info("🚀 Microcontroller API Server starting", "port", port)
		logger.Info("📡 Endpoints available:")
		logger.Info("  POST /api/v1/sensor-data - Receive sensor data")
		logger.Info("  POST /api/v1/sensor-data/batch - Receive batch data")
		logger.Info("  GET  /api/v1/devices/{id}/latest - Get latest reading")
		logger.Info("  GET  /api/v1/devices/{id}/status - Get device status")
		logger.Info("  GET  /health - Health check")
		logger.Info("🔐 Bearer token authentication enabled")
		
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server failed to start", "error", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("🛑 Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", "error", err)
	}
	
	logger.Info("✅ Server exited gracefully")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}