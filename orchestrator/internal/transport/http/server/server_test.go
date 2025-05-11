package server_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alexGoLyceum/calculator-service/orchestrator/internal/config"
	"github.com/alexGoLyceum/calculator-service/orchestrator/internal/transport/http/server"
	"github.com/alexGoLyceum/calculator-service/orchestrator/mocks"
	logmock "github.com/alexGoLyceum/calculator-service/pkg/logging/mocks"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewServer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cfg := &config.Config{
		Orchestrator: config.OrchestratorConfig{
			HTTPHost: "127.0.0.1",
			HTTPPort: 0,
		},
	}
	logger := logmock.NewMockLogger(ctrl)
	handler := mocks.NewMockHandler(ctrl)
	jwtManager := mocks.NewMockJWTManager(ctrl)

	srv := server.NewServer(cfg, logger, handler, jwtManager)
	assert.NotNil(t, srv)
}

func TestServer_Start(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cfg := &config.Config{
		Orchestrator: config.OrchestratorConfig{
			HTTPHost: "127.0.0.1",
			HTTPPort: 0,
		},
	}
	logger := logmock.NewMockLogger(ctrl)
	handler := mocks.NewMockHandler(ctrl)
	jwtManager := mocks.NewMockJWTManager(ctrl)

	logger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()

	srv := server.NewServer(cfg, logger, handler, jwtManager)

	s := srv.(*server.Impl)

	s.Echo.GET("/api/v1/ping", func(c echo.Context) error {
		return c.String(http.StatusOK, "pong")
	})

	go func() {
		_ = s.Start()
	}()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ping", nil)
	rec := httptest.NewRecorder()
	s.Echo.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "pong", rec.Body.String())

	_ = s.Echo.Close()
}
