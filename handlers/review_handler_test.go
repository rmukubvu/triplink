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

type ReviewHandlerTestSuite struct {
	suite.Suite
	app *fiber.App
}

func (suite *ReviewHandlerTestSuite) SetupSuite() {
}

func (suite *ReviewHandlerTestSuite) SetupTest() {
	clearTestDB()
	seedTestDB()
	suite.app = fiber.New()

	suite.app.Post("/reviews", CreateReview)
	suite.app.Get("/users/:user_id/reviews", GetUserReviews)
	suite.app.Get("/users/:user_id/reviews-given", GetReviewsByUser)
	suite.app.Get("/reviews/:id", GetReview)
	suite.app.Put("/reviews/:id", UpdateReview)
	suite.app.Delete("/reviews/:id", DeleteReview)
	suite.app.Get("/users/:user_id/rating-summary", GetUserRatingSummary)
}

func (suite *ReviewHandlerTestSuite) TearDownTest() {
	clearTestDB()
}

func (suite *ReviewHandlerTestSuite) TestCreateReview() {
	t := suite.T()

	review := models.Review{
		LoadID:     1,
		ReviewerID: 1,
		RevieweeID: 1, // Reviewee should be the seeded user
		Rating:     5,
		Comment:    "Great service!",
	}
	jsonData, _ := json.Marshal(review)

	req := httptest.NewRequest("POST", "/reviews", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	resp, err := suite.app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 201, resp.StatusCode)
}

func (suite *ReviewHandlerTestSuite) TestGetUserReviews() {
	t := suite.T()
	req := httptest.NewRequest("GET", "/users/1/reviews", nil)
	resp, err := suite.app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func (suite *ReviewHandlerTestSuite) TestGetReviewsByUser() {
	t := suite.T()
	req := httptest.NewRequest("GET", "/users/1/reviews-given", nil)
	resp, err := suite.app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func (suite *ReviewHandlerTestSuite) TestGetReview() {
	t := suite.T()
	req := httptest.NewRequest("GET", "/reviews/1", nil)
	resp, err := suite.app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func (suite *ReviewHandlerTestSuite) TestUpdateReview() {
	t := suite.T()
	review := models.Review{Rating: 4, Comment: "Good"}
	jsonData, _ := json.Marshal(review)

	req := httptest.NewRequest("PUT", "/reviews/1", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	resp, err := suite.app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func (suite *ReviewHandlerTestSuite) TestDeleteReview() {
	t := suite.T()
	req := httptest.NewRequest("DELETE", "/reviews/1", nil)
	resp, err := suite.app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func (suite *ReviewHandlerTestSuite) TestGetUserRatingSummary() {
	t := suite.T()
	req := httptest.NewRequest("GET", "/users/1/rating-summary", nil)
	resp, err := suite.app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestReviewHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(ReviewHandlerTestSuite))
}