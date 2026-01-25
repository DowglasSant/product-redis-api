package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type RateLimitConfig struct {
	RequestsPerWindow int
	WindowSize        time.Duration
	Enabled           bool
}

type RateLimiter struct {
	redis  *redis.Client
	config RateLimitConfig
	logger *zap.Logger
}

func NewRateLimiter(redisClient *redis.Client, config RateLimitConfig, logger *zap.Logger) *RateLimiter {
	return &RateLimiter{
		redis:  redisClient,
		config: config,
		logger: logger,
	}
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !rl.config.Enabled {
			next.ServeHTTP(w, r)
			return
		}

		identifier := rl.getIdentifier(r)
		key := fmt.Sprintf("ratelimit:%s", identifier)

		allowed, remaining, resetTime, err := rl.checkRateLimit(r.Context(), key)
		if err != nil {
			rl.logger.Error("rate limit check failed", zap.Error(err), zap.String("identifier", identifier))
			next.ServeHTTP(w, r)
			return
		}

		w.Header().Set("X-RateLimit-Limit", strconv.Itoa(rl.config.RequestsPerWindow))
		w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
		w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(resetTime, 10))

		if !allowed {
			rl.logger.Warn("rate limit exceeded",
				zap.String("identifier", identifier),
				zap.Int("limit", rl.config.RequestsPerWindow),
			)
			rl.rateLimitExceededResponse(w, resetTime)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (rl *RateLimiter) getIdentifier(r *http.Request) string {
	if user := GetUserFromContext(r.Context()); user != nil && user.Subject != "" {
		return "user:" + user.Subject
	}

	ip := r.Header.Get("X-Real-IP")
	if ip == "" {
		ip = r.Header.Get("X-Forwarded-For")
	}
	if ip == "" {
		ip = r.RemoteAddr
	}

	return "ip:" + ip
}

func (rl *RateLimiter) checkRateLimit(ctx context.Context, key string) (bool, int, int64, error) {
	now := time.Now()
	windowStart := now.Add(-rl.config.WindowSize)
	resetTime := now.Add(rl.config.WindowSize).Unix()

	script := redis.NewScript(`
		local key = KEYS[1]
		local now = tonumber(ARGV[1])
		local window_start = tonumber(ARGV[2])
		local limit = tonumber(ARGV[3])
		local window_size_ms = tonumber(ARGV[4])

		-- Remove old entries outside the window
		redis.call('ZREMRANGEBYSCORE', key, '-inf', window_start)

		-- Count current requests in window
		local current = redis.call('ZCARD', key)

		if current < limit then
			-- Add current request
			redis.call('ZADD', key, now, now .. ':' .. math.random())
			-- Set expiry on the key
			redis.call('PEXPIRE', key, window_size_ms)
			return {1, limit - current - 1}
		else
			return {0, 0}
		end
	`)

	nowMs := now.UnixMilli()
	windowStartMs := windowStart.UnixMilli()
	windowSizeMs := rl.config.WindowSize.Milliseconds()

	result, err := script.Run(ctx, rl.redis, []string{key},
		nowMs,
		windowStartMs,
		rl.config.RequestsPerWindow,
		windowSizeMs,
	).Slice()

	if err != nil {
		return false, 0, resetTime, fmt.Errorf("rate limit script failed: %w", err)
	}

	allowed := result[0].(int64) == 1
	remaining := int(result[1].(int64))

	return allowed, remaining, resetTime, nil
}

func (rl *RateLimiter) rateLimitExceededResponse(w http.ResponseWriter, resetTime int64) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Retry-After", strconv.FormatInt(resetTime-time.Now().Unix(), 10))
	w.WriteHeader(http.StatusTooManyRequests)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"error":       "rate_limit_exceeded",
		"message":     "Too many requests. Please try again later.",
		"retry_after": resetTime - time.Now().Unix(),
	})
}

func (rl *RateLimiter) GetRateLimitInfo(ctx context.Context, identifier string) (int, int, error) {
	key := fmt.Sprintf("ratelimit:%s", identifier)
	now := time.Now()
	windowStart := now.Add(-rl.config.WindowSize)

	pipe := rl.redis.Pipeline()
	pipe.ZRemRangeByScore(ctx, key, "-inf", strconv.FormatInt(windowStart.UnixMilli(), 10))
	countCmd := pipe.ZCard(ctx, key)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, 0, err
	}

	current := int(countCmd.Val())
	remaining := rl.config.RequestsPerWindow - current
	if remaining < 0 {
		remaining = 0
	}

	return current, remaining, nil
}
