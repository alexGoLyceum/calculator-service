package routes

import (
	"github.com/alexGoLyceum/calculator-service/orchestrator/internal/transport/http/handlers"

	"github.com/labstack/echo/v4"
)

func RegisterRoutes(e *echo.Echo, h handlers.Handler) {
	e.POST("/api/v1/register", h.Register)
	e.POST("/api/v1/login", h.Login)
	e.POST("/api/v1/calculate", h.Calculate)
	e.GET("/api/v1/expressions", h.GetExpressions)
	e.GET("/api/v1/expressions/:id", h.GetExpressionByID)
	e.GET("/api/v1/ping", h.Ping)
}
