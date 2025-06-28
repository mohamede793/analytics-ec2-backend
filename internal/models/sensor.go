// internal/models/sensor.go
package models

import (
	"time"
	"fmt"
)

// SensorReading represents a single sensor data point
type SensorReading struct {
	DeviceID    string    `json:"device_id"`
	Temperature float64   `json:"temperature"`
	Humidity    float64   `json:"humidity"`
	Pressure    float64   `json:"pressure,omitempty"`
	Light       float64   `json:"light,omitempty"`
	Motion      bool      `json:"motion,omitempty"`
	Battery     float64   `json:"battery,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
}

// BatchSensorData represents multiple sensor readings from one device
type BatchSensorData struct {
	DeviceID string          `json:"device_id"`
	Readings []SingleReading `json:"readings"`
}

// SingleReading represents one reading in a batch (no device_id needed)
type SingleReading struct {
	Temperature float64   `json:"temperature"`
	Humidity    float64   `json:"humidity"`
	Pressure    float64   `json:"pressure,omitempty"`
	Light       float64   `json:"light,omitempty"`
	Motion      bool      `json:"motion,omitempty"`
	Battery     float64   `json:"battery,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
}

// DeviceStatus represents the current status of a device
type DeviceStatus struct {
	DeviceID      string    `json:"device_id"`
	LastSeen      time.Time `json:"last_seen"`
	TotalReadings int       `json:"total_readings"`
	Status        string    `json:"status"` // "online", "offline", "unknown"
	BatteryLevel  float64   `json:"battery_level,omitempty"`
}

// Validation methods
func (sr *SensorReading) Validate() []string {
	var errors []string
	
	if sr.DeviceID == "" {
		errors = append(errors, "device_id is required")
	}
	
	if sr.Temperature < -100 || sr.Temperature > 100 {
		errors = append(errors, "temperature must be between -100 and 100 celsius")
	}
	
	if sr.Humidity < 0 || sr.Humidity > 100 {
		errors = append(errors, "humidity must be between 0 and 100 percent")
	}
	
	if sr.Battery < 0 || sr.Battery > 100 {
		errors = append(errors, "battery must be between 0 and 100 percent")
	}
	
	if sr.Timestamp.IsZero() {
		sr.Timestamp = time.Now()
	}
	
	return errors
}

func (bsd *BatchSensorData) Validate() []string {
	var errors []string
	
	if bsd.DeviceID == "" {
		errors = append(errors, "device_id is required")
	}
	
	if len(bsd.Readings) == 0 {
		errors = append(errors, "at least one reading is required")
	}
	
	if len(bsd.Readings) > 100 {
		errors = append(errors, "maximum 100 readings per batch")
	}
	
	for i, reading := range bsd.Readings {
		if reading.Temperature < -100 || reading.Temperature > 100 {
			errors = append(errors, fmt.Sprintf("reading %d: temperature out of range", i))
		}
		if reading.Humidity < 0 || reading.Humidity > 100 {
			errors = append(errors, fmt.Sprintf("reading %d: humidity out of range", i))
		}
		if reading.Timestamp.IsZero() {
			bsd.Readings[i].Timestamp = time.Now()
		}
	}
	
	return errors
}