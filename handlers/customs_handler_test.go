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

type CustomsHandlerTestSuite struct {
	suite.Suite
	app *fiber.App
}

func (suite *CustomsHandlerTestSuite) SetupSuite() {
}

func (suite *CustomsHandlerTestSuite) SetupTest() {
	clearTestDB()
	seedTestDB()
	suite.app = fiber.New()

	suite.app.Post("/customs-documents", CreateCustomsDocument)
	suite.app.Get("/loads/:load_id/customs-documents", GetLoadCustomsDocuments)
	suite.app.Get("/customs-documents/:id", GetCustomsDocument)
	suite.app.Put("/customs-documents/:id", UpdateCustomsDocument)
	suite.app.Delete("/customs-documents/:id", DeleteCustomsDocument)
	suite.app.Post("/loads/:load_id/commercial-invoice", GenerateCommercialInvoice)
	suite.app.Post("/loads/:load_id/bill-of-lading", GenerateBillOfLading)
	suite.app.Post("/loads/:load_id/packing-list", GeneratePackingList)
	suite.app.Get("/trips/:trip_id/customs-summary", GetTripCustomsSummary)
}

func (suite *CustomsHandlerTestSuite) TearDownTest() {
	clearTestDB()
}

func (suite *CustomsHandlerTestSuite) TestCreateCustomsDocument() {
	t := suite.T()

	tests := []struct {
		name           string
		document       models.CustomsDocument
		expectedStatus int
	}{
		{
			name: "Valid document",
			document: models.CustomsDocument{
				LoadID:       1,
				DocumentType: "INVOICE",
			},
			expectedStatus: 201,
		},
		{
			name:           "Invalid JSON",
			document:       models.CustomsDocument{},
			expectedStatus: 400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var jsonData []byte
			var err error
			if tt.name != "Invalid JSON" {
				jsonData, err = json.Marshal(tt.document)
				assert.NoError(t, err)
			} else {
				jsonData = []byte(`{"bad": "json"`)
			}

			req := httptest.NewRequest("POST", "/customs-documents", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")

			resp, err := suite.app.Test(req)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}

func (suite *CustomsHandlerTestSuite) TestGetLoadCustomsDocuments() {
	t := suite.T()
	req := httptest.NewRequest("GET", "/loads/1/customs-documents", nil)
	resp, err := suite.app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func (suite *CustomsHandlerTestSuite) TestGetCustomsDocument() {
	t := suite.T()
	req := httptest.NewRequest("GET", "/customs-documents/1", nil)
	resp, err := suite.app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func (suite *CustomsHandlerTestSuite) TestUpdateCustomsDocument() {
	t := suite.T()
	doc := models.CustomsDocument{DocumentType: "UPDATED"}
	jsonData, _ := json.Marshal(doc)

	req := httptest.NewRequest("PUT", "/customs-documents/1", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	resp, err := suite.app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func (suite *CustomsHandlerTestSuite) TestDeleteCustomsDocument() {
	t := suite.T()
	req := httptest.NewRequest("DELETE", "/customs-documents/1", nil)
	resp, err := suite.app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func (suite *CustomsHandlerTestSuite) TestGenerateCommercialInvoice() {
	t := suite.T()
	req := httptest.NewRequest("POST", "/loads/1/commercial-invoice", nil)
	resp, err := suite.app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func (suite *CustomsHandlerTestSuite) TestGenerateBillOfLading() {
	t := suite.T()
	req := httptest.NewRequest("POST", "/loads/1/bill-of-lading", nil)
	resp, err := suite.app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func (suite *CustomsHandlerTestSuite) TestGeneratePackingList() {
	t := suite.T()
	req := httptest.NewRequest("POST", "/loads/1/packing-list", nil)
	resp, err := suite.app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func (suite *CustomsHandlerTestSuite) TestGetTripCustomsSummary() {
	t := suite.T()
	req := httptest.NewRequest("GET", "/trips/1/customs-summary", nil)
	resp, err := suite.app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestCustomsHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(CustomsHandlerTestSuite))
}