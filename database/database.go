package database

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"triplink/backend/models"
)

var DB *gorm.DB

func Connect() {
	dsn := "host=localhost user=wowcard password=password dbname=triplink port=5432 sslmode=disable"
	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		panic("Failed to connect to database!")
	}

	fmt.Println("Database connection successfully opened")

	database.AutoMigrate(&models.User{}, &models.Trip{}, &models.Load{}, &models.Message{}, &models.Transaction{})

	DB = database
}
