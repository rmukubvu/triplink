package handlers

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"
	"triplink/backend/models"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type QuoteHandlerTestSuite struct {
	suite.Suite
	app *fiber.App
}

func (suite *QuoteHandlerTestSuite) SetupSuite() {
}

func (suite *QuoteHandlerTestSuite) SetupTest() {
	clearTestDB()
	seedTestDB()
	suite.app = fiber.New()

	suite.app.Post("/quotes", CreateQuote)
	suite.app.Get("/loads/:load_id/quotes", GetLoadQuotes)
	suite.app.Get("/carriers/:carrier_id/quotes", GetCarrierQuotes)
	suite.app.Post("/quotes/:id/accept", AcceptQuote)
	suite.app.Post("/quotes/:id/reject", RejectQuote)
	suite.app.Put("/quotes/:id", UpdateQuote)
}

func (suite *QuoteHandlerTestSuite) TearDownTest() {
	clearTestDB()
}

func (suite *QuoteHandlerTestSuite) TestCreateQuote() {
	t := suite.T()

	quote := models.Quote{
		LoadID:      1,
		CarrierID:   1,
		QuoteAmount: 1000.00,
		ValidUntil:  time.Now().Add(24 * time.Hour),
	}
	jsonData, _ := json.Marshal(quote)

	req := httptest.NewRequest("POST", "/quotes", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	resp, err := suite.app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 201, resp.StatusCode)
}

func (suite *QuoteHandlerTestSuite) TestGetLoadQuotes() {
	t := suite.T()
	req := httptest.NewRequest("GET", "/loads/1/quotes", nil)
	resp, err := suite.app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func (suite *QuoteHandlerTestSuite) TestGetCarrierQuotes() {
	t := suite.T()
	req := httptest.NewRequest("GET", "/carriers/1/quotes", nil)
	resp, err := suite.app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func (suite *QuoteHandlerTestSuite) TestAcceptQuote() {
	t := suite.T()
	req := httptest.NewRequest("POST", "/quotes/1/accept", nil)
	resp, err := suite.app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func (suite *QuoteHandlerTestSuite) TestRejectQuote() {
	t := suite.T()
	req := httptest.NewRequest("POST", "/quotes/1/reject", nil)
	resp, err := suite.app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func (suite *QuoteHandlerTestSuite) TestUpdateQuote() {
	t := suite.T()
	quote := models.Quote{QuoteAmount: 1200.00}
	jsonData, _ := json.Marshal(quote)

	req := httptest.NewRequest("PUT", "/quotes/1", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	resp, err := suite.app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestQuoteHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(QuoteHandlerTestSuite))
}