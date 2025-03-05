package routes

import (
	"github.com/alexGoLyceum/calculator-service/orchestrator/internal/handlers"
	"github.com/alexGoLyceum/calculator-service/orchestrator/internal/taskmanager"

	"github.com/labstack/echo/v4"
)

func RegisterRoutes(e *echo.Echo, taskManager *taskmanager.TaskManager) {
	e.POST("/api/v1/calculate", func(c echo.Context) error {
		return handlers.CalculateHandler(c, taskManager)
	})
	e.GET("/api/v1/expressions", func(c echo.Context) error {
		return handlers.GetExpressionsHandler(c, taskManager)
	})
	e.GET("/api/v1/expressions/:id", func(c echo.Context) error {
		return handlers.GetExpressionByIDHandler(c, taskManager)
	})
	e.GET("/api/v1/internal/task", func(c echo.Context) error {
		return handlers.GetTaskHandler(c, taskManager)
	})
	e.POST("/api/v1/internal/task", func(c echo.Context) error {
		return handlers.SetTaskResultHandler(c, taskManager)
	})
}
