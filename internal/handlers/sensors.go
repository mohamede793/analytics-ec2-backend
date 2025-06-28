// internal/handlers/sensors.go
package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	
	"microcontroller-api/internal/auth"
	"microcontroller-api/internal/models"
	"microcontroller-api/internal/storage"
	"microcontroller-api/pkg/logger"
	"microcontroller-api/pkg/response"
)

type SensorHandler struct {
	storage *storage.MemoryStorage
	logger  logger.Logger
	auth    *auth.APIKeyAuth
}

func NewSensorHandler(storage *storage.MemoryStorage, logger logger.Logger, auth *auth.APIKeyAuth) *SensorHandler {
	return &SensorHandler{
		storage: storage,
		logger:  logger,
		auth:    auth,
	}
}

// ReceiveSensorData handles POST /api/v1/sensor-data
func (sh *SensorHandler) ReceiveSensorData(w http.ResponseWriter, r *http.Request) {
	// Auth check first - stops here if unauthorized
	if !sh.auth.RequireAuth(w, r) {
		return
	}

	var reading models.SensorReading
	
	// Parse JSON
	if err := json.NewDecoder(r.Body).Decode(&reading); err != nil {
		sh.logger.Error("Failed to parse sensor data JSON", "error", err)
		response.BadRequest(w, "Invalid JSON format")
		return
	}

	// Validate data
	if errors := reading.Validate(); len(errors) > 0 {
		sh.logger.Error("Sensor data validation failed", "errors", errors)
		response.ValidationError(w, strings.Join(errors, "; "))
		return
	}

	// Store data
	if err := sh.storage.StoreSensorReading(&reading); err != nil {
		sh.logger.Error("Failed to store sensor reading", "error", err, "device_id", reading.DeviceID)
		response.InternalError(w, "Failed to store sensor data")
		return
	}

	sh.logger.Info("Sensor data received", 
		"device_id", reading.DeviceID,
		"temperature", reading.Temperature,
		"humidity", reading.Humidity,
		"timestamp", reading.Timestamp)

	response.Success(w, http.StatusCreated, "Sensor data received successfully", reading)
}

// ReceiveBatchData handles POST /api/v1/sensor-data/batch
func (sh *SensorHandler) ReceiveBatchData(w http.ResponseWriter, r *http.Request) {
	// Auth check first
	if !sh.auth.RequireAuth(w, r) {
		return
	}

	var batchData models.BatchSensorData
	
	// Parse JSON
	if err := json.NewDecoder(r.Body).Decode(&batchData); err != nil {
		sh.logger.Error("Failed to parse batch sensor data JSON", "error", err)
		response.BadRequest(w, "Invalid JSON format")
		return
	}

	// Validate data
	if errors := batchData.Validate(); len(errors) > 0 {
		sh.logger.Error("Batch sensor data validation failed", "errors", errors)
		response.ValidationError(w, strings.Join(errors, "; "))
		return
	}

	// Store batch data
	if err := sh.storage.StoreBatchReadings(batchData.DeviceID, batchData.Readings); err != nil {
		sh.logger.Error("Failed to store batch readings", "error", err, "device_id", batchData.DeviceID)
		response.InternalError(w, "Failed to store batch data")
		return
	}

	sh.logger.Info("Batch sensor data received", 
		"device_id", batchData.DeviceID,
		"readings_count", len(batchData.Readings))

	// Return summary
	meta := &response.Meta{
		Count:       len(batchData.Readings),
		ProcessedAt: time.Now().UTC().Format(time.RFC3339),
	}

	response.SuccessWithMeta(w, http.StatusCreated, "Batch data processed successfully", 
		map[string]interface{}{
			"device_id":      batchData.DeviceID,
			"readings_saved": len(batchData.Readings),
		}, meta)
}

// GetLatestReading handles GET /api/v1/devices/{device_id}/latest
func (sh *SensorHandler) GetLatestReading(w http.ResponseWriter, r *http.Request) {
	// Auth check first
	if !sh.auth.RequireAuth(w, r) {
		return
	}

	vars := mux.Vars(r)
	deviceID := vars["device_id"]

	if deviceID == "" {
		response.BadRequest(w, "device_id parameter is required")
		return
	}

	// Get latest reading
	reading, err := sh.storage.GetLatestReading(deviceID)
	if err != nil {
		if storageErr, ok := err.(*storage.StorageError); ok && storageErr.Code == "DEVICE_NOT_FOUND" {
			response.NotFound(w, "Device: "+deviceID)
			return
		}
		sh.logger.Error("Failed to get latest reading", "error", err, "device_id", deviceID)
		response.InternalError(w, "Failed to retrieve latest reading")
		return
	}

	sh.logger.Info("Latest reading retrieved", "device_id", deviceID)
	response.Success(w, http.StatusOK, "Latest reading retrieved successfully", reading)
}

// GetDeviceStatus handles GET /api/v1/devices/{device_id}/status
func (sh *SensorHandler) GetDeviceStatus(w http.ResponseWriter, r *http.Request) {
	// Auth check first
	if !sh.auth.RequireAuth(w, r) {
		return
	}

	vars := mux.Vars(r)
	deviceID := vars["device_id"]

	if deviceID == "" {
		response.BadRequest(w, "device_id parameter is required")
		return
	}

	// Get device status
	status, err := sh.storage.GetDeviceStatus(deviceID)
	if err != nil {
		if storageErr, ok := err.(*storage.StorageError); ok && storageErr.Code == "DEVICE_NOT_FOUND" {
			response.NotFound(w, "Device: "+deviceID)
			return
		}
		sh.logger.Error("Failed to get device status", "error", err, "device_id", deviceID)
		response.InternalError(w, "Failed to retrieve device status")
		return
	}

	sh.logger.Info("Device status retrieved", "device_id", deviceID, "status", status.Status)
	response.Success(w, http.StatusOK, "Device status retrieved successfully", status)
}