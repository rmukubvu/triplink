package handlers

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type UserHandlerTestSuite struct {
	suite.Suite
	app *fiber.App
}

func (suite *UserHandlerTestSuite) SetupSuite() {
}

func (suite *UserHandlerTestSuite) SetupTest() {
	clearTestDB()
	seedTestDB()
	suite.app = fiber.New()

	suite.app.Post("/api/register", Register)
	suite.app.Post("/api/login", Login)
}

func (suite *UserHandlerTestSuite) TearDownTest() {
	clearTestDB()
}

func (suite *UserHandlerTestSuite) TestRegister() {
	t := suite.T()

	user := map[string]string{
		"email":    "newuser@example.com",
		"phone":    "+1987654321",
		"password": "newpassword",
		"role":     "user",
	}
	jsonData, _ := json.Marshal(user)

	req := httptest.NewRequest("POST", "/api/register", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	resp, err := suite.app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func (suite *UserHandlerTestSuite) TestLogin() {
	t := suite.T()

	user := map[string]string{
		"email":    "test@example.com",
		"password": "password",
	}
	jsonData, _ := json.Marshal(user)

	req := httptest.NewRequest("POST", "/api/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	resp, err := suite.app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestUserHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(UserHandlerTestSuite))
}