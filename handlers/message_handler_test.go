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

type MessageHandlerTestSuite struct {
	suite.Suite
	app *fiber.App
}

func (suite *MessageHandlerTestSuite) SetupSuite() {
}

func (suite *MessageHandlerTestSuite) SetupTest() {
	clearTestDB()
	seedTestDB()
	suite.app = fiber.New()

	suite.app.Post("/messages", CreateMessage)
	suite.app.Get("/messages", GetMessages)
	suite.app.Get("/users/:user_id/messages", GetUserMessages)
	suite.app.Get("/messages/conversation/:user1_id/:user2_id", GetConversation)
}

func (suite *MessageHandlerTestSuite) TearDownTest() {
	clearTestDB()
}

func (suite *MessageHandlerTestSuite) TestCreateMessage() {
	t := suite.T()

	message := models.Message{
		SenderID:   1,
		ReceiverID: 2,
		Content:    "Hello",
	}
	jsonData, _ := json.Marshal(message)

	req := httptest.NewRequest("POST", "/messages", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	resp, err := suite.app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 201, resp.StatusCode)
}

func (suite *MessageHandlerTestSuite) TestGetMessages() {
	t := suite.T()

	req := httptest.NewRequest("GET", "/messages", nil)
	resp, err := suite.app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func (suite *MessageHandlerTestSuite) TestGetUserMessages() {
	t := suite.T()

	req := httptest.NewRequest("GET", "/users/1/messages", nil)
	resp, err := suite.app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func (suite *MessageHandlerTestSuite) TestGetConversation() {
	t := suite.T()

	req := httptest.NewRequest("GET", "/messages/conversation/1/2", nil)
	resp, err := suite.app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestMessageHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(MessageHandlerTestSuite))
}