package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/alexGoLyceum/calculator-service/orchestrator/internal/models"
	"github.com/alexGoLyceum/calculator-service/orchestrator/internal/services"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type RegisterRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type RegisterResponse struct {
	Token string `json:"token,omitempty"`
	Error string `json:"error,omitempty"`
}

type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token,omitempty"`
	Error string `json:"error,omitempty"`
}

type CalculateRequest struct {
	Expression string `json:"expression"`
}

type CalculateResponse struct {
	ID    *uuid.UUID `json:"id,omitempty"`
	Error string     `json:"error,omitempty"`
}

type GetExpressionResponse struct {
	Expressions []*models.Expression `json:"expressions"`
}

type GetExpressionByIDResponse struct {
	Expression *models.Expression `json:"expression,omitempty"`
	Error      string             `json:"error,omitempty"`
}

type Handler interface {
	Register(c echo.Context) error
	Login(c echo.Context) error
	Calculate(c echo.Context) error
	GetExpressions(c echo.Context) error
	GetExpressionByID(c echo.Context) error
	Ping(c echo.Context) error
}

type handler struct {
	userService       services.UserService
	expressionService services.ExpressionTaskService
}

func NewHandler(userService services.UserService, expressionService services.ExpressionTaskService) Handler {
	return &handler{
		userService:       userService,
		expressionService: expressionService,
	}
}

func (h *handler) Register(c echo.Context) error {
	var request RegisterRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, RegisterResponse{Error: "invalid request payload"})
	}
	if request.Login == "" || request.Password == "" {
		return c.JSON(http.StatusBadRequest, RegisterResponse{Error: "Login and Password should not be empty"})
	}

	token, err := h.userService.Register(c.Request().Context(), request.Login, request.Password)
	if err != nil {
		if errors.Is(err, services.ErrUserWithLoginAlreadyExists) || errors.Is(err, services.ErrInvalidLogin) || errors.Is(err, services.ErrWeakPassword) {
			return c.JSON(http.StatusUnprocessableEntity, RegisterResponse{Error: err.Error()})
		}
		if errors.Is(err, services.ErrDatabaseUnavailable) {
			return c.JSON(http.StatusServiceUnavailable, RegisterResponse{Error: "service temporarily unavailable"})
		}
		return c.JSON(http.StatusInternalServerError, RegisterResponse{Error: "internal server error"})
	}
	return c.JSON(http.StatusOK, RegisterResponse{Token: token})
}

func (h *handler) Login(c echo.Context) error {
	var request LoginRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, LoginResponse{Error: "invalid request payload"})
	}
	if request.Login == "" || request.Password == "" {
		return c.JSON(http.StatusBadRequest, LoginResponse{Error: "invalid request payload"})
	}

	token, err := h.userService.Authenticate(c.Request().Context(), request.Login, request.Password)
	if err != nil {
		if errors.Is(err, services.ErrUserNotFoundByLogin) || errors.Is(err, services.ErrInvalidPassword) {
			return c.JSON(http.StatusUnauthorized, LoginResponse{Error: err.Error()})
		}
		if errors.Is(err, services.ErrDatabaseUnavailable) {
			return c.JSON(http.StatusServiceUnavailable, LoginResponse{Error: "service temporarily unavailable"})
		}
		return c.JSON(http.StatusInternalServerError, LoginResponse{Error: "internal server error"})
	}
	return c.JSON(http.StatusOK, LoginResponse{Token: token})
}

func (h *handler) Calculate(c echo.Context) error {
	parsedUserID, err := uuid.Parse(c.Get("user_id").(string))
	if err != nil {
		return c.JSON(http.StatusUnauthorized, CalculateResponse{Error: "unauthorized"})
	}

	var request CalculateRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, CalculateResponse{Error: "invalid request payload"})
	}
	if request.Expression == "" {
		return c.JSON(http.StatusBadRequest, CalculateResponse{Error: "invalid request payload"})
	}

	expressionID, err := h.expressionService.CreateExpressionTask(c.Request().Context(), parsedUserID, request.Expression)
	if err != nil {
		if services.IsExpressionError(err) {
			return c.JSON(http.StatusUnprocessableEntity, CalculateResponse{Error: err.Error()})
		}
		if errors.Is(err, services.ErrUnknownUserID) {
			return c.JSON(http.StatusNotFound, CalculateResponse{Error: err.Error()})
		}
		if errors.Is(err, services.ErrDatabaseUnavailable) {
			return c.JSON(http.StatusServiceUnavailable, CalculateResponse{Error: "service temporarily unavailable"})
		}
		return c.JSON(http.StatusInternalServerError, CalculateResponse{Error: "internal server error"})
	}
	return c.JSON(http.StatusCreated, CalculateResponse{ID: &expressionID})
}

func (h *handler) GetExpressions(c echo.Context) error {
	parsedUserID, err := uuid.Parse(c.Get("user_id").(string))
	if err != nil {
		return c.JSON(http.StatusUnauthorized, CalculateResponse{Error: "unauthorized"})
	}

	expressions, err := h.expressionService.GetAllExpressions(c.Request().Context(), parsedUserID)
	if err != nil {
		if errors.Is(err, services.ErrUnknownUserID) {
			return c.JSON(http.StatusNotFound, CalculateResponse{Error: "unauthorized"})
		}
		if errors.Is(err, services.ErrDatabaseUnavailable) {
			return c.JSON(http.StatusServiceUnavailable, CalculateResponse{Error: "service temporarily unavailable"})
		}
		return c.JSON(http.StatusInternalServerError, CalculateResponse{Error: "internal server error"})
	}

	return c.JSON(http.StatusOK, GetExpressionResponse{Expressions: expressions})
}

func (h *handler) GetExpressionByID(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, GetExpressionByIDResponse{Error: "invalid request payload"})
	}
	if id == uuid.Nil {
		return c.JSON(http.StatusBadRequest, GetExpressionByIDResponse{Error: "invalid request payload"})
	}

	parsedUserID, err := uuid.Parse(c.Get("user_id").(string))
	if err != nil {
		return c.JSON(http.StatusUnauthorized, GetExpressionByIDResponse{Error: "unauthorized"})
	}
	ctx := context.WithValue(context.Background(), "user_id", parsedUserID)
	expression, err := h.expressionService.GetExpressionById(ctx, id)

	if err != nil {
		if errors.Is(err, services.ErrUnknownExpressionsID) {
			return c.JSON(http.StatusNotFound, CalculateResponse{Error: err.Error()})
		}
		if errors.Is(err, services.ErrForbidden) {
			return c.JSON(http.StatusForbidden, CalculateResponse{Error: "you are not allowed to access this expression"})
		}
		if errors.Is(err, services.ErrDatabaseUnavailable) {
			return c.JSON(http.StatusServiceUnavailable, CalculateResponse{Error: "service temporarily unavailable"})
		}
		return c.JSON(http.StatusInternalServerError, CalculateResponse{Error: "internal server error"})
	}
	return c.JSON(http.StatusOK, GetExpressionByIDResponse{Expression: expression})
}

func (h *handler) Ping(c echo.Context) error {
	return c.String(http.StatusOK, "pong")
}
