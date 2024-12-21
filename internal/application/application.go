package application

import (
	"github.com/alexGoLyceum/calculator-service/internal/config"
	"github.com/alexGoLyceum/calculator-service/internal/logging"
	"github.com/alexGoLyceum/calculator-service/internal/transport/http"

	"go.uber.org/zap"
)

type Application struct {
	server *http.Server
	logger *zap.Logger
}

func New() *Application {
	cfg, err := config.LoadConfig("configs/config.yaml")
	if err != nil {
		panic(err)
	}

	logger, err := logging.NewLogger(&cfg.Log)
	if err != nil {
		panic(err)
	}

	server := http.NewServer(&cfg.Server, logger)

	return &Application{
		server: server,
		logger: logger,
	}
}

func (app *Application) RunServer() {
	app.server.Logger.Info("Starting server")
	if err := app.server.Start(); err != nil {
		app.logger.Fatal("failed to start server", zap.Error(err))
	}
}
