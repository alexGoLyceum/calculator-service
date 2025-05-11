package app

import (
	"github.com/alexGoLyceum/calculator-service/agent/internal/agent"
	"github.com/alexGoLyceum/calculator-service/agent/internal/config"
	"github.com/alexGoLyceum/calculator-service/pkg/logging"
)

type Application interface {
	Start()
}

type Impl struct {
	Config *config.Config
	Logger logging.Logger
	Agent  agent.Agent
}

func NewApplication() *Impl {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}

	logger, err := logging.NewLogger(&cfg.Log)
	if err != nil {
		panic(err)
	}

	a, err := agent.NewAgent(cfg, logger)
	if err != nil {
		logger.Error("Failed to create agent", logging.Error(err))
		panic(err)
	}

	return &Impl{
		Config: cfg,
		Logger: logger,
		Agent:  a,
	}
}

func (app *Impl) Start() {
	app.Logger.Info("Starting agent")
	if err := app.Agent.Start(); err != nil {
		app.Logger.Error("Failed to start agent", logging.Error(err))
		panic(err)
	}
}
