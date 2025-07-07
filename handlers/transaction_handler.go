package handlers

import (
	"github.com/gofiber/fiber/v2"
	"triplink/backend/database"
	"triplink/backend/models"
)

func CreateTransaction(c *fiber.Ctx) error {
	var transaction models.Transaction

	if err := c.BodyParser(&transaction); err != nil {
		return err
	}

	database.DB.Create(&transaction)

	return c.JSON(transaction)
}

func GetTransactions(c *fiber.Ctx) error {
	var transactions []models.Transaction

	database.DB.Find(&transactions)

	return c.JSON(transactions)
}
