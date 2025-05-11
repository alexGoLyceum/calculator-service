package config_test

import (
	"os"
	"testing"

	"github.com/alexGoLyceum/calculator-service/orchestrator/internal/config"

	"github.com/stretchr/testify/require"
)

func setEnv(t *testing.T, key, value string) {
	t.Helper()
	err := os.Setenv(key, value)
	if err != nil {
		t.Fatalf("Failed to set env var %s: %v", key, err)
	}
	t.Cleanup(func() {
		_ = os.Unsetenv(key)
	})
}

func setValidEnv(t *testing.T) {
	t.Helper()
	setEnv(t, "ORCHESTRATOR_HTTP_HOST", "localhost")
	setEnv(t, "ORCHESTRATOR_HTTP_PORT", "8081")
	setEnv(t, "ORCHESTRATOR_GRPC_HOST", "localhost")
	setEnv(t, "ORCHESTRATOR_GRPC_PORT", "50052")

	setEnv(t, "TIME_ADDITION_MS", "10ms")
	setEnv(t, "TIME_SUBTRACTION_MS", "10ms")
	setEnv(t, "TIME_MULTIPLICATIONS_MS", "10ms")
	setEnv(t, "TIME_DIVISIONS_MS", "10ms")

	setEnv(t, "LOG_LEVEL", "info")
	setEnv(t, "LOG_PATH", "/tmp/log")

	setEnv(t, "POSTGRES_HOST", "localhost")
	setEnv(t, "POSTGRES_PORT", "5432")
	setEnv(t, "POSTGRES_USER", "user")
	setEnv(t, "POSTGRES_PASSWORD", "pass")
	setEnv(t, "POSTGRES_DB", "db")

	setEnv(t, "MIGRATION_DIR", "/migrations")
	setEnv(t, "JWT_SECRET", "supersecret")
	setEnv(t, "JWT_TTL", "15m")
	setEnv(t, "RESET_INTERVAL", "5m")
	setEnv(t, "EXPIRATION_DELAY", "10m")
}

func TestLoadConfig_Success(t *testing.T) {
	setValidEnv(t)

	cfg, err := config.LoadConfig()
	require.NoError(t, err)
	require.NotNil(t, cfg)
	require.Equal(t, "localhost", cfg.Orchestrator.HTTPHost)
	require.Equal(t, 8081, cfg.Orchestrator.HTTPPort)
}

func TestLoadConfig_MissingOrchestratorHost(t *testing.T) {
	setValidEnv(t)
	_ = os.Unsetenv("ORCHESTRATOR_HTTP_HOST")

	_, err := config.LoadConfig()
	require.Error(t, err)
	require.ErrorContains(t, err, "ORCHESTRATOR_HTTP_HOST")
}

func TestLoadConfig_InvalidHTTPPort(t *testing.T) {
	setValidEnv(t)
	setEnv(t, "ORCHESTRATOR_HTTP_PORT", "70000")

	_, err := config.LoadConfig()
	require.Error(t, err)
	require.ErrorContains(t, err, "ORCHESTRATOR_HTTP_PORT must be in range")
}

func TestLoadConfig_InvalidGRPCPort(t *testing.T) {
	setValidEnv(t)
	setEnv(t, "ORCHESTRATOR_GRPC_PORT", "0")

	_, err := config.LoadConfig()
	require.Error(t, err)
	require.ErrorContains(t, err, "ORCHESTRATOR_GRPC_PORT must be in range")
}

func TestLoadConfig_ZeroOperationTime(t *testing.T) {
	setValidEnv(t)
	setEnv(t, "TIME_ADDITION_MS", "0s")

	_, err := config.LoadConfig()
	require.Error(t, err)
	require.ErrorContains(t, err, "TIME_ADDITION_MS must be greater than 0")
}

func TestLoadConfig_MissingLogLevel(t *testing.T) {
	setValidEnv(t)
	_ = os.Unsetenv("LOG_LEVEL")

	_, err := config.LoadConfig()
	require.Error(t, err)
	require.ErrorContains(t, err, "LOG_LEVEL must be set")
}

func TestLoadConfig_MissingPostgresConfig(t *testing.T) {
	setValidEnv(t)
	_ = os.Unsetenv("POSTGRES_HOST")

	_, err := config.LoadConfig()
	require.Error(t, err)
	require.ErrorContains(t, err, "all POSTGRES_* fields must be set")
}

func TestLoadConfig_MissingMigrationDir(t *testing.T) {
	setValidEnv(t)
	_ = os.Unsetenv("MIGRATION_DIR")

	_, err := config.LoadConfig()
	require.Error(t, err)
	require.ErrorContains(t, err, "MIGRATION_DIR must be set")
}

func TestLoadConfig_MissingJwtSecret(t *testing.T) {
	setValidEnv(t)
	_ = os.Unsetenv("JWT_SECRET")

	_, err := config.LoadConfig()
	require.Error(t, err)
	require.ErrorContains(t, err, "JWT_SECRET must be set")
}

func TestLoadConfig_InvalidJwtTTL(t *testing.T) {
	setValidEnv(t)
	setEnv(t, "JWT_TTL", "0s")

	_, err := config.LoadConfig()
	require.Error(t, err)
	require.ErrorContains(t, err, "JWT_TTL must be greater than 0")
}

func TestLoadConfig_InvalidResetInterval(t *testing.T) {
	setValidEnv(t)
	setEnv(t, "RESET_INTERVAL", "0s")

	_, err := config.LoadConfig()
	require.Error(t, err)
	require.ErrorContains(t, err, "RESET_INTERVAL must be greater than 0")
}

func TestLoadConfig_InvalidExpirationDelay(t *testing.T) {
	setValidEnv(t)
	setEnv(t, "EXPIRATION_DELAY", "0s")

	_, err := config.LoadConfig()
	require.Error(t, err)
	require.ErrorContains(t, err, "EXPIRATION_DELAY must be greater than 0")
}

func TestLoadConfig_ZeroSubtractionTime(t *testing.T) {
	setValidEnv(t)
	setEnv(t, "TIME_SUBTRACTION_MS", "0s")

	_, err := config.LoadConfig()
	require.Error(t, err)
	require.ErrorContains(t, err, "TIME_SUBTRACTION_MS must be greater than 0")
}

func TestLoadConfig_ZeroMultiplicationTime(t *testing.T) {
	setValidEnv(t)
	setEnv(t, "TIME_MULTIPLICATIONS_MS", "0s")

	_, err := config.LoadConfig()
	require.Error(t, err)
	require.ErrorContains(t, err, "TIME_MULTIPLICATIONS_MS must be greater than 0")
}

func TestLoadConfig_ZeroDivisionTime(t *testing.T) {
	setValidEnv(t)
	setEnv(t, "TIME_DIVISIONS_MS", "0s")

	_, err := config.LoadConfig()
	require.Error(t, err)
	require.ErrorContains(t, err, "TIME_DIVISIONS_MS must be greater than 0")
}
