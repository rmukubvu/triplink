package middleware

import (
	"crypto/md5"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"triplink/backend/services"
)

// CacheMiddleware provides Redis-based caching for API endpoints
type CacheMiddleware struct {
	redis *services.RedisService
}

// NewCacheMiddleware creates a new cache middleware instance
func NewCacheMiddleware() *CacheMiddleware {
	return &CacheMiddleware{
		redis: services.NewRedisService(),
	}
}

// CacheConfig holds configuration for caching specific endpoints
type CacheConfig struct {
	TTL         time.Duration
	KeyPrefix   string
	VaryBy      []string // Parameters to include in cache key
	SkipAuth    bool     // Whether to skip user-specific caching
	Condition   func(*fiber.Ctx) bool // Condition to enable caching
}

// DefaultCacheConfigs provides default caching configurations for different endpoint types
var DefaultCacheConfigs = map[string]CacheConfig{
	// Analytics endpoints
	"analytics": {
		TTL:       15 * time.Minute,
		KeyPrefix: "api:analytics:",
		VaryBy:    []string{"body", "user_id"},
		SkipAuth:  false,
	},
	// Route optimization endpoints
	"route_optimization": {
		TTL:       30 * time.Minute,
		KeyPrefix: "api:route_opt:",
		VaryBy:    []string{"body"},
		SkipAuth:  true, // Route optimization is not user-specific
	},
	// ML prediction endpoints
	"ml_predictions": {
		TTL:       1 * time.Hour,
		KeyPrefix: "api:ml:",
		VaryBy:    []string{"body"},
		SkipAuth:  true,
	},
	// External API endpoints
	"external_api": {
		TTL:       10 * time.Minute,
		KeyPrefix: "api:external:",
		VaryBy:    []string{"body"},
		SkipAuth:  true,
	},
	// Real-time data (short TTL)
	"realtime": {
		TTL:       2 * time.Minute,
		KeyPrefix: "api:realtime:",
		VaryBy:    []string{"params", "user_id"},
		SkipAuth:  false,
	},
}

// Cache returns a caching middleware with specified configuration
func (cm *CacheMiddleware) Cache(configName string) fiber.Handler {
	config, exists := DefaultCacheConfigs[configName]
	if !exists {
		// Return a no-op middleware if config doesn't exist
		return func(c *fiber.Ctx) error {
			return c.Next()
		}
	}
	
	return cm.CacheWithConfig(config)
}

// CacheWithConfig returns a caching middleware with custom configuration
func (cm *CacheMiddleware) CacheWithConfig(config CacheConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Skip caching for non-GET requests by default
		if c.Method() != "GET" && c.Method() != "POST" {
			return c.Next()
		}
		
		// Apply condition check if specified
		if config.Condition != nil && !config.Condition(c) {
			return c.Next()
		}
		
		// Generate cache key
		cacheKey := cm.generateCacheKey(c, config)
		
		// Try to get from cache
		var cachedResponse CachedResponse
		if err := cm.redis.Get(cacheKey, &cachedResponse); err == nil {
			// Cache hit - return cached response
			c.Set("X-Cache", "HIT")
			c.Set("X-Cache-TTL", strconv.Itoa(int(cachedResponse.TTL)))
			c.Set("Content-Type", cachedResponse.ContentType)
			
			return c.Status(cachedResponse.StatusCode).Send(cachedResponse.Body)
		}
		
		// Cache miss - continue with request processing
		c.Set("X-Cache", "MISS")
		
		// Capture response
		return cm.captureAndCacheResponse(c, cacheKey, config)
	}
}

// CachedResponse represents a cached HTTP response
type CachedResponse struct {
	StatusCode  int                 `json:"status_code"`
	Body        []byte              `json:"body"`
	Headers     map[string]string   `json:"headers"`
	ContentType string              `json:"content_type"`
	TTL         int64               `json:"ttl"`
	CachedAt    time.Time           `json:"cached_at"`
}

// captureAndCacheResponse captures the response and caches it
func (cm *CacheMiddleware) captureAndCacheResponse(c *fiber.Ctx, cacheKey string, config CacheConfig) error {
	// Continue with normal request processing
	if err := c.Next(); err != nil {
		return err
	}
	
	// Only cache successful responses
	if c.Response().StatusCode() >= 200 && c.Response().StatusCode() < 300 {
		cachedResponse := CachedResponse{
			StatusCode:  c.Response().StatusCode(),
			Body:        c.Response().Body(),
			ContentType: string(c.Response().Header.ContentType()),
			TTL:         int64(config.TTL.Seconds()),
			CachedAt:    time.Now(),
		}
		
		// Store in cache (async to not block response)
		go func() {
			cm.redis.Set(cacheKey, cachedResponse, config.TTL)
		}()
		
		c.Set("X-Cache-Status", "STORED")
	}
	
	return nil
}

// generateCacheKey creates a unique cache key based on request parameters
func (cm *CacheMiddleware) generateCacheKey(c *fiber.Ctx, config CacheConfig) string {
	var keyParts []string
	
	// Add prefix
	keyParts = append(keyParts, config.KeyPrefix)
	
	// Add route path
	keyParts = append(keyParts, c.Route().Path)
	
	// Add user ID if not skipping auth
	if !config.SkipAuth {
		if userID := c.Locals("user_id"); userID != nil {
			keyParts = append(keyParts, fmt.Sprintf("user:%v", userID))
		}
	}
	
	// Add variable parts based on configuration
	for _, varyBy := range config.VaryBy {
		switch varyBy {
		case "body":
			if body := c.Body(); len(body) > 0 {
				hash := md5.Sum(body)
				keyParts = append(keyParts, fmt.Sprintf("body:%x", hash))
			}
		case "params":
			if params := c.AllParams(); len(params) > 0 {
				hash := md5.Sum([]byte(fmt.Sprintf("%v", params)))
				keyParts = append(keyParts, fmt.Sprintf("params:%x", hash))
			}
		case "query":
			if query := c.Context().QueryArgs().String(); query != "" {
				hash := md5.Sum([]byte(query))
				keyParts = append(keyParts, fmt.Sprintf("query:%x", hash))
			}
		case "user_id":
			if userID := c.Locals("user_id"); userID != nil {
				keyParts = append(keyParts, fmt.Sprintf("user:%v", userID))
			}
		}
	}
	
	return strings.Join(keyParts, ":")
}

// RateLimitMiddleware provides Redis-based rate limiting
func (cm *CacheMiddleware) RateLimitMiddleware(requests int, window time.Duration) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get client identifier (IP + User ID if available)
		identifier := c.IP()
		if userID := c.Locals("user_id"); userID != nil {
			identifier = fmt.Sprintf("%s:user:%v", identifier, userID)
		}
		
		// Check rate limit
		allowed, remaining, err := cm.redis.CheckRateLimit(identifier, requests, window)
		if err != nil {
			// If Redis is down, allow the request but log the error
			// In production, you might want to fail closed instead
			return c.Next()
		}
		
		// Set rate limit headers
		c.Set("X-RateLimit-Limit", strconv.Itoa(requests))
		c.Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Set("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(window).Unix(), 10))
		
		if !allowed {
			return c.Status(429).JSON(fiber.Map{
				"error": "Rate limit exceeded",
				"retry_after": int64(window.Seconds()),
			})
		}
		
		return c.Next()
	}
}

// CacheInvalidationMiddleware provides cache invalidation for write operations
func (cm *CacheMiddleware) CacheInvalidationMiddleware(patterns []string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Process the request first
		if err := c.Next(); err != nil {
			return err
		}
		
		// Only invalidate cache for successful write operations
		if c.Method() != "GET" && c.Response().StatusCode() >= 200 && c.Response().StatusCode() < 300 {
			// Invalidate cache patterns asynchronously
			go func() {
				for _, pattern := range patterns {
					cm.redis.ClearCacheByPrefix(pattern)
				}
			}()
		}
		
		return nil
	}
}

// SessionMiddleware provides Redis-based session management
func (cm *CacheMiddleware) SessionMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		sessionID := c.Cookies("session_id")
		if sessionID == "" {
			// No session cookie, continue without session
			return c.Next()
		}
		
		// Get session data from Redis
		sessionData, err := cm.redis.GetSession(sessionID)
		if err != nil {
			// Invalid or expired session, clear cookie
			c.Cookie(&fiber.Cookie{
				Name:   "session_id",
				Value:  "",
				MaxAge: -1,
			})
			return c.Next()
		}
		
		// Store session data in context
		c.Locals("session", sessionData)
		if userID, ok := sessionData["user_id"]; ok {
			c.Locals("user_id", userID)
		}
		
		// Continue with request
		if err := c.Next(); err != nil {
			return err
		}
		
		// Update session TTL on successful request
		go func() {
			cm.redis.UpdateSession(sessionID, sessionData)
		}()
		
		return nil
	}
}

// CacheWarmupMiddleware pre-loads cache with commonly requested data
func (cm *CacheMiddleware) CacheWarmupMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// This middleware can be used to warm up cache with frequently accessed data
		// Implementation depends on specific use case
		return c.Next()
	}
}

// HealthCheckHandler provides cache health check endpoint
func (cm *CacheMiddleware) HealthCheckHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		stats, err := cm.redis.GetCacheStats()
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"status": "unhealthy",
				"error":  err.Error(),
			})
		}
		
		// Check Redis connection
		if err := cm.redis.HealthCheck(); err != nil {
			return c.Status(500).JSON(fiber.Map{
				"status": "unhealthy",
				"error":  "Redis connection failed",
			})
		}
		
		return c.JSON(fiber.Map{
			"status": "healthy",
			"cache_stats": stats,
			"timestamp": time.Now(),
		})
	}
}

// GetRedisService returns the underlying Redis service
func (cm *CacheMiddleware) GetRedisService() *services.RedisService {
	return cm.redis
}