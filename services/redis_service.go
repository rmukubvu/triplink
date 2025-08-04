package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// Redis service for caching and session management
type RedisService struct {
	Client *redis.Client
	ctx    context.Context
}

// Cache configuration constants
const (
	// Route optimization cache
	RouteOptimizationCacheTTL = 30 * time.Minute
	RouteOptimizationPrefix   = "route_opt:"
	
	// Analytics cache
	AnalyticsCacheTTL = 15 * time.Minute
	AnalyticsPrefix   = "analytics:"
	
	// ML predictions cache
	MLPredictionsCacheTTL = 1 * time.Hour
	MLPredictionsPrefix   = "ml_pred:"
	
	// External API cache
	ExternalAPICacheTTL = 10 * time.Minute
	ExternalAPIPrefix   = "ext_api:"
	
	// Rate limiting
	RateLimitTTL    = 1 * time.Hour
	RateLimitPrefix = "rate_limit:"
	
	// Session management
	SessionTTL    = 24 * time.Hour
	SessionPrefix = "session:"
	
	// Real-time data cache
	RealtimeCacheTTL = 2 * time.Minute
	RealtimePrefix   = "realtime:"
)

// Cache key structures
type CacheKey struct {
	Prefix string
	ID     string
	Suffix string
}

func (ck CacheKey) String() string {
	if ck.Suffix != "" {
		return fmt.Sprintf("%s%s:%s", ck.Prefix, ck.ID, ck.Suffix)
	}
	return fmt.Sprintf("%s%s", ck.Prefix, ck.ID)
}

// NewRedisService creates a new Redis service instance
func NewRedisService() *RedisService {
	// Redis configuration from environment variables
	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = "localhost"
	}
	
	redisPort := os.Getenv("REDIS_PORT")
	if redisPort == "" {
		redisPort = "6379"
	}
	
	redisPassword := os.Getenv("REDIS_PASSWORD")
	redisDB := os.Getenv("REDIS_DB")
	
	db := 0
	if redisDB != "" {
		if dbNum, err := strconv.Atoi(redisDB); err == nil {
			db = dbNum
		}
	}

	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", redisHost, redisPort),
		Password: redisPassword,
		DB:       db,
		PoolSize: 20,
		ConnMaxIdleTime: 30 * time.Second,
		ConnMaxLifetime: 5 * time.Minute,
	})

	return &RedisService{
		Client: client,
		ctx:    context.Background(),
	}
}

// Health check for Redis connection
func (r *RedisService) HealthCheck() error {
	return r.Client.Ping(r.ctx).Err()
}

// Generic cache operations

// Set stores a value in cache with TTL
func (r *RedisService) Set(key string, value interface{}, ttl time.Duration) error {
	jsonData, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}
	
	return r.Client.Set(r.ctx, key, jsonData, ttl).Err()
}

// Get retrieves a value from cache
func (r *RedisService) Get(key string, dest interface{}) error {
	val, err := r.Client.Get(r.ctx, key).Result()
	if err != nil {
		return err
	}
	
	return json.Unmarshal([]byte(val), dest)
}

// Delete removes a key from cache
func (r *RedisService) Delete(key string) error {
	return r.Client.Del(r.ctx, key).Err()
}

// Exists checks if a key exists in cache
func (r *RedisService) Exists(key string) bool {
	count, err := r.Client.Exists(r.ctx, key).Result()
	return err == nil && count > 0
}

// SetExpiration updates TTL for existing key
func (r *RedisService) SetExpiration(key string, ttl time.Duration) error {
	return r.Client.Expire(r.ctx, key, ttl).Err()
}

// Route Optimization Caching

// CacheRouteOptimization stores route optimization results
func (r *RedisService) CacheRouteOptimization(routeID string, optimization interface{}) error {
	key := CacheKey{Prefix: RouteOptimizationPrefix, ID: routeID}.String()
	return r.Set(key, optimization, RouteOptimizationCacheTTL)
}

// GetCachedRouteOptimization retrieves cached route optimization
func (r *RedisService) GetCachedRouteOptimization(routeID string, dest interface{}) error {
	key := CacheKey{Prefix: RouteOptimizationPrefix, ID: routeID}.String()
	return r.Get(key, dest)
}

// CacheRouteMatrix stores route matrix results
func (r *RedisService) CacheRouteMatrix(origins, destinations []string, matrix interface{}) error {
	matrixID := r.generateMatrixID(origins, destinations)
	key := CacheKey{Prefix: RouteOptimizationPrefix, ID: matrixID, Suffix: "matrix"}.String()
	return r.Set(key, matrix, RouteOptimizationCacheTTL)
}

// GetCachedRouteMatrix retrieves cached route matrix
func (r *RedisService) GetCachedRouteMatrix(origins, destinations []string, dest interface{}) error {
	matrixID := r.generateMatrixID(origins, destinations)
	key := CacheKey{Prefix: RouteOptimizationPrefix, ID: matrixID, Suffix: "matrix"}.String()
	return r.Get(key, dest)
}

// Analytics Caching

// CacheAnalyticsResult stores analytics computation results
func (r *RedisService) CacheAnalyticsResult(analysisType, filterHash string, result interface{}) error {
	key := CacheKey{Prefix: AnalyticsPrefix, ID: analysisType, Suffix: filterHash}.String()
	return r.Set(key, result, AnalyticsCacheTTL)
}

// GetCachedAnalyticsResult retrieves cached analytics result
func (r *RedisService) GetCachedAnalyticsResult(analysisType, filterHash string, dest interface{}) error {
	key := CacheKey{Prefix: AnalyticsPrefix, ID: analysisType, Suffix: filterHash}.String()
	return r.Get(key, dest)
}

// CacheKPIData stores KPI calculations
func (r *RedisService) CacheKPIData(category string, kpiData interface{}) error {
	key := CacheKey{Prefix: AnalyticsPrefix, ID: "kpi", Suffix: category}.String()
	return r.Set(key, kpiData, AnalyticsCacheTTL)
}

// GetCachedKPIData retrieves cached KPI data
func (r *RedisService) GetCachedKPIData(category string, dest interface{}) error {
	key := CacheKey{Prefix: AnalyticsPrefix, ID: "kpi", Suffix: category}.String()
	return r.Get(key, dest)
}

// ML Predictions Caching

// CacheSentimentAnalysis stores sentiment analysis results
func (r *RedisService) CacheSentimentAnalysis(textHash string, result interface{}) error {
	key := CacheKey{Prefix: MLPredictionsPrefix, ID: "sentiment", Suffix: textHash}.String()
	return r.Set(key, result, MLPredictionsCacheTTL)
}

// GetCachedSentimentAnalysis retrieves cached sentiment analysis
func (r *RedisService) GetCachedSentimentAnalysis(textHash string, dest interface{}) error {
	key := CacheKey{Prefix: MLPredictionsPrefix, ID: "sentiment", Suffix: textHash}.String()
	return r.Get(key, dest)
}

// CacheDelayPrediction stores delay prediction results
func (r *RedisService) CacheDelayPrediction(routeHash string, prediction interface{}) error {
	key := CacheKey{Prefix: MLPredictionsPrefix, ID: "delay", Suffix: routeHash}.String()
	return r.Set(key, prediction, MLPredictionsCacheTTL)
}

// GetCachedDelayPrediction retrieves cached delay prediction
func (r *RedisService) GetCachedDelayPrediction(routeHash string, dest interface{}) error {
	key := CacheKey{Prefix: MLPredictionsPrefix, ID: "delay", Suffix: routeHash}.String()
	return r.Get(key, dest)
}

// CacheSatisfactionPrediction stores customer satisfaction predictions
func (r *RedisService) CacheSatisfactionPrediction(tripID string, prediction interface{}) error {
	key := CacheKey{Prefix: MLPredictionsPrefix, ID: "satisfaction", Suffix: tripID}.String()
	return r.Set(key, prediction, MLPredictionsCacheTTL)
}

// GetCachedSatisfactionPrediction retrieves cached satisfaction prediction
func (r *RedisService) GetCachedSatisfactionPrediction(tripID string, dest interface{}) error {
	key := CacheKey{Prefix: MLPredictionsPrefix, ID: "satisfaction", Suffix: tripID}.String()
	return r.Get(key, dest)
}

// External API Caching

// CacheTrafficInfo stores traffic information
func (r *RedisService) CacheTrafficInfo(routeHash string, trafficInfo interface{}) error {
	key := CacheKey{Prefix: ExternalAPIPrefix, ID: "traffic", Suffix: routeHash}.String()
	return r.Set(key, trafficInfo, ExternalAPICacheTTL)
}

// GetCachedTrafficInfo retrieves cached traffic information
func (r *RedisService) GetCachedTrafficInfo(routeHash string, dest interface{}) error {
	key := CacheKey{Prefix: ExternalAPIPrefix, ID: "traffic", Suffix: routeHash}.String()
	return r.Get(key, dest)
}

// CacheWeatherInfo stores weather information
func (r *RedisService) CacheWeatherInfo(locationHash string, weatherInfo interface{}) error {
	key := CacheKey{Prefix: ExternalAPIPrefix, ID: "weather", Suffix: locationHash}.String()
	return r.Set(key, weatherInfo, ExternalAPICacheTTL)
}

// GetCachedWeatherInfo retrieves cached weather information
func (r *RedisService) GetCachedWeatherInfo(locationHash string, dest interface{}) error {
	key := CacheKey{Prefix: ExternalAPIPrefix, ID: "weather", Suffix: locationHash}.String()
	return r.Get(key, dest)
}

// CacheFuelPrices stores fuel price information
func (r *RedisService) CacheFuelPrices(locationHash string, fuelPrices interface{}) error {
	key := CacheKey{Prefix: ExternalAPIPrefix, ID: "fuel", Suffix: locationHash}.String()
	return r.Set(key, fuelPrices, ExternalAPICacheTTL)
}

// GetCachedFuelPrices retrieves cached fuel prices
func (r *RedisService) GetCachedFuelPrices(locationHash string, dest interface{}) error {
	key := CacheKey{Prefix: ExternalAPIPrefix, ID: "fuel", Suffix: locationHash}.String()
	return r.Get(key, dest)
}

// Rate Limiting

// CheckRateLimit implements rate limiting for API endpoints
func (r *RedisService) CheckRateLimit(identifier string, limit int, window time.Duration) (bool, int, error) {
	key := CacheKey{Prefix: RateLimitPrefix, ID: identifier}.String()
	
	// Get current count
	current, err := r.Client.Get(r.ctx, key).Int()
	if err == redis.Nil {
		// First request in window
		err = r.Client.SetNX(r.ctx, key, 1, window).Err()
		if err != nil {
			return false, 0, err
		}
		return true, limit-1, nil
	} else if err != nil {
		return false, 0, err
	}
	
	// Check if limit exceeded
	if current >= limit {
		return false, 0, nil
	}
	
	// Increment counter
	newCount, err := r.Client.Incr(r.ctx, key).Result()
	if err != nil {
		return false, 0, err
	}
	
	return true, limit-int(newCount), nil
}

// Session Management

// CreateSession creates a new user session
func (r *RedisService) CreateSession(sessionID, userID string, sessionData interface{}) error {
	key := CacheKey{Prefix: SessionPrefix, ID: sessionID}.String()
	data := map[string]interface{}{
		"user_id":    userID,
		"created_at": time.Now(),
		"data":       sessionData,
	}
	return r.Set(key, data, SessionTTL)
}

// GetSession retrieves session data
func (r *RedisService) GetSession(sessionID string) (map[string]interface{}, error) {
	key := CacheKey{Prefix: SessionPrefix, ID: sessionID}.String()
	var sessionData map[string]interface{}
	err := r.Get(key, &sessionData)
	return sessionData, err
}

// UpdateSession updates session data and extends TTL
func (r *RedisService) UpdateSession(sessionID string, sessionData interface{}) error {
	key := CacheKey{Prefix: SessionPrefix, ID: sessionID}.String()
	err := r.Set(key, sessionData, SessionTTL)
	if err != nil {
		return err
	}
	return r.SetExpiration(key, SessionTTL)
}

// DeleteSession removes a session
func (r *RedisService) DeleteSession(sessionID string) error {
	key := CacheKey{Prefix: SessionPrefix, ID: sessionID}.String()
	return r.Delete(key)
}

// Real-time Data Caching

// CacheRealtimeLocation stores real-time vehicle location
func (r *RedisService) CacheRealtimeLocation(vehicleID string, location interface{}) error {
	key := CacheKey{Prefix: RealtimePrefix, ID: "location", Suffix: vehicleID}.String()
	return r.Set(key, location, RealtimeCacheTTL)
}

// GetCachedRealtimeLocation retrieves cached vehicle location
func (r *RedisService) GetCachedRealtimeLocation(vehicleID string, dest interface{}) error {
	key := CacheKey{Prefix: RealtimePrefix, ID: "location", Suffix: vehicleID}.String()
	return r.Get(key, dest)
}

// CacheRealtimeMetrics stores real-time operational metrics
func (r *RedisService) CacheRealtimeMetrics(metricType string, metrics interface{}) error {
	key := CacheKey{Prefix: RealtimePrefix, ID: "metrics", Suffix: metricType}.String()
	return r.Set(key, metrics, RealtimeCacheTTL)
}

// GetCachedRealtimeMetrics retrieves cached real-time metrics
func (r *RedisService) GetCachedRealtimeMetrics(metricType string, dest interface{}) error {
	key := CacheKey{Prefix: RealtimePrefix, ID: "metrics", Suffix: metricType}.String()
	return r.Get(key, dest)
}

// Cache Statistics and Management

// GetCacheStats returns cache statistics
func (r *RedisService) GetCacheStats() (map[string]interface{}, error) {
	info, err := r.Client.Info(r.ctx, "memory", "stats").Result()
	if err != nil {
		return nil, err
	}
	
	// Get key counts by prefix
	keyCounts := make(map[string]int64)
	prefixes := []string{
		RouteOptimizationPrefix,
		AnalyticsPrefix,
		MLPredictionsPrefix,
		ExternalAPIPrefix,
		RateLimitPrefix,
		SessionPrefix,
		RealtimePrefix,
	}
	
	for _, prefix := range prefixes {
		pattern := prefix + "*"
		keys, err := r.Client.Keys(r.ctx, pattern).Result()
		if err == nil {
			keyCounts[prefix] = int64(len(keys))
		}
	}
	
	return map[string]interface{}{
		"redis_info":  info,
		"key_counts":  keyCounts,
		"total_keys":  r.getTotalKeyCount(),
		"uptime":      r.getUptime(),
	}, nil
}

// ClearCache clears all cache data (use with caution)
func (r *RedisService) ClearCache() error {
	return r.Client.FlushDB(r.ctx).Err()
}

// ClearCacheByPrefix clears cache data by prefix
func (r *RedisService) ClearCacheByPrefix(prefix string) error {
	pattern := prefix + "*"
	keys, err := r.Client.Keys(r.ctx, pattern).Result()
	if err != nil {
		return err
	}
	
	if len(keys) > 0 {
		return r.Client.Del(r.ctx, keys...).Err()
	}
	
	return nil
}

// Helper functions

// generateMatrixID creates a unique ID for route matrix
func (r *RedisService) generateMatrixID(origins, destinations []string) string {
	// Create a hash of origins and destinations
	data := fmt.Sprintf("%v%v", origins, destinations)
	return fmt.Sprintf("%x", data) // Simple hash for demo
}

// getTotalKeyCount gets total number of keys in database
func (r *RedisService) getTotalKeyCount() int64 {
	size, err := r.Client.DBSize(r.ctx).Result()
	if err != nil {
		return -1
	}
	return size
}

// getUptime gets Redis server uptime
func (r *RedisService) getUptime() string {
	_, err := r.Client.Info(r.ctx, "server").Result()
	if err != nil {
		return "unknown"
	}
	
	// Parse uptime from info string (simplified)
	lines := []string{}
	for _, line := range lines {
		if len(line) > 0 && line[0] != '#' {
			// Parse server info - simplified for demo
			break
		}
	}
	
	return "available" // Simplified return
}

// Close closes the Redis connection
func (r *RedisService) Close() error {
	return r.Client.Close()
}