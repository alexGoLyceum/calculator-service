package server

import (
	"strconv"

	"github.com/alexGoLyceum/calculator-service/orchestrator/internal/auth"
	"github.com/alexGoLyceum/calculator-service/orchestrator/internal/config"
	"github.com/alexGoLyceum/calculator-service/orchestrator/internal/transport/http/handlers"
	"github.com/alexGoLyceum/calculator-service/orchestrator/internal/transport/http/middlewares"
	"github.com/alexGoLyceum/calculator-service/orchestrator/internal/transport/http/routes"
	"github.com/alexGoLyceum/calculator-service/pkg/logging"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Server interface {
	Start() error
}

type Impl struct {
	config *config.Config
	logger logging.Logger
	Echo   *echo.Echo
}

func NewServer(cfg *config.Config, logger logging.Logger, handler handlers.Handler, JWTManager auth.JWTManager) Server {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	e.Use(middlewares.RequestLoggerConfig(logger))
	e.Use(middlewares.JWTMiddleware(JWTManager))
	e.Use(middlewares.SetupCORS())
	e.Use(middleware.Recover())

	routes.RegisterRoutes(e, handler)

	return &Impl{
		config: cfg,
		logger: logger,
		Echo:   e,
	}
}

func (s *Impl) Start() error {
	address := s.config.Orchestrator.HTTPHost + ":" + strconv.Itoa(s.config.Orchestrator.HTTPPort)
	return s.Echo.Start(address)
}
