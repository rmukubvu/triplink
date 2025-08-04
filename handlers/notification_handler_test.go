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

type NotificationHandlerTestSuite struct {
	suite.Suite
	app *fiber.App
}

func (suite *NotificationHandlerTestSuite) SetupSuite() {
}

func (suite *NotificationHandlerTestSuite) SetupTest() {
	clearTestDB()
	seedTestDB()
	suite.app = fiber.New()

	suite.app.Post("/notifications", CreateNotification)
	suite.app.Get("/users/:user_id/notifications", GetUserNotifications)
	suite.app.Put("/notifications/:id/read", MarkNotificationAsRead)
	suite.app.Put("/users/:user_id/notifications/read-all", MarkAllNotificationsAsRead)
	suite.app.Delete("/notifications/:id", DeleteNotification)
	suite.app.Get("/users/:user_id/notifications/count", GetNotificationCounts)
}

func (suite *NotificationHandlerTestSuite) TearDownTest() {
	clearTestDB()
}

func (suite *NotificationHandlerTestSuite) TestCreateNotification() {
	t := suite.T()

	notification := models.Notification{
		UserID:  1,
		Title:   "New Notification",
		Message: "This is a new test notification",
	}
	jsonData, _ := json.Marshal(notification)

	req := httptest.NewRequest("POST", "/notifications", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	resp, err := suite.app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 201, resp.StatusCode)
}

func (suite *NotificationHandlerTestSuite) TestGetUserNotifications() {
	t := suite.T()

	req := httptest.NewRequest("GET", "/users/1/notifications", nil)
	resp, err := suite.app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func (suite *NotificationHandlerTestSuite) TestMarkNotificationAsRead() {
	t := suite.T()
	req := httptest.NewRequest("PUT", "/notifications/1/read", nil)
	resp, err := suite.app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func (suite *NotificationHandlerTestSuite) TestMarkAllNotificationsAsRead() {
	t := suite.T()
	req := httptest.NewRequest("PUT", "/users/1/notifications/read-all", nil)
	resp, err := suite.app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func (suite *NotificationHandlerTestSuite) TestDeleteNotification() {
	t := suite.T()
	req := httptest.NewRequest("DELETE", "/notifications/1", nil)
	resp, err := suite.app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func (suite *NotificationHandlerTestSuite) TestGetNotificationCounts() {
	t := suite.T()
	req := httptest.NewRequest("GET", "/users/1/notifications/count", nil)
	resp, err := suite.app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestNotificationHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(NotificationHandlerTestSuite))
}