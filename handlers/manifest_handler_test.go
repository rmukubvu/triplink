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

type ManifestHandlerTestSuite struct {
	suite.Suite
	app *fiber.App
}

func (suite *ManifestHandlerTestSuite) SetupSuite() {
}

func (suite *ManifestHandlerTestSuite) SetupTest() {
	clearTestDB()
	seedTestDB()
	suite.app = fiber.New()

	suite.app.Post("/trips/:trip_id/manifest", GenerateManifest)
	suite.app.Get("/trips/:trip_id/manifest", GetTripManifest)
	suite.app.Get("/manifests/:id", GetManifest)
	suite.app.Get("/manifests/:id/detailed", GetDetailedManifest)
	suite.app.Put("/manifests/:id/document", UpdateManifestDocument)
}

func (suite *ManifestHandlerTestSuite) TearDownTest() {
	clearTestDB()
}

func (suite *ManifestHandlerTestSuite) TestGenerateManifest() {
	t := suite.T()
	req := httptest.NewRequest("POST", "/trips/1/manifest", nil)
	resp, err := suite.app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode) // Trip should exist now
}

func (suite *ManifestHandlerTestSuite) TestGetTripManifest() {
	t := suite.T()
	req := httptest.NewRequest("GET", "/trips/1/manifest", nil)
	resp, err := suite.app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode) // Manifest should exist now
}

func (suite *ManifestHandlerTestSuite) TestGetManifest() {
	t := suite.T()
	req := httptest.NewRequest("GET", "/manifests/1", nil)
	resp, err := suite.app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode) // Manifest should exist now
}

func (suite *ManifestHandlerTestSuite) TestGetDetailedManifest() {
	t := suite.T()
	req := httptest.NewRequest("GET", "/manifests/1/detailed", nil)
	resp, err := suite.app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode) // Manifest should exist now
}

func (suite *ManifestHandlerTestSuite) TestUpdateManifestDocument() {
	t := suite.T()
	data := map[string]string{"document_url": "http://example.com/doc.pdf"}
	jsonData, _ := json.Marshal(data)

	req := httptest.NewRequest("PUT", "/manifests/1/document", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	resp, err := suite.app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode) // Manifest should exist now
}

func TestManifestHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(ManifestHandlerTestSuite))
}