package http

import (
	"strconv"

	"github.com/alexGoLyceum/calculator-service/internal/config"
	"github.com/alexGoLyceum/calculator-service/internal/transport/http/middlewares"
	"github.com/alexGoLyceum/calculator-service/internal/transport/http/routes"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"go.uber.org/zap"
)

type Server struct {
	Config *config.ServerConfig
	Logger *zap.Logger
	Echo   *echo.Echo
}

func NewServer(config *config.ServerConfig, logger *zap.Logger) *Server {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	e.Use(middlewares.RequestLoggerWithZapConfig(logger))
	e.Use(middleware.Recover())

	routes.RegisterRoutes(e)

	return &Server{
		Config: config,
		Echo:   e,
		Logger: logger,
	}
}

func (s *Server) Start() error {
	address := s.Config.Host + ":" + strconv.Itoa(s.Config.Port)
	if err := s.Echo.Start(address); err != nil {
		return err
	}
	return nil
}
