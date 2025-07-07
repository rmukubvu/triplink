package handlers

import (
	"github.com/gofiber/fiber/v2"
	"triplink/backend/database"
	"triplink/backend/models"
)

func CreateTrip(c *fiber.Ctx) error {
	var trip models.Trip

	if err := c.BodyParser(&trip); err != nil {
		return err
	}

	userID := c.Locals("user_id").(float64)
	trip.UserID = uint(userID)

	database.DB.Create(&trip)

	return c.JSON(trip)
}

func GetTrips(c *fiber.Ctx) error {
	var trips []models.Trip

	database.DB.Find(&trips)

	return c.JSON(trips)
}

func GetTrip(c *fiber.Ctx) error {
	id := c.Params("id")
	var trip models.Trip

	database.DB.First(&trip, id)

	return c.JSON(trip)
}
