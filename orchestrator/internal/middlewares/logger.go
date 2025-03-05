package middlewares

import (
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

var ignoredPaths = map[string]bool{
	"/api/v1/internal/task": true,
	"/api/v1/health":        true,
}

func RequestLoggerWithZapConfig(logger *zap.Logger) echo.MiddlewareFunc {
	return middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:    true,
		LogStatus: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			path := c.Request().URL.RequestURI()
			if ignoredPaths[path] {
				return nil
			}

			startTime := time.Now()
			logger.Info("request",
				zap.String("URI", v.URI),
				zap.Int("status", v.Status),
				zap.String("method", c.Request().Method),
				zap.String("path", path),
				zap.String("user-agent", c.Request().UserAgent()),
				zap.String("referer", c.Request().Referer()),
				zap.Duration("duration", time.Since(startTime)),
			)
			return nil
		},
	})
}
