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
	Email    string `gorm:"unique" json:"email"`
	Phone    string `gorm:"unique" json:"phone"`
	Password string `json:"-"`    // Exclude from JSON output
	Role     string `json:"role"` // CARRIER, SHIPPER, ADMIN
}

type Trip struct {
	BaseModel
	UserID         uint    `json:"user_id"`
	Origin         string  `json:"origin"`
	Destination    string  `json:"destination"`
	Departure      string  `json:"departure"`
	AvailableSpace float64 `json:"available_space"`
	Status         string  `json:"status"` // PENDING, IN_PROGRESS, COMPLETED
	Loads          []Load  `json:"loads,omitempty" gorm:"foreignKey:TripID"`
}

type Load struct {
	BaseModel
	TripID      uint    `json:"trip_id"`
	ShipperID   uint    `json:"shipper_id"`
	Description string  `json:"description"`
	Weight      float64 `json:"weight"`
	Value       float64 `json:"value"`
	Status      string  `json:"status"` // BOOKED, PICKED_UP, IN_TRANSIT, DELIVERED
}

type Message struct {
	BaseModel
	SenderID   uint   `json:"sender_id"`
	ReceiverID uint   `json:"receiver_id"`
	Content    string `json:"content"`
}

type Transaction struct {
	BaseModel
	LoadID uint    `json:"load_id"`
	Amount float64 `json:"amount"`
	Status string  `json:"status"` // PENDING, COMPLETED
}
