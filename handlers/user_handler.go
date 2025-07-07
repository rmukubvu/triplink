package handlers

import (
	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
	"time"
	"triplink/backend/auth"
	"triplink/backend/database"
	"triplink/backend/models"
)

// ErrorResponse represents a generic error response
type ErrorResponse struct {
	Message string `json:"message"`
}

// LoginSuccessResponse represents a successful login response
type LoginSuccessResponse struct {
	Message string `json:"message"`
	Token   string `json:"token"`
}

// Register godoc
// @Summary Register a new user
// @Description Register a new user with email, phone, password, and role
// @Tags Auth
// @Accept json
// @Produce json
// @Param user body map[string]string true "User registration data"
// @Success 200 {object} models.User
// @Failure 400 {object} ErrorResponse
// @Router /api/register [post]
func Register(c *fiber.Ctx) error {
	var data map[string]string

	if err := c.BodyParser(&data); err != nil {
		return err
	}

	password, _ := bcrypt.GenerateFromPassword([]byte(data["password"]), 14)

	user := models.User{
		Email:    data["email"],
		Phone:    data["phone"],
		Password: string(password),
		Role:     data["role"],
	}

	database.DB.Create(&user)

	return c.JSON(user)
}

// Login godoc
// @Summary Login a user
// @Description Login a user with email and password
// @Tags Auth
// @Accept json
// @Produce json
// @Param user body map[string]string true "User login data"
// @Success 200 {object} LoginSuccessResponse
// @Failure 404 {object} ErrorResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} string "Internal Server Error"
// @Router /api/login [post]
func Login(c *fiber.Ctx) error {
	var data map[string]string

	if err := c.BodyParser(&data); err != nil {
		return err
	}

	var user models.User

	database.DB.Where("email = ?", data["email"]).First(&user)

	if user.ID == 0 {
		c.Status(fiber.StatusNotFound)
		return c.JSON(fiber.Map{
			"message": "user not found",
		})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(data["password"])); err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"message": "incorrect password",
		})
	}

	token, err := auth.GenerateJWT(user.ID, user.Role)

	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	c.Cookie(&fiber.Cookie{
		Name:     "jwt",
		Value:    token,
		Expires:  time.Now().Add(time.Hour * 24),
		HTTPOnly: true,
	})

	return c.JSON(fiber.Map{
		"message": "success",
		"token":   token,
	})
}
