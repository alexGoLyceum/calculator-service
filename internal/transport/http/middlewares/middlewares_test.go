package middlewares_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alexGoLyceum/calculator-service/internal/transport/http/middlewares"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestRequestLoggerWithZapConfig(t *testing.T) {
	logger, _ := zap.NewProduction()
	e := echo.New()
	e.Use(middlewares.RequestLoggerWithZapConfig(logger))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	})
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "test", rec.Body.String())
}
