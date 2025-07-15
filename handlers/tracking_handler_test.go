package handlers

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"triplink/backend/services"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// Test data setup
func setupTestApp() *fiber.App {
	app := fiber.New()

	// Add tracking routes
	app.Post("/trips/:trip_id/tracking/location", UpdateTripLocation)
	app.Get("/trips/:trip_id/tracking/current", GetCurrentTripLocation)
	app.Get("/trips/:trip_id/tracking/history", GetTripTrackingHistory)
	app.Put("/trips/:trip_id/tracking/status", UpdateTripStatus)
	app.Get("/trips/:trip_id/tracking/eta", GetTripETA)
	app.Get("/loads/:load_id/tracking", GetLoadTracking)
	app.Get("/users/:user_id/tracking/active", GetUserActiveTrackings)
	app.Get("/mobile/trips/:trip_id/tracking", GetLightweightTracking)

	return app
}

// Test UpdateTripLocation endpoint
func TestUpdateTripLocation(t *testing.T) {
	app := setupTestApp()

	tests := []struct {
		name           string
		tripID         string
		locationData   services.LocationUpdate
		expectedStatus int
	}{
		{
			name:   "Valid location update",
			tripID: "1",
			locationData: services.LocationUpdate{
				Latitude:  40.7128,
				Longitude: -74.0060,
				Source:    "GPS",
			},
			expectedStatus: 200,
		},
		{
			name:   "Invalid coordinates",
			tripID: "1",
			locationData: services.LocationUpdate{
				Latitude:  91.0,
				Longitude: -74.0060,
				Source:    "GPS",
			},
			expectedStatus: 500, // Would fail in service validation
		},
		{
			name:           "Invalid trip ID",
			tripID:         "invalid",
			locationData:   services.LocationUpdate{},
			expectedStatus: 400,
		},
		{
			name:   "Missing coordinates",
			tripID: "1",
			locationData: services.LocationUpdate{
				Source: "GPS",
			},
			expectedStatus: 400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, _ := json.Marshal(tt.locationData)
			req := httptest.NewRequest("POST", "/trips/"+tt.tripID+"/tracking/location", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}

// Test GetCurrentTripLocation endpoint
func TestGetCurrentTripLocation(t *testing.T) {
	app := setupTestApp()

	tests := []struct {
		name           string
		tripID         string
		expectedStatus int
	}{
		{
			name:           "Valid trip ID",
			tripID:         "1",
			expectedStatus: 404, // No data in test environment
		},
		{
			name:           "Invalid trip ID",
			tripID:         "invalid",
			expectedStatus: 400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/trips/"+tt.tripID+"/tracking/current", nil)
			resp, err := app.Test(req)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}

// Test UpdateTripStatus endpoint
func TestUpdateTripStatus(t *testing.T) {
	app := setupTestApp()

	tests := []struct {
		name           string
		tripID         string
		statusData     map[string]string
		expectedStatus int
	}{
		{
			name:   "Valid status update",
			tripID: "1",
			statusData: map[string]string{
				"status": "ACTIVE",
			},
			expectedStatus: 400, // Would fail due to missing trip in test DB
		},
		{
			name:   "Invalid status",
			tripID: "1",
			statusData: map[string]string{
				"status": "INVALID_STATUS",
			},
			expectedStatus: 400,
		},
		{
			name:           "Missing status",
			tripID:         "1",
			statusData:     map[string]string{},
			expectedStatus: 400,
		},
		{
			name:           "Invalid trip ID",
			tripID:         "invalid",
			statusData:     map[string]string{"status": "ACTIVE"},
			expectedStatus: 400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, _ := json.Marshal(tt.statusData)
			req := httptest.NewRequest("PUT", "/trips/"+tt.tripID+"/tracking/status", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}

// Test GetTripTrackingHistory endpoint
func TestGetTripTrackingHistory(t *testing.T) {
	app := setupTestApp()

	tests := []struct {
		name           string
		tripID         string
		queryParams    string
		expectedStatus int
	}{
		{
			name:           "Valid request with default params",
			tripID:         "1",
			queryParams:    "",
			expectedStatus: 404, // No trip in test DB
		},
		{
			name:           "Valid request with limit",
			tripID:         "1",
			queryParams:    "?limit=10",
			expectedStatus: 404,
		},
		{
			name:           "Valid request with offset",
			tripID:         "1",
			queryParams:    "?offset=5",
			expectedStatus: 404,
		},
		{
			name:           "Invalid trip ID",
			tripID:         "invalid",
			queryParams:    "",
			expectedStatus: 400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/trips/"+tt.tripID+"/tracking/history"+tt.queryParams, nil)
			resp, err := app.Test(req)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}

// Test GetLoadTracking endpoint
func TestGetLoadTracking(t *testing.T) {
	app := setupTestApp()

	tests := []struct {
		name           string
		loadID         string
		expectedStatus int
	}{
		{
			name:           "Valid load ID",
			loadID:         "1",
			expectedStatus: 404, // No load in test DB
		},
		{
			name:           "Invalid load ID",
			loadID:         "invalid",
			expectedStatus: 400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/loads/"+tt.loadID+"/tracking", nil)
			resp, err := app.Test(req)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}

// Test GetUserActiveTrackings endpoint
func TestGetUserActiveTrackings(t *testing.T) {
	app := setupTestApp()

	tests := []struct {
		name           string
		userID         string
		role           string
		expectedStatus int
	}{
		{
			name:           "Valid user ID",
			userID:         "1",
			role:           "",
			expectedStatus: 404, // No user in test DB
		},
		{
			name:           "Valid user ID with role",
			userID:         "1",
			role:           "CARRIER",
			expectedStatus: 404,
		},
		{
			name:           "Invalid user ID",
			userID:         "invalid",
			role:           "",
			expectedStatus: 400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/users/" + tt.userID + "/tracking/active"
			if tt.role != "" {
				url += "?role=" + tt.role
			}

			req := httptest.NewRequest("GET", url, nil)
			resp, err := app.Test(req)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}

// Test GetLightweightTracking endpoint (mobile)
func TestGetLightweightTracking(t *testing.T) {
	app := setupTestApp()

	tests := []struct {
		name           string
		tripID         string
		expectedStatus int
	}{
		{
			name:           "Valid trip ID",
			tripID:         "1",
			expectedStatus: 404, // No trip in test DB
		},
		{
			name:           "Invalid trip ID",
			tripID:         "invalid",
			expectedStatus: 400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/mobile/trips/"+tt.tripID+"/tracking", nil)
			resp, err := app.Test(req)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}

// Test JSON response parsing
func TestJSONResponseParsing(t *testing.T) {
	app := setupTestApp()

	// Test error response format
	req := httptest.NewRequest("GET", "/trips/invalid/tracking/current", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var errorResponse map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&errorResponse)
	assert.NoError(t, err)
	assert.Contains(t, errorResponse, "error")
	assert.Equal(t, "Invalid trip ID", errorResponse["error"])
}

// Test request validation
func TestRequestValidation(t *testing.T) {
	app := setupTestApp()

	// Test malformed JSON
	req := httptest.NewRequest("POST", "/trips/1/tracking/location", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	// Test empty request body
	req = httptest.NewRequest("POST", "/trips/1/tracking/location", bytes.NewBuffer([]byte{}))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
}

// Performance test for tracking endpoints
func BenchmarkUpdateTripLocation(b *testing.B) {
	app := setupTestApp()
	locationData := services.LocationUpdate{
		Latitude:  40.7128,
		Longitude: -74.0060,
		Source:    "GPS",
	}
	jsonData, _ := json.Marshal(locationData)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/trips/1/tracking/location", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		app.Test(req)
	}
}

// Test concurrent requests
func TestConcurrentRequests(t *testing.T) {
	app := setupTestApp()

	// Test concurrent location updates
	concurrency := 10
	done := make(chan bool, concurrency)

	for i := 0; i < concurrency; i++ {
		go func() {
			locationData := services.LocationUpdate{
				Latitude:  40.7128,
				Longitude: -74.0060,
				Source:    "GPS",
			}
			jsonData, _ := json.Marshal(locationData)
			req := httptest.NewRequest("POST", "/trips/1/tracking/location", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			assert.NoError(t, err)
			assert.NotNil(t, resp)

			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < concurrency; i++ {
		<-done
	}
}
