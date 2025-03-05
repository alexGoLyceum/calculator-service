package server_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alexGoLyceum/calculator-service/orchestrator/internal/config"
	"github.com/alexGoLyceum/calculator-service/orchestrator/internal/server"
	"github.com/alexGoLyceum/calculator-service/pkg/logging"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewOrchestrator(t *testing.T) {
	cfg := &config.Config{
		Orchestrator: config.OrchestratorConfig{Host: "localhost", Port: 8080},
		TaskManager:  config.TaskManagerConfig{},
		Log:          logging.LogConfig{Level: "debug", Path: "logfile.log"},
	}
	logger, _ := zap.NewDevelopment()

	orchestrator := server.NewOrchestrator(cfg, logger)

	assert.NotNil(t, orchestrator)
	assert.NotNil(t, orchestrator.Echo)
	assert.Equal(t, cfg, orchestrator.Config)
	assert.Equal(t, logger, orchestrator.Logger)
	assert.NotNil(t, orchestrator.TaskManager)
}

func TestOrchestrator_Start(t *testing.T) {
	cfg := &config.Config{
		Orchestrator: config.OrchestratorConfig{
			Host: "localhost",
			Port: 8080,
		},
	}

	logger, _ := zap.NewProduction()

	orchestrator := server.NewOrchestrator(cfg, logger)

	go func() {
		if err := orchestrator.Start(); err != nil {
			t.Errorf("server failed to start: %v", err)
			return
		}
	}()

	rec := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "http://localhost:8080/api/v1/expressions", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	orchestrator.Echo.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}
