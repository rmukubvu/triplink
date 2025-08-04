package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// RedisConfig holds Redis configuration settings
type RedisConfig struct {
	Host             string
	Port             string
	Password         string
	DB               int
	PoolSize         int
	ConnMaxIdleTime  time.Duration
	ConnMaxLifetime  time.Duration
	DialTimeout      time.Duration
	ReadTimeout      time.Duration
	WriteTimeout     time.Duration
	EnableTLS        bool
	TLSCertFile      string
	TLSKeyFile       string
	MaxRetries       int
	MinRetryBackoff  time.Duration
	MaxRetryBackoff  time.Duration
}

// GetRedisConfig returns Redis configuration from environment variables
func GetRedisConfig() *RedisConfig {
	config := &RedisConfig{
		Host:             getEnvString("REDIS_HOST", "localhost"),
		Port:             getEnvString("REDIS_PORT", "6379"),
		Password:         getEnvString("REDIS_PASSWORD", ""),
		DB:               getEnvInt("REDIS_DB", 0),
		PoolSize:         getEnvInt("REDIS_POOL_SIZE", 20),
		ConnMaxIdleTime:  getEnvDuration("REDIS_CONN_MAX_IDLE_TIME", 30*time.Second),
		ConnMaxLifetime:  getEnvDuration("REDIS_CONN_MAX_LIFETIME", 5*time.Minute),
		DialTimeout:      getEnvDuration("REDIS_DIAL_TIMEOUT", 5*time.Second),
		ReadTimeout:      getEnvDuration("REDIS_READ_TIMEOUT", 3*time.Second),
		WriteTimeout:     getEnvDuration("REDIS_WRITE_TIMEOUT", 3*time.Second),
		EnableTLS:        getEnvBool("REDIS_ENABLE_TLS", false),
		TLSCertFile:      getEnvString("REDIS_TLS_CERT_FILE", ""),
		TLSKeyFile:       getEnvString("REDIS_TLS_KEY_FILE", ""),
		MaxRetries:       getEnvInt("REDIS_MAX_RETRIES", 3),
		MinRetryBackoff:  getEnvDuration("REDIS_MIN_RETRY_BACKOFF", 8*time.Millisecond),
		MaxRetryBackoff:  getEnvDuration("REDIS_MAX_RETRY_BACKOFF", 512*time.Millisecond),
	}

	return config
}

// Cache configuration settings
type CacheConfig struct {
	// Default TTL values for different cache types
	AnalyticsTTL        time.Duration
	RouteOptimizationTTL time.Duration
	MLPredictionsTTL    time.Duration
	ExternalAPITTL      time.Duration
	RealtimeTTL         time.Duration
	SessionTTL          time.Duration
	
	// Rate limiting settings
	DefaultRateLimit       int
	DefaultRateLimitWindow time.Duration
	
	// Cache size limits
	MaxCacheSize           int64 // bytes
	MaxKeysPerPattern      int
	
	// Cache warming settings
	EnableCacheWarming     bool
	CacheWarmupInterval    time.Duration
	
	// Monitoring settings
	EnableMetrics          bool
	MetricsInterval        time.Duration
}

// GetCacheConfig returns cache configuration from environment variables
func GetCacheConfig() *CacheConfig {
	return &CacheConfig{
		AnalyticsTTL:           getEnvDuration("CACHE_ANALYTICS_TTL", 15*time.Minute),
		RouteOptimizationTTL:   getEnvDuration("CACHE_ROUTE_OPT_TTL", 30*time.Minute),
		MLPredictionsTTL:       getEnvDuration("CACHE_ML_PREDICTIONS_TTL", 1*time.Hour),
		ExternalAPITTL:         getEnvDuration("CACHE_EXTERNAL_API_TTL", 10*time.Minute),
		RealtimeTTL:            getEnvDuration("CACHE_REALTIME_TTL", 2*time.Minute),
		SessionTTL:             getEnvDuration("CACHE_SESSION_TTL", 24*time.Hour),
		
		DefaultRateLimit:       getEnvInt("RATE_LIMIT_DEFAULT", 100),
		DefaultRateLimitWindow: getEnvDuration("RATE_LIMIT_WINDOW", 1*time.Minute),
		
		MaxCacheSize:           getEnvInt64("CACHE_MAX_SIZE", 1024*1024*1024), // 1GB default
		MaxKeysPerPattern:      getEnvInt("CACHE_MAX_KEYS_PER_PATTERN", 10000),
		
		EnableCacheWarming:     getEnvBool("CACHE_ENABLE_WARMING", false),
		CacheWarmupInterval:    getEnvDuration("CACHE_WARMUP_INTERVAL", 5*time.Minute),
		
		EnableMetrics:          getEnvBool("CACHE_ENABLE_METRICS", true),
		MetricsInterval:        getEnvDuration("CACHE_METRICS_INTERVAL", 1*time.Minute),
	}
}

// Helper functions for environment variable parsing

func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

// Validation functions

// ValidateRedisConfig validates Redis configuration
func (rc *RedisConfig) ValidateRedisConfig() error {
	if rc.Host == "" {
		return fmt.Errorf("Redis host cannot be empty")
	}
	if rc.Port == "" {
		return fmt.Errorf("Redis port cannot be empty")
	}
	if rc.DB < 0 || rc.DB > 15 {
		return fmt.Errorf("Redis DB must be between 0 and 15")
	}
	if rc.PoolSize <= 0 {
		return fmt.Errorf("Redis pool size must be positive")
	}
	return nil
}

// ValidateCacheConfig validates cache configuration
func (cc *CacheConfig) ValidateCacheConfig() error {
	if cc.DefaultRateLimit <= 0 {
		return fmt.Errorf("Default rate limit must be positive")
	}
	if cc.DefaultRateLimitWindow <= 0 {
		return fmt.Errorf("Default rate limit window must be positive")
	}
	if cc.MaxCacheSize <= 0 {
		return fmt.Errorf("Max cache size must be positive")
	}
	return nil
}

// Environment configuration template for Redis
const RedisEnvTemplate = `
# Redis Configuration
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0
REDIS_POOL_SIZE=20
REDIS_CONN_MAX_IDLE_TIME=30s
REDIS_CONN_MAX_LIFETIME=5m
REDIS_DIAL_TIMEOUT=5s
REDIS_READ_TIMEOUT=3s
REDIS_WRITE_TIMEOUT=3s
REDIS_ENABLE_TLS=false
REDIS_TLS_CERT_FILE=
REDIS_TLS_KEY_FILE=
REDIS_MAX_RETRIES=3
REDIS_MIN_RETRY_BACKOFF=8ms
REDIS_MAX_RETRY_BACKOFF=512ms

# Cache Configuration
CACHE_ANALYTICS_TTL=15m
CACHE_ROUTE_OPT_TTL=30m
CACHE_ML_PREDICTIONS_TTL=1h
CACHE_EXTERNAL_API_TTL=10m
CACHE_REALTIME_TTL=2m
CACHE_SESSION_TTL=24h

# Rate Limiting
RATE_LIMIT_DEFAULT=100
RATE_LIMIT_WINDOW=1m

# Cache Limits
CACHE_MAX_SIZE=1073741824  # 1GB in bytes
CACHE_MAX_KEYS_PER_PATTERN=10000

# Cache Features
CACHE_ENABLE_WARMING=false
CACHE_WARMUP_INTERVAL=5m
CACHE_ENABLE_METRICS=true
CACHE_METRICS_INTERVAL=1m

# External API Keys for ML and External Services
HUGGINGFACE_API_KEY=your_huggingface_api_key_here
GOOGLE_MAPS_API_KEY=your_google_maps_api_key_here
OPENWEATHERMAP_API_KEY=your_openweathermap_api_key_here
HERE_API_KEY=your_here_api_key_here
TOLLGURU_API_KEY=your_tollguru_api_key_here
GASBUDDY_API_KEY=your_gasbuddy_api_key_here
DOT_API_KEY=your_dot_511_api_key_here
`