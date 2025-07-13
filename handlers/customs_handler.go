package handlers

import (
	"fmt"
	"time"
	"triplink/backend/database"
	"triplink/backend/models"

	"github.com/gofiber/fiber/v2"
)

// CreateCustomsDocument @Summary Create customs document
// @Description Create a new customs document for a load
// @Tags customs
// @Accept json
// @Produce json
// @Param document body models.CustomsDocument true "Customs document data"
// @Success 201 {object} models.CustomsDocument
// @Router /customs-documents [post]
func CreateCustomsDocument(c *fiber.Ctx) error {
	var document models.CustomsDocument

	if err := c.BodyParser(&document); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Cannot parse JSON",
		})
	}

	result := database.DB.Create(&document)
	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Could not create customs document",
		})
	}

	return c.Status(201).JSON(document)
}

// GetLoadCustomsDocuments @Summary Get customs documents for a load
// @Description Get all customs documents for a specific load
// @Tags customs
// @Produce json
// @Param load_id path int true "Load ID"
// @Success 200 {array} models.CustomsDocument
// @Router /loads/{load_id}/customs-documents [get]
func GetLoadCustomsDocuments(c *fiber.Ctx) error {
	loadID := c.Params("load_id")
	var documents []models.CustomsDocument

	result := database.DB.Where("load_id = ?", loadID).
		Order("created_at DESC").
		Find(&documents)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Could not fetch customs documents",
		})
	}

	return c.JSON(documents)
}

// GetCustomsDocument @Summary Get customs document by ID
// @Description Get a specific customs document by its ID
// @Tags customs
// @Produce JSON
// @Param id path int true "Document ID"
// @Success 200 {object} models.CustomsDocument
// @Router /customs-documents/{id} [get]
func GetCustomsDocument(c *fiber.Ctx) error {
	id := c.Params("id")
	var document models.CustomsDocument

	result := database.DB.First(&document, id)
	if result.Error != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Customs document not found",
		})
	}

	return c.JSON(document)
}

// UpdateCustomsDocument @Summary Update customs document
// @Description Update an existing customs document
// @Tags customs
// @Accept json
// @Produce json
// @Param id path int true "Document ID"
// @Param document body models.CustomsDocument true "Updated document data"
// @Success 200 {object} models.CustomsDocument
// @Router /customs-documents/{id} [put]
func UpdateCustomsDocument(c *fiber.Ctx) error {
	id := c.Params("id")
	var document models.CustomsDocument

	if err := database.DB.First(&document, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Customs document not found",
		})
	}

	if err := c.BodyParser(&document); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Cannot parse JSON",
		})
	}

	database.DB.Save(&document)
	return c.JSON(document)
}

// DeleteCustomsDocument @Summary Delete customs document
// @Description Delete a customs document
// @Tags customs
// @Param id path int true "Document ID"
// @Success 200 {object} map[string]string
// @Router /customs-documents/{id} [delete]
func DeleteCustomsDocument(c *fiber.Ctx) error {
	id := c.Params("id")
	var document models.CustomsDocument

	if err := database.DB.First(&document, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Customs document not found",
		})
	}

	database.DB.Delete(&document)
	return c.JSON(fiber.Map{
		"message": "Customs document deleted successfully",
	})
}

// GenerateCommercialInvoice @Summary Generate commercial invoice
// @Description Generate a commercial invoice for a load
// @Tags customs
// @Produce json
// @Param load_id path int true "Load ID"
// @Success 200 {object} models.CustomsDocument
// @Router /loads/{load_id}/commercial-invoice [post]
func GenerateCommercialInvoice(c *fiber.Ctx) error {
	loadID := c.Params("load_id")

	// Get load details
	var load models.Load
	if err := database.DB.First(&load, loadID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Load not found",
		})
	}

	// Check if commercial invoice already exists
	var existingDoc models.CustomsDocument
	if err := database.DB.Where("load_id = ? AND document_type = ?",
		loadID, "COMMERCIAL_INVOICE").First(&existingDoc).Error; err == nil {
		return c.JSON(existingDoc)
	}

	// Create commercial invoice document
	document := models.CustomsDocument{
		LoadID:           load.ID,
		DocumentType:     "COMMERCIAL_INVOICE",
		DocumentNumber:   generateDocumentNumber("CI", load.ID),
		IssuedDate:       time.Now(),
		IssuingAuthority: "System Generated",
	}

	result := database.DB.Create(&document)
	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Could not create commercial invoice",
		})
	}

	return c.JSON(document)
}

// GenerateBillOfLading @Summary Generate bill of lading
// @Description Generate a bill of lading for a load
// @Tags customs
// @Produce json
// @Param load_id path int true "Load ID"
// @Success 200 {object} models.CustomsDocument
// @Router /loads/{load_id}/bill-of-lading [post]
func GenerateBillOfLading(c *fiber.Ctx) error {
	loadID := c.Params("load_id")

	// Get load details
	var load models.Load
	if err := database.DB.First(&load, loadID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Load not found",
		})
	}

	// Check if BOL already exists
	var existingDoc models.CustomsDocument
	if err := database.DB.Where("load_id = ? AND document_type = ?",
		loadID, "BOL").First(&existingDoc).Error; err == nil {
		return c.JSON(existingDoc)
	}

	// Create bill of lading document
	document := models.CustomsDocument{
		LoadID:           load.ID,
		DocumentType:     "BOL",
		DocumentNumber:   generateDocumentNumber("BOL", load.ID),
		IssuedDate:       time.Now(),
		IssuingAuthority: "System Generated",
	}

	result := database.DB.Create(&document)
	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Could not create bill of lading",
		})
	}

	return c.JSON(document)
}

// GeneratePackingList @Summary Generate packing list
// @Description Generate a packing list for a load
// @Tags customs
// @Produce json
// @Param load_id path int true "Load ID"
// @Success 200 {object} models.CustomsDocument
// @Router /loads/{load_id}/packing-list [post]
func GeneratePackingList(c *fiber.Ctx) error {
	loadID := c.Params("load_id")

	// Get load details
	var load models.Load
	if err := database.DB.First(&load, loadID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Load not found",
		})
	}

	// Check if packing list already exists
	var existingDoc models.CustomsDocument
	if err := database.DB.Where("load_id = ? AND document_type = ?",
		loadID, "PACKING_LIST").First(&existingDoc).Error; err == nil {
		return c.JSON(existingDoc)
	}

	// Create packing list document
	document := models.CustomsDocument{
		LoadID:           load.ID,
		DocumentType:     "PACKING_LIST",
		DocumentNumber:   generateDocumentNumber("PL", load.ID),
		IssuedDate:       time.Now(),
		IssuingAuthority: "System Generated",
	}

	result := database.DB.Create(&document)
	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Could not create packing list",
		})
	}

	return c.JSON(document)
}

// GetTripCustomsSummary @Summary Get customs documents summary for trip
// @Description Get a summary of all customs documents for loads in a trip
// @Tags customs
// @Produce json
// @Param trip_id path int true "Trip ID"
// @Success 200 {object} map[string]interface{}
// @Router /trips/{trip_id}/customs-summary [get]
func GetTripCustomsSummary(c *fiber.Ctx) error {
	tripID := c.Params("trip_id")

	// Get all loads for the trip
	var loads []models.Load
	if err := database.DB.Where("trip_id = ?", tripID).Find(&loads).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Could not fetch loads",
		})
	}

	summary := map[string]interface{}{
		"trip_id":      tripID,
		"total_loads":  len(loads),
		"documents":    make(map[string]int),
		"missing_docs": []map[string]interface{}{},
	}

	documentCounts := map[string]int{
		"COMMERCIAL_INVOICE":    0,
		"PACKING_LIST":          0,
		"BOL":                   0,
		"CUSTOMS_DECLARATION":   0,
		"CERTIFICATE_OF_ORIGIN": 0,
	}

	for _, load := range loads {
		var documents []models.CustomsDocument
		database.DB.Where("load_id = ?", load.ID).Find(&documents)

		existingDocs := make(map[string]bool)
		for _, doc := range documents {
			documentCounts[doc.DocumentType]++
			existingDocs[doc.DocumentType] = true
		}

		// Check for missing required documents
		requiredDocs := []string{"COMMERCIAL_INVOICE", "PACKING_LIST", "BOL"}
		for _, docType := range requiredDocs {
			if !existingDocs[docType] {
				missingDoc := map[string]interface{}{
					"load_id":        load.ID,
					"load_reference": load.BookingReference,
					"document_type":  docType,
				}
				summary["missing_docs"] = append(summary["missing_docs"].([]map[string]interface{}), missingDoc)
			}
		}
	}

	summary["documents"] = documentCounts
	return c.JSON(summary)
}

// Helper function to generate document numbers
func generateDocumentNumber(prefix string, loadID uint) string {
	timestamp := time.Now().Unix()
	return fmt.Sprintf("%s-%d-%d", prefix, loadID, timestamp)
}
