package middlewares_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/alexGoLyceum/calculator-service/orchestrator/internal/transport/http/middlewares"
	"github.com/alexGoLyceum/calculator-service/pkg/logging/mocks"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestRequestLoggerConfig_FullCoverage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks.NewMockLogger(ctrl)

	mockLogger.EXPECT().Info(
		"request",
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
	).Times(1)

	e := echo.New()
	e.Use(middlewares.RequestLoggerConfig(mockLogger))

	e.GET("/hello", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/hello?name=value", strings.NewReader(""))
	req.Header.Set("User-Agent", "unit-test-agent")
	req.Header.Set("Referer", "http://test.local")

	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "ok", rec.Body.String())
}
