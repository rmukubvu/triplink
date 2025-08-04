package models

import (
	"time"

	_ "gorm.io/gorm"
)

type BaseModel struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `gorm:"index" json:"deleted_at,omitempty"`
}

type User struct {
	BaseModel
	Email           string     `gorm:"unique" json:"email"`
	Phone           string     `gorm:"unique" json:"phone"`
	Password        string     `json:"-"`
	Role            string     `json:"role"` // CARRIER, SHIPPER, ADMIN
	FirstName       string     `json:"first_name"`
	LastName        string     `json:"last_name"`
	CompanyName     string     `json:"company_name"`
	ProfileImage    string     `json:"profile_image"`
	IsVerified      bool       `gorm:"default:false" json:"is_verified"`
	Rating          float64    `gorm:"default:0" json:"rating"`
	TotalReviews    int        `gorm:"default:0" json:"total_reviews"`
	DriverLicense   string     `json:"driver_license"`
	LicenseNumber   string     `json:"license_number"`
	LicenseExpiry   *time.Time `json:"license_expiry"`
	BusinessLicense string     `json:"business_license"`
	TaxID           string     `json:"tax_id"`
	Address         string     `json:"address"`
	City            string     `json:"city"`
	State           string     `json:"state"`
	Country         string     `json:"country"`
	PostalCode      string     `json:"postal_code"`
	Vehicles        []Vehicle  `json:"vehicles,omitempty" gorm:"foreignKey:UserID"`
}

type Trip struct {
	BaseModel
	UserID              uint       `json:"user_id"`
	VehicleID           uint       `json:"vehicle_id"`
	OriginAddress       string     `json:"origin_address"`
	OriginCity          string     `json:"origin_city"`
	OriginState         string     `json:"origin_state"`
	OriginCountry       string     `json:"origin_country"`
	OriginLat           float64    `json:"origin_lat"`
	OriginLng           float64    `json:"origin_lng"`
	DestinationAddress  string     `json:"destination_address"`
	DestinationCity     string     `json:"destination_city"`
	DestinationState    string     `json:"destination_state"`
	DestinationCountry  string     `json:"destination_country"`
	DestinationLat      float64    `json:"destination_lat"`
	DestinationLng      float64    `json:"destination_lng"`
	DepartureDate       time.Time  `json:"departure_date"`
	EstimatedArrival    time.Time  `json:"estimated_arrival"`
	ActualDeparture     *time.Time `json:"actual_departure"`
	ActualArrival       *time.Time `json:"actual_arrival"`
	TotalCapacityWeight float64    `json:"total_capacity_weight"`
	TotalCapacityVolume float64    `json:"total_capacity_volume"`
	UsedWeight          float64    `gorm:"default:0" json:"used_weight"`
	UsedVolume          float64    `gorm:"default:0" json:"used_volume"`
	BasePrice           float64    `json:"base_price"`
	PricePerKg          float64    `json:"price_per_kg"`
	PricePerCubicMeter  float64    `json:"price_per_cubic_meter"`
	Status              string     `json:"status"` // PLANNED, ACTIVE, IN_TRANSIT, COMPLETED, CANCELLED
	Notes               string     `json:"notes"`
	IsPublic            bool       `gorm:"default:true" json:"is_public"`
	// Tracking fields
	CurrentLatitude    *float64   `json:"current_latitude"`
	CurrentLongitude   *float64   `json:"current_longitude"`
	LastLocationUpdate *time.Time `json:"last_location_update"`
	TrackingEnabled    bool       `gorm:"default:true" json:"tracking_enabled"`
	// Relationships
	Loads           []Load           `json:"loads,omitempty" gorm:"foreignKey:TripID"`
	Manifest        *Manifest        `json:"manifest,omitempty" gorm:"foreignKey:TripID"`
	TrackingRecords []TrackingRecord `json:"tracking_records,omitempty" gorm:"foreignKey:TripID"`
	TrackingStatus  *TrackingStatus  `json:"tracking_status,omitempty" gorm:"foreignKey:TripID"`
	TrackingEvents  []TrackingEvent  `json:"tracking_events,omitempty" gorm:"foreignKey:TripID"`
}

type Load struct {
	BaseModel
	TripID                uint              `json:"trip_id"`
	ShipperID             uint              `json:"shipper_id"`
	BookingReference      string            `gorm:"unique" json:"booking_reference"`
	Description           string            `json:"description"`
	Category              string            `json:"category"` // ELECTRONICS, MACHINERY, FOOD, TEXTILES, etc.
	HSCode                string            `json:"hs_code"`
	Quantity              int               `json:"quantity"`
	Weight                float64           `json:"weight"`
	Length                float64           `json:"length"`
	Width                 float64           `json:"width"`
	Height                float64           `json:"height"`
	Volume                float64           `json:"volume"`
	Value                 float64           `json:"value"`
	Currency              string            `gorm:"default:USD" json:"currency"`
	PickupAddress         string            `json:"pickup_address"`
	PickupCity            string            `json:"pickup_city"`
	PickupState           string            `json:"pickup_state"`
	PickupCountry         string            `json:"pickup_country"`
	PickupLat             float64           `json:"pickup_lat"`
	PickupLng             float64           `json:"pickup_lng"`
	DeliveryAddress       string            `json:"delivery_address"`
	DeliveryCity          string            `json:"delivery_city"`
	DeliveryState         string            `json:"delivery_state"`
	DeliveryCountry       string            `json:"delivery_country"`
	DeliveryLat           float64           `json:"delivery_lat"`
	DeliveryLng           float64           `json:"delivery_lng"`
	RequestedPickupDate   time.Time         `json:"requested_pickup_date"`
	RequestedDeliveryDate time.Time         `json:"requested_delivery_date"`
	ActualPickupDate      *time.Time        `json:"actual_pickup_date"`
	ActualDeliveryDate    *time.Time        `json:"actual_delivery_date"`
	SpecialInstructions   string            `json:"special_instructions"`
	IsFragile             bool              `gorm:"default:false" json:"is_fragile"`
	IsHazmat              bool              `gorm:"default:false" json:"is_hazmat"`
	RequiresRefrigeration bool              `gorm:"default:false" json:"requires_refrigeration"`
	InsuranceRequired     bool              `gorm:"default:false" json:"insurance_required"`
	InsuranceValue        float64           `json:"insurance_value"`
	AgreedPrice           float64           `json:"agreed_price"`
	Status                string            `json:"status"` // QUOTE_REQUESTED, QUOTED, BOOKED, PICKED_UP, IN_TRANSIT, DELIVERED, CANCELLED
	PickupProof           string            `json:"pickup_proof"`
	DeliveryProof         string            `json:"delivery_proof"`
	CustomsDocuments      []CustomsDocument `json:"customs_documents,omitempty" gorm:"foreignKey:LoadID"`
	Quotes                []Quote           `json:"quotes,omitempty" gorm:"foreignKey:LoadID"`
	// Tracking relationships
	TrackingRecords []TrackingRecord `json:"tracking_records,omitempty" gorm:"foreignKey:LoadID"`
	TrackingStatus  *TrackingStatus  `json:"tracking_status,omitempty" gorm:"foreignKey:LoadID"`
	TrackingEvents  []TrackingEvent  `json:"tracking_events,omitempty" gorm:"foreignKey:LoadID"`
}

type Message struct {
	BaseModel
	SenderID   uint   `json:"sender_id"`
	ReceiverID uint   `json:"receiver_id"`
	Content    string `json:"content"`
}

type Vehicle struct {
	BaseModel
	UserID             uint       `json:"user_id"`
	Make               string     `json:"make"`
	Model              string     `json:"model"`
	Year               int        `json:"year"`
	LicensePlate       string     `gorm:"unique" json:"license_plate"`
	VIN                string     `gorm:"unique" json:"vin"`
	VehicleType        string     `json:"vehicle_type"` // FLATBED, REEFER, DRY_VAN, TANKER, BOX_TRUCK
	LoadCapacityKg     float64    `json:"load_capacity_kg"`
	LoadCapacityM3     float64    `json:"load_capacity_m3"`
	MaxLength          float64    `json:"max_length"`
	MaxWidth           float64    `json:"max_width"`
	MaxHeight          float64    `json:"max_height"`
	HasLiftgate        bool       `gorm:"default:false" json:"has_liftgate"`
	HasStraps          bool       `gorm:"default:false" json:"has_straps"`
	IsRefrigerated     bool       `gorm:"default:false" json:"is_refrigerated"`
	IsHazmatCertified  bool       `gorm:"default:false" json:"is_hazmat_certified"`
	IsFoodGrade        bool       `gorm:"default:false" json:"is_food_grade"`
	InsuranceExpiry    *time.Time `json:"insurance_expiry"`
	RegistrationExpiry *time.Time `json:"registration_expiry"`
	InspectionExpiry   *time.Time `json:"inspection_expiry"`
	IsActive           bool       `gorm:"default:true" json:"is_active"`
	Images             []string   `gorm:"type:text[]" json:"images"`
}

type Manifest struct {
	BaseModel
	TripID             uint      `json:"trip_id"`
	ManifestNumber     string    `gorm:"unique" json:"manifest_number"`
	TotalWeight        float64   `json:"total_weight"`
	TotalVolume        float64   `json:"total_volume"`
	TotalValue         float64   `json:"total_value"`
	LoadCount          int       `json:"load_count"`
	OriginCountry      string    `json:"origin_country"`
	DestinationCountry string    `json:"destination_country"`
	DocumentURL        string    `json:"document_url"`
	GeneratedAt        time.Time `json:"generated_at"`
}

type CustomsDocument struct {
	BaseModel
	LoadID           uint       `json:"load_id"`
	DocumentType     string     `json:"document_type"` // COMMERCIAL_INVOICE, PACKING_LIST, BOL, CUSTOMS_DECLARATION, CERTIFICATE_OF_ORIGIN
	DocumentNumber   string     `json:"document_number"`
	DocumentURL      string     `json:"document_url"`
	IssuedDate       time.Time  `json:"issued_date"`
	ExpiryDate       *time.Time `json:"expiry_date"`
	IssuingAuthority string     `json:"issuing_authority"`
}

type Quote struct {
	BaseModel
	LoadID       uint       `json:"load_id"`
	CarrierID    uint       `json:"carrier_id"`
	QuoteAmount  float64    `json:"quote_amount"`
	Currency     string     `gorm:"default:USD" json:"currency"`
	ValidUntil   time.Time  `json:"valid_until"`
	PickupDate   time.Time  `json:"pickup_date"`
	DeliveryDate time.Time  `json:"delivery_date"`
	Notes        string     `json:"notes"`
	Status       string     `json:"status"` // PENDING, ACCEPTED, REJECTED, EXPIRED
	AcceptedAt   *time.Time `json:"accepted_at"`
}

type Review struct {
	BaseModel
	ReviewerID uint   `json:"reviewer_id"`
	RevieweeID uint   `json:"reviewee_id"`
	LoadID     uint   `json:"load_id"`
	Rating     int    `json:"rating"` // 1-5 stars
	Comment    string `json:"comment"`
	ReviewType string `json:"review_type"` // CARRIER_TO_SHIPPER, SHIPPER_TO_CARRIER
}

type Notification struct {
	BaseModel
	UserID    uint   `json:"user_id"`
	Title     string `json:"title"`
	Message   string `json:"message"`
	Type      string `json:"type"` // QUOTE_RECEIVED, LOAD_BOOKED, PICKUP_SCHEDULED, etc.
	IsRead    bool   `gorm:"default:false" json:"is_read"`
	RelatedID uint   `json:"related_id"` // ID of related load, trip, etc.
	// Delivery tracking
	Deliveries []NotificationDelivery `json:"deliveries,omitempty" gorm:"foreignKey:NotificationID"`
}

// NotificationToken stores device tokens for push notifications
type NotificationToken struct {
	BaseModel
	UserID     uint      `json:"user_id"`
	Token      string    `json:"token" gorm:"uniqueIndex:idx_token_user"`
	DeviceType string    `json:"device_type"` // ios, android
	LastUsed   time.Time `json:"last_used"`
}

// NotificationDelivery tracks notification delivery attempts
type NotificationDelivery struct {
	BaseModel
	NotificationID uint      `json:"notification_id"`
	UserID         uint      `json:"user_id"`
	Success        bool      `json:"success"`
	Provider       string    `json:"provider"`
	SentAt         time.Time `json:"sent_at"`
	Error          string    `json:"error,omitempty"`
}

// NotificationPreferences stores user preferences for notifications
type NotificationPreferences struct {
	BaseModel
	UserID          uint `json:"user_id" gorm:"uniqueIndex"`
	TripDeparture   bool `json:"trip_departure" gorm:"default:true"`
	TripArrival     bool `json:"trip_arrival" gorm:"default:true"`
	Delays          bool `json:"delays" gorm:"default:true"`
	ETAUpdates      bool `json:"eta_updates" gorm:"default:true"`
	LoadStatus      bool `json:"load_status" gorm:"default:true"`
	LocationUpdates bool `json:"location_updates" gorm:"default:false"`
	EmailEnabled    bool `json:"email_enabled" gorm:"default:true"`
	PushEnabled     bool `json:"push_enabled" gorm:"default:true"`
}

type Transaction struct {
	BaseModel
	LoadID         uint       `json:"load_id"`
	PayerID        uint       `json:"payer_id"`
	PayeeID        uint       `json:"payee_id"`
	Amount         float64    `json:"amount"`
	Currency       string     `gorm:"default:USD" json:"currency"`
	PlatformFee    float64    `json:"platform_fee"`
	PaymentMethod  string     `json:"payment_method"`  // CARD, BANK_TRANSFER, WALLET
	PaymentGateway string     `json:"payment_gateway"` // STRIPE, PAYPAL, PAYSTACK
	GatewayTxnID   string     `json:"gateway_txn_id"`
	Status         string     `json:"status"` // PENDING, COMPLETED, FAILED, REFUNDED
	ProcessedAt    *time.Time `json:"processed_at"`
	FailureReason  string     `json:"failure_reason"`
}

// Tracking Models

type TrackingRecord struct {
	BaseModel
	TripID    uint      `json:"trip_id"`
	LoadID    *uint     `json:"load_id,omitempty"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Altitude  *float64  `json:"altitude,omitempty"`
	Speed     *float64  `json:"speed,omitempty"`
	Heading   *float64  `json:"heading,omitempty"`
	Accuracy  *float64  `json:"accuracy,omitempty"`
	Timestamp time.Time `json:"timestamp"`
	Source    string    `json:"source"` // GPS, MANUAL, ESTIMATED
	Status    string    `json:"status"` // ACTIVE, INACTIVE
}

type TrackingStatus struct {
	BaseModel
	TripID            uint       `json:"trip_id"`
	LoadID            *uint      `json:"load_id,omitempty"`
	CurrentStatus     string     `json:"current_status"`
	PreviousStatus    string     `json:"previous_status"`
	StatusChangedAt   time.Time  `json:"status_changed_at"`
	EstimatedArrival  *time.Time `json:"estimated_arrival"`
	DelayMinutes      *int       `json:"delay_minutes"`
	DelayReason       string     `json:"delay_reason"`
	NextMilestone     string     `json:"next_milestone"`
	CompletionPercent float64    `json:"completion_percent"`
}

type TrackingEvent struct {
	BaseModel
	TripID      uint      `json:"trip_id"`
	LoadID      *uint     `json:"load_id,omitempty"`
	EventType   string    `json:"event_type"` // DEPARTURE, ARRIVAL, DELAY, MILESTONE
	EventData   string    `json:"event_data"` // JSON data specific to event
	Location    string    `json:"location"`
	Latitude    *float64  `json:"latitude"`
	Longitude   *float64  `json:"longitude"`
	Timestamp   time.Time `json:"timestamp"`
	Description string    `json:"description"`
}
