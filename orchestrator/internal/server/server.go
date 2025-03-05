package server

import (
	"strconv"

	"github.com/alexGoLyceum/calculator-service/orchestrator/internal/config"
	"github.com/alexGoLyceum/calculator-service/orchestrator/internal/middlewares"
	"github.com/alexGoLyceum/calculator-service/orchestrator/internal/routes"
	"github.com/alexGoLyceum/calculator-service/orchestrator/internal/taskmanager"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

type Orchestrator struct {
	Config      *config.Config
	Logger      *zap.Logger
	TaskManager *taskmanager.TaskManager
	Echo        *echo.Echo
}

func NewOrchestrator(cfg *config.Config, logger *zap.Logger) *Orchestrator {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	e.Use(middlewares.RequestLoggerWithZapConfig(logger))
	e.Use(middlewares.SetupCORS())
	e.Use(middleware.Recover())

	taskManager := taskmanager.NewTaskManager(&cfg.TaskManager)

	routes.RegisterRoutes(e, taskManager)

	return &Orchestrator{
		Config:      cfg,
		Logger:      logger,
		TaskManager: taskManager,
		Echo:        e,
	}
}

func (o *Orchestrator) Start() error {
	go o.TaskManager.CheckExpiredTasks()

	address := o.Config.Orchestrator.Host + ":" + strconv.Itoa(o.Config.Orchestrator.Port)
	if err := o.Echo.Start(address); err != nil {
		return err
	}
	return nil
}
