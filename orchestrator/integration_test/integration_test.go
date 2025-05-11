package integration_test

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/alexGoLyceum/calculator-service/orchestrator/internal/auth"
	"github.com/alexGoLyceum/calculator-service/orchestrator/internal/repository"
	"github.com/alexGoLyceum/calculator-service/orchestrator/internal/services"
	pb "github.com/alexGoLyceum/calculator-service/orchestrator/internal/transport/grpc/proto"
	"github.com/alexGoLyceum/calculator-service/orchestrator/internal/transport/grpc/server"
	"github.com/alexGoLyceum/calculator-service/orchestrator/internal/transport/http/handlers"
	"github.com/alexGoLyceum/calculator-service/orchestrator/internal/transport/http/routes"
	"github.com/alexGoLyceum/calculator-service/orchestrator/mocks"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

const (
	dbName   = "testdb"
	user     = "user"
	password = "password"
	port     = "5432"
	dsn      = "postgres://%s:%s@localhost:%s/%s?sslmode=disable"
)

var (
	pool        *dockertest.Pool
	resource    *dockertest.Resource
	grpcServer  *grpc.Server
	httpServer  *echo.Echo
	testDB      *pgxpool.Pool
	userService services.UserService
	exprService services.ExpressionTaskService
	jwtManager  auth.JWTManager
)

type IntegrationTestSuite struct {
	suite.Suite
}

func TestIntegrationSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(IntegrationTestSuite))
}

func (s *IntegrationTestSuite) SetupSuite() {
	var err error
	pool, err = dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	resource, err = pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "15-alpine",
		Env: []string{
			"POSTGRES_USER=" + user,
			"POSTGRES_PASSWORD=" + password,
			"POSTGRES_DB=" + dbName,
		},
		ExposedPorts: []string{port},
		PortBindings: map[docker.Port][]docker.PortBinding{
			"5432/tcp": {{HostIP: "0.0.0.0", HostPort: "5432"}},
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	if err := pool.Retry(func() error {
		var err error
		testDB, err = pgxpool.New(context.Background(), fmt.Sprintf(dsn, user, password, resource.GetPort("5432/tcp"), dbName))
		if err != nil {
			return err
		}
		return testDB.Ping(context.Background())
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	if err := applyMigrations(); err != nil {
		log.Fatalf("Failed to apply migrations: %v", err)
	}

	cfg := &services.OperationTimesMS{
		Addition:       500 * time.Millisecond,
		Subtraction:    500 * time.Millisecond,
		Multiplication: 1 * time.Second,
		Division:       1 * time.Second,
	}

	ctrl := gomock.NewController(s.T())
	defer ctrl.Finish()

	mockPool := mocks.NewMockPool(ctrl)
	mockPool.EXPECT().Begin(gomock.Any()).Return(mocks.NewMockTx(ctrl), nil).AnyTimes()
	mockPool.EXPECT().Ping(gomock.Any()).Return(nil).AnyTimes()
	mockPool.EXPECT().Close().AnyTimes()

	jwtSecret := []byte("test-secret-key")
	jwtManager = auth.NewJWTManager(jwtSecret, 24*time.Hour)

	mockConn := mocks.NewMockDatabaseConnection(ctrl)
	mockConn.EXPECT().BeginTx(gomock.Any()).Return(mocks.NewMockTx(ctrl), nil).AnyTimes()
	mockConn.EXPECT().Close().Return(nil).AnyTimes()
	mockConn.EXPECT().Connect(gomock.Any()).Return(nil).AnyTimes()
	mockConn.EXPECT().IsDatabaseUnavailableErr(gomock.Any()).Return(false).AnyTimes()
	mockConn.EXPECT().IsForeignKeyErr(gomock.Any()).Return(false).AnyTimes()
	mockConn.EXPECT().IsNoRowsErr(gomock.Any()).Return(false).AnyTimes()
	mockConn.EXPECT().IsUniqueViolationErr(gomock.Any()).Return(false).AnyTimes()

	repo := repository.NewRepositoryImpl(mockConn)
	userService = services.NewUserService(repo, jwtManager)
	exprService = services.NewExpressionTaskService(repo, &services.OperationTimesMS{
		Addition:       cfg.Addition,
		Subtraction:    cfg.Subtraction,
		Multiplication: cfg.Multiplication,
		Division:       cfg.Division,
	})

	grpcServer = grpc.NewServer()
	grpcListener := bufconn.Listen(1024 * 1024)
	orchestratorServer := server.NewServer(exprService, "localhost", 50051)
	pb.RegisterOrchestratorServiceServer(grpcServer, orchestratorServer)

	go func() {
		if err := grpcServer.Serve(grpcListener); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	httpServer = echo.New()
	httpServer.HideBanner = true
	httpServer.HidePort = true
	httpServer.Use(middleware.Logger())
	httpServer.Use(middleware.Recover())

	handler := handlers.NewHandler(userService, exprService)
	routes.RegisterRoutes(httpServer, handler)

	go func() {
		if err := httpServer.Start(":8080"); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	time.Sleep(2 * time.Second)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	if grpcServer != nil {
		grpcServer.Stop()
	}
	if httpServer != nil {
		httpServer.Close()
	}

	if err := pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}
}

func applyMigrations() error {
	conn, err := sql.Open("postgres", fmt.Sprintf(dsn, user, password, resource.GetPort("5432/tcp"), dbName))
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY,
			login TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL
		);

		CREATE TABLE IF NOT EXISTS expressions (
			id UUID PRIMARY KEY,
			user_id UUID REFERENCES users(id) ON DELETE CASCADE,
			expression TEXT NOT NULL,
			status TEXT NOT NULL,
			result FLOAT
		);

		CREATE TABLE IF NOT EXISTS tasks (
			id UUID PRIMARY KEY,
			expression_id UUID REFERENCES expressions(id) ON DELETE CASCADE,
			arg1_value FLOAT,
			arg1_task_id UUID REFERENCES tasks(id) ON DELETE SET NULL,
			arg2_value FLOAT,
			arg2_task_id UUID REFERENCES tasks(id) ON DELETE SET NULL,
			operator TEXT NOT NULL,
			operation_time TIMESTAMP,
			final_task BOOLEAN NOT NULL,
			status TEXT NOT NULL
		);
	`)
	return err
}
