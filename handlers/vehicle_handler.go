package handlers

import (
	"strconv"
	"triplink/backend/database"
	"triplink/backend/models"

	"github.com/gofiber/fiber/v2"
)

// CreateVehicle @Summary Create a new vehicle
// @Description Create a new vehicle for a carrier
// @Tags vehicles
// @Accept json
// @Produce json
// @Param vehicle body models.Vehicle true "Vehicle data"
// @Success 201 {object} models.Vehicle
// @Router /vehicles [post]
func CreateVehicle(c *fiber.Ctx) error {
	var vehicle models.Vehicle

	if err := c.BodyParser(&vehicle); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Cannot parse JSON",
		})
	}

	result := database.DB.Create(&vehicle)
	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Could not create vehicle",
		})
	}

	return c.Status(201).JSON(vehicle)
}

// GetUserVehicles @Summary Get all vehicles for a user
// @Description Get all vehicles belonging to a specific user
// @Tags vehicles
// @Produce json
// @Param user_id path int true "User ID"
// @Success 200 {array} models.Vehicle
// @Router /users/{user_id}/vehicles [get]
func GetUserVehicles(c *fiber.Ctx) error {
	userID := c.Params("user_id")
	var vehicles []models.Vehicle

	result := database.DB.Where("user_id = ? AND is_active = ?", userID, true).Find(&vehicles)
	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Could not fetch vehicles",
		})
	}

	return c.JSON(vehicles)
}

// GetVehicle @Summary Get vehicle by ID
// @Description Get a specific vehicle by its ID
// @Tags vehicles
// @Produce json
// @Param id path int true "Vehicle ID"
// @Success 200 {object} models.Vehicle
// @Router /vehicles/{id} [get]
func GetVehicle(c *fiber.Ctx) error {
	id := c.Params("id")
	var vehicle models.Vehicle

	result := database.DB.First(&vehicle, id)
	if result.Error != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Vehicle not found",
		})
	}

	return c.JSON(vehicle)
}

// UpdateVehicle @Summary Update vehicle
// @Description Update an existing vehicle
// @Tags vehicles
// @Accept json
// @Produce json
// @Param id path int true "Vehicle ID"
// @Param vehicle body models.Vehicle true "Updated vehicle data"
// @Success 200 {object} models.Vehicle
// @Router /vehicles/{id} [put]
func UpdateVehicle(c *fiber.Ctx) error {
	id := c.Params("id")
	var vehicle models.Vehicle

	if err := database.DB.First(&vehicle, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Vehicle not found",
		})
	}

	if err := c.BodyParser(&vehicle); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Cannot parse JSON",
		})
	}

	database.DB.Save(&vehicle)
	return c.JSON(vehicle)
}

// DeleteVehicle @Summary Delete vehicle
// @Description Soft delete a vehicle (set is_active to false)
// @Tags vehicles
// @Param id path int true "Vehicle ID"
// @Success 200 {object} map[string]string
// @Router /vehicles/{id} [delete]
func DeleteVehicle(c *fiber.Ctx) error {
	id := c.Params("id")
	var vehicle models.Vehicle

	if err := database.DB.First(&vehicle, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Vehicle not found",
		})
	}

	database.DB.Model(&vehicle).Update("is_active", false)
	return c.JSON(fiber.Map{
		"message": "Vehicle deleted successfully",
	})
}

// SearchVehicles @Summary Search vehicles
// @Description Search for available vehicles by type, capacity, and location
// @Tags vehicles
// @Produce json
// @Param vehicle_type query string false "Vehicle type"
// @Param min_capacity_kg query number false "Minimum weight capacity"
// @Param min_capacity_m3 query number false "Minimum volume capacity"
// @Param has_liftgate query boolean false "Requires liftgate"
// @Param is_refrigerated query boolean false "Requires refrigeration"
// @Param is_hazmat_certified query boolean false "Requires hazmat certification"
// @Success 200 {array} models.Vehicle
// @Router /vehicles/search [get]
func SearchVehicles(c *fiber.Ctx) error {
	query := database.DB.Where("is_active = ?", true)

	if vehicleType := c.Query("vehicle_type"); vehicleType != "" {
		query = query.Where("vehicle_type = ?", vehicleType)
	}

	if minCapacityKg := c.Query("min_capacity_kg"); minCapacityKg != "" {
		if capacity, err := strconv.ParseFloat(minCapacityKg, 64); err == nil {
			query = query.Where("load_capacity_kg >= ?", capacity)
		}
	}

	if minCapacityM3 := c.Query("min_capacity_m3"); minCapacityM3 != "" {
		if capacity, err := strconv.ParseFloat(minCapacityM3, 64); err == nil {
			query = query.Where("load_capacity_m3 >= ?", capacity)
		}
	}

	if c.Query("has_liftgate") == "true" {
		query = query.Where("has_liftgate = ?", true)
	}

	if c.Query("is_refrigerated") == "true" {
		query = query.Where("is_refrigerated = ?", true)
	}

	if c.Query("is_hazmat_certified") == "true" {
		query = query.Where("is_hazmat_certified = ?", true)
	}

	var vehicles []models.Vehicle
	result := query.Find(&vehicles)
	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Could not search vehicles",
		})
	}

	return c.JSON(vehicles)
}
