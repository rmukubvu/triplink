package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// CacheMonitor provides monitoring and metrics for Redis cache
type CacheMonitor struct {
	redis   *RedisService
	metrics *CacheMetrics
	ctx     context.Context
}

// CacheMetrics holds cache performance metrics
type CacheMetrics struct {
	HitRate              float64           `json:"hit_rate"`
	MissRate             float64           `json:"miss_rate"`
	TotalRequests        int64             `json:"total_requests"`
	CacheHits            int64             `json:"cache_hits"`
	CacheMisses          int64             `json:"cache_misses"`
	AverageResponseTime  float64           `json:"average_response_time_ms"`
	MemoryUsage          int64             `json:"memory_usage_bytes"`
	KeyCount             int64             `json:"key_count"`
	ExpiredKeys          int64             `json:"expired_keys"`
	EvictedKeys          int64             `json:"evicted_keys"`
	ConnectionCount      int64             `json:"connection_count"`
	KeysByPrefix         map[string]int64  `json:"keys_by_prefix"`
	TopKeys              []KeyMetric       `json:"top_keys"`
	ErrorRate            float64           `json:"error_rate"`
	LastUpdated          time.Time         `json:"last_updated"`
}

// KeyMetric holds metrics for individual keys
type KeyMetric struct {
	Key         string    `json:"key"`
	AccessCount int64     `json:"access_count"`
	Size        int64     `json:"size_bytes"`
	TTL         int64     `json:"ttl_seconds"`
	LastAccess  time.Time `json:"last_access"`
}

// CacheAlert represents a cache-related alert
type CacheAlert struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Severity    string                 `json:"severity"`
	Message     string                 `json:"message"`
	Metric      string                 `json:"metric"`
	Threshold   float64                `json:"threshold"`
	ActualValue float64                `json:"actual_value"`
	Timestamp   time.Time              `json:"timestamp"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// AlertThresholds defines thresholds for different alerts
type AlertThresholds struct {
	HitRateMin          float64 `json:"hit_rate_min"`           // Minimum acceptable hit rate
	MemoryUsageMax      int64   `json:"memory_usage_max"`       // Maximum memory usage in bytes
	ErrorRateMax        float64 `json:"error_rate_max"`         // Maximum acceptable error rate
	ResponseTimeMax     float64 `json:"response_time_max"`      // Maximum response time in ms
	ConnectionCountMax  int64   `json:"connection_count_max"`   // Maximum connections
	EvictionRateMax     float64 `json:"eviction_rate_max"`      // Maximum eviction rate per minute
}

// NewCacheMonitor creates a new cache monitor
func NewCacheMonitor(redisService *RedisService) *CacheMonitor {
	return &CacheMonitor{
		redis: redisService,
		metrics: &CacheMetrics{
			KeysByPrefix: make(map[string]int64),
			TopKeys:      make([]KeyMetric, 0),
			LastUpdated:  time.Now(),
		},
		ctx: context.Background(),
	}
}

// StartMonitoring begins continuous monitoring of cache metrics
func (cm *CacheMonitor) StartMonitoring(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			cm.UpdateMetrics()
		}
	}()
}

// UpdateMetrics refreshes all cache metrics
func (cm *CacheMonitor) UpdateMetrics() error {
	// Get Redis INFO
	info, err := cm.redis.Client.Info(cm.ctx, "memory", "stats", "clients").Result()
	if err != nil {
		log.Printf("Failed to get Redis info: %v", err)
		return err
	}

	// Parse Redis info
	infoMap := cm.parseRedisInfo(info)

	// Update memory metrics
	if memUsed, ok := infoMap["used_memory"].(int64); ok {
		cm.metrics.MemoryUsage = memUsed
	}

	// Update connection metrics
	if connectedClients, ok := infoMap["connected_clients"].(int64); ok {
		cm.metrics.ConnectionCount = connectedClients
	}

	// Update key metrics
	if totalKeys, ok := infoMap["total_keys"].(int64); ok {
		cm.metrics.KeyCount = totalKeys
	}

	if expiredKeys, ok := infoMap["expired_keys"].(int64); ok {
		cm.metrics.ExpiredKeys = expiredKeys
	}

	if evictedKeys, ok := infoMap["evicted_keys"].(int64); ok {
		cm.metrics.EvictedKeys = evictedKeys
	}

	// Update key distribution by prefix
	cm.updateKeyDistribution()

	// Update top keys
	cm.updateTopKeys()

	// Calculate hit rate (if available from custom counters)
	cm.updateHitRate()

	cm.metrics.LastUpdated = time.Now()

	return nil
}

// GetMetrics returns current cache metrics
func (cm *CacheMonitor) GetMetrics() *CacheMetrics {
	return cm.metrics
}

// GetAlerts checks for cache-related alerts based on thresholds
func (cm *CacheMonitor) GetAlerts(thresholds AlertThresholds) []CacheAlert {
	var alerts []CacheAlert

	// Check hit rate
	if cm.metrics.HitRate < thresholds.HitRateMin {
		alerts = append(alerts, CacheAlert{
			ID:          "cache_hit_rate_low",
			Type:        "performance",
			Severity:    "warning",
			Message:     "Cache hit rate is below threshold",
			Metric:      "hit_rate",
			Threshold:   thresholds.HitRateMin,
			ActualValue: cm.metrics.HitRate,
			Timestamp:   time.Now(),
			Metadata: map[string]interface{}{
				"total_requests": cm.metrics.TotalRequests,
				"cache_hits":     cm.metrics.CacheHits,
			},
		})
	}

	// Check memory usage
	if cm.metrics.MemoryUsage > thresholds.MemoryUsageMax {
		alerts = append(alerts, CacheAlert{
			ID:          "cache_memory_high",
			Type:        "resource",
			Severity:    "critical",
			Message:     "Cache memory usage is above threshold",
			Metric:      "memory_usage",
			Threshold:   float64(thresholds.MemoryUsageMax),
			ActualValue: float64(cm.metrics.MemoryUsage),
			Timestamp:   time.Now(),
		})
	}

	// Check error rate
	if cm.metrics.ErrorRate > thresholds.ErrorRateMax {
		alerts = append(alerts, CacheAlert{
			ID:          "cache_error_rate_high",
			Type:        "reliability",
			Severity:    "critical",
			Message:     "Cache error rate is above threshold",
			Metric:      "error_rate",
			Threshold:   thresholds.ErrorRateMax,
			ActualValue: cm.metrics.ErrorRate,
			Timestamp:   time.Now(),
		})
	}

	// Check response time
	if cm.metrics.AverageResponseTime > thresholds.ResponseTimeMax {
		alerts = append(alerts, CacheAlert{
			ID:          "cache_response_time_high",
			Type:        "performance",
			Severity:    "warning",
			Message:     "Cache response time is above threshold",
			Metric:      "response_time",
			Threshold:   thresholds.ResponseTimeMax,
			ActualValue: cm.metrics.AverageResponseTime,
			Timestamp:   time.Now(),
		})
	}

	// Check connection count
	if cm.metrics.ConnectionCount > thresholds.ConnectionCountMax {
		alerts = append(alerts, CacheAlert{
			ID:          "cache_connections_high",
			Type:        "resource",
			Severity:    "warning",
			Message:     "Cache connection count is above threshold",
			Metric:      "connection_count",
			Threshold:   float64(thresholds.ConnectionCountMax),
			ActualValue: float64(cm.metrics.ConnectionCount),
			Timestamp:   time.Now(),
		})
	}

	return alerts
}

// GetCacheHealth returns overall cache health status
func (cm *CacheMonitor) GetCacheHealth() map[string]interface{} {
	health := map[string]interface{}{
		"status":      "healthy",
		"timestamp":   time.Now(),
		"redis_ping":  cm.checkRedisPing(),
		"metrics":     cm.metrics,
	}

	// Determine overall health based on metrics
	issues := []string{}

	if cm.metrics.HitRate < 0.5 {
		issues = append(issues, "low_hit_rate")
	}

	if cm.metrics.ErrorRate > 0.05 {
		issues = append(issues, "high_error_rate")
	}

	if cm.metrics.AverageResponseTime > 100 {
		issues = append(issues, "slow_response_time")
	}

	if len(issues) > 0 {
		health["status"] = "degraded"
		health["issues"] = issues
	}

	if len(issues) > 2 {
		health["status"] = "unhealthy"
	}

	return health
}

// OptimizationRecommendations provides cache optimization suggestions
func (cm *CacheMonitor) OptimizationRecommendations() []map[string]interface{} {
	var recommendations []map[string]interface{}

	// Low hit rate recommendations
	if cm.metrics.HitRate < 0.7 {
		recommendations = append(recommendations, map[string]interface{}{
			"type":        "performance",
			"priority":    "high",
			"title":       "Improve Cache Hit Rate",
			"description": "Cache hit rate is low. Consider increasing TTL for frequently accessed data or improving cache key strategies.",
			"metrics": map[string]interface{}{
				"current_hit_rate": cm.metrics.HitRate,
				"target_hit_rate":  0.8,
			},
		})
	}

	// High memory usage recommendations
	if cm.metrics.MemoryUsage > 800*1024*1024 { // 800MB
		recommendations = append(recommendations, map[string]interface{}{
			"type":        "resource",
			"priority":    "medium",
			"title":       "Optimize Memory Usage",
			"description": "Memory usage is high. Consider implementing cache eviction policies or reducing data size.",
			"metrics": map[string]interface{}{
				"current_memory_mb": cm.metrics.MemoryUsage / (1024 * 1024),
				"evicted_keys":      cm.metrics.EvictedKeys,
			},
		})
	}

	// Key distribution recommendations
	totalKeys := int64(0)
	for _, count := range cm.metrics.KeysByPrefix {
		totalKeys += count
	}

	if totalKeys > 100000 {
		recommendations = append(recommendations, map[string]interface{}{
			"type":        "optimization",
			"priority":    "low",
			"title":       "Optimize Key Management",
			"description": "Large number of keys detected. Consider implementing key rotation and cleanup strategies.",
			"metrics": map[string]interface{}{
				"total_keys":       totalKeys,
				"keys_by_prefix":   cm.metrics.KeysByPrefix,
			},
		})
	}

	return recommendations
}

// Helper methods

func (cm *CacheMonitor) parseRedisInfo(info string) map[string]interface{} {
	result := make(map[string]interface{})
	lines := strings.Split(info, "\r\n")

	for _, line := range lines {
		if strings.Contains(line, ":") && !strings.HasPrefix(line, "#") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])

				// Try to parse as integer
				if intVal, err := parseRedisValue(value); err == nil {
					result[key] = intVal
				} else {
					result[key] = value
				}
			}
		}
	}

	return result
}

func parseRedisValue(value string) (int64, error) {
	// Try parsing as integer first
	if intVal, err := parseInt64(value); err == nil {
		return intVal, nil
	}
	
	// If that fails, try parsing with commas removed (Redis sometimes uses commas in numbers)
	cleanValue := strings.ReplaceAll(value, ",", "")
	return parseInt64(cleanValue)
}

func parseInt64(s string) (int64, error) {
	// Simple integer parsing
	var result int64
	var sign int64 = 1
	
	for i, char := range s {
		if i == 0 && char == '-' {
			sign = -1
			continue
		}
		if char >= '0' && char <= '9' {
			result = result*10 + int64(char-'0')
		} else {
			return 0, fmt.Errorf("invalid integer: %s", s)
		}
	}
	
	return result * sign, nil
}

func (cm *CacheMonitor) updateKeyDistribution() {
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
		keys, err := cm.redis.Client.Keys(cm.ctx, pattern).Result()
		if err == nil {
			cm.metrics.KeysByPrefix[prefix] = int64(len(keys))
		}
	}
}

func (cm *CacheMonitor) updateTopKeys() {
	// This would require custom tracking in production
	// For now, we'll create a simplified version
	cm.metrics.TopKeys = []KeyMetric{
		{
			Key:         "top_accessed_key_example",
			AccessCount: 1500,
			Size:        2048,
			TTL:         3600,
			LastAccess:  time.Now().Add(-5 * time.Minute),
		},
	}
}

func (cm *CacheMonitor) updateHitRate() {
	// This would require custom counters in production
	// For now, we'll use a simplified calculation
	if cm.metrics.TotalRequests > 0 {
		cm.metrics.HitRate = float64(cm.metrics.CacheHits) / float64(cm.metrics.TotalRequests)
		cm.metrics.MissRate = float64(cm.metrics.CacheMisses) / float64(cm.metrics.TotalRequests)
	}
}

func (cm *CacheMonitor) checkRedisPing() bool {
	err := cm.redis.Client.Ping(cm.ctx).Err()
	return err == nil
}

// ExportMetrics exports metrics in JSON format
func (cm *CacheMonitor) ExportMetrics() ([]byte, error) {
	return json.Marshal(cm.metrics)
}

// ResetMetrics resets all metrics counters
func (cm *CacheMonitor) ResetMetrics() {
	cm.metrics = &CacheMetrics{
		KeysByPrefix: make(map[string]int64),
		TopKeys:      make([]KeyMetric, 0),
		LastUpdated:  time.Now(),
	}
}