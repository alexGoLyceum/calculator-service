package config

import (
	"errors"
	"fmt"
	"path"

	"github.com/alexGoLyceum/calculator-service/pkg/logging"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type OrchestratorConfig struct {
	Host string
	Port int
}

type TaskManagerConfig struct {
	TimeAdditionMS        int
	TimeSubtractionMS     int
	TimeMultiplicationsMS int
	TimeDivisionsMS       int
}

type Config struct {
	Orchestrator OrchestratorConfig
	TaskManager  TaskManagerConfig
	Log          logging.LogConfig
}

func LoadConfig(cfgPath string) (*Config, error) {
	if cfgPath != "" {
		if err := godotenv.Load(path.Clean(cfgPath)); err != nil {
			return nil, fmt.Errorf("failed to load .env file: %w", err)
		}
	}

	viper.AutomaticEnv()

	viper.SetDefault("ORCHESTRATOR_HOST", "localhost")
	viper.SetDefault("ORCHESTRATOR_PORT", 8080)

	viper.SetDefault("TIME_ADDITION_MS", 1000)
	viper.SetDefault("TIME_SUBTRACTION_MS", 1000)
	viper.SetDefault("TIME_MULTIPLICATIONS_MS", 1000)
	viper.SetDefault("TIME_DIVISIONS_MS", 1000)

	viper.SetDefault("LOG_LEVEL", "debug")
	viper.SetDefault("LOG_PATH", "logs.log")

	orchestrator := OrchestratorConfig{
		Host: viper.GetString("ORCHESTRATOR_HOST"),
		Port: viper.GetInt("ORCHESTRATOR_PORT"),
	}

	if orchestrator.Port <= 0 {
		return nil, errors.New("ORCHESTRATOR_PORT must be greater than 0")
	}

	taskManager := TaskManagerConfig{
		TimeAdditionMS:        viper.GetInt("TIME_ADDITION_MS"),
		TimeSubtractionMS:     viper.GetInt("TIME_SUBTRACTION_MS"),
		TimeMultiplicationsMS: viper.GetInt("TIME_MULTIPLICATIONS_MS"),
		TimeDivisionsMS:       viper.GetInt("TIME_DIVISIONS_MS"),
	}

	if taskManager.TimeAdditionMS <= 0 || taskManager.TimeSubtractionMS <= 0 ||
		taskManager.TimeMultiplicationsMS <= 0 || taskManager.TimeDivisionsMS <= 0 {
		return nil, errors.New("invalid task manager configuration: all time values must be positive")
	}

	logger := logging.LogConfig{
		Level: viper.GetString("LOG_LEVEL"),
		Path:  viper.GetString("LOG_PATH"),
	}

	return &Config{Orchestrator: orchestrator, TaskManager: taskManager, Log: logger}, nil
}
