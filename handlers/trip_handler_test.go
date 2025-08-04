package handlers

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"triplink/backend/models"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type TripHandlerTestSuite struct {
	suite.Suite
	app *fiber.App
}

func (suite *TripHandlerTestSuite) SetupSuite() {
}

func (suite *TripHandlerTestSuite) SetupTest() {
	clearTestDB()
	seedTestDB()
	suite.app = fiber.New()

	suite.app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", float64(1)) // Set a dummy user ID for testing
		return c.Next()
	})

	suite.app.Post("/trips", CreateTrip)
	suite.app.Get("/trips", GetTrips)
	suite.app.Get("/trips/:id", GetTrip)
}

func (suite *TripHandlerTestSuite) TearDownTest() {
	clearTestDB()
}

func (suite *TripHandlerTestSuite) TestCreateTrip() {
	t := suite.T()

	trip := models.Trip{
		OriginAddress:      "123 Main St",
		DestinationAddress: "456 Oak Ave",
	}
	jsonData, _ := json.Marshal(trip)

	req := httptest.NewRequest("POST", "/trips", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	resp, err := suite.app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func (suite *TripHandlerTestSuite) TestGetTrips() {
	t := suite.T()

	req := httptest.NewRequest("GET", "/trips", nil)
	resp, err := suite.app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func (suite *TripHandlerTestSuite) TestGetTrip() {
	t := suite.T()

	req := httptest.NewRequest("GET", "/trips/1", nil)
	resp, err := suite.app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestTripHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(TripHandlerTestSuite))
}