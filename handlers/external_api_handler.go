package handlers

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"triplink/backend/services"
)

// External API handler for fuel, toll, and construction data

// @Summary Get fuel prices along route
// @Tags External APIs
// @Accept json
// @Produce json
// @Param request body services.RouteOptimizationRequest true "Route request"
// @Success 200 {array} services.FuelStation
// @Router /api/external/fuel-prices [post]
func GetFuelPricesAlongRoute(c *fiber.Ctx) error {
	var request services.RouteOptimizationRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Initialize fuel API service
	fuelService := services.NewFuelAPIService()

	// Convert route to waypoints (simplified - in real implementation would decode polyline)
	waypoints := []services.Coordinate{
		{Latitude: 34.0522, Longitude: -118.2437}, // Mock origin coordinates
		{Latitude: 33.9425, Longitude: -118.4081}, // Mock destination coordinates
	}

	// Add intermediate waypoints if provided
	for range request.Waypoints {
		// In real implementation, would geocode waypoint addresses
		waypoints = append(waypoints, services.Coordinate{
			Latitude:  34.0000,
			Longitude: -118.0000,
		})
	}

	stations, err := fuelService.GetFuelPricesAlongRoute(waypoints)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get fuel prices"})
	}

	return c.JSON(stations)
}

// @Summary Get toll costs for route
// @Tags External APIs
// @Accept json
// @Produce json
// @Param request body services.RouteOptimizationRequest true "Route request"
// @Success 200 {object} services.TollInfo
// @Router /api/external/toll-costs [post]
func GetTollCosts(c *fiber.Ctx) error {
	var request services.RouteOptimizationRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Initialize toll API service
	tollService := services.NewTollAPIService()

	tollInfo, err := tollService.GetTollRates(request.Origin, request.Destination)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get toll costs"})
	}

	return c.JSON(tollInfo)
}

// @Summary Get construction alerts along route
// @Tags External APIs
// @Accept json
// @Produce json
// @Param request body services.RouteOptimizationRequest true "Route request"
// @Success 200 {array} services.ConstructionAlert
// @Router /api/external/construction-alerts [post]
func GetConstructionAlerts(c *fiber.Ctx) error {
	var request services.RouteOptimizationRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Initialize DOT API service
	dotService := services.NewDOTAPIService()

	// Create bounding box from route (simplified)
	bounds := services.BoundingBox{
		NorthEast: services.Coordinate{Latitude: 35.0000, Longitude: -117.0000},
		SouthWest: services.Coordinate{Latitude: 33.0000, Longitude: -119.0000},
	}

	alerts, err := dotService.GetConstructionAlerts(bounds)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get construction alerts"})
	}

	return c.JSON(alerts)
}

// @Summary Get comprehensive route analysis with all external data
// @Tags External APIs
// @Accept json
// @Produce json
// @Param request body services.RouteOptimizationRequest true "Route request"
// @Success 200 {object} fiber.Map
// @Router /api/external/route-analysis [post]
func GetComprehensiveRouteAnalysis(c *fiber.Ctx) error {
	var request services.RouteOptimizationRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Initialize all external API services
	googleMaps, openWeather, _, fuelService, tollService, dotService := services.NewExternalAPIServices()

	// Get traffic conditions
	trafficInfo, err := googleMaps.GetTrafficConditions(request.Origin, request.Destination)
	if err != nil {
		trafficInfo = nil // Continue with other data
	}

	// Get weather conditions
	// Mock coordinates for weather lookup
	originCoord := services.Coordinate{Latitude: 34.0522, Longitude: -118.2437}
	weatherInfo, err := openWeather.GetCurrentWeather(originCoord.Latitude, originCoord.Longitude)
	if err != nil {
		weatherInfo = nil // Continue with other data
	}

	// Get fuel prices
	waypoints := []services.Coordinate{originCoord}
	fuelStations, err := fuelService.GetFuelPricesAlongRoute(waypoints)
	if err != nil {
		fuelStations = []services.FuelStation{} // Continue with empty array
	}

	// Get toll costs
	tollInfo, err := tollService.GetTollRates(request.Origin, request.Destination)
	if err != nil {
		tollInfo = nil // Continue with other data
	}

	// Get construction alerts
	bounds := services.BoundingBox{
		NorthEast: services.Coordinate{Latitude: 35.0000, Longitude: -117.0000},
		SouthWest: services.Coordinate{Latitude: 33.0000, Longitude: -119.0000},
	}
	constructionAlerts, err := dotService.GetConstructionAlerts(bounds)
	if err != nil {
		constructionAlerts = []services.ConstructionAlert{} // Continue with empty array
	}

	// Get road closures
	roadClosures, err := dotService.GetRoadClosures(bounds)
	if err != nil {
		roadClosures = []services.RoadClosure{} // Continue with empty array
	}

	// Compile comprehensive analysis response
	analysis := fiber.Map{
		"route": fiber.Map{
			"origin":      request.Origin,
			"destination": request.Destination,
			"waypoints":   request.Waypoints,
			"options":     request.Options,
		},
		"traffic": trafficInfo,
		"weather": weatherInfo,
		"fuel": fiber.Map{
			"stations_count": len(fuelStations),
			"cheapest_station": func() *services.FuelStation {
				if len(fuelStations) > 0 {
					cheapest := &fuelStations[0]
					for i := range fuelStations {
						if fuelStations[i].RegularPrice < cheapest.RegularPrice && fuelStations[i].RegularPrice > 0 {
							cheapest = &fuelStations[i]
						}
					}
					return cheapest
				}
				return nil
			}(),
			"all_stations": fuelStations,
		},
		"tolls": tollInfo,
		"construction": fiber.Map{
			"alerts_count":      len(constructionAlerts),
			"road_closures_count": len(roadClosures),
			"active_alerts":     constructionAlerts,
			"road_closures":     roadClosures,
		},
		"recommendations": generateRouteRecommendations(trafficInfo, weatherInfo, constructionAlerts, tollInfo),
		"risk_assessment": assessRouteRisk(trafficInfo, weatherInfo, constructionAlerts),
	}

	return c.JSON(analysis)
}

// Helper functions for comprehensive analysis
func generateRouteRecommendations(traffic *services.TrafficInfo, weather *services.WeatherCondition, construction []services.ConstructionAlert, tolls *services.TollInfo) []string {
	var recommendations []string

	// Traffic-based recommendations
	if traffic != nil {
		if traffic.CongestionLevel == "heavy" {
			recommendations = append(recommendations, "Heavy traffic detected - consider departing 30 minutes earlier")
		}
		if traffic.DelayMinutes > 20 {
			recommendations = append(recommendations, "Significant delays expected - explore alternative routes")
		}
	}

	// Weather-based recommendations
	if weather != nil {
		if weather.Condition == "rain" || weather.Precipitation > 0 {
			recommendations = append(recommendations, "Rainy conditions - allow extra time and drive safely")
		}
		if weather.WindSpeed > 50 {
			recommendations = append(recommendations, "High winds - use caution with high-profile vehicles")
		}
		if weather.Visibility < 5 {
			recommendations = append(recommendations, "Poor visibility - reduce speed and increase following distance")
		}
	}

	// Construction-based recommendations
	if len(construction) > 0 {
		recommendations = append(recommendations, "Active construction zones - expect delays and lane changes")
		for _, alert := range construction {
			if alert.Severity == "high" {
				recommendations = append(recommendations,
					fmt.Sprintf("Major construction on %s - consider alternative route", alert.RoadName))
			}
		}
	}

	// Toll-based recommendations
	if tolls != nil && tolls.TotalCost > 20 {
		recommendations = append(recommendations, "High toll costs - consider toll-free alternative routes")
	}

	// Default recommendation if no specific issues
	if len(recommendations) == 0 {
		recommendations = append(recommendations, "Route conditions are favorable - safe travels!")
	}

	return recommendations
}

func assessRouteRisk(traffic *services.TrafficInfo, weather *services.WeatherCondition, construction []services.ConstructionAlert) string {
	riskScore := 0

	// Traffic risk factors
	if traffic != nil {
		switch traffic.CongestionLevel {
		case "heavy":
			riskScore += 3
		case "moderate":
			riskScore += 2
		case "light":
			riskScore += 1
		}
		
		if traffic.DelayMinutes > 30 {
			riskScore += 2
		}
	}

	// Weather risk factors
	if weather != nil {
		if weather.Precipitation > 2.0 {
			riskScore += 2
		}
		if weather.WindSpeed > 60 {
			riskScore += 2
		}
		if weather.Visibility < 2 {
			riskScore += 3
		}
	}

	// Construction risk factors
	highSeverityCount := 0
	for _, alert := range construction {
		if alert.Severity == "high" {
			highSeverityCount++
		}
	}
	riskScore += highSeverityCount

	// Determine overall risk level
	if riskScore >= 7 {
		return "high"
	} else if riskScore >= 4 {
		return "moderate"
	} else {
		return "low"
	}
}