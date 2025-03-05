package app

import (
	"log"
	"path"

	"go.uber.org/zap"

	"github.com/alexGoLyceum/calculator-service/agent/internal/agentsmanager"
	"github.com/alexGoLyceum/calculator-service/agent/internal/client"
	"github.com/alexGoLyceum/calculator-service/agent/internal/config"
	"github.com/alexGoLyceum/calculator-service/pkg/logging"
)

type Application struct {
	Config        *config.Config
	Logger        *zap.Logger
	AgentsManager *agentsmanager.AgentsManager
}

func NewApplication() *Application {
	cfg, err := config.LoadConfig(path.Clean(".env"))
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	logger, err := logging.NewLogger(&cfg.Log)
	if err != nil {
		log.Fatalf("failed to create logger: %v", err)
	}

	client.SetUpUrl(cfg.Orchestrator)

	a := agentsmanager.NewAgentsManager(&cfg.Agent, logger)

	return &Application{
		Config:        cfg,
		Logger:        logger,
		AgentsManager: a,
	}
}

func (app *Application) Start() {
	app.Logger.Info("Starting agents")
	app.AgentsManager.Start()
}
