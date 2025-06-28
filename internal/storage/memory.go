// internal/storage/memory.go
package storage

import (
	"sync"
	"time"

	"github.com/mohamede793/analytics-ec2-backend/internal/models"
)

// MemoryStorage provides in-memory storage for sensor data
type MemoryStorage struct {
	readings map[string][]*models.SensorReading // deviceID -> readings
	devices  map[string]*models.DeviceStatus    // deviceID -> status
	mu       sync.RWMutex
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		readings: make(map[string][]*models.SensorReading),
		devices:  make(map[string]*models.DeviceStatus),
	}
}

// StoreSensorReading stores a single sensor reading
func (ms *MemoryStorage) StoreSensorReading(reading *models.SensorReading) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	// Store the reading
	ms.readings[reading.DeviceID] = append(ms.readings[reading.DeviceID], reading)

	// Update device status
	ms.updateDeviceStatus(reading.DeviceID, reading.Timestamp, reading.Battery)

	return nil
}

// StoreBatchReadings stores multiple sensor readings
func (ms *MemoryStorage) StoreBatchReadings(deviceID string, readings []models.SingleReading) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	var lastTimestamp time.Time
	var lastBattery float64

	// Convert and store each reading
	for _, reading := range readings {
		sensorReading := &models.SensorReading{
			DeviceID:    deviceID,
			Temperature: reading.Temperature,
			Humidity:    reading.Humidity,
			Pressure:    reading.Pressure,
			Light:       reading.Light,
			Motion:      reading.Motion,
			Battery:     reading.Battery,
			Timestamp:   reading.Timestamp,
		}

		ms.readings[deviceID] = append(ms.readings[deviceID], sensorReading)

		// Track latest timestamp and battery for device status
		if reading.Timestamp.After(lastTimestamp) {
			lastTimestamp = reading.Timestamp
			lastBattery = reading.Battery
		}
	}

	// Update device status with latest info
	ms.updateDeviceStatus(deviceID, lastTimestamp, lastBattery)

	return nil
}

// GetLatestReading gets the most recent reading for a device
func (ms *MemoryStorage) GetLatestReading(deviceID string) (*models.SensorReading, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	readings, exists := ms.readings[deviceID]
	if !exists || len(readings) == 0 {
		return nil, ErrDeviceNotFound
	}

	// Return the latest reading (last in slice)
	return readings[len(readings)-1], nil
}

// GetDeviceStatus gets the current status of a device
func (ms *MemoryStorage) GetDeviceStatus(deviceID string) (*models.DeviceStatus, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	status, exists := ms.devices[deviceID]
	if !exists {
		return nil, ErrDeviceNotFound
	}

	// Update online/offline status based on last seen time
	status.Status = ms.calculateDeviceStatus(status.LastSeen)

	return status, nil
}

// GetAllDevices returns status for all known devices
func (ms *MemoryStorage) GetAllDevices() ([]*models.DeviceStatus, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	var devices []*models.DeviceStatus
	for _, status := range ms.devices {
		// Update online/offline status
		status.Status = ms.calculateDeviceStatus(status.LastSeen)
		devices = append(devices, status)
	}

	return devices, nil
}

// Private helper methods
func (ms *MemoryStorage) updateDeviceStatus(deviceID string, timestamp time.Time, battery float64) {
	status, exists := ms.devices[deviceID]
	if !exists {
		status = &models.DeviceStatus{
			DeviceID:      deviceID,
			TotalReadings: 0,
		}
		ms.devices[deviceID] = status
	}

	status.LastSeen = timestamp
	status.TotalReadings++
	if battery > 0 {
		status.BatteryLevel = battery
	}
	status.Status = ms.calculateDeviceStatus(timestamp)
}

func (ms *MemoryStorage) calculateDeviceStatus(lastSeen time.Time) string {
	timeSince := time.Since(lastSeen)
	
	if timeSince < 5*time.Minute {
		return "online"
	} else if timeSince < 30*time.Minute {
		return "idle"
	} else {
		return "offline"
	}
}

// Storage errors
var (
	ErrDeviceNotFound = NewStorageError("DEVICE_NOT_FOUND", "Device not found")
)

// StorageError represents storage-related errors
type StorageError struct {
	Code    string
	Message string
}

func (e *StorageError) Error() string {
	return e.Message
}

func NewStorageError(code, message string) *StorageError {
	return &StorageError{
		Code:    code,
		Message: message,
	}
}