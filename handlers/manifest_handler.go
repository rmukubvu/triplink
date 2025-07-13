package handlers

import (
	"fmt"
	"time"
	"triplink/backend/database"
	"triplink/backend/models"

	"github.com/gofiber/fiber/v2"
)

// GenerateManifest @Summary Generate manifest for a trip
// @Description Generate a consolidated manifest for all loads in a trip
// @Tags manifests
// @Produce json
// @Param trip_id path int true "Trip ID"
// @Success 200 {object} models.Manifest
// @Router /trips/{trip_id}/manifest [post]
func GenerateManifest(c *fiber.Ctx) error {
	tripID := c.Params("trip_id")

	// Get trip with loads
	var trip models.Trip
	if err := database.DB.Preload("Loads").First(&trip, tripID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Trip not found",
		})
	}

	if len(trip.Loads) == 0 {
		return c.Status(400).JSON(fiber.Map{
			"error": "No loads found for this trip",
		})
	}

	// Check if manifest already exists
	var existingManifest models.Manifest
	if err := database.DB.Where("trip_id = ?", tripID).First(&existingManifest).Error; err == nil {
		return c.JSON(existingManifest)
	}

	// Calculate totals
	var totalWeight, totalVolume, totalValue float64
	loadCount := 0
	originCountry := ""
	destinationCountry := ""

	for _, load := range trip.Loads {
		if load.Status == "BOOKED" || load.Status == "PICKED_UP" || load.Status == "IN_TRANSIT" {
			totalWeight += load.Weight
			totalVolume += load.Volume
			totalValue += load.Value
			loadCount++

			if originCountry == "" {
				originCountry = load.PickupCountry
			}
			if destinationCountry == "" {
				destinationCountry = load.DeliveryCountry
			}
		}
	}

	// Generate manifest number
	manifestNumber := fmt.Sprintf("MAN-%d-%d", trip.ID, time.Now().Unix())

	// Create manifest
	manifest := models.Manifest{
		TripID:             trip.ID,
		ManifestNumber:     manifestNumber,
		TotalWeight:        totalWeight,
		TotalVolume:        totalVolume,
		TotalValue:         totalValue,
		LoadCount:          loadCount,
		OriginCountry:      originCountry,
		DestinationCountry: destinationCountry,
		GeneratedAt:        time.Now(),
	}

	result := database.DB.Create(&manifest)
	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Could not create manifest",
		})
	}

	return c.JSON(manifest)
}

// GetTripManifest @Summary Get manifest by trip ID
// @Description Get the manifest for a specific trip
// @Tags manifests
// @Produce json
// @Param trip_id path int true "Trip ID"
// @Success 200 {object} models.Manifest
// @Router /trips/{trip_id}/manifest [get]
func GetTripManifest(c *fiber.Ctx) error {
	tripID := c.Params("trip_id")
	var manifest models.Manifest

	result := database.DB.Where("trip_id = ?", tripID).First(&manifest)
	if result.Error != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Manifest not found",
		})
	}

	return c.JSON(manifest)
}

// GetManifest @Summary Get manifest by ID
// @Description Get a specific manifest by its ID
// @Tags manifests
// @Produce json
// @Param id path int true "Manifest ID"
// @Success 200 {object} models.Manifest
// @Router /manifests/{id} [get]
func GetManifest(c *fiber.Ctx) error {
	id := c.Params("id")
	var manifest models.Manifest

	result := database.DB.First(&manifest, id)
	if result.Error != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Manifest not found",
		})
	}

	return c.JSON(manifest)
}

// GetDetailedManifest @Summary Get detailed manifest data
// @Description Get manifest with detailed load information
// @Tags manifests
// @Produce json
// @Param id path int true "Manifest ID"
// @Success 200 {object} map[string]interface{}
// @Router /manifests/{id}/detailed [get]
func GetDetailedManifest(c *fiber.Ctx) error {
	id := c.Params("id")
	var manifest models.Manifest

	if err := database.DB.First(&manifest, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Manifest not found",
		})
	}

	// Get trip details
	var trip models.Trip
	if err := database.DB.First(&trip, manifest.TripID).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Could not fetch trip details",
		})
	}

	// Get vehicle and user separately
	var vehicle models.Vehicle
	var user models.User
	database.DB.First(&vehicle, trip.VehicleID)
	database.DB.First(&user, trip.UserID)

	// Get loads with shipper info and customs documents
	var loads []models.Load
	if err := database.DB.Preload("CustomsDocuments").
		Where("trip_id = ? AND (status = ? OR status = ? OR status = ?)",
			manifest.TripID, "BOOKED", "PICKED_UP", "IN_TRANSIT").
		Find(&loads).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Could not fetch load details",
		})
	}

	// Build detailed manifest response
	detailedManifest := map[string]interface{}{
		"manifest": manifest,
		"trip": map[string]interface{}{
			"id":                  trip.ID,
			"origin_address":      trip.OriginAddress,
			"destination_address": trip.DestinationAddress,
			"departure_date":      trip.DepartureDate,
			"estimated_arrival":   trip.EstimatedArrival,
			"vehicle":             vehicle,
			"carrier":             user,
		},
		"loads": loads,
		"summary": map[string]interface{}{
			"total_loads":  len(loads),
			"total_weight": manifest.TotalWeight,
			"total_volume": manifest.TotalVolume,
			"total_value":  manifest.TotalValue,
		},
	}

	return c.JSON(detailedManifest)
}

// UpdateManifestDocument @Summary Update manifest document URL
// @Description Update the document URL for a manifest (after PDF generation)
// @Tags manifests
// @Accept json
// @Produce json
// @Param id path int true "Manifest ID"
// @Param data body map[string]string true "Document URL"
// @Success 200 {object} models.Manifest
// @Router /manifests/{id}/document [put]
func UpdateManifestDocument(c *fiber.Ctx) error {
	id := c.Params("id")
	var manifest models.Manifest

	if err := database.DB.First(&manifest, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Manifest not found",
		})
	}

	var data map[string]string
	if err := c.BodyParser(&data); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Cannot parse JSON",
		})
	}

	if documentURL, exists := data["document_url"]; exists {
		manifest.DocumentURL = documentURL
		database.DB.Save(&manifest)
	}

	return c.JSON(manifest)
}
