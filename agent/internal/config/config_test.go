package config_test

import (
	"os"
	"testing"

	"github.com/alexGoLyceum/calculator-service/agent/internal/config"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfigFromFile(t *testing.T) {
	err := os.WriteFile("./env_test.env", []byte(`
COMPUTING_POWER=100
LOG_LEVEL=info
LOG_PATH=/var/log
ORCHESTRATOR_HOST=127.0.0.1
ORCHESTRATOR_PORT=9090
`), 0644)
	require.NoError(t, err)
	defer os.Remove("./env_test.env")

	err = godotenv.Load("./env_test.env")
	require.NoError(t, err)

	cfg, err := config.LoadConfig("./env_test.env")
	require.NoError(t, err)

	assert.Equal(t, 100, cfg.Agent.ComputingPower)
	assert.Equal(t, "info", cfg.Log.Level)
	assert.Equal(t, "/var/log", cfg.Log.Path)
	assert.Equal(t, "127.0.0.1", cfg.Orchestrator.Host)
	assert.Equal(t, 9090, cfg.Orchestrator.Port)
}

func TestLoadConfigFromEnv(t *testing.T) {
	os.Setenv("COMPUTING_POWER", "200")
	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("LOG_PATH", "/tmp/logs")
	os.Setenv("ORCHESTRATOR_HOST", "192.168.1.1")
	os.Setenv("ORCHESTRATOR_PORT", "8081")
	defer func() {
		os.Unsetenv("COMPUTING_POWER")
		os.Unsetenv("LOG_LEVEL")
		os.Unsetenv("LOG_PATH")
		os.Unsetenv("ORCHESTRATOR_HOST")
		os.Unsetenv("ORCHESTRATOR_PORT")
	}()

	cfg, err := config.LoadConfig("")
	require.NoError(t, err)

	assert.Equal(t, 200, cfg.Agent.ComputingPower)
	assert.Equal(t, "debug", cfg.Log.Level)
	assert.Equal(t, "/tmp/logs", cfg.Log.Path)
	assert.Equal(t, "192.168.1.1", cfg.Orchestrator.Host)
	assert.Equal(t, 8081, cfg.Orchestrator.Port)
}

func TestLoadConfigFileNotFound(t *testing.T) {
	cfg, err := config.LoadConfig("./nonexistent.env")
	require.Error(t, err)
	assert.Nil(t, cfg)
	assert.Equal(t, "failed to load .env file: open nonexistent.env: no such file or directory", err.Error())
}

func TestInvalidComputingPower(t *testing.T) {
	os.Setenv("COMPUTING_POWER", "-5")
	os.Setenv("ORCHESTRATOR_HOST", "localhost")
	os.Setenv("ORCHESTRATOR_PORT", "8080")
	os.Setenv("LOG_LEVEL", "info")
	os.Setenv("LOG_PATH", "/var/log")
	defer func() {
		os.Unsetenv("COMPUTING_POWER")
		os.Unsetenv("ORCHESTRATOR_HOST")
		os.Unsetenv("ORCHESTRATOR_PORT")
		os.Unsetenv("LOG_LEVEL")
		os.Unsetenv("LOG_PATH")
	}()

	cfg, err := config.LoadConfig("")
	require.Error(t, err)
	assert.Nil(t, cfg)
	assert.Equal(t, "COMPUTING_POWER must be greater than 0", err.Error())
}

func TestInvalidOrchestratorPort(t *testing.T) {
	os.Setenv("COMPUTING_POWER", "10")
	os.Setenv("ORCHESTRATOR_HOST", "localhost")
	os.Setenv("ORCHESTRATOR_PORT", "-1")
	os.Setenv("LOG_LEVEL", "info")
	os.Setenv("LOG_PATH", "/var/log")
	defer func() {
		os.Unsetenv("COMPUTING_POWER")
		os.Unsetenv("ORCHESTRATOR_HOST")
		os.Unsetenv("ORCHESTRATOR_PORT")
		os.Unsetenv("LOG_LEVEL")
		os.Unsetenv("LOG_PATH")
	}()

	cfg, err := config.LoadConfig("")
	require.Error(t, err)
	assert.Nil(t, cfg)
	assert.Equal(t, "invalid orchestrator configuration: check ORCHESTRATOR_HOST and ORCHESTRATOR_PORT", err.Error())
}

func TestInvalidOrchestratorPortAndComputingPower(t *testing.T) {
	os.Setenv("COMPUTING_POWER", "-5")
	os.Setenv("ORCHESTRATOR_HOST", "localhost")
	os.Setenv("ORCHESTRATOR_PORT", "-1")
	os.Setenv("LOG_LEVEL", "info")
	os.Setenv("LOG_PATH", "/var/log")
	defer func() {
		os.Unsetenv("COMPUTING_POWER")
		os.Unsetenv("ORCHESTRATOR_HOST")
		os.Unsetenv("ORCHESTRATOR_PORT")
		os.Unsetenv("LOG_LEVEL")
		os.Unsetenv("LOG_PATH")
	}()

	cfg, err := config.LoadConfig("")
	require.Error(t, err)
	assert.Nil(t, cfg)
	assert.Equal(t, "COMPUTING_POWER must be greater than 0", err.Error())
}
