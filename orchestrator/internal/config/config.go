package config

import (
	"errors"
	"fmt"
	"time"

	"github.com/alexGoLyceum/calculator-service/orchestrator/internal/database/postgres"
	"github.com/alexGoLyceum/calculator-service/orchestrator/internal/services"
	"github.com/alexGoLyceum/calculator-service/pkg/logging"

	"github.com/spf13/viper"
)

type Config struct {
	Orchestrator     OrchestratorConfig
	Database         *postgres.Config
	Log              *logging.LoggerConfig
	OperationTimesMs *services.OperationTimesMS
	MigrationDir     string
	JwtSecret        []byte
	JwtTTL           time.Duration
	ResetInterval    time.Duration
	ExpirationDelay  time.Duration
}

type OrchestratorConfig struct {
	HTTPHost string
	HTTPPort int
	GRPCHost string
	GRPCPort int
}

func LoadConfig() (*Config, error) {
	viper.AutomaticEnv()

	orchestrator := OrchestratorConfig{
		HTTPHost: viper.GetString("ORCHESTRATOR_HTTP_HOST"),
		HTTPPort: viper.GetInt("ORCHESTRATOR_HTTP_PORT"),
		GRPCHost: viper.GetString("ORCHESTRATOR_GRPC_HOST"),
		GRPCPort: viper.GetInt("ORCHESTRATOR_GRPC_PORT"),
	}

	if err := validateOrchestrator(orchestrator); err != nil {
		return nil, err
	}

	operationTimesMS := services.OperationTimesMS{
		Addition:       viper.GetDuration("TIME_ADDITION_MS"),
		Subtraction:    viper.GetDuration("TIME_SUBTRACTION_MS"),
		Multiplication: viper.GetDuration("TIME_MULTIPLICATIONS_MS"),
		Division:       viper.GetDuration("TIME_DIVISIONS_MS"),
	}

	if err := validateOperationTimes(operationTimesMS); err != nil {
		return nil, err
	}

	logger := &logging.LoggerConfig{
		Level:             viper.GetString("LOG_LEVEL"),
		FilePath:          viper.GetString("LOG_PATH"),
		EnableFileLogging: viper.GetBool("LOG_ENABLE_FILE_LOGGING"),
	}

	if logger.Level == "" {
		return nil, errors.New("LOG_LEVEL must be set")
	}

	database := &postgres.Config{
		Host:     viper.GetString("POSTGRES_HOST"),
		Port:     viper.GetString("POSTGRES_PORT"),
		Username: viper.GetString("POSTGRES_USER"),
		Password: viper.GetString("POSTGRES_PASSWORD"),
		Database: viper.GetString("POSTGRES_DB"),
	}

	if err := validateDatabase(*database); err != nil {
		return nil, err
	}

	migrationDir := viper.GetString("MIGRATION_DIR")
	if migrationDir == "" {
		return nil, errors.New("MIGRATION_DIR must be set")
	}

	jwtSecretStr := viper.GetString("JWT_SECRET")
	if jwtSecretStr == "" {
		return nil, errors.New("JWT_SECRET must be set")
	}
	jwtSecret := []byte(jwtSecretStr)

	jwtTTL := viper.GetDuration("JWT_TTL")
	if jwtTTL <= 0 {
		return nil, errors.New("JWT_TTL must be greater than 0")
	}

	resetInterval := viper.GetDuration("RESET_INTERVAL")
	if resetInterval <= 0 {
		return nil, errors.New("RESET_INTERVAL must be greater than 0")
	}

	expirationDelay := viper.GetDuration("EXPIRATION_DELAY")
	if expirationDelay <= 0 {
		return nil, errors.New("EXPIRATION_DELAY must be greater than 0")
	}

	return &Config{
		Orchestrator:     orchestrator,
		Database:         database,
		Log:              logger,
		OperationTimesMs: &operationTimesMS,
		MigrationDir:     migrationDir,
		JwtSecret:        jwtSecret,
		JwtTTL:           jwtTTL,
		ResetInterval:    resetInterval,
		ExpirationDelay:  expirationDelay,
	}, nil
}

func validateOrchestrator(cfg OrchestratorConfig) error {
	if cfg.HTTPHost == "" || cfg.GRPCHost == "" {
		return errors.New("ORCHESTRATOR_HTTP_HOST and ORCHESTRATOR_GRPC_HOST must be set")
	}
	if cfg.HTTPPort < 1 || cfg.HTTPPort > 65535 {
		return fmt.Errorf("ORCHESTRATOR_HTTP_PORT must be in range [1, 65535], got %d", cfg.HTTPPort)
	}
	if cfg.GRPCPort < 1 || cfg.GRPCPort > 65535 {
		return fmt.Errorf("ORCHESTRATOR_GRPC_PORT must be in range [1, 65535], got %d", cfg.GRPCPort)
	}
	return nil
}

func validateOperationTimes(times services.OperationTimesMS) error {
	if times.Addition <= 0 {
		return errors.New("TIME_ADDITION_MS must be greater than 0")
	}
	if times.Subtraction <= 0 {
		return errors.New("TIME_SUBTRACTION_MS must be greater than 0")
	}
	if times.Multiplication <= 0 {
		return errors.New("TIME_MULTIPLICATIONS_MS must be greater than 0")
	}
	if times.Division <= 0 {
		return errors.New("TIME_DIVISIONS_MS must be greater than 0")
	}
	return nil
}

func validateDatabase(cfg postgres.Config) error {
	if cfg.Host == "" || cfg.Port == "" || cfg.Username == "" || cfg.Password == "" || cfg.Database == "" {
		return errors.New("all POSTGRES_* fields must be set and non-empty")
	}
	return nil
}
