package middleware

import (
	"time"

	"github.com/eaglepoint/authapi/internal/service"
	"github.com/labstack/echo/v4"
)

// MetricsMiddleware records request latency and error counts
func MetricsMiddleware(metrics *service.Metrics) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			metrics.IncrSessions()
			defer metrics.DecrSessions()

			start := time.Now()
			err := next(c)
			duration := time.Since(start)

			isError := err != nil || c.Response().Status >= 500
			metrics.RecordRequest(duration, isError)

			return err
		}
	}
}
