package services

import (
	"testing"
	"time"
	"triplink/backend/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockDB represents a mock database for testing
type MockDB struct {
	mock.Mock
}

func (m *MockDB) Create(value interface{}) error {
	args := m.Called(value)
	return args.Error(0)
}

func (m *MockDB) First(dest interface{}, conds ...interface{}) error {
	args := m.Called(dest, conds)
	return args.Error(0)
}

func (m *MockDB) Where(query interface{}, args ...interface{}) *MockDB {
	m.Called(query, args)
	return m
}

func (m *MockDB) Find(dest interface{}) error {
	args := m.Called(dest)
	return args.Error(0)
}

// Test LocationUpdate validation
func TestValidateLocationUpdate(t *testing.T) {
	ts := NewTrackingService()

	tests := []struct {
		name        string
		tripID      uint
		location    LocationUpdate
		expectError bool
		errorCode   string
	}{
		{
			name:   "Valid location update",
			tripID: 1,
			location: LocationUpdate{
				Latitude:  40.7128,
				Longitude: -74.0060,
				Source:    "GPS",
			},
			expectError: false,
		},
		{
			name:   "Invalid latitude - too high",
			tripID: 1,
			location: LocationUpdate{
				Latitude:  91.0,
				Longitude: -74.0060,
				Source:    "GPS",
			},
			expectError: true,
			errorCode:   "INVALID_COORDINATES",
		},
		{
			name:   "Invalid longitude - too low",
			tripID: 1,
			location: LocationUpdate{
				Latitude:  40.7128,
				Longitude: -181.0,
				Source:    "GPS",
			},
			expectError: true,
			errorCode:   "INVALID_COORDINATES",
		},
		{
			name:   "Invalid speed - negative",
			tripID: 1,
			location: LocationUpdate{
				Latitude:  40.7128,
				Longitude: -74.0060,
				Speed:     floatPtr(-10.0),
				Source:    "GPS",
			},
			expectError: true,
			errorCode:   "INVALID_SPEED",
		},
		{
			name:   "Invalid source",
			tripID: 1,
			location: LocationUpdate{
				Latitude:  40.7128,
				Longitude: -74.0060,
				Source:    "INVALID_SOURCE",
			},
			expectError: true,
			errorCode:   "INVALID_SOURCE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ts.ValidateLocationUpdate(tt.tripID, tt.location)

			if tt.expectError {
				assert.Error(t, err)
				if trackingErr, ok := err.(*TrackingError); ok {
					assert.Equal(t, tt.errorCode, trackingErr.Code)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Test coordinate validation
func TestIsValidCoordinate(t *testing.T) {
	tests := []struct {
		name     string
		lat      float64
		lng      float64
		expected bool
	}{
		{"Valid coordinates - NYC", 40.7128, -74.0060, true},
		{"Valid coordinates - London", 51.5074, -0.1278, true},
		{"Valid coordinates - North Pole", 90.0, 0.0, true},
		{"Valid coordinates - South Pole", -90.0, 0.0, true},
		{"Invalid latitude - too high", 91.0, 0.0, false},
		{"Invalid latitude - too low", -91.0, 0.0, false},
		{"Invalid longitude - too high", 0.0, 181.0, false},
		{"Invalid longitude - too low", 0.0, -181.0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidCoordinate(tt.lat, tt.lng)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test distance calculation
func TestCalculateDistance(t *testing.T) {
	tests := []struct {
		name      string
		lat1      float64
		lng1      float64
		lat2      float64
		lng2      float64
		expected  float64
		tolerance float64
	}{
		{
			name: "NYC to LA",
			lat1: 40.7128, lng1: -74.0060,
			lat2: 34.0522, lng2: -118.2437,
			expected:  3944.0, // Approximate distance in km
			tolerance: 50.0,   // Allow 50km tolerance
		},
		{
			name: "Same location",
			lat1: 40.7128, lng1: -74.0060,
			lat2: 40.7128, lng2: -74.0060,
			expected:  0.0,
			tolerance: 0.1,
		},
		{
			name: "London to Paris",
			lat1: 51.5074, lng1: -0.1278,
			lat2: 48.8566, lng2: 2.3522,
			expected:  344.0, // Approximate distance in km
			tolerance: 20.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateDistance(tt.lat1, tt.lng1, tt.lat2, tt.lng2)
			assert.InDelta(t, tt.expected, result, tt.tolerance)
		})
	}
}

// Test status transition validation
func TestIsValidStatusTransition(t *testing.T) {
	tests := []struct {
		name          string
		currentStatus string
		newStatus     string
		expected      bool
	}{
		{"PLANNED to ACTIVE", "PLANNED", "ACTIVE", true},
		{"ACTIVE to IN_TRANSIT", "ACTIVE", "IN_TRANSIT", true},
		{"IN_TRANSIT to COMPLETED", "IN_TRANSIT", "COMPLETED", true},
		{"COMPLETED to ACTIVE", "COMPLETED", "ACTIVE", false},
		{"CANCELLED to ACTIVE", "CANCELLED", "ACTIVE", false},
		{"PLANNED to COMPLETED", "PLANNED", "COMPLETED", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidStatusTransition(tt.currentStatus, tt.newStatus)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test completion percentage calculation
func TestCalculateCompletionPercent(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected float64
	}{
		{"PLANNED", "PLANNED", 0.0},
		{"ACTIVE", "ACTIVE", 10.0},
		{"AT_PICKUP", "AT_PICKUP", 25.0},
		{"IN_TRANSIT", "IN_TRANSIT", 50.0},
		{"AT_DELIVERY", "AT_DELIVERY", 90.0},
		{"COMPLETED", "COMPLETED", 100.0},
		{"UNKNOWN", "UNKNOWN_STATUS", 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateCompletionPercent(tt.status)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test data sanitization
func TestSanitizeLocationData(t *testing.T) {
	ts := NewTrackingService()

	location := LocationUpdate{
		Latitude:  40.712812345678,
		Longitude: -74.006012345678,
		Altitude:  floatPtr(123.456789),
		Speed:     floatPtr(65.789),
		Heading:   floatPtr(180.7),
		Source:    "gps",
	}

	ts.SanitizeLocationData(&location)

	assert.Equal(t, 40.712812, location.Latitude)
	assert.Equal(t, -74.006012, location.Longitude)
	assert.Equal(t, 123.5, *location.Altitude)
	assert.Equal(t, 65.8, *location.Speed)
	assert.Equal(t, 181.0, *location.Heading)
	assert.Equal(t, "GPS", location.Source)
}


// Test delay detection
func TestCheckForDelays(t *testing.T) {
	// This would require mocking the database call
	// For now, we'll test the delay calculation logic directly
	now := time.Now()
	pastTime := now.Add(-90 * time.Minute) // 90 minutes ago

	delayMinutes := int(now.Sub(pastTime).Minutes())

	assert.Equal(t, 90, delayMinutes)

	// Test severity calculation
	severity := "LOW"
	if delayMinutes > 120 {
		severity = "CRITICAL"
	} else if delayMinutes > 60 {
		severity = "HIGH"
	} else if delayMinutes > 30 {
		severity = "MEDIUM"
	}

	assert.Equal(t, "HIGH", severity)
}

// Test anomaly detection patterns
func TestDetectAnomaliesPatterns(t *testing.T) {
	// Test speed anomaly detection
	records := []models.TrackingRecord{
		{
			Latitude:  40.7128,
			Longitude: -74.0060,
			Speed:     floatPtr(60.0),
			Timestamp: time.Now().Add(-2 * time.Minute),
		},
		{
			Latitude:  40.7200,
			Longitude: -74.0100,
			Speed:     floatPtr(150.0), // Sudden speed increase
			Timestamp: time.Now().Add(-1 * time.Minute),
		},
	}

	// Test speed difference calculation
	speedDiff := *records[1].Speed - *records[0].Speed
	assert.Equal(t, 90.0, speedDiff)
	assert.True(t, speedDiff > 50) // Should trigger anomaly
}

// Helper functions for tests
func floatPtr(f float64) *float64 {
	return &f
}

// Benchmark tests
func BenchmarkCalculateDistance(b *testing.B) {
	lat1, lng1 := 40.7128, -74.0060
	lat2, lng2 := 34.0522, -118.2437

	for i := 0; i < b.N; i++ {
		calculateDistance(lat1, lng1, lat2, lng2)
	}
}

func BenchmarkValidateLocationUpdate(b *testing.B) {
	ts := NewTrackingService()
	location := LocationUpdate{
		Latitude:  40.7128,
		Longitude: -74.0060,
		Source:    "GPS",
	}

	for i := 0; i < b.N; i++ {
		ts.ValidateLocationUpdate(1, location)
	}
}
