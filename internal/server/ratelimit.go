package server

import (
	"math"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type ipBucket struct {
	tokens     float64
	lastUpdate time.Time
}

// IPRateLimiter implements a token bucket rate limiter per client IP
type IPRateLimiter struct {
	mu          sync.Mutex
	buckets     map[string]*ipBucket
	ratePerSec  float64
	burst       float64
	limit       int // requests per minute
	lastCleanup time.Time
	enabled     bool
}

// NewIPRateLimiter creates a new rate limiter (e.g. 60 req/min, burst 10)
func NewIPRateLimiter(requestsPerMinute int, burst int) *IPRateLimiter {
	if requestsPerMinute <= 0 {
		requestsPerMinute = 60
	}
	if burst <= 0 {
		burst = 10
	}
	return &IPRateLimiter{
		buckets:     make(map[string]*ipBucket),
		ratePerSec:  float64(requestsPerMinute) / 60.0,
		burst:       float64(burst),
		limit:       requestsPerMinute,
		lastCleanup: time.Now(),
		enabled:     true,
	}
}

// SetEnabled toggles rate limiting
func (rl *IPRateLimiter) SetEnabled(enabled bool) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.enabled = enabled
}

// Limit returns the configured rate limit (requests per minute)
func (rl *IPRateLimiter) Limit() int {
	return rl.limit
}

// RateLimitResult contains the outcome of an Allow check
type RateLimitResult struct {
	Allowed    bool
	Limit      int
	Remaining  int
	ResetTime  int64
	RetryAfter int64
}

// Allow checks if a request from the given client IP is allowed
func (rl *IPRateLimiter) Allow(r *http.Request) RateLimitResult {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	ip := GetClientIP(r)
	now := time.Now()

	if !rl.enabled {
		return RateLimitResult{
			Allowed:    true,
			Limit:      rl.limit,
			Remaining:  rl.limit,
			ResetTime:  now.Add(time.Minute).Unix(),
			RetryAfter: 0,
		}
	}

	// Periodic cleanup of stale IPs every 5 minutes
	if now.Sub(rl.lastCleanup) > 5*time.Minute {
		for k, b := range rl.buckets {
			if now.Sub(b.lastUpdate) > 10*time.Minute {
				delete(rl.buckets, k)
			}
		}
		rl.lastCleanup = now
	}

	bucket, exists := rl.buckets[ip]
	if !exists {
		bucket = &ipBucket{
			tokens:     rl.burst,
			lastUpdate: now,
		}
		rl.buckets[ip] = bucket
	}

	elapsed := now.Sub(bucket.lastUpdate).Seconds()
	if elapsed > 0 {
		bucket.tokens += elapsed * rl.ratePerSec
		if bucket.tokens > rl.burst {
			bucket.tokens = rl.burst
		}
		bucket.lastUpdate = now
	}

	res := RateLimitResult{
		Limit: rl.limit,
	}

	if bucket.tokens >= 1.0 {
		bucket.tokens -= 1.0
		res.Allowed = true
		res.Remaining = int(bucket.tokens)
		fillSeconds := (rl.burst - bucket.tokens) / rl.ratePerSec
		res.ResetTime = now.Add(time.Duration(fillSeconds * float64(time.Second))).Unix()
		if res.ResetTime <= now.Unix() {
			res.ResetTime = now.Unix() + 1
		}
	} else {
		res.Allowed = false
		res.Remaining = 0
		waitSeconds := (1.0 - bucket.tokens) / rl.ratePerSec
		if waitSeconds < 1.0 {
			waitSeconds = 1.0
		}
		res.ResetTime = now.Add(time.Duration(waitSeconds * float64(time.Second))).Unix()
		if res.ResetTime <= now.Unix() {
			res.ResetTime = now.Unix() + 1
		}
		res.RetryAfter = int64(math.Ceil(waitSeconds))
		if res.RetryAfter < 1 {
			res.RetryAfter = 1
		}
	}

	return res
}

// SetHeaders sets standard rate limit headers on the response writer
func (res RateLimitResult) SetHeaders(w http.ResponseWriter) {
	w.Header().Set("X-RateLimit-Limit", strconv.Itoa(res.Limit))
	w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(res.Remaining))
	w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(res.ResetTime, 10))
	if !res.Allowed && res.RetryAfter > 0 {
		w.Header().Set("Retry-After", strconv.FormatInt(res.RetryAfter, 10))
	}
}

// GetClientIP extracts client IP addressing reverse proxies and CDNs
func GetClientIP(r *http.Request) string {
	// 1. Cloudflare
	if cfIP := strings.TrimSpace(r.Header.Get("CF-Connecting-IP")); cfIP != "" {
		return cfIP
	}

	// 2. X-Forwarded-For (take the first hop)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			ip := strings.TrimSpace(parts[0])
			if ip != "" {
				return ip
			}
		}
	}

	// 3. X-Real-IP
	if xrip := strings.TrimSpace(r.Header.Get("X-Real-IP")); xrip != "" {
		return xrip
	}

	// 4. RemoteAddr
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil && host != "" {
		return host
	}

	return strings.TrimSpace(r.RemoteAddr)
}
