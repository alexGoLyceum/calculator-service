package middlewares

import (
	"net/http"
	"strings"

	"github.com/alexGoLyceum/calculator-service/orchestrator/internal/auth"

	"github.com/labstack/echo/v4"
)

func JWTMiddleware(tokenManager auth.JWTManager) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			switch c.Path() {
			case "/api/v1/register", "/api/v1/login", "/api/v1/ping":
				return next(c)
			}

			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Authorization header is required"})
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			userID, err := tokenManager.Parse(tokenString)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid token"})
			}

			c.Set("user_id", userID.String())
			return next(c)
		}
	}
}
