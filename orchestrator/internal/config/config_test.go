package config_test

import (
	"os"
	"testing"

	"github.com/alexGoLyceum/calculator-service/orchestrator/internal/config"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfigFromFile(t *testing.T) {
	err := os.WriteFile("./env_test.env", []byte(`
ORCHESTRATOR_HOST=localhost
ORCHESTRATOR_PORT=8080
TIME_ADDITION_MS=10
TIME_SUBTRACTION_MS=20
TIME_MULTIPLICATIONS_MS=30
TIME_DIVISIONS_MS=40
LOG_LEVEL=info
LOG_PATH=/var/log
`), 0644)
	require.NoError(t, err)
	defer os.Remove("./env_test.env")

	err = godotenv.Load("./env_test.env")
	require.NoError(t, err)

	cfg, err := config.LoadConfig("./env_test.env")
	require.NoError(t, err)

	assert.Equal(t, "localhost", cfg.Orchestrator.Host)
	assert.Equal(t, 8080, cfg.Orchestrator.Port)
	assert.Equal(t, 10, cfg.TaskManager.TimeAdditionMS)
	assert.Equal(t, 20, cfg.TaskManager.TimeSubtractionMS)
	assert.Equal(t, 30, cfg.TaskManager.TimeMultiplicationsMS)
	assert.Equal(t, 40, cfg.TaskManager.TimeDivisionsMS)
	assert.Equal(t, "info", cfg.Log.Level)
	assert.Equal(t, "/var/log", cfg.Log.Path)
}

func TestLoadConfigFromEnv(t *testing.T) {
	os.Setenv("ORCHESTRATOR_HOST", "127.0.0.1")
	os.Setenv("ORCHESTRATOR_PORT", "9090")
	os.Setenv("TIME_ADDITION_MS", "15")
	os.Setenv("TIME_SUBTRACTION_MS", "25")
	os.Setenv("TIME_MULTIPLICATIONS_MS", "35")
	os.Setenv("TIME_DIVISIONS_MS", "45")
	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("LOG_PATH", "/tmp/logs")
	defer func() {
		os.Unsetenv("ORCHESTRATOR_HOST")
		os.Unsetenv("ORCHESTRATOR_PORT")
		os.Unsetenv("TIME_ADDITION_MS")
		os.Unsetenv("TIME_SUBTRACTION_MS")
		os.Unsetenv("TIME_MULTIPLICATIONS_MS")
		os.Unsetenv("TIME_DIVISIONS_MS")
		os.Unsetenv("LOG_LEVEL")
		os.Unsetenv("LOG_PATH")
	}()

	cfg, err := config.LoadConfig("")
	require.NoError(t, err)

	assert.Equal(t, "127.0.0.1", cfg.Orchestrator.Host)
	assert.Equal(t, 9090, cfg.Orchestrator.Port)
	assert.Equal(t, 15, cfg.TaskManager.TimeAdditionMS)
	assert.Equal(t, 25, cfg.TaskManager.TimeSubtractionMS)
	assert.Equal(t, 35, cfg.TaskManager.TimeMultiplicationsMS)
	assert.Equal(t, 45, cfg.TaskManager.TimeDivisionsMS)
	assert.Equal(t, "debug", cfg.Log.Level)
	assert.Equal(t, "/tmp/logs", cfg.Log.Path)
}

func TestLoadConfigFileNotFound(t *testing.T) {
	cfg, err := config.LoadConfig("./nonexistent.env")
	require.Error(t, err)
	assert.Nil(t, cfg)
	assert.Equal(t, "failed to load .env file: open nonexistent.env: no such file or directory", err.Error())
}

func TestInvalidOrchestratorPort(t *testing.T) {
	os.Setenv("ORCHESTRATOR_HOST", "localhost")
	os.Setenv("ORCHESTRATOR_PORT", "-1")
	os.Setenv("TIME_ADDITION_MS", "10")
	os.Setenv("TIME_SUBTRACTION_MS", "20")
	os.Setenv("TIME_MULTIPLICATIONS_MS", "30")
	os.Setenv("TIME_DIVISIONS_MS", "40")
	os.Setenv("LOG_LEVEL", "info")
	os.Setenv("LOG_PATH", "/var/log")
	defer func() {
		os.Unsetenv("ORCHESTRATOR_HOST")
		os.Unsetenv("ORCHESTRATOR_PORT")
		os.Unsetenv("TIME_ADDITION_MS")
		os.Unsetenv("TIME_SUBTRACTION_MS")
		os.Unsetenv("TIME_MULTIPLICATIONS_MS")
		os.Unsetenv("TIME_DIVISIONS_MS")
		os.Unsetenv("LOG_LEVEL")
		os.Unsetenv("LOG_PATH")
	}()

	cfg, err := config.LoadConfig("")
	require.Error(t, err)
	assert.Nil(t, cfg)
	assert.Equal(t, "ORCHESTRATOR_PORT must be greater than 0", err.Error())
}

func TestInvalidTaskManagerTimes(t *testing.T) {
	os.Setenv("ORCHESTRATOR_HOST", "localhost")
	os.Setenv("ORCHESTRATOR_PORT", "8080")
	os.Setenv("TIME_ADDITION_MS", "-10")
	os.Setenv("TIME_SUBTRACTION_MS", "20")
	os.Setenv("TIME_MULTIPLICATIONS_MS", "30")
	os.Setenv("TIME_DIVISIONS_MS", "40")
	os.Setenv("LOG_LEVEL", "info")
	os.Setenv("LOG_PATH", "/var/log")
	defer func() {
		os.Unsetenv("ORCHESTRATOR_HOST")
		os.Unsetenv("ORCHESTRATOR_PORT")
		os.Unsetenv("TIME_ADDITION_MS")
		os.Unsetenv("TIME_SUBTRACTION_MS")
		os.Unsetenv("TIME_MULTIPLICATIONS_MS")
		os.Unsetenv("TIME_DIVISIONS_MS")
		os.Unsetenv("LOG_LEVEL")
		os.Unsetenv("LOG_PATH")
	}()

	cfg, err := config.LoadConfig("")
	require.Error(t, err)
	assert.Nil(t, cfg)
	assert.Equal(t, "invalid task manager configuration: all time values must be positive", err.Error())
}

func TestInvalidTaskManagerAllTimes(t *testing.T) {
	os.Setenv("ORCHESTRATOR_HOST", "localhost")
	os.Setenv("ORCHESTRATOR_PORT", "8080")
	os.Setenv("TIME_ADDITION_MS", "0")
	os.Setenv("TIME_SUBTRACTION_MS", "0")
	os.Setenv("TIME_MULTIPLICATIONS_MS", "0")
	os.Setenv("TIME_DIVISIONS_MS", "0")
	os.Setenv("LOG_LEVEL", "info")
	os.Setenv("LOG_PATH", "/var/log")
	defer func() {
		os.Unsetenv("ORCHESTRATOR_HOST")
		os.Unsetenv("ORCHESTRATOR_PORT")
		os.Unsetenv("TIME_ADDITION_MS")
		os.Unsetenv("TIME_SUBTRACTION_MS")
		os.Unsetenv("TIME_MULTIPLICATIONS_MS")
		os.Unsetenv("TIME_DIVISIONS_MS")
		os.Unsetenv("LOG_LEVEL")
		os.Unsetenv("LOG_PATH")
	}()

	cfg, err := config.LoadConfig("")
	require.Error(t, err)
	assert.Nil(t, cfg)
	assert.Equal(t, "invalid task manager configuration: all time values must be positive", err.Error())
}
