package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/LGU-SE-Internal/rcabench/dto"
	"github.com/gin-gonic/gin"
)

// RateLimiter represents a simple in-memory rate limiter
type RateLimiter struct {
	requests map[string][]time.Time
	mutex    sync.RWMutex
	limit    int
	window   time.Duration
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

// Allow checks if a request from the given key is allowed
func (rl *RateLimiter) Allow(key string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	// Get existing requests for this key
	requests, exists := rl.requests[key]
	if !exists {
		requests = []time.Time{}
	}

	// Remove old requests outside the window
	var validRequests []time.Time
	for _, reqTime := range requests {
		if reqTime.After(cutoff) {
			validRequests = append(validRequests, reqTime)
		}
	}

	// Check if we're under the limit
	if len(validRequests) >= rl.limit {
		return false
	}

	// Add current request
	validRequests = append(validRequests, now)
	rl.requests[key] = validRequests

	return true
}

// Cleanup removes old entries to prevent memory leaks
func (rl *RateLimiter) Cleanup() {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	for key, requests := range rl.requests {
		var validRequests []time.Time
		for _, reqTime := range requests {
			if reqTime.After(cutoff) {
				validRequests = append(validRequests, reqTime)
			}
		}

		if len(validRequests) == 0 {
			delete(rl.requests, key)
		} else {
			rl.requests[key] = validRequests
		}
	}
}

// Global rate limiters
var (
	// General API rate limiter: 100 requests per minute
	generalLimiter = NewRateLimiter(100, time.Minute)

	// Authentication rate limiter: 10 login attempts per minute
	authLimiter = NewRateLimiter(10, time.Minute)

	// Strict rate limiter: 20 requests per minute for sensitive operations
	strictLimiter = NewRateLimiter(20, time.Minute)
)

// StartCleanupRoutine starts a goroutine to periodically clean up old rate limit entries
func StartCleanupRoutine() {
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			generalLimiter.Cleanup()
			authLimiter.Cleanup()
			strictLimiter.Cleanup()
		}
	}()
}

// RateLimit creates a rate limiting middleware
func RateLimit(limiter *RateLimiter, keyFunc func(*gin.Context) string) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := keyFunc(c)

		if !limiter.Allow(key) {
			dto.ErrorResponse(c, http.StatusTooManyRequests, "Rate limit exceeded. Please try again later.")
			c.Abort()
			return
		}

		c.Next()
	}
}

// IPBasedRateLimit creates an IP-based rate limiting middleware
func IPBasedRateLimit(limiter *RateLimiter) gin.HandlerFunc {
	return RateLimit(limiter, func(c *gin.Context) string {
		// Try to get real IP from headers first
		if ip := c.GetHeader("X-Forwarded-For"); ip != "" {
			return ip
		}
		if ip := c.GetHeader("X-Real-IP"); ip != "" {
			return ip
		}
		return c.ClientIP()
	})
}

// UserBasedRateLimit creates a user-based rate limiting middleware
func UserBasedRateLimit(limiter *RateLimiter) gin.HandlerFunc {
	return RateLimit(limiter, func(c *gin.Context) string {
		// If user is authenticated, use user ID
		if userID, exists := GetCurrentUserID(c); exists {
			return "user_" + string(rune(userID))
		}

		// Fall back to IP-based limiting for unauthenticated users
		if ip := c.GetHeader("X-Forwarded-For"); ip != "" {
			return "ip_" + ip
		}
		if ip := c.GetHeader("X-Real-IP"); ip != "" {
			return "ip_" + ip
		}
		return "ip_" + c.ClientIP()
	})
}

// APIKeyBasedRateLimit creates an API key-based rate limiting middleware
func APIKeyBasedRateLimit(limiter *RateLimiter) gin.HandlerFunc {
	return RateLimit(limiter, func(c *gin.Context) string {
		// Get API key from header
		apiKey := c.GetHeader("X-API-Key")
		if apiKey != "" {
			return "api_" + apiKey
		}

		// Fall back to IP-based limiting
		return "ip_" + c.ClientIP()
	})
}

// Common rate limiting middlewares
var (
	// General rate limiting: 100 requests per minute per IP
	GeneralRateLimit = IPBasedRateLimit(generalLimiter)

	// Authentication rate limiting: 10 attempts per minute per IP
	AuthRateLimit = IPBasedRateLimit(authLimiter)

	// Strict rate limiting: 20 requests per minute per user
	StrictRateLimit = UserBasedRateLimit(strictLimiter)

	// User-based general rate limiting
	UserRateLimit = UserBasedRateLimit(generalLimiter)
)

// RateLimitConfig allows custom rate limiting configuration
type RateLimitConfig struct {
	Limit   int
	Window  time.Duration
	KeyFunc func(*gin.Context) string
}

// CustomRateLimit creates a custom rate limiting middleware
func CustomRateLimit(config RateLimitConfig) gin.HandlerFunc {
	limiter := NewRateLimiter(config.Limit, config.Window)
	return RateLimit(limiter, config.KeyFunc)
}
