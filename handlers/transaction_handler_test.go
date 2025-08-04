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

type TransactionHandlerTestSuite struct {
	suite.Suite
	app *fiber.App
}

func (suite *TransactionHandlerTestSuite) SetupSuite() {
}

func (suite *TransactionHandlerTestSuite) SetupTest() {
	clearTestDB()
	seedTestDB()
	suite.app = fiber.New()

	suite.app.Post("/transactions", CreateTransaction)
	suite.app.Get("/transactions", GetTransactions)
}

func (suite *TransactionHandlerTestSuite) TearDownTest() {
	clearTestDB()
}

func (suite *TransactionHandlerTestSuite) TestCreateTransaction() {
	t := suite.T()

	transaction := models.Transaction{
		Amount:   100.00,
		Currency: "USD",
		PaymentMethod: "CARD",
		PaymentGateway: "STRIPE",
		Status: "COMPLETED",
	}
	jsonData, _ := json.Marshal(transaction)

	req := httptest.NewRequest("POST", "/transactions", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	resp, err := suite.app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func (suite *TransactionHandlerTestSuite) TestGetTransactions() {
	t := suite.T()

	req := httptest.NewRequest("GET", "/transactions", nil)
	resp, err := suite.app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestTransactionHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(TransactionHandlerTestSuite))
}