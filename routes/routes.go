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

	// Users
	app.Get("/api/users/:user_id/vehicles", handlers.GetUserVehicles)

	// Vehicles
	app.Post("/api/vehicles", auth.Middleware(), handlers.CreateVehicle)
	app.Get("/api/vehicles/:id", handlers.GetVehicle)
	app.Put("/api/vehicles/:id", auth.Middleware(), handlers.UpdateVehicle)
	app.Delete("/api/vehicles/:id", auth.Middleware(), handlers.DeleteVehicle)
	app.Get("/api/vehicles/search", handlers.SearchVehicles)

	// Trips
	app.Get("/api/trips", handlers.GetTrips)
	app.Get("/api/trips/:id", handlers.GetTrip)
	app.Post("/api/trips", auth.Middleware(), handlers.CreateTrip)
	app.Post("/api/trips/:trip_id/manifest", auth.Middleware(), handlers.GenerateManifest)
	app.Get("/api/trips/:trip_id/manifest", handlers.GetTripManifest)
	app.Get("/api/trips/:trip_id/customs-summary", handlers.GetTripCustomsSummary)

	// Loads
	app.Get("/api/loads", handlers.GetLoads)
	app.Get("/api/loads/:id", handlers.GetLoad)
	app.Post("/api/loads", auth.Middleware(), handlers.CreateLoad)
	app.Get("/api/loads/:load_id/quotes", handlers.GetLoadQuotes)
	app.Get("/api/loads/:load_id/customs-documents", handlers.GetLoadCustomsDocuments)
	app.Post("/api/loads/:load_id/commercial-invoice", auth.Middleware(), handlers.GenerateCommercialInvoice)
	app.Post("/api/loads/:load_id/bill-of-lading", auth.Middleware(), handlers.GenerateBillOfLading)
	app.Post("/api/loads/:load_id/packing-list", auth.Middleware(), handlers.GeneratePackingList)

	// Quotes
	app.Post("/api/quotes", auth.Middleware(), handlers.CreateQuote)
	app.Get("/api/carriers/:carrier_id/quotes", handlers.GetCarrierQuotes)
	app.Post("/api/quotes/:id/accept", auth.Middleware(), handlers.AcceptQuote)
	app.Post("/api/quotes/:id/reject", auth.Middleware(), handlers.RejectQuote)
	app.Put("/api/quotes/:id", auth.Middleware(), handlers.UpdateQuote)

	// Manifests
	app.Get("/api/manifests/:id", handlers.GetManifest)
	app.Get("/api/manifests/:id/detailed", handlers.GetDetailedManifest)
	app.Put("/api/manifests/:id/document", auth.Middleware(), handlers.UpdateManifestDocument)

	// Customs Documents
	app.Post("/api/customs-documents", auth.Middleware(), handlers.CreateCustomsDocument)
	app.Get("/api/customs-documents/:id", handlers.GetCustomsDocument)
	app.Put("/api/customs-documents/:id", auth.Middleware(), handlers.UpdateCustomsDocument)
	app.Delete("/api/customs-documents/:id", auth.Middleware(), handlers.DeleteCustomsDocument)

	// Messages
	app.Get("/api/messages", auth.Middleware(), handlers.GetMessages)
	app.Post("/api/messages", auth.Middleware(), handlers.CreateMessage)

	// Transactions
	app.Get("/api/transactions", auth.Middleware(), handlers.GetTransactions)
	app.Post("/api/transactions", auth.Middleware(), handlers.CreateTransaction)

	// Swagger
	app.Get("/swagger/*", swagger.HandlerDefault) // default
}
