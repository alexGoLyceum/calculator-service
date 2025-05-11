package config

import (
	"errors"

	"github.com/alexGoLyceum/calculator-service/pkg/logging"

	"github.com/spf13/viper"
)

type OrchestratorConfig struct {
	Host string
	Port int
}

type Config struct {
	Orchestrator OrchestratorConfig
	Log          logging.LoggerConfig
}

func LoadConfig() (*Config, error) {
	viper.AutomaticEnv()

	orchestrator := OrchestratorConfig{
		Host: viper.GetString("ORCHESTRATOR_HOST"),
		Port: viper.GetInt("ORCHESTRATOR_GRPC_PORT"),
	}

	if orchestrator.Port <= 0 {
		return nil, errors.New("invalid orchestrator configuration: check ORCHESTRATOR_HOST and ORCHESTRATOR_PORT")
	}

	logger := logging.LoggerConfig{
		Level:             viper.GetString("LOG_LEVEL"),
		FilePath:          viper.GetString("LOG_PATH"),
		EnableFileLogging: viper.GetBool("LOG_ENABLE_FILE_LOGGING"),
	}

	return &Config{Orchestrator: orchestrator, Log: logger}, nil
}
