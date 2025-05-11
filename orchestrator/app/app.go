package app

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/alexGoLyceum/calculator-service/orchestrator/internal/auth"
	"github.com/alexGoLyceum/calculator-service/orchestrator/internal/config"
	"github.com/alexGoLyceum/calculator-service/orchestrator/internal/database/migrate"
	"github.com/alexGoLyceum/calculator-service/orchestrator/internal/database/postgres"
	"github.com/alexGoLyceum/calculator-service/orchestrator/internal/repository"
	"github.com/alexGoLyceum/calculator-service/orchestrator/internal/services"
	grpc "github.com/alexGoLyceum/calculator-service/orchestrator/internal/transport/grpc/server"
	"github.com/alexGoLyceum/calculator-service/orchestrator/internal/transport/http/handlers"
	http "github.com/alexGoLyceum/calculator-service/orchestrator/internal/transport/http/server"
	"github.com/alexGoLyceum/calculator-service/pkg/logging"
)

type Application interface {
	Start()
}

type Impl struct {
	Config     *config.Config
	Logger     logging.Logger
	HTTPServer http.Server
	GRPCServer grpc.Server
}

func NewApplication() *Impl {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}

	logger, err := logging.NewLogger(cfg.Log)
	if err != nil {
		panic(err)
	}

	db := postgres.NewPostgresConnection(cfg.Database)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := db.Connect(ctx); err != nil {
		logger.Error("failed to connect to database", logging.Error(err))
		panic(err)
	}
	db.StartMonitoring(context.Background(), 5*time.Second)

	if err := migrate.RunMigrations(cfg.Database, cfg.MigrationDir); err != nil {
		logger.Error("failed to run migrations", logging.Error(err))
		panic(err)
	}

	repo := repository.NewRepositoryImpl(db)
	JWTManager := auth.NewJWTManager(cfg.JwtSecret, cfg.JwtTTL)

	userService := services.NewUserService(repo, JWTManager)
	expressionTaskService := services.NewExpressionTaskService(repo, cfg.OperationTimesMs)
	expressionTaskService.StartExpiredTaskReset(context.Background(), cfg.ResetInterval, cfg.ExpirationDelay)

	handler := handlers.NewHandler(userService, expressionTaskService)
	httpServer := http.NewServer(cfg, logger, handler, JWTManager)
	grpcServer := grpc.NewServer(expressionTaskService, cfg.Orchestrator.GRPCHost, cfg.Orchestrator.GRPCPort)

	return &Impl{
		Config:     cfg,
		Logger:     logger,
		HTTPServer: httpServer,
		GRPCServer: grpcServer,
	}

	return nil
}

func (app *Impl) Start() {
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	var wg sync.WaitGroup
	wg.Add(2)

	app.Logger.Info("Starting orchestrator")
	go func() {
		defer wg.Done()
		if err := app.GRPCServer.Start(); err != nil {
			app.Logger.Error("failed to start GRPC server", logging.Error(err))
			panic(err)
		}
	}()

	go func() {
		defer wg.Done()
		if err := app.HTTPServer.Start(); err != nil {
			app.Logger.Error("failed to start HTTP server", logging.Error(err))
			panic(err)
		}
	}()

	<-done
	wg.Wait()
}
