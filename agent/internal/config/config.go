package config

import (
	"errors"
	"fmt"
	"path"

	"github.com/alexGoLyceum/calculator-service/pkg/logging"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type AgentConfig struct {
	ComputingPower int
}

type OrchestratorConfig struct {
	Host string
	Port int
}

type Config struct {
	Agent        AgentConfig
	Orchestrator OrchestratorConfig
	Log          logging.LogConfig
}

func LoadConfig(cfgPath string) (*Config, error) {
	if cfgPath != "" {
		if err := godotenv.Load(path.Clean(cfgPath)); err != nil {
			return nil, fmt.Errorf("failed to load .env file: %w", err)
		}
	}

	viper.AutomaticEnv()

	viper.SetDefault("COMPUTING_POWER", 5)

	viper.SetDefault("ORCHESTRATOR_HOST", "localhost")
	viper.SetDefault("ORCHESTRATOR_PORT", 8080)

	viper.SetDefault("LOG_LEVEL", "debug")
	viper.SetDefault("LOG_PATH", "logs.log")

	agent := AgentConfig{
		ComputingPower: viper.GetInt("COMPUTING_POWER"),
	}

	if agent.ComputingPower <= 0 {
		return nil, errors.New("COMPUTING_POWER must be greater than 0")
	}

	orchestrator := OrchestratorConfig{
		Host: viper.GetString("ORCHESTRATOR_HOST"),
		Port: viper.GetInt("ORCHESTRATOR_PORT"),
	}

	if orchestrator.Port <= 0 {
		return nil, errors.New("invalid orchestrator configuration: check ORCHESTRATOR_HOST and ORCHESTRATOR_PORT")
	}

	logger := logging.LogConfig{
		Level: viper.GetString("LOG_LEVEL"),
		Path:  viper.GetString("LOG_PATH"),
	}

	return &Config{Agent: agent, Orchestrator: orchestrator, Log: logger}, nil
}
