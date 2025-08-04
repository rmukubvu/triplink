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

type LoadHandlerTestSuite struct {
	suite.Suite
	app *fiber.App
}

func (suite *LoadHandlerTestSuite) SetupSuite() {
}

func (suite *LoadHandlerTestSuite) SetupTest() {
	clearTestDB()
	seedTestDB()
	suite.app = fiber.New()

	suite.app.Post("/loads", CreateLoad)
	suite.app.Get("/loads", GetLoads)
	suite.app.Get("/loads/:id", GetLoad)
}

func (suite *LoadHandlerTestSuite) TearDownTest() {
	clearTestDB()
}

func (suite *LoadHandlerTestSuite) TestCreateLoad() {
	t := suite.T()

	load := models.Load{
		BookingReference: "NEW-TEST-LOAD",
	}
	jsonData, _ := json.Marshal(load)

	req := httptest.NewRequest("POST", "/loads", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	resp, err := suite.app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func (suite *LoadHandlerTestSuite) TestGetLoads() {
	t := suite.T()

	req := httptest.NewRequest("GET", "/loads", nil)
	resp, err := suite.app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func (suite *LoadHandlerTestSuite) TestGetLoad() {
	t := suite.T()

	req := httptest.NewRequest("GET", "/loads/1", nil)
	resp, err := suite.app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestLoadHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(LoadHandlerTestSuite))
}