package handlers

import (
	"triplink/backend/database"
	"triplink/backend/models"

	"github.com/gofiber/fiber/v2"
)

// CreateReview @Summary Create a review
// @Description Create a new review for a user after a completed load
// @Tags reviews
// @Accept json
// @Produce json
// @Param review body models.Review true "Review data"
// @Success 201 {object} models.Review
// @Router /reviews [post]
func CreateReview(c *fiber.Ctx) error {
	var review models.Review

	if err := c.BodyParser(&review); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Cannot parse JSON",
		})
	}

	// Validate that the load exists and is completed
	var load models.Load
	if err := database.DB.First(&load, review.LoadID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Load not found",
		})
	}

	if load.Status != "DELIVERED" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Cannot review incomplete load",
		})
	}

	// Check if review already exists
	var existingReview models.Review
	if err := database.DB.Where("reviewer_id = ? AND reviewee_id = ? AND load_id = ?",
		review.ReviewerID, review.RevieweeID, review.LoadID).First(&existingReview).Error; err == nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Review already exists for this load",
		})
	}

	// Validate rating
	if review.Rating < 1 || review.Rating > 5 {
		return c.Status(400).JSON(fiber.Map{
			"error": "Rating must be between 1 and 5",
		})
	}

	result := database.DB.Create(&review)
	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Could not create review",
		})
	}

	// Update user's overall rating
	updateUserRating(review.RevieweeID)

	return c.Status(201).JSON(review)
}

// GetUserReviews @Summary Get reviews for a user
// @Description Get all reviews for a specific user
// @Tags reviews
// @Produce json
// @Param user_id path int true "User ID"
// @Success 200 {array} models.Review
// @Router /users/{user_id}/reviews [get]
func GetUserReviews(c *fiber.Ctx) error {
	userID := c.Params("user_id")
	var reviews []models.Review

	result := database.DB.Where("reviewee_id = ?", userID).
		Preload("Reviewer").
		Order("created_at DESC").
		Find(&reviews)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Could not fetch reviews",
		})
	}

	return c.JSON(reviews)
}

// GetReviewsByUser @Summary Get reviews by a user
// @Description Get all reviews written by a specific user
// @Tags reviews
// @Produce json
// @Param user_id path int true "User ID"
// @Success 200 {array} models.Review
// @Router /users/{user_id}/reviews-given [get]
func GetReviewsByUser(c *fiber.Ctx) error {
	userID := c.Params("user_id")
	var reviews []models.Review

	result := database.DB.Where("reviewer_id = ?", userID).
		Preload("Reviewee").
		Order("created_at DESC").
		Find(&reviews)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Could not fetch reviews",
		})
	}

	return c.JSON(reviews)
}

// GetReview @Summary Get review by ID
// @Description Get a specific review by its ID
// @Tags reviews
// @Produce json
// @Param id path int true "Review ID"
// @Success 200 {object} models.Review
// @Router /reviews/{id} [get]
func GetReview(c *fiber.Ctx) error {
	id := c.Params("id")
	var review models.Review

	result := database.DB.Preload("Reviewer").
		Preload("Reviewee").
		Preload("Load").
		First(&review, id)

	if result.Error != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Review not found",
		})
	}

	return c.JSON(review)
}

// UpdateReview @Summary Update review
// @Description Update an existing review (only by the reviewer)
// @Tags reviews
// @Accept json
// @Produce json
// @Param id path int true "Review ID"
// @Param review body models.Review true "Updated review data"
// @Success 200 {object} models.Review
// @Router /reviews/{id} [put]
func UpdateReview(c *fiber.Ctx) error {
	id := c.Params("id")
	var review models.Review

	if err := database.DB.First(&review, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Review not found",
		})
	}

	var updateData models.Review
	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Cannot parse JSON",
		})
	}

	// Validate rating
	if updateData.Rating < 1 || updateData.Rating > 5 {
		return c.Status(400).JSON(fiber.Map{
			"error": "Rating must be between 1 and 5",
		})
	}

	// Only allow updating rating and comment
	review.Rating = updateData.Rating
	review.Comment = updateData.Comment

	database.DB.Save(&review)

	// Update user's overall rating
	updateUserRating(review.RevieweeID)

	return c.JSON(review)
}

// DeleteReview @Summary Delete review
// @Description Delete a review (only by the reviewer)
// @Tags reviews
// @Param id path int true "Review ID"
// @Success 200 {object} map[string]string
// @Router /reviews/{id} [delete]
func DeleteReview(c *fiber.Ctx) error {
	id := c.Params("id")
	var review models.Review

	if err := database.DB.First(&review, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Review not found",
		})
	}

	revieweeID := review.RevieweeID
	database.DB.Delete(&review)

	// Update user's overall rating
	updateUserRating(revieweeID)

	return c.JSON(fiber.Map{
		"message": "Review deleted successfully",
	})
}

// GetUserRatingSummary @Summary Get user rating summary
// @Description Get rating summary and statistics for a user
// @Tags reviews
// @Produce json
// @Param user_id path int true "User ID"
// @Success 200 {object} map[string]interface{}
// @Router /users/{user_id}/rating-summary [get]
func GetUserRatingSummary(c *fiber.Ctx) error {
	userID := c.Params("user_id")

	// Get all reviews for the user
	var reviews []models.Review
	database.DB.Where("reviewee_id = ?", userID).Find(&reviews)

	if len(reviews) == 0 {
		return c.JSON(map[string]interface{}{
			"user_id":        userID,
			"average_rating": 0,
			"total_reviews":  0,
			"rating_breakdown": map[string]int{
				"5_star": 0,
				"4_star": 0,
				"3_star": 0,
				"2_star": 0,
				"1_star": 0,
			},
		})
	}

	// Calculate statistics
	var totalRating int
	ratingBreakdown := map[string]int{
		"5_star": 0,
		"4_star": 0,
		"3_star": 0,
		"2_star": 0,
		"1_star": 0,
	}

	for _, review := range reviews {
		totalRating += review.Rating
		switch review.Rating {
		case 5:
			ratingBreakdown["5_star"]++
		case 4:
			ratingBreakdown["4_star"]++
		case 3:
			ratingBreakdown["3_star"]++
		case 2:
			ratingBreakdown["2_star"]++
		case 1:
			ratingBreakdown["1_star"]++
		}
	}

	averageRating := float64(totalRating) / float64(len(reviews))

	summary := map[string]interface{}{
		"user_id":          userID,
		"average_rating":   averageRating,
		"total_reviews":    len(reviews),
		"rating_breakdown": ratingBreakdown,
	}

	return c.JSON(summary)
}

// Helper function to update user's overall rating
func updateUserRating(userID uint) {
	var reviews []models.Review
	database.DB.Where("reviewee_id = ?", userID).Find(&reviews)

	if len(reviews) == 0 {
		database.DB.Model(&models.User{}).Where("id = ?", userID).
			Updates(map[string]interface{}{
				"rating":        0,
				"total_reviews": 0,
			})
		return
	}

	var totalRating int
	for _, review := range reviews {
		totalRating += review.Rating
	}

	averageRating := float64(totalRating) / float64(len(reviews))

	database.DB.Model(&models.User{}).Where("id = ?", userID).
		Updates(map[string]interface{}{
			"rating":        averageRating,
			"total_reviews": len(reviews),
		})
}
