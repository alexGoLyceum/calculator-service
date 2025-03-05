package middlewares_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alexGoLyceum/calculator-service/orchestrator/internal/middlewares"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestRequestLoggerWithZapConfig_IgnoredPath(t *testing.T) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	e := echo.New()
	e.Use(middlewares.RequestLoggerWithZapConfig(logger))
	e.GET("/api/v1/internal/task", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/internal/task", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRequestLoggerWithZapConfig_NotIgnoredPath(t *testing.T) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	e := echo.New()
	e.Use(middlewares.RequestLoggerWithZapConfig(logger))
	e.GET("/api/v1/some/other/path", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/some/other/path", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}
