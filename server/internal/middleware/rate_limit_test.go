package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRateLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var handled int
	router := gin.New()
	router.POST("/login", RateLimit(2), func(ctx *gin.Context) {
		handled++
		ctx.Status(http.StatusCreated)
	})

	for range 2 {
		w := requestFrom(router, "203.0.113.10:1234", "")
		require.Equal(t, http.StatusCreated, w.Code)
	}

	w := requestFrom(router, "203.0.113.10:1234", "")
	assert.Equal(t, http.StatusTooManyRequests, w.Code)
	assert.Equal(t, 2, handled, "a rejected request must not reach the handler")
}

func TestRateLimitUsesRealClientIPBehindCloudRun(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var handled int
	router := gin.New()
	require.NoError(t, router.SetTrustedProxies(CloudRunTrustedProxies))
	router.POST("/login", RateLimit(2), func(ctx *gin.Context) {
		handled++
		ctx.Status(http.StatusCreated)
	})

	// CloudFront appends the real viewer IP after any X-Forwarded-For value that
	// arrived from the viewer. The left-most value is therefore attacker-controlled.
	for _, forwardedFor := range []string{
		"198.51.100.9, 203.0.113.10",
		"192.0.2.42, 203.0.113.10",
	} {
		w := requestFrom(router, "169.254.8.129:8080", forwardedFor)
		require.Equal(t, http.StatusCreated, w.Code)
	}

	w := requestFrom(router, "169.254.8.129:8080", "203.0.113.10")
	assert.Equal(t, http.StatusTooManyRequests, w.Code,
		"forging X-Forwarded-For must not create a new rate-limit bucket")
	assert.Equal(t, 2, handled)
}

func TestRateLimitAllowsDifferentClients(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.POST("/login", RateLimit(1), func(ctx *gin.Context) {
		ctx.Status(http.StatusCreated)
	})

	assert.Equal(t, http.StatusCreated, requestFrom(router, "203.0.113.10:1234", "").Code)
	assert.Equal(t, http.StatusCreated, requestFrom(router, "203.0.113.11:1234", "").Code)
}

func TestRateLimitBoundsTrackedClients(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.POST("/login", newRateLimit(1, 2), func(ctx *gin.Context) {
		ctx.Status(http.StatusCreated)
	})

	assert.Equal(t, http.StatusCreated, requestFrom(router, "203.0.113.10:1234", "").Code)
	assert.Equal(t, http.StatusCreated, requestFrom(router, "203.0.113.11:1234", "").Code)
	assert.Equal(t, http.StatusCreated, requestFrom(router, "203.0.113.12:1234", "").Code)

	// The least-recently-used bucket is evicted when capacity is reached, so
	// this request is admitted rather than retaining an unbounded third entry.
	assert.Equal(t, http.StatusCreated, requestFrom(router, "203.0.113.10:1234", "").Code)
}

func TestRateLimitRequestsPerMinute(t *testing.T) {
	t.Setenv(RateLimitRequestsPerMinuteEnv, "25")
	assert.Equal(t, 25, RateLimitRequestsPerMinute())

	t.Setenv(RateLimitRequestsPerMinuteEnv, "not-a-number")
	assert.Equal(t, defaultRateLimitRequestsPerMinute, RateLimitRequestsPerMinute())
}

func requestFrom(router http.Handler, remoteAddr, forwardedFor string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/login", nil)
	req.RemoteAddr = remoteAddr
	if forwardedFor != "" {
		req.Header.Set("X-Forwarded-For", forwardedFor)
	}
	router.ServeHTTP(w, req)
	return w
}
