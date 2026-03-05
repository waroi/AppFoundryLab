package middleware

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/example/appfoundrylab/backend/services/api-gateway/pkg/httpx"
	"github.com/redis/go-redis/v9"
)

type rateBucket struct {
	count       int
	windowStart time.Time
	lastSeen    time.Time
}

type RateLimiter struct {
	mu      sync.Mutex
	limit   int
	window  time.Duration
	buckets map[string]rateBucket
	cleanup time.Duration
	now     func() time.Time
}

type RedisRateLimiter struct {
	client      *redis.Client
	limit       int
	window      time.Duration
	prefix      string
	failureMode string
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	if limit < 1 {
		limit = 1
	}
	if window <= 0 {
		window = time.Minute
	}

	return &RateLimiter{
		limit:   limit,
		window:  window,
		buckets: make(map[string]rateBucket),
		cleanup: 10 * time.Minute,
		now:     time.Now,
	}
}

func (rl *RateLimiter) Allow(key string) (bool, int, time.Duration) {
	now := rl.now()
	remaining := rl.limit

	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.gc(now)
	bucket, ok := rl.buckets[key]
	if !ok || now.Sub(bucket.windowStart) >= rl.window {
		bucket = rateBucket{count: 0, windowStart: now}
	}

	bucket.count++
	bucket.lastSeen = now
	rl.buckets[key] = bucket

	if bucket.count > rl.limit {
		retryAfter := rl.window - now.Sub(bucket.windowStart)
		if retryAfter < 0 {
			retryAfter = 0
		}
		return false, 0, retryAfter
	}

	remaining = rl.limit - bucket.count
	return true, remaining, rl.window - now.Sub(bucket.windowStart)
}

func (rl *RateLimiter) gc(now time.Time) {
	for key, bucket := range rl.buckets {
		if now.Sub(bucket.lastSeen) > rl.cleanup {
			delete(rl.buckets, key)
		}
	}
}

func (rl *RateLimiter) Middleware(keyFn func(*http.Request) string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := keyFn(r)
			if key == "" {
				key = r.URL.Path
			}

			allowed, remaining, retryAfter := rl.Allow(key)
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(rl.limit))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
			w.Header().Set("X-RateLimit-Window-Seconds", strconv.Itoa(int(rl.window.Seconds())))
			if !allowed {
				w.Header().Set("Retry-After", strconv.Itoa(int(retryAfter.Seconds())+1))
				httpx.WriteError(w, r, http.StatusTooManyRequests, "rate_limit_exceeded", "too many requests", nil)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func RateLimitByIP(limit int, window time.Duration) func(http.Handler) http.Handler {
	rl := NewRateLimiter(limit, window)
	return rl.Middleware(func(r *http.Request) string {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			host = r.RemoteAddr
		}
		return host + ":" + r.URL.Path
	})
}

func NewRedisRateLimiter(client *redis.Client, prefix string, limit int, window time.Duration, failureMode string) *RedisRateLimiter {
	if limit < 1 {
		limit = 1
	}
	if window <= 0 {
		window = time.Minute
	}
	if prefix == "" {
		prefix = "default"
	}
	return &RedisRateLimiter{
		client:      client,
		limit:       limit,
		window:      window,
		prefix:      prefix,
		failureMode: normalizeRedisFailureMode(failureMode),
	}
}

var redisRateScript = redis.NewScript(`
local current = redis.call("INCR", KEYS[1])
if current == 1 then
  redis.call("PEXPIRE", KEYS[1], ARGV[1])
end
local ttl = redis.call("PTTL", KEYS[1])
return {current, ttl}
`)

func (rl *RedisRateLimiter) Allow(ctx context.Context, key string) (bool, int, time.Duration, error) {
	if rl.client == nil {
		return false, 0, 0, fmt.Errorf("redis client is nil")
	}
	if key == "" {
		key = "unknown"
	}
	windowMS := rl.window.Milliseconds()
	if windowMS <= 0 {
		windowMS = time.Minute.Milliseconds()
	}

	redisKey := fmt.Sprintf("ratelimit:%s:%s", rl.prefix, key)
	result, err := redisRateScript.Run(ctx, rl.client, []string{redisKey}, windowMS).Result()
	if err != nil {
		return false, 0, 0, err
	}

	items, ok := result.([]interface{})
	if !ok || len(items) != 2 {
		return false, 0, 0, fmt.Errorf("invalid redis rate-limit script result")
	}

	current, err := toInt64(items[0])
	if err != nil {
		return false, 0, 0, err
	}
	ttlMS, err := toInt64(items[1])
	if err != nil {
		return false, 0, 0, err
	}
	if ttlMS < 0 {
		ttlMS = windowMS
	}

	if int(current) > rl.limit {
		return false, 0, time.Duration(ttlMS) * time.Millisecond, nil
	}
	remaining := rl.limit - int(current)
	return true, remaining, time.Duration(ttlMS) * time.Millisecond, nil
}

func (rl *RedisRateLimiter) Middleware(keyFn func(*http.Request) string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := keyFn(r)
			if key == "" {
				key = r.URL.Path
			}

			allowed, remaining, retryAfter, err := rl.Allow(r.Context(), key)
			if err != nil {
				if rl.failureMode == "closed" {
					httpx.WriteError(w, r, http.StatusServiceUnavailable, "rate_limiter_unavailable", "rate limiter unavailable", nil)
					return
				}
				// Fail-open: keep availability if redis is temporarily unavailable.
				next.ServeHTTP(w, r)
				return
			}

			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(rl.limit))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
			w.Header().Set("X-RateLimit-Window-Seconds", strconv.Itoa(int(rl.window.Seconds())))
			if !allowed {
				w.Header().Set("Retry-After", strconv.Itoa(int(retryAfter.Seconds())+1))
				httpx.WriteError(w, r, http.StatusTooManyRequests, "rate_limit_exceeded", "too many requests", nil)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func RateLimitByIPDistributed(client *redis.Client, prefix string, limit int, window time.Duration) func(http.Handler) http.Handler {
	return RateLimitByIPDistributedWithFailureMode(client, prefix, limit, window, "open")
}

func RateLimitByIPDistributedWithFailureMode(client *redis.Client, prefix string, limit int, window time.Duration, failureMode string) func(http.Handler) http.Handler {
	rl := NewRedisRateLimiter(client, prefix, limit, window, failureMode)
	return rl.Middleware(func(r *http.Request) string {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			host = r.RemoteAddr
		}
		return host + ":" + r.URL.Path
	})
}

func normalizeRedisFailureMode(mode string) string {
	switch mode {
	case "closed":
		return "closed"
	default:
		return "open"
	}
}

func toInt64(value interface{}) (int64, error) {
	switch v := value.(type) {
	case int64:
		return v, nil
	case int:
		return int64(v), nil
	default:
		return 0, fmt.Errorf("unexpected script value type %T", value)
	}
}
