package middleware

import (
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

const (
	// RateLimitRequestsPerMinuteEnv is shared by every route protected by RateLimit.
	RateLimitRequestsPerMinuteEnv = "RATE_LIMIT_REQUESTS_PER_MINUTE"

	defaultRateLimitRequestsPerMinute = 10
	clientIdleTTL                     = 10 * time.Minute
	maxTrackedClients                 = 10_000
)

// CloudRunTrustedProxies contains Cloud Run's link-local reverse-proxy range.
// CloudFront appends the viewer address to X-Forwarded-For, so Gin resolves the
// right-most untrusted address rather than a client-supplied earlier value.
var CloudRunTrustedProxies = []string{"169.254.0.0/16"}

// RateLimitRequestsPerMinute returns the configured per-client limit. Invalid
// and non-positive values fall back to the safe default.
func RateLimitRequestsPerMinute() int {
	value, err := strconv.Atoi(os.Getenv(RateLimitRequestsPerMinuteEnv))
	if err != nil || value <= 0 {
		return defaultRateLimitRequestsPerMinute
	}
	return value
}

type clientLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RateLimit applies an in-process, per-client-IP token bucket. Each Cloud Run
// instance has its own buckets; this intentionally does not provide a shared,
// distributed limit across instances.
func RateLimit(requestsPerMinute int) gin.HandlerFunc {
	return newRateLimit(requestsPerMinute, maxTrackedClients)
}

func newRateLimit(requestsPerMinute, maxClients int) gin.HandlerFunc {
	if requestsPerMinute <= 0 {
		requestsPerMinute = defaultRateLimitRequestsPerMinute
	}
	if maxClients <= 0 {
		maxClients = maxTrackedClients
	}

	limiters := make(map[string]*clientLimiter)
	var mu sync.Mutex
	nextCleanup := time.Now().Add(clientIdleTTL)
	ratePerSecond := rate.Limit(float64(requestsPerMinute) / time.Minute.Seconds())

	return func(ctx *gin.Context) {
		now := time.Now()
		clientIP := ctx.ClientIP()

		mu.Lock()
		if now.After(nextCleanup) {
			for ip, client := range limiters {
				if now.Sub(client.lastSeen) >= clientIdleTTL {
					delete(limiters, ip)
				}
			}
			nextCleanup = now.Add(clientIdleTTL)
		}

		client, ok := limiters[clientIP]
		if !ok {
			if len(limiters) >= maxClients {
				oldestIP := ""
				var oldest time.Time
				for ip, candidate := range limiters {
					if oldestIP == "" || candidate.lastSeen.Before(oldest) {
						oldestIP = ip
						oldest = candidate.lastSeen
					}
				}
				delete(limiters, oldestIP)
			}
			client = &clientLimiter{limiter: rate.NewLimiter(ratePerSecond, requestsPerMinute)}
			limiters[clientIP] = client
		}
		client.lastSeen = now
		allowed := client.limiter.AllowN(now, 1)
		mu.Unlock()

		if !allowed {
			ctx.AbortWithStatus(http.StatusTooManyRequests)
			return
		}

		ctx.Next()
	}
}
