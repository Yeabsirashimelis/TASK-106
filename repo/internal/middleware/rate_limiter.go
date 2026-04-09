package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"golang.org/x/time/rate"
)

type RateLimiter struct {
	limiters sync.Map
	rps      rate.Limit
	burst    int
}

type limiterEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

func NewRateLimiter(rps float64, burst int) *RateLimiter {
	rl := &RateLimiter{
		rps:   rate.Limit(rps),
		burst: burst,
	}
	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Use account ID if authenticated, otherwise use IP
			var key string
			if id, ok := c.Get("account_id").(uuid.UUID); ok {
				key = "account:" + id.String()
			} else {
				key = "ip:" + c.RealIP()
			}

			entry, _ := rl.limiters.LoadOrStore(key, &limiterEntry{
				limiter:  rate.NewLimiter(rl.rps, rl.burst),
				lastSeen: time.Now(),
			})
			le := entry.(*limiterEntry)
			le.lastSeen = time.Now()

			if !le.limiter.Allow() {
				return echo.NewHTTPError(http.StatusTooManyRequests, "rate limit exceeded")
			}
			return next(c)
		}
	}
}

func (rl *RateLimiter) cleanup() {
	for {
		time.Sleep(10 * time.Minute)
		rl.limiters.Range(func(key, value interface{}) bool {
			le := value.(*limiterEntry)
			if time.Since(le.lastSeen) > 10*time.Minute {
				rl.limiters.Delete(key)
			}
			return true
		})
	}
}
