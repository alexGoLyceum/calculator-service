package config_test

import (
	"os"
	"testing"

	"github.com/alexGoLyceum/calculator-service/agent/internal/config"

	"github.com/stretchr/testify/require"
)

func setEnv(t *testing.T, key, value string) {
	t.Helper()
	err := os.Setenv(key, value)
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = os.Unsetenv(key)
	})
}

func setValidEnv(t *testing.T) {
	t.Helper()
	setEnv(t, "ORCHESTRATOR_HOST", "localhost")
	setEnv(t, "ORCHESTRATOR_GRPC_PORT", "9090")

	setEnv(t, "LOG_LEVEL", "info")
	setEnv(t, "LOG_PATH", "/tmp/log")
}

func TestLoadConfig_Success(t *testing.T) {
	setValidEnv(t)

	cfg, err := config.LoadConfig()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	require.Equal(t, "localhost", cfg.Orchestrator.Host)
	require.Equal(t, 9090, cfg.Orchestrator.Port)

	require.Equal(t, "info", cfg.Log.Level)
	require.Equal(t, "/tmp/log", cfg.Log.FilePath)
}

func TestLoadConfig_InvalidPort(t *testing.T) {
	setValidEnv(t)
	setEnv(t, "ORCHESTRATOR_GRPC_PORT", "0")

	_, err := config.LoadConfig()
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid orchestrator configuration")
}
