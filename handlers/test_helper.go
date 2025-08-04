package handlers

import (
	"fmt"
	"os"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"triplink/backend/database"
	"triplink/backend/models"
)

var testDB *gorm.DB

func TestMain(m *testing.M) {
	// Initialize the database connection for tests
	var err error
	testDB, err = gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Migrate the schema
	fmt.Println("Migrating database schema...")
	err = testDB.AutoMigrate(&models.User{}, &models.Trip{}, &models.Load{}, &models.Vehicle{}, &models.Quote{}, &models.Review{}, &models.Transaction{}, &models.Notification{}, &models.CustomsDocument{}, &models.Manifest{}, &models.Message{}, &models.TrackingRecord{}, &models.TrackingEvent{}, &models.TrackingStatus{}, &models.NotificationToken{}, &models.NotificationPreferences{}, &models.NotificationDelivery{})
	if err != nil {
		fmt.Printf("Error migrating database: %v\n", err)
		panic("failed to migrate database")
	}
	fmt.Println("Database schema migrated successfully.")

	code := m.Run()
	os.Exit(code)
}

func clearTestDB(db *gorm.DB) {
	fmt.Println("Clearing test database...")
	if db != nil {
		db.Exec("DELETE FROM users")
		db.Exec("DELETE FROM trips")
		db.Exec("DELETE FROM loads")
		db.Exec("DELETE FROM vehicles")
		db.Exec("DELETE FROM quotes")
		db.Exec("DELETE FROM reviews")
		db.Exec("DELETE FROM transactions")
		db.Exec("DELETE FROM notifications")
		db.Exec("DELETE FROM customs_documents")
		db.Exec("DELETE FROM manifests")
		db.Exec("DELETE FROM messages")
		db.Exec("DELETE FROM tracking_records")
		db.Exec("DELETE FROM tracking_events")
		db.Exec("DELETE FROM tracking_statuses")
		db.Exec("DELETE FROM notification_tokens")
		db.Exec("DELETE FROM notification_preferences")
		db.Exec("DELETE FROM notification_deliveries")
	}
	fmt.Println("Test database cleared.")
}

func seedTestDB(db *gorm.DB) {
	fmt.Println("Seeding test database...")
	// Create a test user
	user := models.User{
		Email: "test@example.com",
		Phone: "+1234567890",
		Password: "password",
		Role: "SHIPPER",
	}
	result := db.Create(&user)
	if result.Error != nil {
		fmt.Printf("Error creating user: %v\n", result.Error)
	}

	// Create a test trip
	trip := models.Trip{
		UserID: user.ID,
		OriginAddress: "123 Main St",
		DestinationAddress: "456 Oak Ave",
		Status: "PLANNED",
		DepartureDate: time.Now(),
		EstimatedArrival: time.Now().Add(time.Hour * 24),
	}
	result = db.Create(&trip)
	if result.Error != nil {
		fmt.Printf("Error creating trip: %v\n", result.Error)
	}

	// Create a test load
	load := models.Load{
		ShipperID: user.ID,
		TripID: trip.ID,
		BookingReference: "TEST-LOAD-001",
		Status: "BOOKED",
		RequestedPickupDate: time.Now(),
		RequestedDeliveryDate: time.Now().Add(time.Hour * 48),
	}
	result = db.Create(&load)
	if result.Error != nil {
		fmt.Printf("Error creating load: %v\n", result.Error)
	}

	// Create a test vehicle
	vehicle := models.Vehicle{
		UserID: user.ID,
		VehicleType: "TRUCK",
		Make: "Volvo",
		Model: "VNL 860",
		IsActive: true,
		LicensePlate: "TEST123",
		VIN: "VINTEST123",
	}
	result = db.Create(&vehicle)
	if result.Error != nil {
		fmt.Printf("Error creating vehicle: %v\n", result.Error)
	}

	// Create a test quote
	quote := models.Quote{
		LoadID: load.ID,
		CarrierID: user.ID,
		QuoteAmount: 100.00,
		Currency: "USD",
		ValidUntil: time.Now().Add(time.Hour * 24),
		PickupDate: time.Now().Add(time.Hour * 2),
		DeliveryDate: time.Now().Add(time.Hour * 26),
		Status: "PENDING",
	}
	result = db.Create(&quote)
	if result.Error != nil {
		fmt.Printf("Error creating quote: %v\n", result.Error)
	}

	// Create a test review
	review := models.Review{
		LoadID: load.ID,
		ReviewerID: user.ID,
		RevieweeID: user.ID,
		Rating: 5,
		Comment: "Great service!",
	}
	result = db.Create(&review)
	if result.Error != nil {
		fmt.Printf("Error creating review: %v\n", result.Error)
	}

	// Create a test transaction
	processedAt := time.Now()
	transaction := models.Transaction{
		LoadID: load.ID,
		PayerID: user.ID,
		PayeeID: user.ID,
		Amount: 100.00,
		Currency: "USD",
		PaymentMethod: "CARD",
		PaymentGateway: "STRIPE",
		Status: "COMPLETED",
		ProcessedAt: &processedAt,
	}
	result = db.Create(&transaction)
	if result.Error != nil {
		fmt.Printf("Error creating transaction: %v\n", result.Error)
	}

	// Create a test notification
	notification := models.Notification{
		UserID: user.ID,
		Title: "Test Notification",
		Message: "This is a test notification",
		Type: "GENERAL",
		IsRead: false,
	}
	result = db.Create(&notification)
	if result.Error != nil {
		fmt.Printf("Error creating notification: %v\n", result.Error)
	}

	// Create a test customs document
	customsDocument := models.CustomsDocument{
		LoadID: load.ID,
		DocumentType: "COMMERCIAL_INVOICE",
		DocumentNumber: "CI-001",
		IssuedDate: time.Now(),
	}
	result = db.Create(&customsDocument)
	if result.Error != nil {
		fmt.Printf("Error creating customs document: %v\n", result.Error)
	}

	// Create a test manifest
	manifest := models.Manifest{
		TripID: trip.ID,
		ManifestNumber: "MAN-001",
		GeneratedAt: time.Now(),
	}
	result = db.Create(&manifest)
	if result.Error != nil {
		fmt.Printf("Error creating manifest: %v\n", result.Error)
	}

	// Create a test message
	message := models.Message{
		SenderID: user.ID,
		ReceiverID: user.ID,
		Content: "Hello, world!",
	}
	result = db.Create(&message)
	if result.Error != nil {
		fmt.Printf("Error creating message: %v\n", result.Error)
	}

	// Create a test tracking record
	trackingRecordTimestamp := time.Now()
	trackingRecord := models.TrackingRecord{
		TripID: trip.ID,
		Latitude: 1.0,
		Longitude: 1.0,
		Timestamp: trackingRecordTimestamp,
	}
	result = db.Create(&trackingRecord)
	if result.Error != nil {
		fmt.Printf("Error creating tracking record: %v\n", result.Error)
	}

	// Create a test tracking event
	trackingEventTimestamp := time.Now()
	trackingEvent := models.TrackingEvent{
		TripID: trip.ID,
		EventType: "LOCATION_UPDATE",
		Timestamp: trackingEventTimestamp,
	}
	result = db.Create(&trackingEvent)
	if result.Error != nil {
		fmt.Printf("Error creating tracking event: %v\n", result.Error)
	}

	// Create a test tracking status
	trackingStatusTimestamp := time.Now()
	trackingStatus := models.TrackingStatus{
		TripID: trip.ID,
		CurrentStatus: "IN_TRANSIT",
		StatusChangedAt: trackingStatusTimestamp,
	}
	result = db.Create(&trackingStatus)
	if result.Error != nil {
		fmt.Printf("Error creating tracking status: %v\n", result.Error)
	}

	// Create a test notification token
	notificationTokenLastUsed := time.Now()
	notificationToken := models.NotificationToken{
		UserID: user.ID,
		Token: "test_token",
		DeviceType: "ios",
		LastUsed: notificationTokenLastUsed,
	}
	result = db.Create(&notificationToken)
	if result.Error != nil {
		fmt.Printf("Error creating notification token: %v\n", result.Error)
	}
}
