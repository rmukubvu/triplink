package routes

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"triplink/backend/auth"
	"triplink/backend/handlers"
	"triplink/backend/middleware"

	swagger "github.com/gofiber/swagger" // swagger handler
)

func Setup(app *fiber.App) {
	// @BasePath /api
	
	// Initialize cache middleware
	cacheMiddleware := middleware.NewCacheMiddleware()
	
	// Add rate limiting middleware globally (100 requests per minute)
	app.Use(cacheMiddleware.RateLimitMiddleware(100, time.Minute))
	
	// Add session middleware
	app.Use(cacheMiddleware.SessionMiddleware())
	
	// Cache health check endpoint
	app.Get("/api/cache/health", cacheMiddleware.HealthCheckHandler())
	
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

	// Analytics Routes with caching
	analyticsGroup := app.Group("/api/analytics", auth.Middleware(), cacheMiddleware.Cache("analytics"))
	analyticsGroup.Post("/on-time-delivery", handlers.GetOnTimeDeliveryAnalytics)
	analyticsGroup.Post("/delivery-performance/by-route", handlers.GetDeliveryPerformanceByRoute)
	analyticsGroup.Post("/delivery-performance/by-driver", handlers.GetDeliveryPerformanceByDriver)
	analyticsGroup.Post("/customer-satisfaction", handlers.GetCustomerSatisfactionAnalytics)
	analyticsGroup.Post("/load-matching", handlers.GetLoadMatchingAnalytics)
	analyticsGroup.Post("/capacity-utilization", handlers.GetCapacityUtilizationAnalytics)
	analyticsGroup.Post("/delay-analysis", handlers.GetDelayAnalysisAnalytics)
	analyticsGroup.Post("/vehicle-capacity", handlers.GetVehicleCapacityData)
	analyticsGroup.Post("/kpis/:category", handlers.GetOperationalKPIs)

	// Route Optimization Routes with caching
	routeOptGroup := app.Group("/api/route-optimization", auth.Middleware(), cacheMiddleware.Cache("route_optimization"))
	routeOptGroup.Post("/optimize", handlers.OptimizeRoute)
	routeOptGroup.Post("/multiple", handlers.GetMultipleRouteOptions)
	routeOptGroup.Post("/compare", handlers.CompareRoutes)
	routeOptGroup.Get("/traffic/:route_id", handlers.GetRealTimeTraffic)
	routeOptGroup.Get("/recommendations/:route_id", handlers.GetRouteRecommendations)

	// External API Routes with caching
	externalGroup := app.Group("/api/external", auth.Middleware(), cacheMiddleware.Cache("external_api"))
	externalGroup.Post("/fuel-prices", handlers.GetFuelPricesAlongRoute)
	externalGroup.Post("/toll-costs", handlers.GetTollCosts)
	externalGroup.Post("/construction-alerts", handlers.GetConstructionAlerts)
	externalGroup.Post("/route-analysis", handlers.GetComprehensiveRouteAnalysis)

	// Machine Learning Routes with caching
	mlGroup := app.Group("/api/ml", auth.Middleware(), cacheMiddleware.Cache("ml_predictions"))
	mlGroup.Post("/sentiment-analysis", handlers.AnalyzeSentiment)
	mlGroup.Post("/predict-delay", handlers.PredictDeliveryDelay)
	mlGroup.Post("/predict-satisfaction", handlers.PredictCustomerSatisfaction)
	mlGroup.Post("/classify-text", handlers.ClassifyText)
	mlGroup.Post("/optimize-route", handlers.OptimizeRouteWithML)
	mlGroup.Post("/batch-analyze-feedback", handlers.BatchAnalyzeFeedback)
	mlGroup.Post("/operations-insights", handlers.GetOperationsMLInsights)

	// Tracking Routes (Phase 4 - Real-time tracking)
	trackingGroup := app.Group("/api/tracking", auth.Middleware())
	
	// Trip Tracking Endpoints
	trackingGroup.Post("/trips/:trip_id/location", handlers.UpdateTripLocation)
	trackingGroup.Get("/trips/:trip_id/current", handlers.GetCurrentTripLocation)
	trackingGroup.Get("/trips/:trip_id/history", handlers.GetTripTrackingHistory)
	trackingGroup.Put("/trips/:trip_id/status", handlers.UpdateTripStatus)
	trackingGroup.Get("/trips/:trip_id/eta", handlers.GetTripETA)
	trackingGroup.Get("/trips/:trip_id/status", handlers.GetTripTrackingStatus)
	trackingGroup.Get("/trips/:trip_id/events", handlers.GetTripTrackingEvents)
	
	// Load Tracking Endpoints
	trackingGroup.Get("/loads/:load_id", handlers.GetLoadTracking)
	trackingGroup.Get("/loads/:load_id/events", handlers.GetLoadTrackingEvents)
	trackingGroup.Put("/loads/:load_id/status", handlers.UpdateLoadStatus)
	trackingGroup.Get("/loads/:load_id/history", handlers.GetLoadTrackingHistory)
	
	// User-specific Tracking Endpoints
	trackingGroup.Get("/users/:user_id/active", handlers.GetUserActiveTrackings)
	trackingGroup.Get("/users/:user_id/shipper-view", handlers.GetShipperTrackingView)
	trackingGroup.Get("/users/:user_id/carrier-view", handlers.GetCarrierTrackingView)
	trackingGroup.Get("/users/:user_id/notifications", handlers.GetUserTrackingNotifications)
	
	// Mobile-optimized Tracking Endpoints
	mobileGroup := app.Group("/api/mobile")
	mobileGroup.Get("/trips/:trip_id/tracking", handlers.GetLightweightTracking)
	mobileGroup.Post("/trips/:trip_id/sync", auth.Middleware(), handlers.SyncOfflineData)
	mobileGroup.Get("/trips/:trip_id/battery-settings", handlers.GetBatteryOptimizedSettings)
	mobileGroup.Put("/users/:user_id/preferences", auth.Middleware(), handlers.UpdateMobileTrackingPreferences)
	mobileGroup.Get("/users/:user_id/tracking/summary", handlers.GetMobileTrackingSummary)
	
	// Analytics and Monitoring Endpoints
	monitoringGroup := app.Group("/api/monitoring", auth.Middleware())
	monitoringGroup.Get("/trips/:trip_id/analytics", handlers.GetTrackingAnalytics)
	monitoringGroup.Get("/tracking/health", handlers.GetSystemHealthMetrics)
	monitoringGroup.Get("/tracking/performance", handlers.GetTrackingPerformanceMetrics)
	monitoringGroup.Get("/tracking/data-quality", handlers.GetDataQualityReport)

	// Swagger
	app.Get("/swagger/*", swagger.HandlerDefault) // default
	app.Static("/docs", "./docs")                 // Serve swagger files directly
}
