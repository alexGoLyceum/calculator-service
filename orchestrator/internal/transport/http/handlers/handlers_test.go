package handlers_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/alexGoLyceum/calculator-service/orchestrator/internal/models"
	"github.com/alexGoLyceum/calculator-service/orchestrator/internal/services"
	"github.com/alexGoLyceum/calculator-service/orchestrator/internal/transport/http/handlers"
	"github.com/alexGoLyceum/calculator-service/orchestrator/mocks"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestHandler_Register(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserService := mocks.NewMockUserService(ctrl)
	mockExpressionService := mocks.NewMockExpressionTaskService(ctrl)
	h := handlers.NewHandler(mockUserService, mockExpressionService)

	validPassword := "ValidPass123!"
	weakPassword := "weak"

	tests := []struct {
		name           string
		requestBody    string
		mockSetup      func()
		expectedStatus int
		expectedBody   string
	}{
		{
			name:        "successful registration",
			requestBody: `{"login":"test","password":"` + validPassword + `"}`,
			mockSetup: func() {
				mockUserService.EXPECT().Register(gomock.Any(), "test", validPassword).Return("token123", nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"token":"token123"}` + "\n",
		},
		{
			name:           "invalid request payload - empty fields",
			requestBody:    `{"login":"","password":""}`,
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Login and Password should not be empty"}` + "\n",
		},
		{
			name:        "weak password",
			requestBody: `{"login":"test","password":"` + weakPassword + `"}`,
			mockSetup: func() {
				mockUserService.EXPECT().Register(gomock.Any(), "test", weakPassword).
					Return("", services.ErrWeakPassword)
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedBody:   `{"error":"password must contain upper and lower case letters, a digit, a special character, and be 8-20 characters long"}` + "\n",
		},
		{
			name:        "user already exists",
			requestBody: `{"login":"test","password":"` + validPassword + `"}`,
			mockSetup: func() {
				mockUserService.EXPECT().Register(gomock.Any(), "test", validPassword).
					Return("", services.ErrUserWithLoginAlreadyExists)
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedBody:   `{"error":"user with this login already exists"}` + "\n",
		},
		{
			name:        "invalid login",
			requestBody: `{"login":"t","password":"` + validPassword + `"}`,
			mockSetup: func() {
				mockUserService.EXPECT().Register(gomock.Any(), "t", validPassword).
					Return("", services.ErrInvalidLogin)
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedBody:   `{"error":"login must be 3–32 characters long"}` + "\n",
		},
		{
			name:        "database unavailable during registration",
			requestBody: `{"login":"test","password":"ValidPass123!"}`,
			mockSetup: func() {
				mockUserService.EXPECT().
					Register(gomock.Any(), "test", "ValidPass123!").
					Return("", services.ErrDatabaseUnavailable)
			},
			expectedStatus: http.StatusServiceUnavailable,
			expectedBody:   `{"error":"service temporarily unavailable"}` + "\n",
		},
		{
			name:        "internal server error during registration",
			requestBody: `{"login":"test","password":"ValidPass123!"}`,
			mockSetup: func() {
				mockUserService.EXPECT().
					Register(gomock.Any(), "test", "ValidPass123!").
					Return("", errors.New("some unexpected error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"internal server error"}` + "\n",
		},
		{
			name:           "bind error - invalid json",
			requestBody:    `invalid json`,
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid request payload"}` + "\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(tt.requestBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := h.Register(c)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)
			assert.Equal(t, tt.expectedBody, rec.Body.String())
		})
	}
}

func TestHandler_Login(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserService := mocks.NewMockUserService(ctrl)
	mockExpressionService := mocks.NewMockExpressionTaskService(ctrl)
	h := handlers.NewHandler(mockUserService, mockExpressionService)

	validPassword := "ValidPass123!"

	tests := []struct {
		name           string
		requestBody    string
		mockSetup      func()
		expectedStatus int
		expectedBody   string
	}{
		{
			name:        "successful login",
			requestBody: `{"login":"test","password":"` + validPassword + `"}`,
			mockSetup: func() {
				mockUserService.EXPECT().Authenticate(gomock.Any(), "test", validPassword).Return("token123", nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"token":"token123"}` + "\n",
		},
		{
			name:           "invalid request payload - empty fields",
			requestBody:    `{"login":"","password":""}`,
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid request payload"}` + "\n",
		},
		{
			name:        "user not found",
			requestBody: `{"login":"test","password":"` + validPassword + `"}`,
			mockSetup: func() {
				mockUserService.EXPECT().Authenticate(gomock.Any(), "test", validPassword).
					Return("", services.ErrUserNotFoundByLogin)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"user with this login does not exist"}` + "\n",
		},
		{
			name:        "invalid password",
			requestBody: `{"login":"test","password":"` + validPassword + `"}`,
			mockSetup: func() {
				mockUserService.EXPECT().Authenticate(gomock.Any(), "test", validPassword).
					Return("", services.ErrInvalidPassword)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"invalid password"}` + "\n",
		},
		{
			name:        "database unavailable during login",
			requestBody: `{"login":"test","password":"ValidPass123!"}`,
			mockSetup: func() {
				mockUserService.EXPECT().
					Authenticate(gomock.Any(), "test", "ValidPass123!").
					Return("", services.ErrDatabaseUnavailable)
			},
			expectedStatus: http.StatusServiceUnavailable,
			expectedBody:   `{"error":"service temporarily unavailable"}` + "\n",
		},
		{
			name:        "internal server error during login",
			requestBody: `{"login":"test","password":"ValidPass123!"}`,
			mockSetup: func() {
				mockUserService.EXPECT().
					Authenticate(gomock.Any(), "test", "ValidPass123!").
					Return("", errors.New("some unexpected error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"internal server error"}` + "\n",
		},
		{
			name:           "bind error - invalid json",
			requestBody:    `invalid json`,
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid request payload"}` + "\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(tt.requestBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := h.Login(c)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)
			assert.Equal(t, tt.expectedBody, rec.Body.String())
		})
	}
}

func TestHandler_Calculate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserService := mocks.NewMockUserService(ctrl)
	mockExpressionService := mocks.NewMockExpressionTaskService(ctrl)
	h := handlers.NewHandler(mockUserService, mockExpressionService)

	testUserID := uuid.New()
	expressionID := uuid.New()

	tests := []struct {
		name           string
		userID         string
		requestBody    string
		mockSetup      func()
		expectedStatus int
		expectedBody   string
	}{
		{
			name:        "successful calculation",
			userID:      testUserID.String(),
			requestBody: `{"expression":"2+2"}`,
			mockSetup: func() {
				mockExpressionService.EXPECT().
					CreateExpressionTask(gomock.Any(), testUserID, "2+2").
					Return(expressionID, nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   `{"id":"` + expressionID.String() + `"}` + "\n",
		},
		{
			name:           "invalid user id",
			userID:         "invalid",
			requestBody:    `{"expression":"2+2"}`,
			mockSetup:      func() {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"unauthorized"}` + "\n",
		},
		{
			name:           "invalid request payload",
			userID:         testUserID.String(),
			requestBody:    `{"expression":""}`,
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid request payload"}` + "\n",
		},
		{
			name:        "expression error",
			userID:      testUserID.String(),
			requestBody: `{"expression":"invalid"}`,
			mockSetup: func() {
				mockExpressionService.EXPECT().
					CreateExpressionTask(gomock.Any(), testUserID, "invalid").
					Return(uuid.Nil, services.ErrInvalidExpression)
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedBody:   `{"error":"invalid expression"}` + "\n",
		},
		{
			name:        "unknown user id",
			userID:      testUserID.String(),
			requestBody: `{"expression":"2+2"}`,
			mockSetup: func() {
				mockExpressionService.EXPECT().
					CreateExpressionTask(gomock.Any(), testUserID, "2+2").
					Return(uuid.Nil, services.ErrUnknownUserID)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"error":"unknown user id"}` + "\n",
		},
		{
			name:        "database unavailable during calculation",
			userID:      testUserID.String(),
			requestBody: `{"expression":"2+2"}`,
			mockSetup: func() {
				mockExpressionService.EXPECT().
					CreateExpressionTask(gomock.Any(), testUserID, "2+2").
					Return(uuid.Nil, services.ErrDatabaseUnavailable)
			},
			expectedStatus: http.StatusServiceUnavailable,
			expectedBody:   `{"error":"service temporarily unavailable"}` + "\n",
		},
		{
			name:           "bind error - invalid json",
			userID:         testUserID.String(),
			requestBody:    `invalid json`,
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid request payload"}` + "\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/calculate", strings.NewReader(tt.requestBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.Set("user_id", tt.userID)

			err := h.Calculate(c)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)
			assert.Equal(t, tt.expectedBody, rec.Body.String())
		})
	}
}

func TestHandler_GetExpressions(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserService := mocks.NewMockUserService(ctrl)
	mockExpressionService := mocks.NewMockExpressionTaskService(ctrl)
	h := handlers.NewHandler(mockUserService, mockExpressionService)

	testUserID := uuid.New()
	expressions := []*models.Expression{
		{
			ID:         uuid.MustParse("b466ca50-0158-494a-b278-43375c6e3c34"),
			UserID:     uuid.Nil,
			Expression: "2+2",
			Status:     "completed",
			Result:     4,
		},
	}

	tests := []struct {
		name           string
		userID         string
		mockSetup      func()
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "successful get expressions",
			userID: testUserID.String(),
			mockSetup: func() {
				mockExpressionService.EXPECT().
					GetAllExpressions(gomock.Any(), testUserID).
					Return(expressions, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"expressions":[{"id":"b466ca50-0158-494a-b278-43375c6e3c34","user_id":"00000000-0000-0000-0000-000000000000","expression":"2+2","status":"completed","result":4}]}` + "\n",
		},
		{
			name:           "invalid user id",
			userID:         "invalid",
			mockSetup:      func() {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"unauthorized"}` + "\n",
		},
		{
			name:   "unknown user id",
			userID: testUserID.String(),
			mockSetup: func() {
				mockExpressionService.EXPECT().
					GetAllExpressions(gomock.Any(), testUserID).
					Return(nil, services.ErrUnknownUserID)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"error":"unauthorized"}` + "\n",
		},
		{
			name:   "database unavailable when getting expressions",
			userID: testUserID.String(),
			mockSetup: func() {
				mockExpressionService.EXPECT().
					GetAllExpressions(gomock.Any(), testUserID).
					Return(nil, services.ErrDatabaseUnavailable)
			},
			expectedStatus: http.StatusServiceUnavailable,
			expectedBody:   `{"error":"service temporarily unavailable"}` + "\n",
		},
		{
			name:   "internal server error when getting expressions",
			userID: testUserID.String(),
			mockSetup: func() {
				mockExpressionService.EXPECT().
					GetAllExpressions(gomock.Any(), testUserID).
					Return(nil, errors.New("some unexpected error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"internal server error"}` + "\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/expressions", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.Set("user_id", tt.userID)

			err := h.GetExpressions(c)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)
			assert.Equal(t, tt.expectedBody, rec.Body.String())
		})
	}
}

func TestHandler_GetExpressionByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserService := mocks.NewMockUserService(ctrl)
	mockExpressionService := mocks.NewMockExpressionTaskService(ctrl)
	h := handlers.NewHandler(mockUserService, mockExpressionService)

	expressionID := uuid.MustParse("b85cdb62-8d5c-435f-b921-35bbf229e822")
	expression := &models.Expression{
		ID:         expressionID,
		UserID:     uuid.Nil,
		Expression: "2+2",
		Status:     "completed",
		Result:     4,
	}

	tests := []struct {
		name           string
		idParam        string
		userID         string
		mockSetup      func()
		expectedStatus int
		expectedBody   string
	}{
		{
			name:    "successful get expression by id",
			idParam: expressionID.String(),
			userID:  uuid.New().String(),
			mockSetup: func() {
				mockExpressionService.EXPECT().
					GetExpressionById(gomock.Any(), expressionID).
					Return(expression, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"expression":{"id":"b85cdb62-8d5c-435f-b921-35bbf229e822","user_id":"00000000-0000-0000-0000-000000000000","expression":"2+2","status":"completed","result":4}}` + "\n",
		},
		{
			name:           "invalid id",
			idParam:        "invalid",
			userID:         uuid.New().String(),
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid request payload"}` + "\n",
		},
		{
			name:    "expression not found",
			idParam: expressionID.String(),
			userID:  uuid.New().String(),
			mockSetup: func() {
				mockExpressionService.EXPECT().
					GetExpressionById(gomock.Any(), expressionID).
					Return(nil, services.ErrUnknownExpressionsID)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"error":"unknown expressions id"}` + "\n",
		},
		{
			name:    "database unavailable when getting expression by id",
			idParam: expressionID.String(),
			userID:  uuid.New().String(),
			mockSetup: func() {
				mockExpressionService.EXPECT().
					GetExpressionById(gomock.Any(), expressionID).
					Return(nil, services.ErrDatabaseUnavailable)
			},
			expectedStatus: http.StatusServiceUnavailable,
			expectedBody:   `{"error":"service temporarily unavailable"}` + "\n",
		},
		{
			name:    "internal server error when getting expression by id",
			idParam: expressionID.String(),
			userID:  uuid.New().String(),
			mockSetup: func() {
				mockExpressionService.EXPECT().
					GetExpressionById(gomock.Any(), expressionID).
					Return(nil, errors.New("some unexpected error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"internal server error"}` + "\n",
		},
		{
			name:           "nil uuid",
			idParam:        "00000000-0000-0000-0000-000000000000",
			userID:         uuid.New().String(),
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid request payload"}` + "\n",
		},
		{
			name:           "invalid user id",
			idParam:        expressionID.String(),
			userID:         "invalid",
			mockSetup:      func() {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"unauthorized"}` + "\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/expressions/"+tt.idParam, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tt.idParam)
			c.Set("user_id", tt.userID)

			err := h.GetExpressionByID(c)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)
			assert.Equal(t, tt.expectedBody, rec.Body.String())
		})
	}
}

func TestHandler_Ping(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserService := mocks.NewMockUserService(ctrl)
	mockExpressionService := mocks.NewMockExpressionTaskService(ctrl)
	h := handlers.NewHandler(mockUserService, mockExpressionService)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.Ping(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "pong", rec.Body.String())
}

func TestNewHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserService := mocks.NewMockUserService(ctrl)
	mockExpressionService := mocks.NewMockExpressionTaskService(ctrl)

	h := handlers.NewHandler(mockUserService, mockExpressionService)

	assert.NotNil(t, h)
	_, ok := h.(handlers.Handler)
	assert.True(t, ok)
}
