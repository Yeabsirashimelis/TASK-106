package middleware

import (
	"net/http"
	"strings"

	"github.com/eaglepoint/authapi/internal/service"
	"github.com/labstack/echo/v4"
)

func JWTAuth(tokenService *service.TokenService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			auth := c.Request().Header.Get("Authorization")
			if auth == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "missing authorization header")
			}

			parts := strings.SplitN(auth, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid authorization header format")
			}

			claims, err := tokenService.ValidateAccessToken(parts[1])
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid or expired token")
			}

			c.Set("claims", claims)
			c.Set("account_id", claims.AccountID)
			c.Set("role", claims.Role)
			return next(c)
		}
	}
}
