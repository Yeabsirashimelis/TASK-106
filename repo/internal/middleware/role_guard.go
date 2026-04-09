package middleware

import (
	"net/http"

	"github.com/eaglepoint/authapi/internal/models"
	"github.com/labstack/echo/v4"
)

func RequireRoles(roles ...models.Role) echo.MiddlewareFunc {
	allowed := make(map[models.Role]bool, len(roles))
	for _, r := range roles {
		allowed[r] = true
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			role, ok := c.Get("role").(models.Role)
			if !ok {
				return echo.NewHTTPError(http.StatusForbidden, "access denied")
			}
			if !allowed[role] {
				return echo.NewHTTPError(http.StatusForbidden, "insufficient permissions")
			}
			return next(c)
		}
	}
}
