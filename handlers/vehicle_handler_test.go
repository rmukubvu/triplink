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

type VehicleHandlerTestSuite struct {
	suite.Suite
	app *fiber.App
}

func (suite *VehicleHandlerTestSuite) SetupSuite() {
}

func (suite *VehicleHandlerTestSuite) SetupTest() {
	clearTestDB()
	seedTestDB()
	suite.app = fiber.New()

	suite.app.Post("/vehicles", CreateVehicle)
	suite.app.Get("/users/:user_id/vehicles", GetUserVehicles)
	suite.app.Get("/vehicles/:id", GetVehicle)
	suite.app.Put("/vehicles/:id", UpdateVehicle)
	suite.app.Delete("/vehicles/:id", DeleteVehicle)
	suite.app.Get("/vehicles/search", SearchVehicles)
}

func (suite *VehicleHandlerTestSuite) TearDownTest() {
	clearTestDB()
}

func (suite *VehicleHandlerTestSuite) TestCreateVehicle() {
	t := suite.T()

	vehicle := models.Vehicle{
		UserID:      1,
		VehicleType: "VAN",
		Make:        "Ford",
		Model:       "Transit",
	}
	jsonData, _ := json.Marshal(vehicle)

	req := httptest.NewRequest("POST", "/vehicles", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	resp, err := suite.app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 201, resp.StatusCode)
}

func (suite *VehicleHandlerTestSuite) TestGetUserVehicles() {
	t := suite.T()
	req := httptest.NewRequest("GET", "/users/1/vehicles", nil)
	resp, err := suite.app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func (suite *VehicleHandlerTestSuite) TestGetVehicle() {
	t := suite.T()
	req := httptest.NewRequest("GET", "/vehicles/1", nil)
	resp, err := suite.app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func (suite *VehicleHandlerTestSuite) TestUpdateVehicle() {
	t := suite.T()
	vehicle := models.Vehicle{Make: "Mercedes"}
	jsonData, _ := json.Marshal(vehicle)

	req := httptest.NewRequest("PUT", "/vehicles/1", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	resp, err := suite.app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func (suite *VehicleHandlerTestSuite) TestDeleteVehicle() {
	t := suite.T()
	req := httptest.NewRequest("DELETE", "/vehicles/1", nil)
	resp, err := suite.app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func (suite *VehicleHandlerTestSuite) TestSearchVehicles() {
	t := suite.T()
	req := httptest.NewRequest("GET", "/vehicles/search?vehicle_type=TRUCK", nil)
	resp, err := suite.app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestVehicleHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(VehicleHandlerTestSuite))
}