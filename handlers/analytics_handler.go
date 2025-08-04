package handlers

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"triplink/backend/database"
	"triplink/backend/models"
	"triplink/backend/services"
)

// Analytics request/response structures
type AnalyticsFilters struct {
	DateRange   *DateRange `json:"date_range,omitempty"`
	VehicleIDs  []uint     `json:"vehicle_ids,omitempty"`
	DriverIDs   []uint     `json:"driver_ids,omitempty"`
	RouteIDs    []uint     `json:"route_ids,omitempty"`
	CustomerIDs []uint     `json:"customer_ids,omitempty"`
}

type DateRange struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

// On-Time Delivery Analytics
type OnTimeDeliveryMetrics struct {
	TotalDeliveries     int     `json:"total_deliveries"`
	OnTimeDeliveries    int     `json:"on_time_deliveries"`
	OnTimePercentage    float64 `json:"on_time_percentage"`
	AverageDelay        float64 `json:"average_delay"` // minutes
	EarlyDeliveries     int     `json:"early_deliveries"`
	LateDeliveries      int     `json:"late_deliveries"`
	CriticalDelays      int     `json:"critical_delays"`
	AverageDeliveryTime float64 `json:"average_delivery_time"` // hours
	OnTimeImprovement   float64 `json:"on_time_improvement"`   // percentage change
}

type DeliveryPerformanceByRoute struct {
	RouteID          string  `json:"route_id"`
	RouteName        string  `json:"route_name"`
	Origin           string  `json:"origin"`
	Destination      string  `json:"destination"`
	Distance         float64 `json:"distance"`
	TotalDeliveries  int     `json:"total_deliveries"`
	OnTimeDeliveries int     `json:"on_time_deliveries"`
	OnTimePercentage float64 `json:"on_time_percentage"`
	AverageDelay     float64 `json:"average_delay"`
	Improvement      float64 `json:"improvement"`
	RiskLevel        string  `json:"risk_level"`
}

type DeliveryPerformanceByDriver struct {
	DriverID             string  `json:"driver_id"`
	DriverName           string  `json:"driver_name"`
	TotalDeliveries      int     `json:"total_deliveries"`
	OnTimeDeliveries     int     `json:"on_time_deliveries"`
	OnTimePercentage     float64 `json:"on_time_percentage"`
	AverageDelay         float64 `json:"average_delay"`
	EarlyDeliveries      int     `json:"early_deliveries"`
	LateDeliveries       int     `json:"late_deliveries"`
	PerformanceRating    float64 `json:"performance_rating"`
	ImprovementTrend     string  `json:"improvement_trend"`
	TrainingRecommended  bool    `json:"training_recommended"`
}

// Customer Satisfaction Analytics
type CustomerSatisfactionMetrics struct {
	OverallScore              float64 `json:"overall_score"`
	TotalResponses            int     `json:"total_responses"`
	ResponseRate              float64 `json:"response_rate"`
	ScoreImprovement          float64 `json:"score_improvement"`
	NPSScore                  int     `json:"nps_score"`
	SatisfiedCustomers        int     `json:"satisfied_customers"`
	NeutralCustomers          int     `json:"neutral_customers"`
	DissatisfiedCustomers     int     `json:"dissatisfied_customers"`
	RetentionRate             float64 `json:"retention_rate"`
	ChurnRate                 float64 `json:"churn_rate"`
	AverageResponseTime       float64 `json:"average_response_time"`
	ResolutionRate            float64 `json:"resolution_rate"`
}

type CustomerFeedback struct {
	FeedbackID     string    `json:"feedback_id"`
	CustomerID     string    `json:"customer_id"`
	CustomerName   string    `json:"customer_name"`
	TripID         string    `json:"trip_id"`
	RouteName      string    `json:"route_name"`
	DriverName     string    `json:"driver_name"`
	Rating         int       `json:"rating"`
	Category       string    `json:"category"`
	Sentiment      string    `json:"sentiment"`
	Comment        string    `json:"comment"`
	SubmittedAt    time.Time `json:"submitted_at"`
	Status         string    `json:"status"`
	Priority       string    `json:"priority"`
	ResponseTime   float64   `json:"response_time"`
	Resolved       bool      `json:"resolved"`
	FollowUpRequired bool    `json:"follow_up_required"`
}

// Load Matching Analytics
type LoadMatchingMetrics struct {
	TotalLoads              int     `json:"total_loads"`
	MatchedLoads            int     `json:"matched_loads"`
	MatchingRate            float64 `json:"matching_rate"`
	AverageMatchTime        float64 `json:"average_match_time"` // hours
	OptimalMatches          int     `json:"optimal_matches"`
	SuboptimalMatches       int     `json:"suboptimal_matches"`
	UnmatchedLoads          int     `json:"unmatched_loads"`
	UtilizationRate         float64 `json:"utilization_rate"`
	RevenueEfficiency       float64 `json:"revenue_efficiency"`
	CostEfficiency          float64 `json:"cost_efficiency"`
	ImprovementOpportunity  float64 `json:"improvement_opportunity"`
}

type LoadMatchingData struct {
	LoadID           string    `json:"load_id"`
	LoadType         string    `json:"load_type"`
	Origin           string    `json:"origin"`
	Destination      string    `json:"destination"`
	Weight           float64   `json:"weight"`
	Volume           float64   `json:"volume"`
	Revenue          float64   `json:"revenue"`
	Distance         float64   `json:"distance"`
	PickupDate       time.Time `json:"pickup_date"`
	DeliveryDate     time.Time `json:"delivery_date"`
	Status           string    `json:"status"`
	MatchedVehicleID *string   `json:"matched_vehicle_id,omitempty"`
	MatchedDriverID  *string   `json:"matched_driver_id,omitempty"`
	MatchScore       float64   `json:"match_score"`
	MatchTime        float64   `json:"match_time"`
	MatchQuality     string    `json:"match_quality"`
	CustomerPriority string    `json:"customer_priority"`
	ProfitMargin     float64   `json:"profit_margin"`
}

// Capacity Utilization Analytics
type CapacityMetrics struct {
	TotalCapacity         float64 `json:"total_capacity"`
	UtilizedCapacity      float64 `json:"utilized_capacity"`
	UtilizationRate       float64 `json:"utilization_rate"`
	AvailableCapacity     float64 `json:"available_capacity"`
	PeakUtilization       float64 `json:"peak_utilization"`
	AverageUtilization    float64 `json:"average_utilization"`
	CapacityEfficiency    float64 `json:"capacity_efficiency"`
	RevenuePerCapacityUnit float64 `json:"revenue_per_capacity_unit"`
	CostPerCapacityUnit   float64 `json:"cost_per_capacity_unit"`
	CapacityTrend         string  `json:"capacity_trend"`
	DemandVsCapacity      float64 `json:"demand_vs_capacity"`
	ForecastedDemand      float64 `json:"forecasted_demand"`
}

type VehicleCapacityData struct {
	VehicleID          string    `json:"vehicle_id"`
	VehicleNumber      string    `json:"vehicle_number"`
	VehicleType        string    `json:"vehicle_type"`
	MaxCapacity        float64   `json:"max_capacity"`
	MaxVolume          float64   `json:"max_volume"`
	CurrentLoad        float64   `json:"current_load"`
	CurrentVolume      float64   `json:"current_volume"`
	WeightUtilization  float64   `json:"weight_utilization"`
	VolumeUtilization  float64   `json:"volume_utilization"`
	OverallUtilization float64   `json:"overall_utilization"`
	Trips              int       `json:"trips"`
	Revenue            float64   `json:"revenue"`
	OperatingCosts     float64   `json:"operating_costs"`
	Profitability      float64   `json:"profitability"`
	Efficiency         string    `json:"efficiency"`
	Location           string    `json:"location"`
	Status             string    `json:"status"`
	LastUpdated        time.Time `json:"last_updated"`
	UtilizationTrend   string    `json:"utilization_trend"`
}

// Delay Analysis Analytics
type DelayMetrics struct {
	TotalDelays            int     `json:"total_delays"`
	AverageDelayDuration   float64 `json:"average_delay_duration"` // minutes
	DelayFrequency         float64 `json:"delay_frequency"`        // delays per 100 trips
	TotalDelayTime         float64 `json:"total_delay_time"`       // total minutes lost
	CostOfDelays           float64 `json:"cost_of_delays"`
	CustomerImpact         float64 `json:"customer_impact"`
	OnTimePerformance      float64 `json:"on_time_performance"`
	DelayTrend             string  `json:"delay_trend"`
	RecurrentDelays        int     `json:"recurrent_delays"`
	PreventableDelays      int     `json:"preventable_delays"`
	MitigatedDelays        int     `json:"mitigated_delays"`
	ImprovementOpportunity float64 `json:"improvement_opportunity"`
}

type DelayIncident struct {
	IncidentID         string    `json:"incident_id"`
	TripID             string    `json:"trip_id"`
	VehicleID          string    `json:"vehicle_id"`
	DriverID           string    `json:"driver_id"`
	RouteID            string    `json:"route_id"`
	Origin             string    `json:"origin"`
	Destination        string    `json:"destination"`
	ScheduledDeparture time.Time `json:"scheduled_departure"`
	ActualDeparture    time.Time `json:"actual_departure"`
	ScheduledArrival   time.Time `json:"scheduled_arrival"`
	ActualArrival      time.Time `json:"actual_arrival"`
	DelayDuration      float64   `json:"delay_duration"` // minutes
	DelayType          string    `json:"delay_type"`
	Severity           string    `json:"severity"`
	CustomerNotified   bool      `json:"customer_notified"`
	ResolutionTime     float64   `json:"resolution_time"`
	Preventable        bool      `json:"preventable"`
	Recurrent          bool      `json:"recurrent"`
	CostImpact         float64   `json:"cost_impact"`
	CustomerImpact     string    `json:"customer_impact"`
	Status             string    `json:"status"`
	ReportedAt         time.Time `json:"reported_at"`
	ResolvedAt         *time.Time `json:"resolved_at,omitempty"`
}

// Handler functions

// @Summary Get on-time delivery analytics
// @Tags Analytics
// @Accept json
// @Produce json
// @Param filters body AnalyticsFilters true "Analytics filters"
// @Success 200 {object} OnTimeDeliveryMetrics
// @Router /api/analytics/on-time-delivery [post]
func GetOnTimeDeliveryAnalytics(c *fiber.Ctx) error {
	var filters AnalyticsFilters
	if err := c.BodyParser(&filters); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Initialize Redis service for caching
	redisService := services.NewRedisService()
	
	// Generate cache key based on filters
	filterHash := generateFilterHash(filters)
	
	// Try to get from cache first
	var cachedMetrics OnTimeDeliveryMetrics
	if err := redisService.GetCachedAnalyticsResult("on_time_delivery", filterHash, &cachedMetrics); err == nil {
		c.Set("X-Cache", "HIT")
		return c.JSON(cachedMetrics)
	}

	// Build query based on filters
	query := database.DB.Table("trips")
	
	if filters.DateRange != nil {
		if filters.DateRange.Start != "" {
			query = query.Where("departure_date >= ?", filters.DateRange.Start)
		}
		if filters.DateRange.End != "" {
			query = query.Where("departure_date <= ?", filters.DateRange.End)
		}
	}

	if len(filters.VehicleIDs) > 0 {
		query = query.Where("vehicle_id IN ?", filters.VehicleIDs)
	}

	// Calculate metrics from database
	var totalDeliveries int64
	var onTimeDeliveries int64
	var lateDeliveries int64
	var earlyDeliveries int64

	// Get total completed trips
	query.Where("status = ?", "COMPLETED").Count(&totalDeliveries)

	// Count on-time deliveries (delivered within 30 minutes of estimated arrival)
	database.DB.Table("trips").
		Where("status = ?", "COMPLETED").
		Where("ABS(EXTRACT(EPOCH FROM (actual_arrival - estimated_arrival))/60) <= 30").
		Count(&onTimeDeliveries)

	// Count late deliveries
	database.DB.Table("trips").
		Where("status = ?", "COMPLETED").
		Where("actual_arrival > estimated_arrival + INTERVAL '30 minutes'").
		Count(&lateDeliveries)

	// Count early deliveries
	database.DB.Table("trips").
		Where("status = ?", "COMPLETED").
		Where("actual_arrival < estimated_arrival - INTERVAL '30 minutes'").
		Count(&earlyDeliveries)

	// Calculate on-time percentage
	var onTimePercentage float64
	if totalDeliveries > 0 {
		onTimePercentage = (float64(onTimeDeliveries) / float64(totalDeliveries)) * 100
	}

	// Calculate average delay
	var avgDelayResult struct {
		AverageDelay float64
	}
	database.DB.Raw(`
		SELECT AVG(EXTRACT(EPOCH FROM (actual_arrival - estimated_arrival))/60) as average_delay
		FROM trips 
		WHERE status = 'COMPLETED' 
		AND actual_arrival > estimated_arrival
	`).Scan(&avgDelayResult)

	metrics := OnTimeDeliveryMetrics{
		TotalDeliveries:     int(totalDeliveries),
		OnTimeDeliveries:    int(onTimeDeliveries),
		OnTimePercentage:    onTimePercentage,
		AverageDelay:        avgDelayResult.AverageDelay,
		EarlyDeliveries:     int(earlyDeliveries),
		LateDeliveries:      int(lateDeliveries),
		CriticalDelays:      0, // Calculate critical delays (>2 hours)
		AverageDeliveryTime: 0, // Calculate from trip duration
		OnTimeImprovement:   0, // Calculate trend
	}

	// Cache the result
	go func() {
		redisService.CacheAnalyticsResult("on_time_delivery", filterHash, metrics)
	}()
	
	c.Set("X-Cache", "MISS")
	return c.JSON(metrics)
}

// @Summary Get delivery performance by route
// @Tags Analytics
// @Accept json
// @Produce json
// @Param filters body AnalyticsFilters true "Analytics filters"
// @Success 200 {array} DeliveryPerformanceByRoute
// @Router /api/analytics/delivery-performance/by-route [post]
func GetDeliveryPerformanceByRoute(c *fiber.Ctx) error {
	var filters AnalyticsFilters
	if err := c.BodyParser(&filters); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Mock data for now - replace with actual database queries
	routes := []DeliveryPerformanceByRoute{
		{
			RouteID:          "RT001",
			RouteName:        "LA-Phoenix Express",
			Origin:           "Los Angeles, CA",
			Destination:      "Phoenix, AZ",
			Distance:         372,
			TotalDeliveries:  95,
			OnTimeDeliveries: 87,
			OnTimePercentage: 91.6,
			AverageDelay:     12,
			Improvement:      3.2,
			RiskLevel:        "low",
		},
		{
			RouteID:          "RT002",
			RouteName:        "SF-Sacramento",
			Origin:           "San Francisco, CA",
			Destination:      "Sacramento, CA",
			Distance:         87,
			TotalDeliveries:  78,
			OnTimeDeliveries: 65,
			OnTimePercentage: 83.3,
			AverageDelay:     28,
			Improvement:      -1.8,
			RiskLevel:        "medium",
		},
	}

	return c.JSON(routes)
}

// @Summary Get delivery performance by driver
// @Tags Analytics
// @Accept json
// @Produce json
// @Param filters body AnalyticsFilters true "Analytics filters"
// @Success 200 {array} DeliveryPerformanceByDriver
// @Router /api/analytics/delivery-performance/by-driver [post]
func GetDeliveryPerformanceByDriver(c *fiber.Ctx) error {
	var filters AnalyticsFilters
	if err := c.BodyParser(&filters); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Query database for driver performance
	var drivers []DeliveryPerformanceByDriver

	// Get users who are drivers (carriers) and their trip performance
	rows, err := database.DB.Raw(`
		SELECT 
			u.id as driver_id,
			CONCAT(u.first_name, ' ', u.last_name) as driver_name,
			COUNT(t.id) as total_deliveries,
			COUNT(CASE WHEN ABS(EXTRACT(EPOCH FROM (t.actual_arrival - t.estimated_arrival))/60) <= 30 THEN 1 END) as on_time_deliveries,
			AVG(CASE WHEN t.actual_arrival > t.estimated_arrival THEN EXTRACT(EPOCH FROM (t.actual_arrival - t.estimated_arrival))/60 ELSE 0 END) as average_delay,
			COUNT(CASE WHEN t.actual_arrival < t.estimated_arrival - INTERVAL '30 minutes' THEN 1 END) as early_deliveries,
			COUNT(CASE WHEN t.actual_arrival > t.estimated_arrival + INTERVAL '30 minutes' THEN 1 END) as late_deliveries,
			u.rating as performance_rating
		FROM users u
		JOIN trips t ON u.id = t.user_id
		WHERE u.role = 'CARRIER' AND t.status = 'COMPLETED'
		GROUP BY u.id, u.first_name, u.last_name, u.rating
	`).Rows()

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database query failed"})
	}
	defer rows.Close()

	for rows.Next() {
		var driver DeliveryPerformanceByDriver
		var totalDeliveries, onTimeDeliveries, earlyDeliveries, lateDeliveries int
		var averageDelay, performanceRating float64
		var driverName string
		var driverID string

		err := rows.Scan(
			&driverID, &driverName, &totalDeliveries, &onTimeDeliveries,
			&averageDelay, &earlyDeliveries, &lateDeliveries, &performanceRating,
		)
		if err != nil {
			continue
		}

		var onTimePercentage float64
		if totalDeliveries > 0 {
			onTimePercentage = (float64(onTimeDeliveries) / float64(totalDeliveries)) * 100
		}

		driver = DeliveryPerformanceByDriver{
			DriverID:             driverID,
			DriverName:           driverName,
			TotalDeliveries:      totalDeliveries,
			OnTimeDeliveries:     onTimeDeliveries,
			OnTimePercentage:     onTimePercentage,
			AverageDelay:         averageDelay,
			EarlyDeliveries:      earlyDeliveries,
			LateDeliveries:       lateDeliveries,
			PerformanceRating:    performanceRating,
			ImprovementTrend:     "stable", // Calculate based on historical data
			TrainingRecommended:  onTimePercentage < 80,
		}

		drivers = append(drivers, driver)
	}

	return c.JSON(drivers)
}

// @Summary Get customer satisfaction analytics
// @Tags Analytics
// @Accept json
// @Produce json
// @Param filters body AnalyticsFilters true "Analytics filters"
// @Success 200 {object} CustomerSatisfactionMetrics
// @Router /api/analytics/customer-satisfaction [post]
func GetCustomerSatisfactionAnalytics(c *fiber.Ctx) error {
	var filters AnalyticsFilters
	if err := c.BodyParser(&filters); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Calculate satisfaction metrics from reviews table
	var totalReviews int64
	var avgRating float64
	var satisfiedCount, neutralCount, dissatisfiedCount int64

	database.DB.Model(&models.Review{}).Count(&totalReviews)

	database.DB.Model(&models.Review{}).Select("AVG(rating)").Row().Scan(&avgRating)

	database.DB.Model(&models.Review{}).Where("rating >= 4").Count(&satisfiedCount)
	database.DB.Model(&models.Review{}).Where("rating = 3").Count(&neutralCount)
	database.DB.Model(&models.Review{}).Where("rating <= 2").Count(&dissatisfiedCount)

	// Calculate NPS score (simplified)
	var npsScore int
	if totalReviews > 0 {
		promoters := float64(satisfiedCount) / float64(totalReviews) * 100
		detractors := float64(dissatisfiedCount) / float64(totalReviews) * 100
		npsScore = int(promoters - detractors)
	}

	metrics := CustomerSatisfactionMetrics{
		OverallScore:          avgRating,
		TotalResponses:        int(totalReviews),
		ResponseRate:          75.0, // Mock data - calculate based on trips vs reviews
		ScoreImprovement:      0.3,  // Mock data - calculate trend
		NPSScore:              npsScore,
		SatisfiedCustomers:    int(satisfiedCount),
		NeutralCustomers:      int(neutralCount),
		DissatisfiedCustomers: int(dissatisfiedCount),
		RetentionRate:         94.2, // Mock data
		ChurnRate:             5.8,  // Mock data
		AverageResponseTime:   2.4,  // Mock data
		ResolutionRate:        87.3, // Mock data
	}

	return c.JSON(metrics)
}

// @Summary Get load matching efficiency analytics
// @Tags Analytics
// @Accept json
// @Produce json
// @Param filters body AnalyticsFilters true "Analytics filters"
// @Success 200 {object} LoadMatchingMetrics
// @Router /api/analytics/load-matching [post]
func GetLoadMatchingAnalytics(c *fiber.Ctx) error {
	var filters AnalyticsFilters
	if err := c.BodyParser(&filters); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Calculate load matching metrics
	var totalLoads, matchedLoads int64

	query := database.DB.Model(&models.Load{})
	
	if filters.DateRange != nil {
		if filters.DateRange.Start != "" {
			query = query.Where("created_at >= ?", filters.DateRange.Start)
		}
		if filters.DateRange.End != "" {
			query = query.Where("created_at <= ?", filters.DateRange.End)
		}
	}

	query.Count(&totalLoads)
	query.Where("trip_id IS NOT NULL").Count(&matchedLoads)

	var matchingRate float64
	if totalLoads > 0 {
		matchingRate = (float64(matchedLoads) / float64(totalLoads)) * 100
	}

	metrics := LoadMatchingMetrics{
		TotalLoads:             int(totalLoads),
		MatchedLoads:           int(matchedLoads),
		MatchingRate:           matchingRate,
		AverageMatchTime:       3.4,  // Mock data - calculate from load creation to trip assignment
		OptimalMatches:         int(matchedLoads * 60 / 100), // Mock: 60% optimal
		SuboptimalMatches:      int(matchedLoads * 40 / 100), // Mock: 40% suboptimal
		UnmatchedLoads:         int(totalLoads - matchedLoads),
		UtilizationRate:        78.3, // Mock data
		RevenueEfficiency:      87.5, // Mock data
		CostEfficiency:         82.1, // Mock data
		ImprovementOpportunity: 15.8, // Mock data
	}

	return c.JSON(metrics)
}

// @Summary Get capacity utilization analytics
// @Tags Analytics
// @Accept json
// @Produce json
// @Param filters body AnalyticsFilters true "Analytics filters"
// @Success 200 {object} CapacityMetrics
// @Router /api/analytics/capacity-utilization [post]
func GetCapacityUtilizationAnalytics(c *fiber.Ctx) error {
	var filters AnalyticsFilters
	if err := c.BodyParser(&filters); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Calculate capacity metrics from vehicles and trips
	var totalCapacity, utilizedCapacity float64

	// Get total fleet capacity
	database.DB.Model(&models.Vehicle{}).
		Where("is_active = ?", true).
		Select("SUM(load_capacity_kg)").
		Row().Scan(&totalCapacity)

	// Calculate current utilization from active trips
	database.DB.Raw(`
		SELECT COALESCE(SUM(t.used_weight), 0)
		FROM trips t 
		JOIN vehicles v ON t.vehicle_id = v.id
		WHERE t.status IN ('ACTIVE', 'IN_TRANSIT') AND v.is_active = true
	`).Row().Scan(&utilizedCapacity)

	var utilizationRate float64
	if totalCapacity > 0 {
		utilizationRate = (utilizedCapacity / totalCapacity) * 100
	}

	metrics := CapacityMetrics{
		TotalCapacity:          totalCapacity,
		UtilizedCapacity:       utilizedCapacity,
		UtilizationRate:        utilizationRate,
		AvailableCapacity:      totalCapacity - utilizedCapacity,
		PeakUtilization:        94.2, // Mock data - calculate from historical data
		AverageUtilization:     76.8, // Mock data
		CapacityEfficiency:     82.1, // Mock data
		RevenuePerCapacityUnit: 1.85, // Mock data
		CostPerCapacityUnit:    1.32, // Mock data
		CapacityTrend:          "increasing",
		DemandVsCapacity:       85.7, // Mock data
		ForecastedDemand:       1920000, // Mock data
	}

	return c.JSON(metrics)
}

// @Summary Get delay analysis analytics
// @Tags Analytics
// @Accept json
// @Produce json
// @Param filters body AnalyticsFilters true "Analytics filters"
// @Success 200 {object} DelayMetrics
// @Router /api/analytics/delay-analysis [post]
func GetDelayAnalysisAnalytics(c *fiber.Ctx) error {
	var filters AnalyticsFilters
	if err := c.BodyParser(&filters); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Calculate delay metrics
	var totalDelays int64
	var avgDelayDuration float64
	var totalDelayTime float64

	query := database.DB.Table("trips").Where("status = 'COMPLETED'")
	
	if filters.DateRange != nil {
		if filters.DateRange.Start != "" {
			query = query.Where("departure_date >= ?", filters.DateRange.Start)
		}
		if filters.DateRange.End != "" {
			query = query.Where("departure_date <= ?", filters.DateRange.End)
		}
	}

	// Count trips with delays
	query.Where("actual_arrival > estimated_arrival + INTERVAL '30 minutes'").Count(&totalDelays)

	// Calculate average delay duration
	database.DB.Raw(`
		SELECT 
			AVG(EXTRACT(EPOCH FROM (actual_arrival - estimated_arrival))/60) as avg_delay,
			SUM(EXTRACT(EPOCH FROM (actual_arrival - estimated_arrival))/60) as total_delay
		FROM trips 
		WHERE status = 'COMPLETED' 
		AND actual_arrival > estimated_arrival + INTERVAL '30 minutes'
	`).Row().Scan(&avgDelayDuration, &totalDelayTime)

	// Calculate delay frequency (delays per 100 trips)
	var totalTrips int64
	database.DB.Model(&models.Trip{}).Where("status = 'COMPLETED'").Count(&totalTrips)
	
	var delayFrequency float64
	if totalTrips > 0 {
		delayFrequency = (float64(totalDelays) / float64(totalTrips)) * 100
	}

	// Calculate on-time performance
	var onTimeTrips int64
	database.DB.Table("trips").
		Where("status = 'COMPLETED'").
		Where("ABS(EXTRACT(EPOCH FROM (actual_arrival - estimated_arrival))/60) <= 30").
		Count(&onTimeTrips)

	var onTimePerformance float64
	if totalTrips > 0 {
		onTimePerformance = (float64(onTimeTrips) / float64(totalTrips)) * 100
	}

	metrics := DelayMetrics{
		TotalDelays:            int(totalDelays),
		AverageDelayDuration:   avgDelayDuration,
		DelayFrequency:         delayFrequency,
		TotalDelayTime:         totalDelayTime,
		CostOfDelays:           85000, // Mock data - calculate based on delay cost
		CustomerImpact:         23.5,  // Mock data
		OnTimePerformance:      onTimePerformance,
		DelayTrend:             "improving", // Mock data - calculate trend
		RecurrentDelays:        89,   // Mock data
		PreventableDelays:      156,  // Mock data
		MitigatedDelays:        67,   // Mock data
		ImprovementOpportunity: 32.4, // Mock data
	}

	return c.JSON(metrics)
}

// @Summary Get vehicle capacity data
// @Tags Analytics
// @Accept json
// @Produce json
// @Param filters body AnalyticsFilters true "Analytics filters"
// @Success 200 {array} VehicleCapacityData
// @Router /api/analytics/vehicle-capacity [post]
func GetVehicleCapacityData(c *fiber.Ctx) error {
	var filters AnalyticsFilters
	if err := c.BodyParser(&filters); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	var vehicles []VehicleCapacityData

	// Query vehicles with their current utilization
	rows, err := database.DB.Raw(`
		SELECT 
			v.id,
			v.license_plate,
			v.vehicle_type,
			v.load_capacity_kg,
			v.load_capacity_m3,
			COALESCE(SUM(CASE WHEN t.status IN ('ACTIVE', 'IN_TRANSIT') THEN t.used_weight ELSE 0 END), 0) as current_load,
			COALESCE(SUM(CASE WHEN t.status IN ('ACTIVE', 'IN_TRANSIT') THEN t.used_volume ELSE 0 END), 0) as current_volume,
			COUNT(CASE WHEN t.status = 'COMPLETED' THEN 1 END) as completed_trips,
			v.is_active
		FROM vehicles v
		LEFT JOIN trips t ON v.id = t.vehicle_id
		WHERE v.is_active = true
		GROUP BY v.id, v.license_plate, v.vehicle_type, v.load_capacity_kg, v.load_capacity_m3, v.is_active
	`).Rows()

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database query failed"})
	}
	defer rows.Close()

	for rows.Next() {
		var vehicle VehicleCapacityData
		var vehicleID, licensePlate, vehicleType string
		var maxCapacity, maxVolume, currentLoad, currentVolume float64
		var trips int
		var isActive bool

		err := rows.Scan(
			&vehicleID, &licensePlate, &vehicleType, &maxCapacity, &maxVolume,
			&currentLoad, &currentVolume, &trips, &isActive,
		)
		if err != nil {
			continue
		}

		// Calculate utilization percentages
		var weightUtilization, volumeUtilization, overallUtilization float64
		if maxCapacity > 0 {
			weightUtilization = (currentLoad / maxCapacity) * 100
		}
		if maxVolume > 0 {
			volumeUtilization = (currentVolume / maxVolume) * 100
		}
		overallUtilization = (weightUtilization + volumeUtilization) / 2

		// Determine efficiency rating
		var efficiency string
		if overallUtilization >= 85 {
			efficiency = "excellent"
		} else if overallUtilization >= 70 {
			efficiency = "good"
		} else if overallUtilization >= 60 {
			efficiency = "average"
		} else {
			efficiency = "poor"
		}

		// Determine status
		var status string
		if isActive && currentLoad > 0 {
			status = "active"
		} else if isActive {
			status = "idle"
		} else {
			status = "unavailable"
		}

		vehicle = VehicleCapacityData{
			VehicleID:          vehicleID,
			VehicleNumber:      licensePlate,
			VehicleType:        vehicleType,
			MaxCapacity:        maxCapacity,
			MaxVolume:          maxVolume,
			CurrentLoad:        currentLoad,
			CurrentVolume:      currentVolume,
			WeightUtilization:  weightUtilization,
			VolumeUtilization:  volumeUtilization,
			OverallUtilization: overallUtilization,
			Trips:              trips,
			Revenue:            0,     // Mock data - calculate from completed trips
			OperatingCosts:     0,     // Mock data
			Profitability:      0,     // Mock data
			Efficiency:         efficiency,
			Location:           "Unknown", // Mock data - get from tracking
			Status:             status,
			LastUpdated:        time.Now(),
			UtilizationTrend:   "stable", // Mock data
		}

		vehicles = append(vehicles, vehicle)
	}

	return c.JSON(vehicles)
}

// Helper function to generate filter hash for caching
func generateFilterHash(filters AnalyticsFilters) string {
	filterBytes, _ := json.Marshal(filters)
	hash := md5.Sum(filterBytes)
	return fmt.Sprintf("%x", hash)
}

// Additional helper functions for complex analytics queries would go here...

// @Summary Get operational KPIs
// @Tags Analytics
// @Accept json
// @Produce json
// @Param category path string true "KPI Category"
// @Param filters body AnalyticsFilters true "Analytics filters"
// @Success 200 {object} map[string]interface{}
// @Router /api/analytics/kpis/{category} [post]
func GetOperationalKPIs(c *fiber.Ctx) error {
	category := c.Params("category")
	
	var filters AnalyticsFilters
	if err := c.BodyParser(&filters); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Return category-specific KPIs
	switch category {
	case "delivery":
		return c.JSON(fiber.Map{
			"on_time_percentage": 87.3,
			"average_delay": 23.5,
			"customer_satisfaction": 4.2,
		})
	case "utilization":
		return c.JSON(fiber.Map{
			"fleet_utilization": 78.3,
			"capacity_efficiency": 82.1,
			"load_matching_rate": 81.7,
		})
	case "financial":
		return c.JSON(fiber.Map{
			"revenue_per_mile": 2.35,
			"profit_margin": 28.4,
			"cost_efficiency": 82.1,
		})
	default:
		return c.Status(400).JSON(fiber.Map{"error": "Invalid KPI category"})
	}
}