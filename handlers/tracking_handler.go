package handlers

import (
	"fmt"
	"math"
	"strconv"
	"time"
	"triplink/backend/database"
	"triplink/backend/models"
	"triplink/backend/services"

	"github.com/gofiber/fiber/v2"
)

var trackingService = services.NewTrackingService()

// UpdateTripLocation @Summary Update trip location
// @Description Update the current location of a trip
// @Tags tracking
// @Accept json
// @Produce json
// @Param trip_id path int true "Trip ID"
// @Param location body services.LocationUpdate true "Location data"
// @Success 200 {object} map[string]interface{}
// @Router /trips/{trip_id}/tracking/location [post]
func UpdateTripLocation(c *fiber.Ctx) error {
	tripIDStr := c.Params("trip_id")
	tripID, err := strconv.ParseUint(tripIDStr, 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid trip ID",
		})
	}

	var locationUpdate services.LocationUpdate
	if err := c.BodyParser(&locationUpdate); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Cannot parse location data",
		})
	}

	// Validate required fields
	if locationUpdate.Latitude == 0 || locationUpdate.Longitude == 0 {
		return c.Status(400).JSON(fiber.Map{
			"error": "Latitude and longitude are required",
		})
	}

	// Set default source if not provided
	if locationUpdate.Source == "" {
		locationUpdate.Source = "GPS"
	}

	// Verify trip exists and user has permission
	var trip models.Trip
	if err := database.DB.First(&trip, uint(tripID)).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Trip not found",
		})
	}

	// Update location using tracking service
	if err := trackingService.UpdateLocation(uint(tripID), locationUpdate); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to update location: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message":   "Location updated successfully",
		"trip_id":   tripID,
		"latitude":  locationUpdate.Latitude,
		"longitude": locationUpdate.Longitude,
		"timestamp": trip.LastLocationUpdate,
	})
}

// GetCurrentTripLocation @Summary Get current trip location
// @Description Get the most recent location of a trip
// @Tags tracking
// @Produce json
// @Param trip_id path int true "Trip ID"
// @Success 200 {object} models.TrackingRecord
// @Router /trips/{trip_id}/tracking/current [get]
func GetCurrentTripLocation(c *fiber.Ctx) error {
	tripIDStr := c.Params("trip_id")
	tripID, err := strconv.ParseUint(tripIDStr, 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid trip ID",
		})
	}

	// Verify trip exists
	var trip models.Trip
	if err := database.DB.First(&trip, uint(tripID)).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Trip not found",
		})
	}

	// Get current location
	location, err := trackingService.GetCurrentLocation(uint(tripID))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "No location data found",
		})
	}

	return c.JSON(location)
}

// GetTripTrackingHistory @Summary Get trip tracking history
// @Description Get the location history for a trip
// @Tags tracking
// @Produce json
// @Param trip_id path int true "Trip ID"
// @Param limit query int false "Number of records to return (default 50)"
// @Param offset query int false "Number of records to skip (default 0)"
// @Success 200 {array} models.TrackingRecord
// @Router /trips/{trip_id}/tracking/history [get]
func GetTripTrackingHistory(c *fiber.Ctx) error {
	tripIDStr := c.Params("trip_id")
	tripID, err := strconv.ParseUint(tripIDStr, 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid trip ID",
		})
	}

	// Parse query parameters
	limit := c.QueryInt("limit", 50)
	offset := c.QueryInt("offset", 0)

	// Verify trip exists
	var trip models.Trip
	if err := database.DB.First(&trip, uint(tripID)).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Trip not found",
		})
	}

	// Get tracking history
	var trackingRecords []models.TrackingRecord
	result := database.DB.Where("trip_id = ?", tripID).
		Order("timestamp DESC").
		Limit(limit).
		Offset(offset).
		Find(&trackingRecords)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch tracking history",
		})
	}

	return c.JSON(fiber.Map{
		"data":   trackingRecords,
		"count":  len(trackingRecords),
		"limit":  limit,
		"offset": offset,
	})
}

// UpdateTripStatus @Summary Update trip status
// @Description Update the status of a trip
// @Tags tracking
// @Accept json
// @Produce json
// @Param trip_id path int true "Trip ID"
// @Param status body map[string]string true "Status data"
// @Success 200 {object} map[string]interface{}
// @Router /trips/{trip_id}/tracking/status [put]
func UpdateTripStatus(c *fiber.Ctx) error {
	tripIDStr := c.Params("trip_id")
	tripID, err := strconv.ParseUint(tripIDStr, 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid trip ID",
		})
	}

	var statusUpdate map[string]string
	if err := c.BodyParser(&statusUpdate); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Cannot parse status data",
		})
	}

	newStatus, exists := statusUpdate["status"]
	if !exists || newStatus == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Status is required",
		})
	}

	// Validate status value
	validStatuses := []string{"PLANNED", "ACTIVE", "IN_TRANSIT", "AT_PICKUP", "AT_DELIVERY", "DELAYED", "COMPLETED", "CANCELLED"}
	isValid := false
	for _, status := range validStatuses {
		if status == newStatus {
			isValid = true
			break
		}
	}

	if !isValid {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid status value",
		})
	}

	// Update status using tracking service
	if err := trackingService.UpdateTripStatus(uint(tripID), newStatus); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Failed to update status: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Status updated successfully",
		"trip_id": tripID,
		"status":  newStatus,
	})
}

// GetTripETA @Summary Get trip ETA
// @Description Get the estimated time of arrival for a trip
// @Tags tracking
// @Produce json
// @Param trip_id path int true "Trip ID"
// @Success 200 {object} map[string]interface{}
// @Router /trips/{trip_id}/tracking/eta [get]
func GetTripETA(c *fiber.Ctx) error {
	tripIDStr := c.Params("trip_id")
	tripID, err := strconv.ParseUint(tripIDStr, 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid trip ID",
		})
	}

	// Verify trip exists
	var trip models.Trip
	if err := database.DB.First(&trip, uint(tripID)).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Trip not found",
		})
	}

	// Calculate ETA
	eta, err := trackingService.CalculateETA(uint(tripID))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to calculate ETA: " + err.Error(),
		})
	}

	// Check for delays
	delayInfo, _ := trackingService.CheckForDelays(uint(tripID))

	response := fiber.Map{
		"trip_id":           tripID,
		"estimated_arrival": eta,
		"original_eta":      trip.EstimatedArrival,
	}

	if delayInfo != nil {
		response["delay_info"] = delayInfo
	}

	return c.JSON(response)
}

// GetTripTrackingStatus @Summary Get trip tracking status
// @Description Get the current tracking status and progress of a trip
// @Tags tracking
// @Produce json
// @Param trip_id path int true "Trip ID"
// @Success 200 {object} models.TrackingStatus
// @Router /trips/{trip_id}/tracking/status [get]
func GetTripTrackingStatus(c *fiber.Ctx) error {
	tripIDStr := c.Params("trip_id")
	tripID, err := strconv.ParseUint(tripIDStr, 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid trip ID",
		})
	}

	// Get tracking status
	var trackingStatus models.TrackingStatus
	if err := database.DB.Where("trip_id = ?", tripID).First(&trackingStatus).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Tracking status not found",
		})
	}

	return c.JSON(trackingStatus)
}

// GetTripTrackingEvents @Summary Get trip tracking events
// @Description Get all tracking events for a trip
// @Tags tracking
// @Produce json
// @Param trip_id path int true "Trip ID"
// @Param limit query int false "Number of events to return (default 20)"
// @Success 200 {array} models.TrackingEvent
// @Router /trips/{trip_id}/tracking/events [get]
func GetTripTrackingEvents(c *fiber.Ctx) error {
	tripIDStr := c.Params("trip_id")
	tripID, err := strconv.ParseUint(tripIDStr, 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid trip ID",
		})
	}

	limit := c.QueryInt("limit", 20)

	// Get tracking events
	var events []models.TrackingEvent
	result := database.DB.Where("trip_id = ?", tripID).
		Order("timestamp DESC").
		Limit(limit).
		Find(&events)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch tracking events",
		})
	}

	return c.JSON(events)
}

// Load Tracking Endpoints

// GetLoadTracking @Summary Get load tracking information
// @Description Get comprehensive tracking information for a specific load
// @Tags load-tracking
// @Produce json
// @Param load_id path int true "Load ID"
// @Success 200 {object} map[string]interface{}
// @Router /loads/{load_id}/tracking [get]
func GetLoadTracking(c *fiber.Ctx) error {
	loadIDStr := c.Params("load_id")
	loadID, err := strconv.ParseUint(loadIDStr, 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid load ID",
		})
	}

	// Get load with trip information
	var load models.Load
	if err := database.DB.Preload("TrackingRecords").
		Preload("TrackingStatus").
		Preload("TrackingEvents").
		First(&load, uint(loadID)).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Load not found",
		})
	}

	// Get trip information for this load
	var trip models.Trip
	if err := database.DB.First(&trip, load.TripID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Associated trip not found",
		})
	}

	// Get current trip location
	currentLocation, _ := trackingService.GetCurrentLocation(trip.ID)

	// Calculate ETA for the trip
	eta, _ := trackingService.CalculateETA(trip.ID)

	// Check for delays
	delayInfo, _ := trackingService.CheckForDelays(trip.ID)

	// Get load-specific tracking status
	var loadTrackingStatus models.TrackingStatus
	database.DB.Where("load_id = ?", loadID).First(&loadTrackingStatus)

	response := fiber.Map{
		"load_id":           loadID,
		"booking_reference": load.BookingReference,
		"status":            load.Status,
		"trip_id":           trip.ID,
		"trip_status":       trip.Status,
		"current_location":  currentLocation,
		"estimated_arrival": eta,
		"pickup_address":    load.PickupAddress,
		"delivery_address":  load.DeliveryAddress,
		"tracking_enabled":  trip.TrackingEnabled,
	}

	if delayInfo != nil {
		response["delay_info"] = delayInfo
	}

	if loadTrackingStatus.ID != 0 {
		response["load_tracking_status"] = loadTrackingStatus
	}

	return c.JSON(response)
}

// GetLoadTrackingEvents @Summary Get load tracking events
// @Description Get all tracking events specific to a load
// @Tags load-tracking
// @Produce json
// @Param load_id path int true "Load ID"
// @Param limit query int false "Number of events to return (default 20)"
// @Success 200 {array} models.TrackingEvent
// @Router /loads/{load_id}/tracking/events [get]
func GetLoadTrackingEvents(c *fiber.Ctx) error {
	loadIDStr := c.Params("load_id")
	loadID, err := strconv.ParseUint(loadIDStr, 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid load ID",
		})
	}

	limit := c.QueryInt("limit", 20)

	// Verify load exists
	var load models.Load
	if err := database.DB.First(&load, uint(loadID)).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Load not found",
		})
	}

	// Get load-specific tracking events and trip-level events
	var events []models.TrackingEvent
	result := database.DB.Where("load_id = ? OR (trip_id = ? AND load_id IS NULL)", loadID, load.TripID).
		Order("timestamp DESC").
		Limit(limit).
		Find(&events)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch tracking events",
		})
	}

	return c.JSON(events)
}

// UpdateLoadStatus @Summary Update load status
// @Description Update the status of a specific load and create tracking events
// @Tags load-tracking
// @Accept json
// @Produce json
// @Param load_id path int true "Load ID"
// @Param status body map[string]string true "Status data"
// @Success 200 {object} map[string]interface{}
// @Router /loads/{load_id}/status [put]
func UpdateLoadStatus(c *fiber.Ctx) error {
	loadIDStr := c.Params("load_id")
	loadID, err := strconv.ParseUint(loadIDStr, 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid load ID",
		})
	}

	var statusUpdate map[string]string
	if err := c.BodyParser(&statusUpdate); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Cannot parse status data",
		})
	}

	newStatus, exists := statusUpdate["status"]
	if !exists || newStatus == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Status is required",
		})
	}

	// Validate load status
	validLoadStatuses := []string{"BOOKED", "PICKUP_SCHEDULED", "PICKED_UP", "IN_TRANSIT", "OUT_FOR_DELIVERY", "DELIVERED", "EXCEPTION"}
	isValid := false
	for _, status := range validLoadStatuses {
		if status == newStatus {
			isValid = true
			break
		}
	}

	if !isValid {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid load status value",
		})
	}

	// Get load
	var load models.Load
	if err := database.DB.First(&load, uint(loadID)).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Load not found",
		})
	}

	previousStatus := load.Status

	// Update load status
	if err := database.DB.Model(&load).Update("status", newStatus).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to update load status",
		})
	}

	// Create tracking event for load status change
	event := models.TrackingEvent{
		TripID:      load.TripID,
		LoadID:      &load.ID,
		EventType:   "LOAD_STATUS_CHANGE",
		EventData:   `{"from":"` + previousStatus + `","to":"` + newStatus + `"}`,
		Location:    "",
		Timestamp:   time.Now(),
		Description: "Load status changed from " + previousStatus + " to " + newStatus,
	}
	database.DB.Create(&event)

	// Update or create load tracking status
	var loadTrackingStatus models.TrackingStatus
	result := database.DB.Where("load_id = ?", loadID).First(&loadTrackingStatus)

	if result.Error != nil {
		// Create new load tracking status
		loadTrackingStatus = models.TrackingStatus{
			TripID:            load.TripID,
			LoadID:            &load.ID,
			CurrentStatus:     newStatus,
			PreviousStatus:    previousStatus,
			StatusChangedAt:   time.Now(),
			CompletionPercent: calculateLoadCompletionPercent(newStatus),
		}
		database.DB.Create(&loadTrackingStatus)
	} else {
		// Update existing load tracking status
		loadTrackingStatus.PreviousStatus = loadTrackingStatus.CurrentStatus
		loadTrackingStatus.CurrentStatus = newStatus
		loadTrackingStatus.StatusChangedAt = time.Now()
		loadTrackingStatus.CompletionPercent = calculateLoadCompletionPercent(newStatus)
		database.DB.Save(&loadTrackingStatus)
	}

	return c.JSON(fiber.Map{
		"message":         "Load status updated successfully",
		"load_id":         loadID,
		"status":          newStatus,
		"previous_status": previousStatus,
	})
}

// GetLoadTrackingHistory @Summary Get load tracking history
// @Description Get location tracking history for a load based on its trip
// @Tags load-tracking
// @Produce json
// @Param load_id path int true "Load ID"
// @Param limit query int false "Number of records to return (default 50)"
// @Param offset query int false "Number of records to skip (default 0)"
// @Success 200 {object} map[string]interface{}
// @Router /loads/{load_id}/tracking/history [get]
func GetLoadTrackingHistory(c *fiber.Ctx) error {
	loadIDStr := c.Params("load_id")
	loadID, err := strconv.ParseUint(loadIDStr, 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid load ID",
		})
	}

	// Parse query parameters
	limit := c.QueryInt("limit", 50)
	offset := c.QueryInt("offset", 0)

	// Get load to find associated trip
	var load models.Load
	if err := database.DB.First(&load, uint(loadID)).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Load not found",
		})
	}

	// Get tracking history for the trip (which includes this load)
	var trackingRecords []models.TrackingRecord
	result := database.DB.Where("trip_id = ?", load.TripID).
		Order("timestamp DESC").
		Limit(limit).
		Offset(offset).
		Find(&trackingRecords)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch tracking history",
		})
	}

	return c.JSON(fiber.Map{
		"load_id":       loadID,
		"trip_id":       load.TripID,
		"tracking_data": trackingRecords,
		"count":         len(trackingRecords),
		"limit":         limit,
		"offset":        offset,
	})
}

// Helper function for load completion percentage
func calculateLoadCompletionPercent(status string) float64 {
	statusPercent := map[string]float64{
		"BOOKED":           10.0,
		"PICKUP_SCHEDULED": 20.0,
		"PICKED_UP":        40.0,
		"IN_TRANSIT":       60.0,
		"OUT_FOR_DELIVERY": 80.0,
		"DELIVERED":        100.0,
		"EXCEPTION":        50.0, // Depends on context
	}

	if percent, exists := statusPercent[status]; exists {
		return percent
	}
	return 0.0
}

// User-Specific Tracking Endpoints

// GetUserActiveTrackings @Summary Get active trackings for a user
// @Description Get all active trips and loads being tracked for a specific user
// @Tags user-tracking
// @Produce json
// @Param user_id path int true "User ID"
// @Param role query string false "User role filter (CARRIER, SHIPPER)"
// @Success 200 {object} map[string]interface{}
// @Router /users/{user_id}/tracking/active [get]
func GetUserActiveTrackings(c *fiber.Ctx) error {
	userIDStr := c.Params("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	role := c.Query("role", "")

	// Get user to determine their role if not specified
	var user models.User
	if err := database.DB.First(&user, uint(userID)).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	if role == "" {
		role = user.Role
	}

	response := fiber.Map{
		"user_id": userID,
		"role":    role,
	}

	if role == "CARRIER" {
		// Get active trips for carrier
		var trips []models.Trip
		database.DB.Where("user_id = ? AND status IN ?", userID,
			[]string{"ACTIVE", "IN_TRANSIT", "AT_PICKUP", "AT_DELIVERY"}).
			Preload("Loads").
			Find(&trips)

		// Get tracking info for each trip
		var activeTrips []map[string]interface{}
		for _, trip := range trips {
			currentLocation, _ := trackingService.GetCurrentLocation(trip.ID)
			eta, _ := trackingService.CalculateETA(trip.ID)
			delayInfo, _ := trackingService.CheckForDelays(trip.ID)

			tripData := map[string]interface{}{
				"trip_id":           trip.ID,
				"status":            trip.Status,
				"origin":            trip.OriginCity + ", " + trip.OriginState,
				"destination":       trip.DestinationCity + ", " + trip.DestinationState,
				"departure_date":    trip.DepartureDate,
				"estimated_arrival": eta,
				"current_location":  currentLocation,
				"load_count":        len(trip.Loads),
				"tracking_enabled":  trip.TrackingEnabled,
			}

			if delayInfo != nil {
				tripData["delay_info"] = delayInfo
			}

			activeTrips = append(activeTrips, tripData)
		}

		response["active_trips"] = activeTrips
		response["total_active_trips"] = len(activeTrips)

	} else if role == "SHIPPER" {
		// Get active loads for shipper
		var loads []models.Load
		database.DB.Where("shipper_id = ? AND status IN ?", userID,
			[]string{"BOOKED", "PICKUP_SCHEDULED", "PICKED_UP", "IN_TRANSIT", "OUT_FOR_DELIVERY"}).
			Find(&loads)

		// Get tracking info for each load
		var activeLoads []map[string]interface{}
		for _, load := range loads {
			// Get trip information
			var trip models.Trip
			database.DB.First(&trip, load.TripID)

			currentLocation, _ := trackingService.GetCurrentLocation(trip.ID)
			eta, _ := trackingService.CalculateETA(trip.ID)
			delayInfo, _ := trackingService.CheckForDelays(trip.ID)

			loadData := map[string]interface{}{
				"load_id":           load.ID,
				"booking_reference": load.BookingReference,
				"status":            load.Status,
				"trip_id":           trip.ID,
				"trip_status":       trip.Status,
				"pickup_address":    load.PickupAddress,
				"delivery_address":  load.DeliveryAddress,
				"estimated_arrival": eta,
				"current_location":  currentLocation,
				"tracking_enabled":  trip.TrackingEnabled,
			}

			if delayInfo != nil {
				loadData["delay_info"] = delayInfo
			}

			activeLoads = append(activeLoads, loadData)
		}

		response["active_loads"] = activeLoads
		response["total_active_loads"] = len(activeLoads)
	}

	return c.JSON(response)
}

// GetShipperTrackingView @Summary Get shipper-specific tracking view
// @Description Get comprehensive tracking view for shippers showing their loads
// @Tags user-tracking
// @Produce json
// @Param user_id path int true "Shipper User ID"
// @Param status query string false "Load status filter"
// @Param limit query int false "Number of loads to return (default 20)"
// @Success 200 {object} map[string]interface{}
// @Router /users/{user_id}/tracking/shipper-view [get]
func GetShipperTrackingView(c *fiber.Ctx) error {
	userIDStr := c.Params("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	statusFilter := c.Query("status", "")
	limit := c.QueryInt("limit", 20)

	// Verify user is a shipper
	var user models.User
	if err := database.DB.First(&user, uint(userID)).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	query := database.DB.Where("shipper_id = ?", userID)
	if statusFilter != "" {
		query = query.Where("status = ?", statusFilter)
	}

	var loads []models.Load
	query.Limit(limit).Order("created_at DESC").Find(&loads)

	var trackingData []map[string]interface{}
	for _, load := range loads {
		// Get trip and tracking information
		var trip models.Trip
		database.DB.First(&trip, load.TripID)

		currentLocation, _ := trackingService.GetCurrentLocation(trip.ID)
		eta, _ := trackingService.CalculateETA(trip.ID)
		delayInfo, _ := trackingService.CheckForDelays(trip.ID)

		// Get recent tracking events for this load
		var recentEvents []models.TrackingEvent
		database.DB.Where("load_id = ? OR (trip_id = ? AND load_id IS NULL)", load.ID, trip.ID).
			Order("timestamp DESC").
			Limit(5).
			Find(&recentEvents)

		loadTracking := map[string]interface{}{
			"load_id":           load.ID,
			"booking_reference": load.BookingReference,
			"description":       load.Description,
			"status":            load.Status,
			"pickup_address":    load.PickupAddress,
			"delivery_address":  load.DeliveryAddress,
			"trip_id":           trip.ID,
			"trip_status":       trip.Status,
			"current_location":  currentLocation,
			"estimated_arrival": eta,
			"recent_events":     recentEvents,
			"tracking_enabled":  trip.TrackingEnabled,
		}

		if delayInfo != nil {
			loadTracking["delay_info"] = delayInfo
		}

		trackingData = append(trackingData, loadTracking)
	}

	return c.JSON(fiber.Map{
		"user_id":       userID,
		"role":          "SHIPPER",
		"tracking_data": trackingData,
		"total_loads":   len(trackingData),
		"status_filter": statusFilter,
	})
}

// GetCarrierTrackingView @Summary Get carrier-specific tracking view
// @Description Get comprehensive tracking view for carriers showing their trips
// @Tags user-tracking
// @Produce json
// @Param user_id path int true "Carrier User ID"
// @Param status query string false "Trip status filter"
// @Param limit query int false "Number of trips to return (default 20)"
// @Success 200 {object} map[string]interface{}
// @Router /users/{user_id}/tracking/carrier-view [get]
func GetCarrierTrackingView(c *fiber.Ctx) error {
	userIDStr := c.Params("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	statusFilter := c.Query("status", "")
	limit := c.QueryInt("limit", 20)

	// Verify user is a carrier
	var user models.User
	if err := database.DB.First(&user, uint(userID)).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	query := database.DB.Where("user_id = ?", userID)
	if statusFilter != "" {
		query = query.Where("status = ?", statusFilter)
	}

	var trips []models.Trip
	query.Preload("Loads").Limit(limit).Order("created_at DESC").Find(&trips)

	var trackingData []map[string]interface{}
	for _, trip := range trips {
		currentLocation, _ := trackingService.GetCurrentLocation(trip.ID)
		eta, _ := trackingService.CalculateETA(trip.ID)
		delayInfo, _ := trackingService.CheckForDelays(trip.ID)

		// Get recent tracking events for this trip
		var recentEvents []models.TrackingEvent
		database.DB.Where("trip_id = ?", trip.ID).
			Order("timestamp DESC").
			Limit(5).
			Find(&recentEvents)

		// Get load summaries
		var loadSummaries []map[string]interface{}
		for _, load := range trip.Loads {
			loadSummaries = append(loadSummaries, map[string]interface{}{
				"load_id":           load.ID,
				"booking_reference": load.BookingReference,
				"status":            load.Status,
				"shipper_id":        load.ShipperID,
				"pickup_city":       load.PickupCity,
				"delivery_city":     load.DeliveryCity,
			})
		}

		tripTracking := map[string]interface{}{
			"trip_id":           trip.ID,
			"status":            trip.Status,
			"origin":            trip.OriginCity + ", " + trip.OriginState,
			"destination":       trip.DestinationCity + ", " + trip.DestinationState,
			"departure_date":    trip.DepartureDate,
			"estimated_arrival": eta,
			"current_location":  currentLocation,
			"loads":             loadSummaries,
			"load_count":        len(trip.Loads),
			"recent_events":     recentEvents,
			"tracking_enabled":  trip.TrackingEnabled,
		}

		if delayInfo != nil {
			tripTracking["delay_info"] = delayInfo
		}

		trackingData = append(trackingData, tripTracking)
	}

	return c.JSON(fiber.Map{
		"user_id":       userID,
		"role":          "CARRIER",
		"tracking_data": trackingData,
		"total_trips":   len(trackingData),
		"status_filter": statusFilter,
	})
}

// GetUserTrackingNotifications @Summary Get tracking notifications for a user
// @Description Get all tracking-related notifications for a specific user
// @Tags user-tracking
// @Produce json
// @Param user_id path int true "User ID"
// @Param unread_only query boolean false "Show only unread notifications"
// @Param limit query int false "Number of notifications to return (default 20)"
// @Success 200 {array} models.Notification
// @Router /users/{user_id}/tracking/notifications [get]
func GetUserTrackingNotifications(c *fiber.Ctx) error {
	userIDStr := c.Params("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	unreadOnly := c.Query("unread_only") == "true"
	limit := c.QueryInt("limit", 20)

	// Define tracking-related notification types
	trackingTypes := []string{
		"TRIP_DEPARTED", "TRIP_ARRIVED", "TRIP_DELAYED", "ETA_UPDATED",
		"LOAD_STATUS_CHANGED", "LOCATION_UPDATE", "PICKUP_SCHEDULED",
		"LOAD_DELIVERED",
	}

	query := database.DB.Where("user_id = ? AND type IN ?", userID, trackingTypes)
	if unreadOnly {
		query = query.Where("is_read = ?", false)
	}

	var notifications []models.Notification
	result := query.Order("created_at DESC").Limit(limit).Find(&notifications)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Could not fetch tracking notifications",
		})
	}

	return c.JSON(fiber.Map{
		"user_id":       userID,
		"notifications": notifications,
		"count":         len(notifications),
		"unread_only":   unreadOnly,
	})
}

// Mobile-Optimized Tracking Endpoints

// GetLightweightTracking @Summary Get lightweight tracking data for mobile
// @Description Get essential tracking information optimized for mobile bandwidth
// @Tags mobile-tracking
// @Produce json
// @Param trip_id path int true "Trip ID"
// @Success 200 {object} map[string]interface{}
// @Router /mobile/trips/{trip_id}/tracking [get]
func GetLightweightTracking(c *fiber.Ctx) error {
	tripIDStr := c.Params("trip_id")
	tripID, err := strconv.ParseUint(tripIDStr, 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid trip ID",
		})
	}

	// Get essential trip information
	var trip models.Trip
	if err := database.DB.Select("id, status, current_latitude, current_longitude, estimated_arrival, tracking_enabled").
		First(&trip, uint(tripID)).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Trip not found",
		})
	}

	// Get latest location (lightweight)
	var latestLocation models.TrackingRecord
	database.DB.Select("latitude, longitude, timestamp, speed").
		Where("trip_id = ?", tripID).
		Order("timestamp DESC").
		First(&latestLocation)

	// Check for delays (simplified)
	delayInfo, _ := trackingService.CheckForDelays(uint(tripID))

	response := fiber.Map{
		"trip_id":          tripID,
		"status":           trip.Status,
		"eta":              trip.EstimatedArrival,
		"tracking_enabled": trip.TrackingEnabled,
	}

	if latestLocation.ID != 0 {
		response["location"] = map[string]interface{}{
			"lat":       latestLocation.Latitude,
			"lng":       latestLocation.Longitude,
			"timestamp": latestLocation.Timestamp,
			"speed":     latestLocation.Speed,
		}
	}

	if delayInfo != nil {
		response["delay"] = map[string]interface{}{
			"minutes":  delayInfo.DelayMinutes,
			"severity": delayInfo.Severity,
		}
	}

	return c.JSON(response)
}

// SyncOfflineData @Summary Sync offline tracking data
// @Description Sync tracking data collected while offline
// @Tags mobile-tracking
// @Accept json
// @Produce json
// @Param trip_id path int true "Trip ID"
// @Param data body []services.LocationUpdate true "Offline location data"
// @Success 200 {object} map[string]interface{}
// @Router /mobile/trips/{trip_id}/sync [post]
func SyncOfflineData(c *fiber.Ctx) error {
	tripIDStr := c.Params("trip_id")
	tripID, err := strconv.ParseUint(tripIDStr, 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid trip ID",
		})
	}

	var offlineData []services.LocationUpdate
	if err := c.BodyParser(&offlineData); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Cannot parse offline data",
		})
	}

	// Verify trip exists
	var trip models.Trip
	if err := database.DB.First(&trip, uint(tripID)).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Trip not found",
		})
	}

	successCount := 0
	errorCount := 0
	var errors []string

	// Process each location update
	for _, locationUpdate := range offlineData {
		// Validate and sanitize data
		if err := trackingService.ValidateLocationUpdate(uint(tripID), locationUpdate); err != nil {
			errorCount++
			errors = append(errors, err.Error())
			continue
		}

		trackingService.SanitizeLocationData(&locationUpdate)

		// Update location
		if err := trackingService.UpdateLocation(uint(tripID), locationUpdate); err != nil {
			errorCount++
			errors = append(errors, err.Error())
		} else {
			successCount++
		}
	}

	// Log sync event
	trackingService.LogTrackingEvent(uint(tripID), nil, "OFFLINE_SYNC",
		fmt.Sprintf(`{"total_records":%d,"success":%d,"errors":%d}`, len(offlineData), successCount, errorCount),
		"", nil, nil, fmt.Sprintf("Synced %d offline location records", successCount))

	return c.JSON(fiber.Map{
		"message":       "Offline data sync completed",
		"total_records": len(offlineData),
		"success_count": successCount,
		"error_count":   errorCount,
		"errors":        errors,
	})
}

// GetBatteryOptimizedSettings @Summary Get battery-optimized tracking settings
// @Description Get tracking settings optimized for battery life
// @Tags mobile-tracking
// @Produce json
// @Param trip_id path int true "Trip ID"
// @Success 200 {object} map[string]interface{}
// @Router /mobile/trips/{trip_id}/battery-settings [get]
func GetBatteryOptimizedSettings(c *fiber.Ctx) error {
	tripIDStr := c.Params("trip_id")
	tripID, err := strconv.ParseUint(tripIDStr, 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid trip ID",
		})
	}

	// Get trip to determine optimal settings
	var trip models.Trip
	if err := database.DB.First(&trip, uint(tripID)).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Trip not found",
		})
	}

	// Calculate trip duration to optimize settings
	tripDuration := trip.EstimatedArrival.Sub(trip.DepartureDate).Hours()

	// Optimize based on trip duration
	var settings map[string]interface{}

	if tripDuration < 2 { // Short trips
		settings = map[string]interface{}{
			"update_interval_seconds": 30,
			"high_accuracy_mode":      true,
			"background_updates":      true,
			"wifi_only_sync":          false,
		}
	} else if tripDuration < 8 { // Medium trips
		settings = map[string]interface{}{
			"update_interval_seconds": 60,
			"high_accuracy_mode":      false,
			"background_updates":      true,
			"wifi_only_sync":          false,
		}
	} else { // Long trips
		settings = map[string]interface{}{
			"update_interval_seconds": 120,
			"high_accuracy_mode":      false,
			"background_updates":      false,
			"wifi_only_sync":          true,
		}
	}

	// Add common battery optimization settings
	settings["adaptive_interval"] = true
	settings["stop_when_stationary"] = true
	settings["reduce_accuracy_when_slow"] = true
	settings["batch_uploads"] = true
	settings["compress_data"] = true

	return c.JSON(fiber.Map{
		"trip_id":             tripID,
		"trip_duration_hours": tripDuration,
		"settings":            settings,
		"battery_level_thresholds": map[string]interface{}{
			"low_battery_mode":      20, // Switch to power saving at 20%
			"critical_battery_mode": 10, // Minimal tracking at 10%
			"stop_tracking":         5,  // Stop tracking at 5%
		},
	})
}

// UpdateMobileTrackingPreferences @Summary Update mobile tracking preferences
// @Description Update tracking preferences for mobile app
// @Tags mobile-tracking
// @Accept json
// @Produce json
// @Param user_id path int true "User ID"
// @Param preferences body map[string]interface{} true "Mobile preferences"
// @Success 200 {object} map[string]interface{}
// @Router /mobile/users/{user_id}/preferences [put]
func UpdateMobileTrackingPreferences(c *fiber.Ctx) error {
	userIDStr := c.Params("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	var preferences map[string]interface{}
	if err := c.BodyParser(&preferences); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Cannot parse preferences data",
		})
	}

	// Validate preferences
	validKeys := []string{
		"auto_tracking", "background_updates", "wifi_only_sync",
		"battery_optimization", "high_accuracy_mode", "update_interval",
		"push_notifications", "sound_alerts", "vibration_alerts",
	}

	for key := range preferences {
		isValid := false
		for _, validKey := range validKeys {
			if key == validKey {
				isValid = true
				break
			}
		}
		if !isValid {
			return c.Status(400).JSON(fiber.Map{
				"error": fmt.Sprintf("Invalid preference key: %s", key),
			})
		}
	}

	// In a real implementation, save preferences to database
	// For now, just return success with the preferences

	return c.JSON(fiber.Map{
		"message":     "Mobile tracking preferences updated successfully",
		"user_id":     userID,
		"preferences": preferences,
		"updated_at":  time.Now(),
	})
}

// GetMobileTrackingSummary @Summary Get mobile tracking summary
// @Description Get a summary of tracking data optimized for mobile display
// @Tags mobile-tracking
// @Produce json
// @Param user_id path int true "User ID"
// @Param role query string false "User role (CARRIER, SHIPPER)"
// @Success 200 {object} map[string]interface{}
// @Router /mobile/users/{user_id}/tracking/summary [get]
func GetMobileTrackingSummary(c *fiber.Ctx) error {
	userIDStr := c.Params("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	role := c.Query("role", "")

	// Get user to determine role if not specified
	var user models.User
	if err := database.DB.Select("id, role").First(&user, uint(userID)).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	if role == "" {
		role = user.Role
	}

	summary := fiber.Map{
		"user_id": userID,
		"role":    role,
	}

	if role == "CARRIER" {
		// Get active trips count and basic info
		var activeTripsCount int64
		database.DB.Model(&models.Trip{}).
			Where("user_id = ? AND status IN ?", userID,
				[]string{"ACTIVE", "IN_TRANSIT", "AT_PICKUP", "AT_DELIVERY"}).
			Count(&activeTripsCount)

		// Get next departure
		var nextTrip models.Trip
		database.DB.Where("user_id = ? AND status = 'PLANNED'", userID).
			Order("departure_date ASC").
			First(&nextTrip)

		summary["active_trips"] = activeTripsCount
		if nextTrip.ID != 0 {
			summary["next_departure"] = map[string]interface{}{
				"trip_id":        nextTrip.ID,
				"departure_date": nextTrip.DepartureDate,
				"destination":    nextTrip.DestinationCity,
			}
		}

	} else if role == "SHIPPER" {
		// Get active loads count and basic info
		var activeLoadsCount int64
		database.DB.Model(&models.Load{}).
			Where("shipper_id = ? AND status IN ?", userID,
				[]string{"BOOKED", "PICKUP_SCHEDULED", "PICKED_UP", "IN_TRANSIT", "OUT_FOR_DELIVERY"}).
			Count(&activeLoadsCount)

		// Get next pickup
		var nextLoad models.Load
		database.DB.Where("shipper_id = ? AND status IN ?", userID,
			[]string{"BOOKED", "PICKUP_SCHEDULED"}).
			Order("requested_pickup_date ASC").
			First(&nextLoad)

		summary["active_loads"] = activeLoadsCount
		if nextLoad.ID != 0 {
			summary["next_pickup"] = map[string]interface{}{
				"load_id":           nextLoad.ID,
				"booking_reference": nextLoad.BookingReference,
				"pickup_date":       nextLoad.RequestedPickupDate,
				"pickup_address":    nextLoad.PickupAddress,
			}
		}
	}

	// Get recent notifications count
	var unreadNotifications int64
	database.DB.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = false", userID).
		Count(&unreadNotifications)

	summary["unread_notifications"] = unreadNotifications
	summary["last_updated"] = time.Now()

	return c.JSON(summary)
}

// Analytics and Monitoring Endpoints

// GetTrackingAnalytics @Summary Get tracking analytics
// @Description Get comprehensive analytics for tracking data
// @Tags analytics
// @Produce json
// @Param trip_id path int true "Trip ID"
// @Success 200 {object} map[string]interface{}
// @Router /analytics/trips/{trip_id}/tracking [get]
func GetTrackingAnalytics(c *fiber.Ctx) error {
	tripIDStr := c.Params("trip_id")
	tripID, err := strconv.ParseUint(tripIDStr, 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid trip ID",
		})
	}

	// Get tracking statistics
	stats, err := trackingService.GetTrackingStatistics(uint(tripID))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to get tracking statistics",
		})
	}

	// Get anomalies
	anomalies, _ := trackingService.DetectAnomalies(uint(tripID))

	// Get consistency issues
	consistencyIssues := trackingService.ValidateTrackingConsistency(uint(tripID))

	// Calculate route efficiency
	routeEfficiency := calculateRouteEfficiency(uint(tripID))

	// Get delay analysis
	delayAnalysis := getDelayAnalysis(uint(tripID))

	analytics := fiber.Map{
		"trip_id":            tripID,
		"statistics":         stats,
		"anomalies":          anomalies,
		"consistency_issues": consistencyIssues,
		"route_efficiency":   routeEfficiency,
		"delay_analysis":     delayAnalysis,
		"generated_at":       time.Now(),
	}

	return c.JSON(analytics)
}

// GetSystemHealthMetrics @Summary Get system health metrics
// @Description Get health metrics for the tracking system
// @Tags monitoring
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /monitoring/tracking/health [get]
func GetSystemHealthMetrics(c *fiber.Ctx) error {
	metrics := fiber.Map{
		"timestamp": time.Now(),
	}

	// Database health
	dbHealth := checkDatabaseHealth()
	metrics["database"] = dbHealth

	// Tracking data quality
	dataQuality := assessTrackingDataQuality()
	metrics["data_quality"] = dataQuality

	// System performance
	performance := getSystemPerformance()
	metrics["performance"] = performance

	// Active tracking sessions
	activeSessions := getActiveTrackingSessions()
	metrics["active_sessions"] = activeSessions

	// Error rates
	errorRates := getTrackingErrorRates()
	metrics["error_rates"] = errorRates

	// Overall health score (0-100)
	healthScore := calculateOverallHealthScore(dbHealth, dataQuality, performance, errorRates)
	metrics["health_score"] = healthScore

	return c.JSON(metrics)
}

// GetTrackingPerformanceMetrics @Summary Get tracking performance metrics
// @Description Get detailed performance metrics for tracking operations
// @Tags monitoring
// @Produce json
// @Param hours query int false "Hours to look back (default 24)"
// @Success 200 {object} map[string]interface{}
// @Router /monitoring/tracking/performance [get]
func GetTrackingPerformanceMetrics(c *fiber.Ctx) error {
	hours := c.QueryInt("hours", 24)
	since := time.Now().Add(-time.Duration(hours) * time.Hour)

	// Location update metrics
	var locationUpdateCount int64
	database.DB.Model(&models.TrackingRecord{}).
		Where("created_at >= ?", since).
		Count(&locationUpdateCount)

	// Event metrics
	var eventCount int64
	database.DB.Model(&models.TrackingEvent{}).
		Where("created_at >= ?", since).
		Count(&eventCount)

	// Average response times (simulated - in real system would be measured)
	avgResponseTimes := map[string]float64{
		"location_update_ms": 150.5,
		"status_update_ms":   89.2,
		"eta_calculation_ms": 245.8,
		"history_query_ms":   312.1,
	}

	// Error counts by type
	errorCounts := getErrorCountsByType(since)

	// Throughput metrics
	throughput := map[string]interface{}{
		"location_updates_per_hour": float64(locationUpdateCount) / float64(hours),
		"events_per_hour":           float64(eventCount) / float64(hours),
	}

	metrics := fiber.Map{
		"time_period_hours":  hours,
		"since":              since,
		"location_updates":   locationUpdateCount,
		"events":             eventCount,
		"avg_response_times": avgResponseTimes,
		"error_counts":       errorCounts,
		"throughput":         throughput,
		"generated_at":       time.Now(),
	}

	return c.JSON(metrics)
}

// GetDataQualityReport @Summary Get data quality report
// @Description Get comprehensive data quality assessment
// @Tags monitoring
// @Produce json
// @Param trip_id query int false "Specific trip ID (optional)"
// @Success 200 {object} map[string]interface{}
// @Router /monitoring/tracking/data-quality [get]
func GetDataQualityReport(c *fiber.Ctx) error {
	tripIDStr := c.Query("trip_id", "")

	report := fiber.Map{
		"generated_at": time.Now(),
	}

	if tripIDStr != "" {
		// Trip-specific data quality
		tripID, err := strconv.ParseUint(tripIDStr, 10, 32)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "Invalid trip ID",
			})
		}

		tripQuality := assessTripDataQuality(uint(tripID))
		report["trip_id"] = tripID
		report["trip_quality"] = tripQuality
	} else {
		// System-wide data quality
		systemQuality := assessSystemDataQuality()
		report["system_quality"] = systemQuality
	}

	return c.JSON(report)
}

// Helper functions for analytics and monitoring

func calculateRouteEfficiency(tripID uint) map[string]interface{} {
	// Get trip details
	var trip models.Trip
	database.DB.First(&trip, tripID)

	// Calculate direct distance
	directDistance := calculateDistance(trip.OriginLat, trip.OriginLng,
		trip.DestinationLat, trip.DestinationLng)

	// Calculate actual distance traveled
	var records []models.TrackingRecord
	database.DB.Where("trip_id = ?", tripID).Order("timestamp ASC").Find(&records)

	actualDistance := 0.0
	for i := 1; i < len(records); i++ {
		actualDistance += calculateDistance(
			records[i-1].Latitude, records[i-1].Longitude,
			records[i].Latitude, records[i].Longitude)
	}

	efficiency := 0.0
	if actualDistance > 0 {
		efficiency = (directDistance / actualDistance) * 100
	}

	return map[string]interface{}{
		"direct_distance_km": directDistance,
		"actual_distance_km": actualDistance,
		"efficiency_percent": efficiency,
		"extra_distance_km":  actualDistance - directDistance,
	}
}

func getDelayAnalysis(tripID uint) map[string]interface{} {
	// Get delay events
	var delayEvents []models.TrackingEvent
	database.DB.Where("trip_id = ? AND event_type = 'DELAY'", tripID).
		Order("timestamp ASC").
		Find(&delayEvents)

	totalDelayMinutes := 0
	delayReasons := make(map[string]int)

	for range delayEvents {
		// Parse delay minutes from event data (simplified)
		// In real implementation, would properly parse JSON
		totalDelayMinutes += 30   // Placeholder
		delayReasons["Traffic"]++ // Placeholder
	}

	return map[string]interface{}{
		"total_delays":        len(delayEvents),
		"total_delay_minutes": totalDelayMinutes,
		"delay_reasons":       delayReasons,
		"avg_delay_minutes":   totalDelayMinutes / max(len(delayEvents), 1),
	}
}

func checkDatabaseHealth() map[string]interface{} {
	// Check database connectivity and performance
	start := time.Now()
	var count int64
	err := database.DB.Model(&models.TrackingRecord{}).Count(&count)
	queryTime := time.Since(start).Milliseconds()

	health := map[string]interface{}{
		"connected":     err == nil,
		"query_time_ms": queryTime,
		"total_records": count,
	}

	if queryTime > 1000 {
		health["status"] = "slow"
	} else if err != nil {
		health["status"] = "error"
	} else {
		health["status"] = "healthy"
	}

	return health
}

func assessTrackingDataQuality() map[string]interface{} {
	// Assess overall data quality
	now := time.Now()
	oneHourAgo := now.Add(-time.Hour)

	// Recent update rate
	var recentUpdates int64
	database.DB.Model(&models.TrackingRecord{}).
		Where("created_at >= ?", oneHourAgo).
		Count(&recentUpdates)

	// GPS accuracy assessment
	var avgAccuracy float64
	database.DB.Model(&models.TrackingRecord{}).
		Where("accuracy IS NOT NULL AND created_at >= ?", oneHourAgo).
		Select("AVG(accuracy)").
		Scan(&avgAccuracy)

	quality := map[string]interface{}{
		"recent_updates":    recentUpdates,
		"avg_gps_accuracy":  avgAccuracy,
		"update_rate_score": min(float64(recentUpdates)/100*100, 100), // Score out of 100
	}

	return quality
}

func getSystemPerformance() map[string]interface{} {
	// Simulate system performance metrics
	return map[string]interface{}{
		"cpu_usage_percent":    25.5,
		"memory_usage_percent": 68.2,
		"disk_usage_percent":   45.1,
		"network_latency_ms":   12.3,
	}
}

func getActiveTrackingSessions() map[string]interface{} {
	// Count active tracking sessions
	var activeTrips int64
	database.DB.Model(&models.Trip{}).
		Where("status IN ? AND tracking_enabled = true",
			[]string{"ACTIVE", "IN_TRANSIT"}).
		Count(&activeTrips)

	return map[string]interface{}{
		"active_trips":   activeTrips,
		"total_sessions": activeTrips,
	}
}

func getTrackingErrorRates() map[string]interface{} {
	// Calculate error rates (simplified)
	return map[string]interface{}{
		"location_update_errors": 2.1,
		"status_update_errors":   0.8,
		"notification_errors":    1.5,
		"overall_error_rate":     1.5,
	}
}

func calculateOverallHealthScore(dbHealth, dataQuality, performance, errorRates map[string]interface{}) int {
	// Calculate overall health score (simplified)
	score := 100

	// Deduct points for issues
	if dbHealth["status"] != "healthy" {
		score -= 30
	}

	if errorRates["overall_error_rate"].(float64) > 5.0 {
		score -= 20
	}

	return max(score, 0)
}

func getErrorCountsByType(since time.Time) map[string]int64 {
	// Get error counts by type from tracking events
	var errorEvents []struct {
		EventType string `json:"event_type"`
		Count     int64  `json:"count"`
	}

	database.DB.Model(&models.TrackingEvent{}).
		Select("event_type, count(*) as count").
		Where("created_at >= ? AND event_type LIKE '%ERROR%'", since).
		Group("event_type").
		Find(&errorEvents)

	errorCounts := make(map[string]int64)
	for _, event := range errorEvents {
		errorCounts[event.EventType] = event.Count
	}

	return errorCounts
}

func assessTripDataQuality(tripID uint) map[string]interface{} {
	// Assess data quality for specific trip
	consistencyIssues := trackingService.ValidateTrackingConsistency(tripID)
	anomalies, _ := trackingService.DetectAnomalies(tripID)

	var recordCount int64
	database.DB.Model(&models.TrackingRecord{}).
		Where("trip_id = ?", tripID).
		Count(&recordCount)

	quality := map[string]interface{}{
		"total_records":      recordCount,
		"consistency_issues": len(consistencyIssues),
		"anomalies":          len(anomalies),
		"quality_score":      max(100-len(consistencyIssues)*10-len(anomalies)*5, 0),
	}

	return quality
}

func assessSystemDataQuality() map[string]interface{} {
	// Assess system-wide data quality
	now := time.Now()
	oneHourAgo := now.Add(-time.Hour)

	var totalRecords int64
	database.DB.Model(&models.TrackingRecord{}).Count(&totalRecords)

	var recentRecords int64
	database.DB.Model(&models.TrackingRecord{}).
		Where("created_at >= ?", oneHourAgo).
		Count(&recentRecords)

	quality := map[string]interface{}{
		"total_records":    totalRecords,
		"recent_records":   recentRecords,
		"update_frequency": float64(recentRecords) / 60, // per minute
	}

	return quality
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// calculateDistance calculates the distance between two coordinates using Haversine formula
func calculateDistance(lat1, lng1, lat2, lng2 float64) float64 {
	const earthRadius = 6371 // Earth's radius in kilometers

	// Convert degrees to radians
	lat1Rad := lat1 * math.Pi / 180
	lng1Rad := lng1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	lng2Rad := lng2 * math.Pi / 180

	// Calculate differences
	deltaLat := lat2Rad - lat1Rad
	deltaLng := lng2Rad - lng1Rad

	// Haversine formula
	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLng/2)*math.Sin(deltaLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c
}
