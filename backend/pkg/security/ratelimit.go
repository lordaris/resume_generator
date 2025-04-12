package security

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

// Default rate limit parameters
const (
	DefaultRateLimit    = 100
	DefaultRateInterval = time.Minute
)

// ErrRateLimitExceeded is returned when the rate limit is exceeded
var ErrRateLimitExceeded = errors.New("rate limit exceeded")

// RateLimiterConfig contains configuration options for rate limiting
type RateLimiterConfig struct {
	// Redis is the Redis client to use for rate limiting
	Redis *redis.Client
	// Limit is the maximum number of requests per interval
	Limit int
	// Interval is the time period for the limit
	Interval time.Duration
	// SkipSuccessfulAuth determines if successful authentication requests should bypass rate limiting
	SkipSuccessfulAuth bool
}

// RateLimiter provides rate limiting functionality
type RateLimiter struct {
	redis    *redis.Client
	limit    int
	interval time.Duration
	skipAuth bool
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(config RateLimiterConfig) *RateLimiter {
	if config.Redis == nil {
		panic("Redis client is required for rate limiting")
	}

	// Set defaults if not provided
	if config.Limit <= 0 {
		config.Limit = DefaultRateLimit
	}
	if config.Interval <= 0 {
		config.Interval = DefaultRateInterval
	}

	return &RateLimiter{
		redis:    config.Redis,
		limit:    config.Limit,
		interval: config.Interval,
		skipAuth: config.SkipSuccessfulAuth,
	}
}

// getIPAddress extracts the real IP address from the request
func getIPAddress(r *http.Request) string {
	// Check for X-Forwarded-For header first
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For can contain multiple IPs, we want the first one
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			clientIP := strings.TrimSpace(ips[0])
			return clientIP
		}
	}

	// Check for X-Real-IP header
	if xrip := r.Header.Get("X-Real-IP"); xrip != "" {
		return xrip
	}

	// Extract IP from RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		// If there's an error, just use RemoteAddr as-is
		return r.RemoteAddr
	}
	return ip
}

// getLimitKey generates a Redis key for rate limiting
func (rl *RateLimiter) getLimitKey(r *http.Request) string {
	ip := getIPAddress(r)
	path := r.URL.Path
	return fmt.Sprintf("ratelimit:%s:%s", ip, path)
}

// CheckRateLimit checks if the request is within the rate limit
func (rl *RateLimiter) CheckRateLimit(ctx context.Context, r *http.Request) (int, error) {
	key := rl.getLimitKey(r)
	now := time.Now().Unix()
	windowStart := now - int64(rl.interval.Seconds())

	// Remove old entries (outside the current window)
	err := rl.redis.ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(windowStart, 10)).Err()
	if err != nil {
		log.Error().Err(err).Msg("Failed to remove old rate limit entries")
		// Allow the request to proceed if we can't communicate with Redis
		return 0, nil
	}

	// Count existing requests in the current window
	count, err := rl.redis.ZCard(ctx, key).Result()
	if err != nil {
		log.Error().Err(err).Msg("Failed to count rate limit entries")
		// Allow the request to proceed if we can't communicate with Redis
		return 0, nil
	}

	// Check if the rate limit has been exceeded
	if count >= int64(rl.limit) {
		return int(count), ErrRateLimitExceeded
	}

	// Add the current request to the sorted set with the current timestamp as score
	err = rl.redis.ZAdd(ctx, key, redis.Z{
		Score:  float64(now),
		Member: now,
	}).Err()
	if err != nil {
		log.Error().Err(err).Msg("Failed to add rate limit entry")
	}

	// Set expiration for the key to the rate limit interval + 1 minute
	err = rl.redis.Expire(ctx, key, rl.interval+time.Minute).Err()
	if err != nil {
		log.Error().Err(err).Msg("Failed to set rate limit key expiration")
	}

	return int(count) + 1, nil
}

// Middleware provides rate limiting middleware
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip rate limiting for OPTIONS requests (for CORS preflight)
		if r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}

		// Check rate limit
		count, err := rl.CheckRateLimit(r.Context(), r)
		if err != nil {
			if errors.Is(err, ErrRateLimitExceeded) {
				// Set rate limit headers
				w.Header().Set("X-RateLimit-Limit", strconv.Itoa(rl.limit))
				w.Header().Set("X-RateLimit-Remaining", "0")
				w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(rl.interval).Unix(), 10))
				w.Header().Set("Retry-After", strconv.Itoa(int(rl.interval.Seconds())))

				// Return 429 Too Many Requests
				w.WriteHeader(http.StatusTooManyRequests)
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{"error":"Rate limit exceeded","status":429}`))
				return
			}

			// For other errors, log and continue
			log.Error().Err(err).Msg("Rate limiting error")
		}

		// Set rate limit headers on successful requests too
		w.Header().Set("X-RateLimit-Limit", strconv.Itoa(rl.limit))
		w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(rl.limit-count))
		w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(rl.interval).Unix(), 10))

		next.ServeHTTP(w, r)
	})
}
