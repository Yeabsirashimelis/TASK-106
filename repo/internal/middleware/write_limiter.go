package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// WriteLimiter enforces per-account write operation limits (anti-spam).
// Default: 60 writes per minute per account.
type WriteLimiter struct {
	mu       sync.Mutex
	counters map[uuid.UUID]*writeCounter
	limit    int
	window   time.Duration
}

type writeCounter struct {
	count     int
	windowEnd time.Time
}

func NewWriteLimiter(limit int, window time.Duration) *WriteLimiter {
	wl := &WriteLimiter{
		counters: make(map[uuid.UUID]*writeCounter),
		limit:    limit,
		window:   window,
	}
	go wl.cleanup()
	return wl
}

func (wl *WriteLimiter) Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Only limit write operations
			method := c.Request().Method
			if method != http.MethodPost && method != http.MethodPut && method != http.MethodDelete && method != http.MethodPatch {
				return next(c)
			}

			accountID, ok := c.Get("account_id").(uuid.UUID)
			if !ok {
				return next(c)
			}

			wl.mu.Lock()
			now := time.Now()
			counter, exists := wl.counters[accountID]
			if !exists || now.After(counter.windowEnd) {
				counter = &writeCounter{
					count:     0,
					windowEnd: now.Add(wl.window),
				}
				wl.counters[accountID] = counter
			}

			if counter.count >= wl.limit {
				wl.mu.Unlock()
				return echo.NewHTTPError(http.StatusTooManyRequests, "write operation limit exceeded, please try again later")
			}

			counter.count++
			wl.mu.Unlock()

			return next(c)
		}
	}
}

func (wl *WriteLimiter) cleanup() {
	for {
		time.Sleep(5 * time.Minute)
		wl.mu.Lock()
		now := time.Now()
		for id, counter := range wl.counters {
			if now.After(counter.windowEnd) {
				delete(wl.counters, id)
			}
		}
		wl.mu.Unlock()
	}
}
