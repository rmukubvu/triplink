package routes

import (
	"github.com/gofiber/fiber/v2"
	"triplink/backend/auth"
	"triplink/backend/handlers"

	swagger "github.com/gofiber/swagger" // swagger handler
)

func Setup(app *fiber.App) {
	// @BasePath /api
	// Auth
	app.Post("/api/register", handlers.Register)
	app.Post("/api/login", handlers.Login)

	// Trips
	app.Get("/api/trips", handlers.GetTrips)
	app.Get("/api/trips/:id", handlers.GetTrip)
	app.Post("/api/trips", auth.Middleware(), handlers.CreateTrip)

	// Loads
	app.Get("/api/loads", handlers.GetLoads)
	app.Get("/api/loads/:id", handlers.GetLoad)
	app.Post("/api/loads", auth.Middleware(), handlers.CreateLoad)

	// Messages
	app.Get("/api/messages", auth.Middleware(), handlers.GetMessages)
	app.Post("/api/messages", auth.Middleware(), handlers.CreateMessage)

	// Transactions
	app.Get("/api/transactions", auth.Middleware(), handlers.GetTransactions)
	app.Post("/api/transactions", auth.Middleware(), handlers.CreateTransaction)

	// Swagger
	app.Get("/swagger/*", swagger.HandlerDefault) // default
}
