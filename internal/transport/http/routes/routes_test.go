package routes_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/alexGoLyceum/calculator-service/internal/transport/http/handlers"
	"github.com/alexGoLyceum/calculator-service/internal/transport/http/routes"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestRegisterRoutes(t *testing.T) {
	e := echo.New()
	routes.RegisterRoutes(e)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/calculate", strings.NewReader(`{"expression":"1+1"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	handler := handlers.CalculateHandler

	if assert.NoError(t, handler(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), `"result":"2"`)
	}
}
