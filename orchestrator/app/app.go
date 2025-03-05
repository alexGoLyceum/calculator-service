package app

import (
	"log"
	"path"

	"github.com/alexGoLyceum/calculator-service/orchestrator/internal/config"
	"github.com/alexGoLyceum/calculator-service/orchestrator/internal/server"
	"github.com/alexGoLyceum/calculator-service/pkg/logging"

	"go.uber.org/zap"
)

type Application struct {
	Config       *config.Config
	Logger       *zap.Logger
	Orchestrator *server.Orchestrator
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

	o := server.NewOrchestrator(cfg, logger)

	return &Application{
		Config:       cfg,
		Logger:       logger,
		Orchestrator: o,
	}
}

func (app *Application) Start() {
	app.Logger.Info("starting orchestrator")
	if err := app.Orchestrator.Start(); err != nil {
		app.Logger.Error("failed to start orchestrator", zap.Error(err))
		panic(err)
	}
}
