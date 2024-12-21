package middlewares

import (
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

func RequestLoggerWithZapConfig(logger *zap.Logger) echo.MiddlewareFunc {
	return middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:    true,
		LogStatus: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			startTime := time.Now()
			logger.Info("request",
				zap.String("URI", v.URI),
				zap.Int("status", v.Status),
				zap.String("method", c.Request().Method),
				zap.String("path", c.Request().URL.RequestURI()),
				zap.String("user-agent", c.Request().UserAgent()),
				zap.String("referer", c.Request().Referer()),
				zap.Duration("duration", time.Since(startTime)),
			)
			return nil
		},
	})
}
