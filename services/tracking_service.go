package services

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
	"triplink/backend/models"
	
	"gorm.io/gorm"
)

// LocationUpdate represents a location update request
type LocationUpdate struct {
	Latitude  float64  `json:"latitude"`
	Longitude float64  `json:"longitude"`
	Altitude  *float64 `json:"altitude,omitempty"`
	Speed     *float64 `json:"speed,omitempty"`
	Heading   *float64 `json:"heading,omitempty"`
	Accuracy  *float64 `json:"accuracy,omitempty"`
	Source    string   `json:"source"`
}

// DelayInfo represents delay information
type DelayInfo struct {
	DelayMinutes int    `json:"delay_minutes"`
	Reason       string `json:"reason"`
	Severity     string `json:"severity"` // LOW, MEDIUM, HIGH, CRITICAL
}

// TrackingService provides tracking-related operations
type TrackingService struct{
	db *gorm.DB
}

// NewTrackingService creates a new tracking service instance
func NewTrackingService(db *gorm.DB) *TrackingService {
	return &TrackingService{db: db}
}

// UpdateLocation updates the location for a trip
func (ts *TrackingService) UpdateLocation(tripID uint, location LocationUpdate) error {
	// Validate coordinates
	if !isValidCoordinate(location.Latitude, location.Longitude) {
		return errors.New("invalid coordinates")
	}

	// Create tracking record
	trackingRecord := models.TrackingRecord{
		TripID:    tripID,
		Latitude:  location.Latitude,
		Longitude: location.Longitude,
		Altitude:  location.Altitude,
		Speed:     location.Speed,
		Heading:   location.Heading,
		Accuracy:  location.Accuracy,
		Timestamp: time.Now(),
		Source:    location.Source,
		Status:    "ACTIVE",
	}

	// Save tracking record
	if err := ts.db.Create(&trackingRecord).Error; err != nil {
		return err
	}

	// Update trip's current location
	now := time.Now()
	err := ts.db.Model(&models.Trip{}).Where("id = ?", tripID).Updates(map[string]interface{}{
		"current_latitude":     location.Latitude,
		"current_longitude":    location.Longitude,
		"last_location_update": &now,
	}).Error

	if err != nil {
		return err
	}

	// Update ETA based on new location
	_, err = ts.CalculateETA(tripID)
	return err
}

// GetCurrentLocation retrieves the most recent location for a trip
func (ts *TrackingService) GetCurrentLocation(tripID uint) (*models.TrackingRecord, error) {
	var trackingRecord models.TrackingRecord
	err := ts.db.Where("trip_id = ?", tripID).Order("timestamp DESC").First(&trackingRecord).Error

	if err != nil {
		return nil, err
	}

	return &trackingRecord, nil
}

// CalculateETA calculates estimated time of arrival based on current location
func (ts *TrackingService) CalculateETA(tripID uint) (*time.Time, error) {
	var trip models.Trip
	if err := ts.db.First(&trip, tripID).Error; err != nil {
		return nil, err
	}

	// If no current location, return original estimated arrival
	if trip.CurrentLatitude == nil || trip.CurrentLongitude == nil {
		return &trip.EstimatedArrival, nil
	}

	// Calculate distance to destination
	distance := calculateDistance(*trip.CurrentLatitude, *trip.CurrentLongitude,
		trip.DestinationLat, trip.DestinationLng)

	// Estimate average speed (default 60 km/h if no recent speed data)
	avgSpeed := 60.0

	// Get recent tracking records to calculate average speed
	var recentRecords []models.TrackingRecord
	ts.db.Where("trip_id = ? AND speed IS NOT NULL", tripID).
		Order("timestamp DESC").
		Limit(5).
		Find(&recentRecords)

	if len(recentRecords) > 0 {
		totalSpeed := 0.0
		for _, record := range recentRecords {
			if record.Speed != nil {
				totalSpeed += *record.Speed
			}
		}
		avgSpeed = totalSpeed / float64(len(recentRecords))

		// Ensure minimum speed to avoid division by zero
		if avgSpeed < 10 {
			avgSpeed = 30 // Default to 30 km/h for city driving
		}
	}

	// Calculate ETA
	hoursToDestination := distance / avgSpeed
	eta := time.Now().Add(time.Duration(hoursToDestination * float64(time.Hour)))

	// Update trip's estimated arrival
	ts.db.Model(&trip).Update("estimated_arrival", eta)

	return &eta, nil
}

// UpdateTripStatus updates the status of a trip with validation
func (ts *TrackingService) UpdateTripStatus(tripID uint, newStatus string) error {
	var trip models.Trip
	if err := ts.db.First(&trip, tripID).Error; err != nil {
		return err
	}

	// Validate status transition
	if !isValidStatusTransition(trip.Status, newStatus) {
		return errors.New("invalid status transition")
	}

	// Update trip status
	previousStatus := trip.Status
	now := time.Now()

	err := ts.db.Model(&trip).Updates(map[string]interface{}{
		"status": newStatus,
	}).Error

	if err != nil {
		return err
	}

	// Create or update tracking status record
	var trackingStatus models.TrackingStatus
	result := ts.db.Where("trip_id = ?", tripID).First(&trackingStatus)

	if result.Error != nil {
		// Create new tracking status
		trackingStatus = models.TrackingStatus{
			TripID:            tripID,
			CurrentStatus:     newStatus,
			PreviousStatus:    previousStatus,
			StatusChangedAt:   now,
			CompletionPercent: calculateCompletionPercent(newStatus),
		}
		ts.db.Create(&trackingStatus)
	} else {
		// Update existing tracking status
		trackingStatus.PreviousStatus = trackingStatus.CurrentStatus
		trackingStatus.CurrentStatus = newStatus
		trackingStatus.StatusChangedAt = now
		trackingStatus.CompletionPercent = calculateCompletionPercent(newStatus)
		ts.db.Save(&trackingStatus)
	}

	// Create tracking event
	event := models.TrackingEvent{
		TripID:      tripID,
		EventType:   "STATUS_CHANGE",
		EventData:   `{"from":"` + previousStatus + `","to":"` + newStatus + `"}`,
		Location:    "",
		Timestamp:   now,
		Description: "Trip status changed from " + previousStatus + " to " + newStatus,
	}
	ts.db.Create(&event)

	return nil
}

// CheckForDelays checks if a trip is delayed and returns delay information
func (ts *TrackingService) CheckForDelays(tripID uint) (*DelayInfo, error) {
	var trip models.Trip
	if err := ts.db.First(&trip, tripID).Error; err != nil {
		return nil, err
	}

	now := time.Now()

	// Check if trip is delayed
	if now.After(trip.EstimatedArrival) {
		delayMinutes := int(now.Sub(trip.EstimatedArrival).Minutes())

		severity := "LOW"
		if delayMinutes > 120 {
			severity = "CRITICAL"
		} else if delayMinutes > 60 {
			severity = "HIGH"
		} else if delayMinutes > 30 {
			severity = "MEDIUM"
		}

		return &DelayInfo{
			DelayMinutes: delayMinutes,
			Reason:       "Behind schedule",
			Severity:     severity,
		}, nil
	}

	return nil, nil // No delay
}

// ProcessDelayAlerts checks for delays and sends notifications if thresholds are exceeded
func (ts *TrackingService) ProcessDelayAlerts(tripID uint) error {
	delayInfo, err := ts.CheckForDelays(tripID)
	if err != nil {
		return err
	}

	if delayInfo == nil {
		return nil // No delay
	}

	// Check if we should send an alert based on delay thresholds
	shouldAlert := false
	alertThresholds := []int{30, 60, 120, 240} // Alert at 30min, 1hr, 2hr, 4hr delays

	for _, threshold := range alertThresholds {
		if delayInfo.DelayMinutes >= threshold {
			// Check if we've already sent an alert for this threshold
			if !ts.hasDelayAlertBeenSent(tripID, threshold) {
				shouldAlert = true
				ts.markDelayAlertSent(tripID, threshold)
				break
			}
		}
	}

	if shouldAlert {
		// Get all shippers on this trip and send delay notifications
		var loads []models.Load
		if err := ts.db.Where("trip_id = ?", tripID).Find(&loads).Error; err != nil {
			return err
		}

		// Create delay notification for each shipper
		for _, load := range loads {
			notification := models.Notification{
				UserID:    load.ShipperID,
				Title:     "Shipment Delayed",
				Message:   fmt.Sprintf("Your shipment is delayed by %d minutes due to %s", delayInfo.DelayMinutes, delayInfo.Reason),
				Type:      "TRIP_DELAYED",
				RelatedID: tripID,
			}
			ts.db.Create(&notification)
		}

		// Create tracking event for delay
		event := models.TrackingEvent{
			TripID:      tripID,
			EventType:   "DELAY",
			EventData:   fmt.Sprintf(`{"delay_minutes":%d,"reason":"%s","severity":"%s"}`, delayInfo.DelayMinutes, delayInfo.Reason, delayInfo.Severity),
			Location:    "",
			Timestamp:   time.Now(),
			Description: fmt.Sprintf("Trip delayed by %d minutes - %s", delayInfo.DelayMinutes, delayInfo.Reason),
		}
		ts.db.Create(&event)

		// Update tracking status with delay information
		ts.updateDelayStatus(tripID, delayInfo.DelayMinutes, delayInfo.Reason)
	}

	return nil
}

// updateDelayStatus updates the tracking status with delay information
func (ts *TrackingService) updateDelayStatus(tripID uint, delayMinutes int, reason string) error {
	var trackingStatus models.TrackingStatus
	result := ts.db.Where("trip_id = ?", tripID).First(&trackingStatus)

	if result.Error != nil {
		// Create new tracking status with delay info
		trackingStatus = models.TrackingStatus{
			TripID:          tripID,
			CurrentStatus:   "DELAYED",
			DelayMinutes:    &delayMinutes,
			DelayReason:     reason,
			StatusChangedAt: time.Now(),
		}
		return ts.db.Create(&trackingStatus).Error
	} else {
		// Update existing tracking status
		trackingStatus.DelayMinutes = &delayMinutes
		trackingStatus.DelayReason = reason
		if trackingStatus.CurrentStatus != "DELAYED" {
			trackingStatus.PreviousStatus = trackingStatus.CurrentStatus
			trackingStatus.CurrentStatus = "DELAYED"
			trackingStatus.StatusChangedAt = time.Now()
		}
		return ts.db.Save(&trackingStatus).Error
	}
}

// hasDelayAlertBeenSent checks if a delay alert has already been sent for a specific threshold
func (ts *TrackingService) hasDelayAlertBeenSent(tripID uint, thresholdMinutes int) bool {
	var count int64
	ts.db.Model(&models.TrackingEvent{}).
		Where("trip_id = ? AND event_type = 'DELAY' AND event_data LIKE ?",
			tripID, fmt.Sprintf("%%\"delay_minutes\":%d%%", thresholdMinutes)).
		Count(&count)
	return count > 0
}

// markDelayAlertSent marks that a delay alert has been sent for a specific threshold
func (ts *TrackingService) markDelayAlertSent(tripID uint, thresholdMinutes int) {
	// This is implicitly handled by creating the tracking event in ProcessDelayAlerts
	// The event creation serves as the marker that an alert was sent
}

// DetectAnomalies detects unusual patterns in tracking data that might indicate issues
func (ts *TrackingService) DetectAnomalies(tripID uint) ([]string, error) {
	var anomalies []string

	// Get recent tracking records
	var records []models.TrackingRecord
	err := ts.db.Where("trip_id = ?", tripID).
		Order("timestamp DESC").
		Limit(10).
		Find(&records).Error

	if err != nil || len(records) < 2 {
		return anomalies, err
	}

	// Check for unusual speed patterns
	for i := 0; i < len(records)-1; i++ {
		current := records[i]
		previous := records[i+1]

		if current.Speed != nil && previous.Speed != nil {
			speedDiff := math.Abs(*current.Speed - *previous.Speed)

			// Flag sudden speed changes > 50 km/h
			if speedDiff > 50 {
				anomalies = append(anomalies, fmt.Sprintf("Sudden speed change detected: %.1f km/h difference", speedDiff))
			}

			// Flag unusually high speeds > 120 km/h
			if *current.Speed > 120 {
				anomalies = append(anomalies, fmt.Sprintf("High speed detected: %.1f km/h", *current.Speed))
			}
		}

		// Check for location jumps (teleportation detection)
		distance := calculateDistance(current.Latitude, current.Longitude, previous.Latitude, previous.Longitude)
		timeDiff := current.Timestamp.Sub(previous.Timestamp).Hours()

		if timeDiff > 0 {
			impliedSpeed := distance / timeDiff

			// Flag impossible speeds > 200 km/h for ground transport
			if impliedSpeed > 200 {
				anomalies = append(anomalies, fmt.Sprintf("Impossible speed detected: %.1f km/h between locations", impliedSpeed))
			}
		}
	}

	// Check for long periods without updates
	if len(records) > 0 {
		lastUpdate := records[0].Timestamp
		hoursSinceUpdate := time.Since(lastUpdate).Hours()

		if hoursSinceUpdate > 4 {
			anomalies = append(anomalies, fmt.Sprintf("No location updates for %.1f hours", hoursSinceUpdate))
		}
	}

	return anomalies, nil
}

// TrackingFilters represents filters for tracking history queries
type TrackingFilters struct {
	StartDate  *time.Time `json:"start_date,omitempty"`
	EndDate    *time.Time `json:"end_date,omitempty"`
	EventTypes []string   `json:"event_types,omitempty"`
	Source     string     `json:"source,omitempty"`
	Status     string     `json:"status,omitempty"`
	Limit      int        `json:"limit,omitempty"`
	Offset     int        `json:"offset,omitempty"`
}

// GetTrackingHistory retrieves tracking history with optional filters
func (ts *TrackingService) GetTrackingHistory(tripID uint, filters TrackingFilters) ([]models.TrackingRecord, error) {
	query := ts.db.Where("trip_id = ?", tripID)

	// Apply filters
	if filters.StartDate != nil {
		query = query.Where("timestamp >= ?", *filters.StartDate)
	}
	if filters.EndDate != nil {
		query = query.Where("timestamp <= ?", *filters.EndDate)
	}
	if filters.Source != "" {
		query = query.Where("source = ?", filters.Source)
	}
	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}

	// Set default limit if not specified
	limit := filters.Limit
	if limit <= 0 {
		limit = 100
	}

	offset := filters.Offset
	if offset < 0 {
		offset = 0
	}

	var records []models.TrackingRecord
	err := query.Order("timestamp DESC").
		Limit(limit).
		Offset(offset).
		Find(&records).Error

	return records, err
}

// GetTrackingEvents retrieves tracking events with optional filters
func (ts *TrackingService) GetTrackingEvents(tripID uint, eventTypes []string, limit int) ([]models.TrackingEvent, error) {
	query := ts.db.Where("trip_id = ?", tripID)

	if len(eventTypes) > 0 {
		query = query.Where("event_type IN ?", eventTypes)
	}

	if limit <= 0 {
		limit = 50
	}

	var events []models.TrackingEvent
	err := query.Order("timestamp DESC").
		Limit(limit).
		Find(&events).Error

	return events, err
}

// LogTrackingEvent creates a new tracking event with comprehensive logging
func (ts *TrackingService) LogTrackingEvent(tripID uint, loadID *uint, eventType string, eventData string, location string, latitude *float64, longitude *float64, description string) error {
	event := models.TrackingEvent{
		TripID:      tripID,
		LoadID:      loadID,
		EventType:   eventType,
		EventData:   eventData,
		Location:    location,
		Latitude:    latitude,
		Longitude:   longitude,
		Timestamp:   time.Now(),
		Description: description,
	}

	return ts.db.Create(&event).Error
}

// GetAuditTrail provides a comprehensive audit trail of all tracking activities
func (ts *TrackingService) GetAuditTrail(tripID uint, includeSystemEvents bool) (map[string]interface{}, error) {
	// Get all tracking records
	var trackingRecords []models.TrackingRecord
	ts.db.Where("trip_id = ?", tripID).
		Order("timestamp ASC").
		Find(&trackingRecords)

	// Get all tracking events
	var trackingEvents []models.TrackingEvent
	query := ts.db.Where("trip_id = ?", tripID)
	if !includeSystemEvents {
		query = query.Where("event_type NOT IN ?", []string{"SYSTEM_UPDATE", "AUTO_CALCULATION"})
	}
	query.Order("timestamp ASC").Find(&trackingEvents)

	// Get status changes
	var statusChanges []models.TrackingStatus
	ts.db.Where("trip_id = ?", tripID).
		Order("status_changed_at ASC").
		Find(&statusChanges)

	// Combine all activities into a timeline
	timeline := make([]map[string]interface{}, 0)

	// Add tracking records to timeline
	for _, record := range trackingRecords {
		timeline = append(timeline, map[string]interface{}{
			"timestamp": record.Timestamp,
			"type":      "location_update",
			"latitude":  record.Latitude,
			"longitude": record.Longitude,
			"speed":     record.Speed,
			"source":    record.Source,
			"accuracy":  record.Accuracy,
		})
	}

	// Add tracking events to timeline
	for _, event := range trackingEvents {
		timeline = append(timeline, map[string]interface{}{
			"timestamp":   event.Timestamp,
			"type":        "event",
			"event_type":  event.EventType,
			"description": event.Description,
			"location":    event.Location,
			"event_data":  event.EventData,
		})
	}

	// Add status changes to timeline
	for _, status := range statusChanges {
		timeline = append(timeline, map[string]interface{}{
			"timestamp":          status.StatusChangedAt,
			"type":               "status_change",
			"current_status":     status.CurrentStatus,
			"previous_status":    status.PreviousStatus,
			"completion_percent": status.CompletionPercent,
			"delay_minutes":      status.DelayMinutes,
			"delay_reason":       status.DelayReason,
		})
	}

	// Sort timeline by timestamp
	// Note: In a real implementation, you'd sort this slice by timestamp

	return map[string]interface{}{
		"trip_id":              tripID,
		"timeline":             timeline,
		"total_records":        len(trackingRecords),
		"total_events":         len(trackingEvents),
		"total_status_changes": len(statusChanges),
		"generated_at":         time.Now(),
	}, nil
}

// CleanupOldTrackingData removes old tracking data based on retention policies
func (ts *TrackingService) CleanupOldTrackingData(retentionDays int) error {
	cutoffDate := time.Now().AddDate(0, 0, -retentionDays)

	// Delete old tracking records
	result := ts.db.Where("timestamp < ?", cutoffDate).Delete(&models.TrackingRecord{})
	if result.Error != nil {
		return result.Error
	}

	// Delete old tracking events (keep critical events longer)
	criticalEvents := []string{"DEPARTURE", "ARRIVAL", "DELAY", "EXCEPTION"}
	ts.db.Where("timestamp < ? AND event_type NOT IN ?", cutoffDate, criticalEvents).Delete(&models.TrackingEvent{})

	// Log cleanup activity
	ts.LogTrackingEvent(0, nil, "SYSTEM_CLEANUP",
		fmt.Sprintf(`{"retention_days":%d,"records_deleted":%d}`, retentionDays, result.RowsAffected),
		"", nil, nil, fmt.Sprintf("Cleaned up tracking data older than %d days", retentionDays))

	return nil
}

// GetTrackingStatistics provides statistics about tracking data
func (ts *TrackingService) GetTrackingStatistics(tripID uint) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Count tracking records
	var recordCount int64
	ts.db.Model(&models.TrackingRecord{}).Where("trip_id = ?", tripID).Count(&recordCount)
	stats["total_location_updates"] = recordCount

	// Count tracking events by type
	var eventStats []struct {
		EventType string `json:"event_type"`
		Count     int64  `json:"count"`
	}
	ts.db.Model(&models.TrackingEvent{}).
		Select("event_type, count(*) as count").
		Where("trip_id = ?", tripID).
		Group("event_type").
		Find(&eventStats)
	stats["events_by_type"] = eventStats

	// Get first and last location updates
	var firstRecord, lastRecord models.TrackingRecord
	ts.db.Where("trip_id = ?", tripID).Order("timestamp ASC").First(&firstRecord)
	ts.db.Where("trip_id = ?", tripID).Order("timestamp DESC").First(&lastRecord)

	if firstRecord.ID != 0 {
		stats["first_update"] = firstRecord.Timestamp
		stats["last_update"] = lastRecord.Timestamp
		stats["tracking_duration_hours"] = lastRecord.Timestamp.Sub(firstRecord.Timestamp).Hours()
	}

	// Calculate average update frequency
	if recordCount > 1 {
		duration := lastRecord.Timestamp.Sub(firstRecord.Timestamp)
		avgInterval := duration / time.Duration(recordCount-1)
		stats["average_update_interval_minutes"] = avgInterval.Minutes()
	}

	// Get speed statistics
	var speedStats struct {
		MaxSpeed *float64 `json:"max_speed"`
		MinSpeed *float64 `json:"min_speed"`
		AvgSpeed *float64 `json:"avg_speed"`
	}
	ts.db.Model(&models.TrackingRecord{}).
		Select("MAX(speed) as max_speed, MIN(speed) as min_speed, AVG(speed) as avg_speed").
		Where("trip_id = ? AND speed IS NOT NULL", tripID).
		Scan(&speedStats)

	if speedStats.MaxSpeed != nil {
		stats["max_speed_kmh"] = *speedStats.MaxSpeed
		stats["min_speed_kmh"] = *speedStats.MinSpeed
		stats["avg_speed_kmh"] = *speedStats.AvgSpeed
	}

	return stats, nil
}

// ExportTrackingData exports tracking data in various formats
func (ts *TrackingService) ExportTrackingData(tripID uint, format string) (interface{}, error) {
	// Get all tracking data
	var records []models.TrackingRecord
	ts.db.Where("trip_id = ?", tripID).Order("timestamp ASC").Find(&records)

	var events []models.TrackingEvent
	ts.db.Where("trip_id = ?", tripID).Order("timestamp ASC").Find(&events)

	switch format {
	case "json":
		return map[string]interface{}{
			"trip_id":          tripID,
			"tracking_records": records,
			"tracking_events":  events,
			"exported_at":      time.Now(),
		}, nil
	case "csv":
		// In a real implementation, you'd convert to CSV format
		return "CSV export not implemented in this example", nil
	case "gpx":
		// In a real implementation, you'd convert to GPX format for GPS devices
		return "GPX export not implemented in this example", nil
	default:
		return nil, errors.New("unsupported export format")
	}
}

// Validation and Error Handling

// TrackingError represents a tracking-specific error
type TrackingError struct {
	Code      string    `json:"code"`
	Message   string    `json:"message"`
	Details   string    `json:"details,omitempty"`
	Timestamp time.Time `json:"timestamp"`
	TripID    *uint     `json:"trip_id,omitempty"`
	LoadID    *uint     `json:"load_id,omitempty"`
}

func (e TrackingError) Error() string {
	return fmt.Sprintf("[%s] %s: %s", e.Code, e.Message, e.Details)
}

// NewTrackingError creates a new tracking error
func NewTrackingError(code, message, details string, tripID *uint, loadID *uint) *TrackingError {
	return &TrackingError{
		Code:      code,
		Message:   message,
		Details:   details,
		Timestamp: time.Now(),
		TripID:    tripID,
		LoadID:    loadID,
	}
}

// ValidateLocationUpdate validates location update data
func (ts *TrackingService) ValidateLocationUpdate(tripID uint, location LocationUpdate) error {
	// Validate coordinates
	if !isValidCoordinate(location.Latitude, location.Longitude) {
		return NewTrackingError("INVALID_COORDINATES",
			"Invalid GPS coordinates",
			fmt.Sprintf("Latitude: %.6f, Longitude: %.6f", location.Latitude, location.Longitude),
			&tripID, nil)
	}

	// Validate altitude if provided
	if location.Altitude != nil && (*location.Altitude < -500 || *location.Altitude > 10000) {
		return NewTrackingError("INVALID_ALTITUDE",
			"Altitude out of reasonable range",
			fmt.Sprintf("Altitude: %.2f meters", *location.Altitude),
			&tripID, nil)
	}

	// Validate speed if provided
	if location.Speed != nil && (*location.Speed < 0 || *location.Speed > 300) {
		return NewTrackingError("INVALID_SPEED",
			"Speed out of reasonable range",
			fmt.Sprintf("Speed: %.2f km/h", *location.Speed),
			&tripID, nil)
	}

	// Validate heading if provided
	if location.Heading != nil && (*location.Heading < 0 || *location.Heading >= 360) {
		return NewTrackingError("INVALID_HEADING",
			"Heading must be between 0 and 359 degrees",
			fmt.Sprintf("Heading: %.2f degrees", *location.Heading),
			&tripID, nil)
	}

	// Validate accuracy if provided
	if location.Accuracy != nil && (*location.Accuracy < 0 || *location.Accuracy > 1000) {
		return NewTrackingError("INVALID_ACCURACY",
			"GPS accuracy out of reasonable range",
			fmt.Sprintf("Accuracy: %.2f meters", *location.Accuracy),
			&tripID, nil)
	}

	// Validate source
	validSources := []string{"GPS", "MANUAL", "ESTIMATED", "NETWORK", "PASSIVE"}
	isValidSource := false
	for _, validSource := range validSources {
		if location.Source == validSource {
			isValidSource = true
			break
		}
	}
	if !isValidSource {
		return NewTrackingError("INVALID_SOURCE",
			"Invalid location source",
			fmt.Sprintf("Source: %s", location.Source),
			&tripID, nil)
	}

	return nil
}

// ValidateStatusTransition validates if a status transition is allowed
func (ts *TrackingService) ValidateStatusTransition(tripID uint, currentStatus, newStatus string) error {
	if !isValidStatusTransition(currentStatus, newStatus) {
		return NewTrackingError("INVALID_STATUS_TRANSITION",
			"Status transition not allowed",
			fmt.Sprintf("From: %s, To: %s", currentStatus, newStatus),
			&tripID, nil)
	}
	return nil
}

// SanitizeLocationData cleans and normalizes location data
func (ts *TrackingService) SanitizeLocationData(location *LocationUpdate) {
	// Round coordinates to 6 decimal places (approximately 0.1 meter precision)
	location.Latitude = math.Round(location.Latitude*1000000) / 1000000
	location.Longitude = math.Round(location.Longitude*1000000) / 1000000

	// Round altitude to 1 decimal place if provided
	if location.Altitude != nil {
		rounded := math.Round(*location.Altitude*10) / 10
		location.Altitude = &rounded
	}

	// Round speed to 1 decimal place if provided
	if location.Speed != nil {
		rounded := math.Round(*location.Speed*10) / 10
		location.Speed = &rounded
	}

	// Round heading to nearest degree if provided
	if location.Heading != nil {
		rounded := math.Round(*location.Heading)
		location.Heading = &rounded
	}

	// Round accuracy to 1 decimal place if provided
	if location.Accuracy != nil {
		rounded := math.Round(*location.Accuracy*10) / 10
		location.Accuracy = &rounded
	}

	// Normalize source to uppercase
	location.Source = strings.ToUpper(location.Source)
}

// RetryLocationUpdate implements retry logic for failed location updates
func (ts *TrackingService) RetryLocationUpdate(tripID uint, location LocationUpdate, maxRetries int) error {
	var lastError error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff: wait 2^attempt seconds
			waitTime := time.Duration(math.Pow(2, float64(attempt))) * time.Second
			time.Sleep(waitTime)
		}

		err := ts.UpdateLocation(tripID, location)
		if err == nil {
			// Success
			if attempt > 0 {
				// Log successful retry
				ts.LogTrackingEvent(tripID, nil, "RETRY_SUCCESS",
					fmt.Sprintf(`{"attempt":%d,"max_retries":%d}`, attempt+1, maxRetries+1),
					"", nil, nil, fmt.Sprintf("Location update succeeded after %d retries", attempt))
			}
			return nil
		}

		lastError = err

		// Log retry attempt
		ts.LogTrackingEvent(tripID, nil, "RETRY_ATTEMPT",
			fmt.Sprintf(`{"attempt":%d,"error":"%s"}`, attempt+1, err.Error()),
			"", nil, nil, fmt.Sprintf("Location update retry attempt %d failed", attempt+1))
	}

	// All retries failed
	ts.LogTrackingEvent(tripID, nil, "RETRY_FAILED",
		fmt.Sprintf(`{"max_retries":%d,"final_error":"%s"}`, maxRetries+1, lastError.Error()),
		"", nil, nil, fmt.Sprintf("Location update failed after %d attempts", maxRetries+1))

	return NewTrackingError("UPDATE_FAILED_AFTER_RETRIES",
		"Location update failed after multiple attempts",
		lastError.Error(),
		&tripID, nil)
}

// ValidateTrackingConsistency checks for data consistency issues
func (ts *TrackingService) ValidateTrackingConsistency(tripID uint) []string {
	var issues []string

	// Get recent tracking records
	var records []models.TrackingRecord
	ts.db.Where("trip_id = ?", tripID).
		Order("timestamp DESC").
		Limit(10).
		Find(&records)

	if len(records) < 2 {
		return issues
	}

	// Check for timestamp consistency
	for i := 0; i < len(records)-1; i++ {
		current := records[i]
		next := records[i+1]

		// Records should be in chronological order
		if current.Timestamp.Before(next.Timestamp) {
			issues = append(issues, fmt.Sprintf("Timestamp order inconsistency detected at record %d", current.ID))
		}

		// Check for duplicate timestamps
		if current.Timestamp.Equal(next.Timestamp) {
			issues = append(issues, fmt.Sprintf("Duplicate timestamp detected: %s", current.Timestamp.Format(time.RFC3339)))
		}

		// Check for unrealistic location jumps
		distance := calculateDistance(current.Latitude, current.Longitude, next.Latitude, next.Longitude)
		timeDiff := current.Timestamp.Sub(next.Timestamp).Hours()

		if timeDiff > 0 {
			impliedSpeed := distance / timeDiff
			if impliedSpeed > 200 { // Unrealistic for ground transport
				issues = append(issues, fmt.Sprintf("Unrealistic location jump detected: %.1f km in %.2f hours (%.1f km/h)",
					distance, timeDiff, impliedSpeed))
			}
		}
	}

	// Check for stale data
	if len(records) > 0 {
		lastUpdate := records[0].Timestamp
		hoursSinceUpdate := time.Since(lastUpdate).Hours()

		if hoursSinceUpdate > 6 {
			issues = append(issues, fmt.Sprintf("Stale tracking data: last update %.1f hours ago", hoursSinceUpdate))
		}
	}

	return issues
}

// RecoverFromTrackingError attempts to recover from tracking errors
func (ts *TrackingService) RecoverFromTrackingError(tripID uint, errorType string) error {
	switch errorType {
	case "STALE_DATA":
		// Try to get fresh location data
		return ts.requestLocationUpdate(tripID)
	case "INVALID_COORDINATES":
		// Mark location as estimated and use last known good location
		return ts.useLastKnownLocation(tripID)
	case "DATABASE_ERROR":
		// Attempt to reconnect and retry
		return ts.retryDatabaseOperation(tripID)
	default:
		return errors.New("unknown error type, cannot recover")
	}
}

// requestLocationUpdate requests a fresh location update
func (ts *TrackingService) requestLocationUpdate(tripID uint) error {
	// In a real implementation, this would trigger a request to the mobile app or GPS device
	ts.LogTrackingEvent(tripID, nil, "LOCATION_UPDATE_REQUESTED",
		`{"reason":"stale_data_recovery"}`,
		"", nil, nil, "Requested fresh location update due to stale data")
	return nil
}

// useLastKnownLocation falls back to the last known good location
func (ts *TrackingService) useLastKnownLocation(tripID uint) error {
	var lastGoodRecord models.TrackingRecord
	err := ts.db.Where("trip_id = ? AND status = 'ACTIVE'", tripID).
		Order("timestamp DESC").
		First(&lastGoodRecord).Error

	if err != nil {
		return err
	}

	// Create estimated location record
	estimatedLocation := LocationUpdate{
		Latitude:  lastGoodRecord.Latitude,
		Longitude: lastGoodRecord.Longitude,
		Source:    "ESTIMATED",
	}

	return ts.UpdateLocation(tripID, estimatedLocation)
}

// retryDatabaseOperation retries database operations
func (ts *TrackingService) retryDatabaseOperation(tripID uint) error {
	// In a real implementation, this would attempt to reconnect to the database
	ts.LogTrackingEvent(tripID, nil, "DATABASE_RETRY",
		`{"reason":"connection_recovery"}`,
		"", nil, nil, "Attempting database operation retry")
	return nil
}

// Helper functions

// isValidCoordinate validates latitude and longitude values
func isValidCoordinate(lat, lng float64) bool {
	return lat >= -90 && lat <= 90 && lng >= -180 && lng <= 180
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

// isValidStatusTransition validates if a status transition is allowed
func isValidStatusTransition(currentStatus, newStatus string) bool {
	validTransitions := map[string][]string{
		"PLANNED":     {"ACTIVE", "CANCELLED"},
		"ACTIVE":      {"IN_TRANSIT", "AT_PICKUP", "CANCELLED"},
		"AT_PICKUP":   {"IN_TRANSIT", "ACTIVE"},
		"IN_TRANSIT":  {"AT_DELIVERY", "DELAYED", "COMPLETED"},
		"AT_DELIVERY": {"COMPLETED", "IN_TRANSIT"},
		"DELAYED":     {"IN_TRANSIT", "AT_DELIVERY", "COMPLETED"},
		"COMPLETED":   {}, // Terminal state
		"CANCELLED":   {}, // Terminal state
	}

	allowedTransitions, exists := validTransitions[currentStatus]
	if !exists {
		return false
	}

	for _, allowed := range allowedTransitions {
		if allowed == newStatus {
			return true
		}
	}

	return false
}

// calculateCompletionPercent calculates completion percentage based on status
func calculateCompletionPercent(status string) float64 {
	statusPercent := map[string]float64{
		"PLANNED":     0.0,
		"ACTIVE":      10.0,
		"AT_PICKUP":   25.0,
		"IN_TRANSIT":  50.0,
		"AT_DELIVERY": 90.0,
		"COMPLETED":   100.0,
		"DELAYED":     50.0, // Same as IN_TRANSIT
		"CANCELLED":   0.0,
	}

	if percent, exists := statusPercent[status]; exists {
		return percent
	}
	return 0.0
}
