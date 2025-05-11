package middlewares_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alexGoLyceum/calculator-service/orchestrator/internal/transport/http/middlewares"
	"github.com/alexGoLyceum/calculator-service/orchestrator/mocks"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestJWTMiddleware(t *testing.T) {
	tests := []struct {
		name               string
		path               string
		authHeader         string
		mockParse          func(m *mocks.MockJWTManager)
		expectedStatusCode int
	}{
		{
			name:               "Skip auth for /register",
			path:               "/api/v1/register",
			authHeader:         "",
			mockParse:          nil,
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "Skip auth for /login",
			path:               "/api/v1/login",
			authHeader:         "",
			mockParse:          nil,
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "Missing Authorization header",
			path:               "/api/v1/protected",
			authHeader:         "",
			mockParse:          nil,
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:       "Invalid token",
			path:       "/api/v1/protected",
			authHeader: "Bearer invalidtoken",
			mockParse: func(m *mocks.MockJWTManager) {
				m.EXPECT().Parse("invalidtoken").Return(uuid.UUID{}, errors.New("invalid token"))
			},
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:       "Valid token",
			path:       "/api/v1/protected",
			authHeader: "Bearer validtoken",
			mockParse: func(m *mocks.MockJWTManager) {
				m.EXPECT().Parse("validtoken").Return(uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"), nil)
			},
			expectedStatusCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockJWT := mocks.NewMockJWTManager(ctrl)
			if tt.mockParse != nil {
				tt.mockParse(mockJWT)
			}

			e := echo.New()
			e.GET(tt.path, func(c echo.Context) error {
				return c.String(http.StatusOK, "ok")
			}, middlewares.JWTMiddleware(mockJWT))

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			req.Header.Set("Authorization", tt.authHeader)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatusCode, rec.Code)
		})
	}
}
