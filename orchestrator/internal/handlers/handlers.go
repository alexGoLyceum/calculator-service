package handlers

import (
	"net/http"

	"github.com/alexGoLyceum/calculator-service/orchestrator/internal/taskmanager"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type CalculateRequest struct {
	Expression string `json:"expression"`
}

type CalculateResponse struct {
	ID    *uuid.UUID `json:"id,omitempty"`
	Error string     `json:"error,omitempty"`
}

type GetExpressionResponse struct {
	Expressions []taskmanager.Expression `json:"expressions"`
}

type GetExpressionByIDResponse struct {
	Expression *taskmanager.Expression `json:"expression,omitempty"`
	Error      string                  `json:"error,omitempty"`
}

type GetTaskResponse struct {
	Task taskmanager.Task `json:"task"`
}

type SetTaskResultRequest struct {
	Task   taskmanager.Task `json:"task"`
	Result float64          `json:"result"`
}

type SetTaskResultResponse struct {
	Error string `json:"error"`
}

func CalculateHandler(c echo.Context, tm *taskmanager.TaskManager) error {
	var request CalculateRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, CalculateResponse{Error: "Invalid JSON"})
	}

	if err := tm.ValidateExpression(request.Expression); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, CalculateResponse{Error: err.Error()})
	}

	expressionID, err := tm.CreateAndStoreTasksFromExpression(request.Expression)
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, CalculateResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusCreated, CalculateResponse{ID: &expressionID})
}

func GetExpressionsHandler(c echo.Context, tm *taskmanager.TaskManager) error {
	return c.JSON(http.StatusOK, GetExpressionResponse{Expressions: tm.GetExpressions()})
}

func GetExpressionByIDHandler(c echo.Context, tm *taskmanager.TaskManager) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, GetExpressionByIDResponse{Error: "Invalid ID"})
	}

	expression, err := tm.GetExpressionByID(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, GetExpressionByIDResponse{Error: err.Error()})
	}
	return c.JSON(http.StatusOK, GetExpressionByIDResponse{Expression: expression})
}

func GetTaskHandler(c echo.Context, tm *taskmanager.TaskManager) error {
	task := tm.GetTask()
	if task == nil {
		return c.NoContent(http.StatusNotFound)
	}
	return c.JSON(http.StatusOK, GetTaskResponse{Task: *task})
}

func SetTaskResultHandler(c echo.Context, tm *taskmanager.TaskManager) error {
	var request SetTaskResultRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, SetTaskResultResponse{Error: "Invalid JSON"})
	}

	if err := tm.SetResult(request.Task, request.Result); err != nil {
		return c.JSON(http.StatusNotFound, SetTaskResultResponse{Error: err.Error()})
	}

	return c.NoContent(http.StatusOK)
}
