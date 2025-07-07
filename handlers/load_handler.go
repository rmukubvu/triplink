package handlers

import (
	"github.com/gofiber/fiber/v2"
	"triplink/backend/database"
	"triplink/backend/models"
)

func CreateLoad(c *fiber.Ctx) error {
	var load models.Load

	if err := c.BodyParser(&load); err != nil {
		return err
	}

	database.DB.Create(&load)

	return c.JSON(load)
}

func GetLoads(c *fiber.Ctx) error {
	var loads []models.Load

	database.DB.Find(&loads)

	return c.JSON(loads)
}

func GetLoad(c *fiber.Ctx) error {
	id := c.Params("id")
	var load models.Load

	database.DB.First(&load, id)

	return c.JSON(load)
}
