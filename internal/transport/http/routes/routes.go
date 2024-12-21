package routes

import (
	"github.com/alexGoLyceum/calculator-service/internal/transport/http/handlers"

	"github.com/labstack/echo/v4"
)

func RegisterRoutes(e *echo.Echo) {
	e.POST("/api/v1/calculate", handlers.CalculateHandler)
}
