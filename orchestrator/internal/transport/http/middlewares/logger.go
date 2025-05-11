package middlewares

import (
	"time"

	"github.com/alexGoLyceum/calculator-service/pkg/logging"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func RequestLoggerConfig(logger logging.Logger) echo.MiddlewareFunc {
	return middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:    true,
		LogStatus: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			path := c.Request().URL.RequestURI()
			startTime := time.Now()
			logger.Info("request",
				logging.String("URI", v.URI),
				logging.Int("status", v.Status),
				logging.String("method", c.Request().Method),
				logging.String("path", path),
				logging.String("user-agent", c.Request().UserAgent()),
				logging.String("referer", c.Request().Referer()),
				logging.Duration("duration", time.Since(startTime)),
			)
			return nil
		},
	})
}
