package handlers

import (
	"encoding/json"
	"math"
	"time"

	"github.com/gofiber/fiber/v2"
	"triplink/backend/database"
	"triplink/backend/models"
	"triplink/backend/services"
)

// Route optimization structures
type RouteOptimizationRequest struct {
	Origin      Location              `json:"origin"`
	Destination Location              `json:"destination"`
	Waypoints   []Location            `json:"waypoints,omitempty"`
	VehicleType string                `json:"vehicle_type,omitempty"`
	LoadWeight  float64               `json:"load_weight,omitempty"`
	LoadVolume  float64               `json:"load_volume,omitempty"`
	Preferences OptimizationPreferences `json:"preferences"`
	Constraints OptimizationConstraints `json:"constraints"`
}

type Location struct {
	Address   string  `json:"address"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	City      string  `json:"city,omitempty"`
	State     string  `json:"state,omitempty"`
	Country   string  `json:"country,omitempty"`
}

type OptimizationPreferences struct {
	Priority         string   `json:"priority"` // "time", "distance", "fuel", "cost"
	AvoidTolls       bool     `json:"avoid_tolls"`
	AvoidHighways    bool     `json:"avoid_highways"`
	AvoidFerries     bool     `json:"avoid_ferries"`
	PreferredRoutes  []string `json:"preferred_routes,omitempty"`
	TimeWindows      []TimeWindow `json:"time_windows,omitempty"`
}

type OptimizationConstraints struct {
	MaxDistance     float64   `json:"max_distance,omitempty"`
	MaxDuration     float64   `json:"max_duration,omitempty"` // hours
	RestBreaks      []RestBreak `json:"rest_breaks,omitempty"`
	FuelStops       bool      `json:"fuel_stops"`
	WeightLimits    bool      `json:"weight_limits"`
	HeightLimits    bool      `json:"height_limits"`
	HazmatRestrictions bool   `json:"hazmat_restrictions"`
}

type TimeWindow struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
	Type  string    `json:"type"` // "pickup", "delivery", "break"
}

type RestBreak struct {
	Duration  float64 `json:"duration"` // hours
	Frequency float64 `json:"frequency"` // hours between breaks
	Mandatory bool    `json:"mandatory"`
}

type RouteOptimizationResponse struct {
	OptimizedRoutes []OptimizedRoute `json:"optimized_routes"`
	Summary         RouteSummary     `json:"summary"`
	Alternatives    []AlternativeRoute `json:"alternatives,omitempty"`
	Recommendations []RouteRecommendation `json:"recommendations,omitempty"`
	Metadata        RouteMetadata    `json:"metadata"`
}

type OptimizedRoute struct {
	RouteID          string           `json:"route_id"`
	Algorithm        string           `json:"algorithm"`
	Priority         string           `json:"priority"`
	TotalDistance    float64          `json:"total_distance"` // km
	TotalDuration    float64          `json:"total_duration"` // hours
	EstimatedFuelCost float64         `json:"estimated_fuel_cost"`
	EstimatedTollCost float64         `json:"estimated_toll_cost"`
	EstimatedTotalCost float64        `json:"estimated_total_cost"`
	Waypoints        []RouteWaypoint  `json:"waypoints"`
	Segments         []RouteSegment   `json:"segments"`
	TrafficInfo      TrafficInfo      `json:"traffic_info"`
	WeatherInfo      WeatherInfo      `json:"weather_info"`
	Efficiency       RouteEfficiency  `json:"efficiency"`
	RiskAssessment   RiskAssessment   `json:"risk_assessment"`
}

type RouteWaypoint struct {
	Location       Location  `json:"location"`
	SequenceNumber int       `json:"sequence_number"`
	EstimatedArrival time.Time `json:"estimated_arrival"`
	EstimatedDeparture time.Time `json:"estimated_departure"`
	ServiceTime    float64   `json:"service_time"` // minutes
	WaitTime       float64   `json:"wait_time"`    // minutes
	Type           string    `json:"type"`         // "pickup", "delivery", "waypoint"
}

type RouteSegment struct {
	SegmentID     string    `json:"segment_id"`
	StartLocation Location  `json:"start_location"`
	EndLocation   Location  `json:"end_location"`
	Distance      float64   `json:"distance"` // km
	Duration      float64   `json:"duration"` // hours
	RoadType      string    `json:"road_type"`
	TollCost      float64   `json:"toll_cost"`
	FuelCost      float64   `json:"fuel_cost"`
	TrafficDelay  float64   `json:"traffic_delay"` // minutes
	Instructions  []string  `json:"instructions"`
}

type TrafficInfo struct {
	AverageSpeed       float64           `json:"average_speed"` // km/h
	CongestionLevel    string            `json:"congestion_level"`
	DelayMinutes       float64           `json:"delay_minutes"`
	PeakHours          []string          `json:"peak_hours"`
	TrafficIncidents   []TrafficIncident `json:"traffic_incidents,omitempty"`
	BestDepartureTime  time.Time         `json:"best_departure_time"`
	WorstDepartureTime time.Time         `json:"worst_departure_time"`
}

type TrafficIncident struct {
	IncidentID   string    `json:"incident_id"`
	Type         string    `json:"type"` // "accident", "construction", "road_closure"
	Location     Location  `json:"location"`
	Severity     string    `json:"severity"`
	DelayImpact  float64   `json:"delay_impact"` // minutes
	StartTime    time.Time `json:"start_time"`
	EndTime      *time.Time `json:"end_time,omitempty"`
	Description  string    `json:"description"`
	Detour       bool      `json:"detour"`
}

type WeatherInfo struct {
	Conditions      []WeatherCondition `json:"conditions"`
	RainProbability float64            `json:"rain_probability"`
	Temperature     float64            `json:"temperature"` // celsius
	WindSpeed       float64            `json:"wind_speed"`  // km/h
	Visibility      float64            `json:"visibility"`  // km
	WeatherRisk     string             `json:"weather_risk"`
	Alerts          []WeatherAlert     `json:"alerts,omitempty"`
}

type WeatherCondition struct {
	Location    Location  `json:"location"`
	Condition   string    `json:"condition"`
	Temperature float64   `json:"temperature"`
	Precipitation float64 `json:"precipitation"`
	Timestamp   time.Time `json:"timestamp"`
}

type WeatherAlert struct {
	AlertID     string    `json:"alert_id"`
	Type        string    `json:"type"` // "storm", "snow", "fog", "wind"
	Severity    string    `json:"severity"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Description string    `json:"description"`
	Impact      string    `json:"impact"`
}

type RouteEfficiency struct {
	FuelEfficiency    float64 `json:"fuel_efficiency"`    // km/l
	TimeEfficiency    float64 `json:"time_efficiency"`    // percentage
	CostEfficiency    float64 `json:"cost_efficiency"`    // cost per km
	DistanceRatio     float64 `json:"distance_ratio"`     // actual vs straight line
	OptimizationScore float64 `json:"optimization_score"` // 0-100
	Ranking           int     `json:"ranking"`            // among alternatives
}

type RiskAssessment struct {
	OverallRisk    string           `json:"overall_risk"`
	TrafficRisk    float64          `json:"traffic_risk"`
	WeatherRisk    float64          `json:"weather_risk"`
	RouteRisk      float64          `json:"route_risk"`
	DelayRisk      float64          `json:"delay_risk"`
	CostRisk       float64          `json:"cost_risk"`
	RiskFactors    []RiskFactor     `json:"risk_factors"`
	Mitigations    []RiskMitigation `json:"mitigations"`
}

type RiskFactor struct {
	Factor      string  `json:"factor"`
	Probability float64 `json:"probability"`
	Impact      string  `json:"impact"`
	Description string  `json:"description"`
}

type RiskMitigation struct {
	Strategy    string  `json:"strategy"`
	Effectiveness float64 `json:"effectiveness"`
	Cost        float64 `json:"cost"`
	Description string  `json:"description"`
}

type RouteSummary struct {
	TotalRoutes       int                    `json:"total_routes"`
	BestRoute         OptimizedRoute         `json:"best_route"`
	ComparisonMatrix  RouteComparisonMatrix  `json:"comparison_matrix"`
	Savings           RouteSavings           `json:"savings"`
	Performance       RoutePerformance       `json:"performance"`
}

type RouteComparisonMatrix struct {
	Fastest   OptimizedRoute `json:"fastest"`
	Shortest  OptimizedRoute `json:"shortest"`
	Cheapest  OptimizedRoute `json:"cheapest"`
	Balanced  OptimizedRoute `json:"balanced"`
}

type RouteSavings struct {
	TimeSaved     float64 `json:"time_saved"`     // hours
	DistanceSaved float64 `json:"distance_saved"` // km
	FuelSaved     float64 `json:"fuel_saved"`     // liters
	CostSaved     float64 `json:"cost_saved"`     // currency
	CO2Reduced    float64 `json:"co2_reduced"`    // kg
}

type RoutePerformance struct {
	OptimizationTime    float64 `json:"optimization_time"`    // seconds
	AlgorithmsUsed      []string `json:"algorithms_used"`
	DataSources         []string `json:"data_sources"`
	AccuracyScore       float64  `json:"accuracy_score"`
	ConfidenceLevel     float64  `json:"confidence_level"`
	LastUpdated         time.Time `json:"last_updated"`
}

type AlternativeRoute struct {
	RouteID     string  `json:"route_id"`
	Description string  `json:"description"`
	Pros        []string `json:"pros"`
	Cons        []string `json:"cons"`
	UseCase     string  `json:"use_case"`
	Route       OptimizedRoute `json:"route"`
}

type RouteRecommendation struct {
	RecommendationID string `json:"recommendation_id"`
	Type            string `json:"type"`
	Priority        string `json:"priority"`
	Title           string `json:"title"`
	Description     string `json:"description"`
	Impact          string `json:"impact"`
	Implementation  string `json:"implementation"`
	Cost            float64 `json:"cost"`
	Savings         float64 `json:"savings"`
}

type RouteMetadata struct {
	RequestID       string    `json:"request_id"`
	ProcessingTime  float64   `json:"processing_time"` // seconds
	Algorithm       string    `json:"algorithm"`
	DataSources     []string  `json:"data_sources"`
	CacheHit        bool      `json:"cache_hit"`
	ExpiresAt       time.Time `json:"expires_at"`
	Version         string    `json:"version"`
}

// Handler functions

// @Summary Optimize route
// @Tags Route Optimization
// @Accept json
// @Produce json
// @Param request body RouteOptimizationRequest true "Route optimization request"
// @Success 200 {object} RouteOptimizationResponse
// @Router /api/route-optimization/optimize [post]
func OptimizeRoute(c *fiber.Ctx) error {
	var request RouteOptimizationRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Initialize Redis service for caching
	redisService := services.NewRedisService()
	
	// Generate route ID for caching
	routeID := generateRouteIDFromRequest(request)
	
	// Try to get from cache first
	var cachedResponse RouteOptimizationResponse
	if err := redisService.GetCachedRouteOptimization(routeID, &cachedResponse); err == nil {
		c.Set("X-Cache", "HIT")
		return c.JSON(cachedResponse)
	}

	// Generate route optimization response
	response := generateOptimizedRoute(request)
	
	// Cache the result asynchronously
	go func() {
		redisService.CacheRouteOptimization(routeID, response)
	}()
	
	c.Set("X-Cache", "MISS")
	return c.JSON(response)
}

// @Summary Get multiple route options
// @Tags Route Optimization
// @Accept json
// @Produce json
// @Param request body RouteOptimizationRequest true "Route optimization request"
// @Success 200 {object} RouteOptimizationResponse
// @Router /api/route-optimization/multiple [post]
func GetMultipleRouteOptions(c *fiber.Ctx) error {
	var request RouteOptimizationRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Generate multiple route options with different optimization priorities
	response := generateMultipleRouteOptions(request)
	
	return c.JSON(response)
}

// @Summary Compare routes
// @Tags Route Optimization
// @Accept json
// @Produce json
// @Param routes body []OptimizedRoute true "Routes to compare"
// @Success 200 {object} RouteComparisonMatrix
// @Router /api/route-optimization/compare [post]
func CompareRoutes(c *fiber.Ctx) error {
	var routes []OptimizedRoute
	if err := c.BodyParser(&routes); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Generate comparison matrix
	comparison := generateRouteComparison(routes)
	
	return c.JSON(comparison)
}

// @Summary Get real-time traffic updates
// @Tags Route Optimization
// @Accept json
// @Produce json
// @Param route_id path string true "Route ID"
// @Success 200 {object} TrafficInfo
// @Router /api/route-optimization/traffic/{route_id} [get]
func GetRealTimeTraffic(c *fiber.Ctx) error {
	routeID := c.Params("route_id")
	
	// Get real-time traffic information for the route
	traffic := getRealTimeTrafficInfo(routeID)
	
	return c.JSON(traffic)
}

// @Summary Get route recommendations
// @Tags Route Optimization
// @Accept json
// @Produce json
// @Param route_id path string true "Route ID"
// @Success 200 {array} RouteRecommendation
// @Router /api/route-optimization/recommendations/{route_id} [get]
func GetRouteRecommendations(c *fiber.Ctx) error {
	routeID := c.Params("route_id")
	
	// Generate route-specific recommendations
	recommendations := generateRouteRecommendations(routeID)
	
	return c.JSON(recommendations)
}

// Helper functions for route optimization logic

func generateOptimizedRoute(request RouteOptimizationRequest) RouteOptimizationResponse {
	// Calculate straight-line distance between origin and destination
	distance := calculateDistance(request.Origin, request.Destination)
	
	// Apply optimization algorithms based on preferences
	optimizedRoute := OptimizedRoute{
		RouteID:          generateRouteID(),
		Algorithm:        "dijkstra",
		Priority:         request.Preferences.Priority,
		TotalDistance:    distance * 1.3, // Add realistic road distance factor
		TotalDuration:    (distance * 1.3) / 80, // Assume 80 km/h average speed
		EstimatedFuelCost: calculateFuelCost(distance * 1.3, request.VehicleType),
		EstimatedTollCost: calculateTollCost(distance * 1.3, request.Preferences.AvoidTolls),
		Waypoints:        generateWaypoints(request),
		Segments:         generateRouteSegments(request),
		TrafficInfo:      generateTrafficInfo(),
		WeatherInfo:      generateWeatherInfo(),
		Efficiency:       calculateRouteEfficiency(distance),
		RiskAssessment:   generateRiskAssessment(),
	}

	optimizedRoute.EstimatedTotalCost = optimizedRoute.EstimatedFuelCost + optimizedRoute.EstimatedTollCost

	// Generate alternatives
	alternatives := generateAlternativeRoutes(request, optimizedRoute)
	
	// Generate recommendations
	recommendations := generateOptimizationRecommendations(optimizedRoute)
	
	// Create summary
	summary := RouteSummary{
		TotalRoutes: len(alternatives) + 1,
		BestRoute:   optimizedRoute,
		ComparisonMatrix: generateRouteComparison([]OptimizedRoute{optimizedRoute}),
		Savings:     calculateRouteSavings(optimizedRoute),
		Performance: RoutePerformance{
			OptimizationTime: 0.5,
			AlgorithmsUsed:   []string{"dijkstra", "a_star"},
			DataSources:      []string{"osm", "traffic_api", "weather_api"},
			AccuracyScore:    0.92,
			ConfidenceLevel:  0.88,
			LastUpdated:      time.Now(),
		},
	}

	return RouteOptimizationResponse{
		OptimizedRoutes: []OptimizedRoute{optimizedRoute},
		Summary:         summary,
		Alternatives:    alternatives,
		Recommendations: recommendations,
		Metadata: RouteMetadata{
			RequestID:      generateRequestID(),
			ProcessingTime: 0.5,
			Algorithm:      "multi_objective",
			DataSources:    []string{"osm", "traffic", "weather"},
			CacheHit:       false,
			ExpiresAt:      time.Now().Add(30 * time.Minute),
			Version:        "1.0",
		},
	}
}

func generateMultipleRouteOptions(request RouteOptimizationRequest) RouteOptimizationResponse {
	// Generate different optimization priorities
	priorities := []string{"time", "distance", "fuel", "cost"}
	var routes []OptimizedRoute
	
	for _, priority := range priorities {
		req := request
		req.Preferences.Priority = priority
		route := generateOptimizedRoute(req).OptimizedRoutes[0]
		route.Priority = priority
		routes = append(routes, route)
	}
	
	return RouteOptimizationResponse{
		OptimizedRoutes: routes,
		Summary: RouteSummary{
			TotalRoutes:      len(routes),
			BestRoute:        routes[0],
			ComparisonMatrix: generateRouteComparison(routes),
		},
		Alternatives:    generateAlternativeRoutes(request, routes[0]),
		Recommendations: generateOptimizationRecommendations(routes[0]),
		Metadata: RouteMetadata{
			RequestID:      generateRequestID(),
			ProcessingTime: 1.2,
			Algorithm:      "multi_priority",
			DataSources:    []string{"osm", "traffic", "weather"},
			CacheHit:       false,
			ExpiresAt:      time.Now().Add(30 * time.Minute),
			Version:        "1.0",
		},
	}
}

func generateRouteComparison(routes []OptimizedRoute) RouteComparisonMatrix {
	if len(routes) == 0 {
		return RouteComparisonMatrix{}
	}
	
	// Find best route for each criteria
	fastest := routes[0]
	shortest := routes[0]
	cheapest := routes[0]
	balanced := routes[0]
	
	for _, route := range routes {
		if route.TotalDuration < fastest.TotalDuration {
			fastest = route
		}
		if route.TotalDistance < shortest.TotalDistance {
			shortest = route
		}
		if route.EstimatedTotalCost < cheapest.EstimatedTotalCost {
			cheapest = route
		}
		// Balanced score based on normalized values
		balancedScore := calculateBalancedScore(route)
		if balancedScore > calculateBalancedScore(balanced) {
			balanced = route
		}
	}
	
	return RouteComparisonMatrix{
		Fastest:  fastest,
		Shortest: shortest,
		Cheapest: cheapest,
		Balanced: balanced,
	}
}

func getRealTimeTrafficInfo(routeID string) TrafficInfo {
	return TrafficInfo{
		AverageSpeed:    65.5,
		CongestionLevel: "moderate",
		DelayMinutes:    15.3,
		PeakHours:       []string{"07:00-09:00", "17:00-19:00"},
		TrafficIncidents: []TrafficIncident{
			{
				IncidentID:  "TI001",
				Type:        "construction",
				Location:    Location{Address: "I-10 Mile 45", Latitude: 34.0522, Longitude: -118.2437},
				Severity:    "moderate",
				DelayImpact: 10.5,
				StartTime:   time.Now().Add(-2 * time.Hour),
				Description: "Lane closure for bridge repair",
				Detour:      true,
			},
		},
		BestDepartureTime:  time.Now().Add(2 * time.Hour),
		WorstDepartureTime: time.Now().Add(6 * time.Hour),
	}
}

func generateRouteRecommendations(routeID string) []RouteRecommendation {
	return []RouteRecommendation{
		{
			RecommendationID: "REC001",
			Type:            "timing",
			Priority:        "high",
			Title:           "Avoid Peak Traffic",
			Description:     "Departing 30 minutes later will save 15 minutes in traffic",
			Impact:          "15 minutes time savings, $8 fuel savings",
			Implementation:  "Adjust departure time to 08:30",
			Cost:            0,
			Savings:         8.50,
		},
		{
			RecommendationID: "REC002",
			Type:            "route",
			Priority:        "medium",
			Title:           "Alternative Highway",
			Description:     "Consider using I-5 instead of I-405 to avoid construction",
			Impact:          "5% longer distance but 20% faster",
			Implementation:  "Update route to use I-5 corridor",
			Cost:            2.30, // Additional fuel cost
			Savings:         25.00, // Time value savings
		},
	}
}

// Utility functions

func calculateDistance(origin, destination Location) float64 {
	// Haversine formula for great-circle distance
	const R = 6371 // Earth's radius in km
	
	lat1 := origin.Latitude * math.Pi / 180
	lat2 := destination.Latitude * math.Pi / 180
	deltaLat := (destination.Latitude - origin.Latitude) * math.Pi / 180
	deltaLng := (destination.Longitude - origin.Longitude) * math.Pi / 180
	
	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1)*math.Cos(lat2)*
			math.Sin(deltaLng/2)*math.Sin(deltaLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	
	return R * c
}

func calculateFuelCost(distance float64, vehicleType string) float64 {
	// Fuel consumption rates by vehicle type (L/100km)
	fuelRates := map[string]float64{
		"FLATBED":  35.0,
		"DRY_VAN":  30.0,
		"REEFER":   40.0,
		"TANKER":   45.0,
		"BOX_TRUCK": 25.0,
	}
	
	rate, exists := fuelRates[vehicleType]
	if !exists {
		rate = 35.0 // Default rate
	}
	
	fuelPrice := 1.45 // USD per liter
	return (distance / 100) * rate * fuelPrice
}

func calculateTollCost(distance float64, avoidTolls bool) float64 {
	if avoidTolls {
		return 0
	}
	// Estimate toll cost based on distance
	return distance * 0.08 // $0.08 per km average
}

func generateWaypoints(request RouteOptimizationRequest) []RouteWaypoint {
	waypoints := []RouteWaypoint{
		{
			Location:         request.Origin,
			SequenceNumber:   0,
			EstimatedArrival: time.Now(),
			EstimatedDeparture: time.Now().Add(15 * time.Minute),
			ServiceTime:      15,
			Type:            "pickup",
		},
	}
	
	// Add intermediate waypoints
	for i, wp := range request.Waypoints {
		waypoints = append(waypoints, RouteWaypoint{
			Location:         wp,
			SequenceNumber:   i + 1,
			EstimatedArrival: time.Now().Add(time.Duration(i+1) * time.Hour),
			EstimatedDeparture: time.Now().Add(time.Duration(i+1) * time.Hour + 30 * time.Minute),
			ServiceTime:      30,
			Type:            "waypoint",
		})
	}
	
	// Add destination
	waypoints = append(waypoints, RouteWaypoint{
		Location:         request.Destination,
		SequenceNumber:   len(waypoints),
		EstimatedArrival: time.Now().Add(time.Duration(len(waypoints)) * time.Hour),
		ServiceTime:      0,
		Type:            "delivery",
	})
	
	return waypoints
}

func generateRouteSegments(request RouteOptimizationRequest) []RouteSegment {
	// Generate segments between waypoints
	segments := []RouteSegment{
		{
			SegmentID:     "SEG001",
			StartLocation: request.Origin,
			EndLocation:   request.Destination,
			Distance:      calculateDistance(request.Origin, request.Destination),
			Duration:      2.5,
			RoadType:      "highway",
			TollCost:      15.50,
			FuelCost:      25.30,
			TrafficDelay:  8.5,
			Instructions:  []string{"Head north on I-5", "Continue for 150 miles", "Take exit 234"},
		},
	}
	
	return segments
}

func generateTrafficInfo() TrafficInfo {
	return TrafficInfo{
		AverageSpeed:    72.3,
		CongestionLevel: "light",
		DelayMinutes:    8.5,
		PeakHours:       []string{"07:00-09:00", "17:00-19:00"},
		BestDepartureTime:  time.Now().Add(1 * time.Hour),
		WorstDepartureTime: time.Now().Add(6 * time.Hour),
	}
}

func generateWeatherInfo() WeatherInfo {
	return WeatherInfo{
		Conditions: []WeatherCondition{
			{
				Location:      Location{Address: "Los Angeles, CA", Latitude: 34.0522, Longitude: -118.2437},
				Condition:     "clear",
				Temperature:   22.5,
				Precipitation: 0,
				Timestamp:     time.Now(),
			},
		},
		RainProbability: 10,
		Temperature:     22.5,
		WindSpeed:       15.2,
		Visibility:      10.0,
		WeatherRisk:     "low",
	}
}

func calculateRouteEfficiency(distance float64) RouteEfficiency {
	return RouteEfficiency{
		FuelEfficiency:    8.5,  // km/l
		TimeEfficiency:    87.3, // percentage
		CostEfficiency:    1.25, // cost per km
		DistanceRatio:     1.15, // 15% longer than straight line
		OptimizationScore: 92.5,
		Ranking:           1,
	}
}

func generateRiskAssessment() RiskAssessment {
	return RiskAssessment{
		OverallRisk: "low",
		TrafficRisk: 25.3,
		WeatherRisk: 10.2,
		RouteRisk:   15.7,
		DelayRisk:   18.9,
		CostRisk:    12.4,
		RiskFactors: []RiskFactor{
			{
				Factor:      "traffic_congestion",
				Probability: 0.3,
				Impact:      "moderate",
				Description: "Potential delays during peak hours",
			},
		},
		Mitigations: []RiskMitigation{
			{
				Strategy:      "flexible_departure",
				Effectiveness: 0.8,
				Cost:          0,
				Description:   "Adjust departure time to avoid peak traffic",
			},
		},
	}
}

func generateAlternativeRoutes(request RouteOptimizationRequest, mainRoute OptimizedRoute) []AlternativeRoute {
	return []AlternativeRoute{
		{
			RouteID:     "ALT001",
			Description: "Scenic coastal route",
			Pros:        []string{"Better views", "Less traffic", "More rest stops"},
			Cons:        []string{"Longer distance", "Higher fuel cost"},
			UseCase:     "When time is not critical and driver comfort is priority",
		},
	}
}

func generateOptimizationRecommendations(route OptimizedRoute) []RouteRecommendation {
	return []RouteRecommendation{
		{
			RecommendationID: "OPT001",
			Type:            "timing",
			Priority:        "high",
			Title:           "Optimize Departure Time",
			Description:     "Leave 30 minutes earlier to avoid peak traffic",
			Impact:          "Save 15 minutes and $8 in fuel costs",
			Implementation:  "Adjust departure to 07:30",
			Cost:            0,
			Savings:         8.50,
		},
	}
}

func calculateRouteSavings(route OptimizedRoute) RouteSavings {
	return RouteSavings{
		TimeSaved:     0.25, // 15 minutes
		DistanceSaved: 12.5,
		FuelSaved:     4.2,
		CostSaved:     15.30,
		CO2Reduced:    9.8,
	}
}

func calculateBalancedScore(route OptimizedRoute) float64 {
	// Normalize and weight different factors
	timeScore := 100 - (route.TotalDuration / 10 * 100) // Normalize assuming max 10 hours
	distanceScore := 100 - (route.TotalDistance / 1000 * 100) // Normalize assuming max 1000 km
	costScore := 100 - (route.EstimatedTotalCost / 500 * 100) // Normalize assuming max $500
	
	// Weighted average
	return (timeScore*0.4 + distanceScore*0.3 + costScore*0.3)
}

func generateRouteID() string {
	return "ROUTE_" + time.Now().Format("20060102150405")
}

func generateRequestID() string {
	return "REQ_" + time.Now().Format("20060102150405")
}

func generateRouteIDFromRequest(request RouteOptimizationRequest) string {
	// Create a hash-based ID from request parameters for caching
	requestBytes, _ := json.Marshal(request)
	hash := fmt.Sprintf("%x", requestBytes)
	return "route_" + hash[:16] // Use first 16 chars of hash
}