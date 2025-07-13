package handlers

import (
	"time"
	"triplink/backend/database"
	"triplink/backend/models"

	"github.com/gofiber/fiber/v2"
)

// CreateQuote @Summary Create a quote
// @Description Create a new quote for a load
// @Tags quotes
// @Accept json
// @Produce json
// @Param quote body models.Quote true "Quote data"
// @Success 201 {object} models.Quote
// @Router /quotes [post]
func CreateQuote(c *fiber.Ctx) error {
	var quote models.Quote

	if err := c.BodyParser(&quote); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Cannot parse JSON",
		})
	}

	quote.Status = "PENDING"

	result := database.DB.Create(&quote)
	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Could not create quote",
		})
	}

	return c.Status(201).JSON(quote)
}

// GetLoadQuotes @Summary Get quotes for a load
// @Description Get all quotes for a specific load
// @Tags quotes
// @Produce json
// @Param load_id path int true "Load ID"
// @Success 200 {array} models.Quote
// @Router /loads/{load_id}/quotes [get]
func GetLoadQuotes(c *fiber.Ctx) error {
	loadID := c.Params("load_id")
	var quotes []models.Quote

	result := database.DB.Where("load_id = ?", loadID).
		Preload("Carrier").
		Order("created_at DESC").
		Find(&quotes)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Could not fetch quotes",
		})
	}

	return c.JSON(quotes)
}

// GetCarrierQuotes @Summary Get quotes by carrier
// @Description Get all quotes created by a specific carrier
// @Tags quotes
// @Produce json
// @Param carrier_id path int true "Carrier ID"
// @Success 200 {array} models.Quote
// @Router /carriers/{carrier_id}/quotes [get]
func GetCarrierQuotes(c *fiber.Ctx) error {
	carrierID := c.Params("carrier_id")
	var quotes []models.Quote

	result := database.DB.Where("carrier_id = ?", carrierID).
		Preload("Load").
		Order("created_at DESC").
		Find(&quotes)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Could not fetch quotes",
		})
	}

	return c.JSON(quotes)
}

// AcceptQuote @Summary Accept a quote
// @Description Accept a quote and create a booking
// @Tags quotes
// @Param id path int true "Quote ID"
// @Success 200 {object} models.Quote
// @Router /quotes/{id}/accept [post]
func AcceptQuote(c *fiber.Ctx) error {
	id := c.Params("id")
	var quote models.Quote

	if err := database.DB.First(&quote, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Quote not found",
		})
	}

	if quote.Status != "PENDING" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Quote is no longer available",
		})
	}

	if time.Now().After(quote.ValidUntil) {
		return c.Status(400).JSON(fiber.Map{
			"error": "Quote has expired",
		})
	}

	// Start transaction
	tx := database.DB.Begin()

	// Update quote status
	now := time.Now()
	quote.Status = "ACCEPTED"
	quote.AcceptedAt = &now
	if err := tx.Save(&quote).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{
			"error": "Could not accept quote",
		})
	}

	// Update load status and agreed price
	var load models.Load
	if err := tx.First(&load, quote.LoadID).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{
			"error": "Load not found",
		})
	}

	load.Status = "BOOKED"
	load.AgreedPrice = quote.QuoteAmount
	if err := tx.Save(&load).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{
			"error": "Could not update load",
		})
	}

	// Reject other pending quotes for this load
	if err := tx.Model(&models.Quote{}).
		Where("load_id = ? AND id != ? AND status = ?", quote.LoadID, quote.ID, "PENDING").
		Update("status", "REJECTED").Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{
			"error": "Could not reject other quotes",
		})
	}

	tx.Commit()
	return c.JSON(quote)
}

// RejectQuote @Summary Reject a quote
// @Description Reject a quote
// @Tags quotes
// @Param id path int true "Quote ID"
// @Success 200 {object} models.Quote
// @Router /quotes/{id}/reject [post]
func RejectQuote(c *fiber.Ctx) error {
	id := c.Params("id")
	var quote models.Quote

	if err := database.DB.First(&quote, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Quote not found",
		})
	}

	if quote.Status != "PENDING" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Quote cannot be rejected",
		})
	}

	quote.Status = "REJECTED"
	database.DB.Save(&quote)

	return c.JSON(quote)
}

// UpdateQuote @Summary Update quote
// @Description Update an existing quote (only if pending)
// @Tags quotes
// @Accept json
// @Produce json
// @Param id path int true "Quote ID"
// @Param quote body models.Quote true "Updated quote data"
// @Success 200 {object} models.Quote
// @Router /quotes/{id} [put]
func UpdateQuote(c *fiber.Ctx) error {
	id := c.Params("id")
	var quote models.Quote

	if err := database.DB.First(&quote, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Quote not found",
		})
	}

	if quote.Status != "PENDING" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Cannot update non-pending quote",
		})
	}

	var updateData models.Quote
	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Cannot parse JSON",
		})
	}

	// Only allow updating certain fields
	quote.QuoteAmount = updateData.QuoteAmount
	quote.Currency = updateData.Currency
	quote.ValidUntil = updateData.ValidUntil
	quote.PickupDate = updateData.PickupDate
	quote.DeliveryDate = updateData.DeliveryDate
	quote.Notes = updateData.Notes

	database.DB.Save(&quote)
	return c.JSON(quote)
}
