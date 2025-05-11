package app_test

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/alexGoLyceum/calculator-service/agent/app"
	"github.com/alexGoLyceum/calculator-service/agent/internal/config"
	"github.com/alexGoLyceum/calculator-service/agent/mocks"
	logmock "github.com/alexGoLyceum/calculator-service/pkg/logging/mocks"

	"go.uber.org/mock/gomock"
)

func TestStart(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := logmock.NewMockLogger(ctrl)
	mockAgent := mocks.NewMockAgent(ctrl)

	mockLogger.EXPECT().Info("Starting agent").Times(1)
	mockAgent.EXPECT().Start().Return(nil).Times(1)

	a := &app.Impl{
		Config: &config.Config{},
		Logger: mockLogger,
		Agent:  mockAgent,
	}

	a.Start()
}

func TestStart_WithError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := logmock.NewMockLogger(ctrl)
	mockAgent := mocks.NewMockAgent(ctrl)

	testErr := fmt.Errorf("test error")
	mockLogger.EXPECT().Info("Starting agent").Times(1)
	mockAgent.EXPECT().Start().Return(testErr).Times(1)
	mockLogger.EXPECT().Error("Failed to start agent", gomock.Any()).Times(1)

	a := &app.Impl{
		Config: &config.Config{},
		Logger: mockLogger,
		Agent:  mockAgent,
	}

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when agent start fails")
		}
	}()

	a.Start()
}

func TestNewApplication_PanicOnInvalidOrchestratorConfig(t *testing.T) {
	logDir := t.TempDir()
	logPath := filepath.Join(logDir, "app.log")
	t.Setenv("ORCHESTRATOR_HOST", "localhost")
	t.Setenv("ORCHESTRATOR_GRPC_PORT", "0")
	t.Setenv("LOG_LEVEL", "info")
	t.Setenv("LOG_PATH", logPath)

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on missing orchestrator config")
		}
	}()

	_ = app.NewApplication()
}

func TestNewApplication_PanicOnInvalidLoggerConfig(t *testing.T) {
	logDir := t.TempDir()
	logPath := filepath.Join(logDir, "app.log")
	t.Setenv("ORCHESTRATOR_HOST", "localhost")
	t.Setenv("ORCHESTRATOR_GRPC_PORT", "0")
	t.Setenv("LOG_LEVEL", "unknown")
	t.Setenv("LOG_PATH", logPath)

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on missing orchestrator config")
		}
	}()

	_ = app.NewApplication()
}
