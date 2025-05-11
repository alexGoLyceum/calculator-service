package agent

import (
	"context"
	"fmt"

	"github.com/alexGoLyceum/calculator-service/agent/internal/client"
	"github.com/alexGoLyceum/calculator-service/agent/internal/config"
	"github.com/alexGoLyceum/calculator-service/agent/internal/tasks"
	"github.com/alexGoLyceum/calculator-service/pkg/logging"
)

type Agent interface {
	Start() error
}

type Impl struct {
	Config *config.Config
	Logger logging.Logger
	Client client.Client
}

func NewAgent(cfg *config.Config, logger logging.Logger) (*Impl, error) {
	grpcClient, err := client.NewClient(cfg.Orchestrator)
	if err != nil {
		return nil, fmt.Errorf("failed to create grpc client: %w", err)
	}
	return &Impl{
		Config: cfg,
		Logger: logger,
		Client: grpcClient,
	}, nil
}

func (a *Impl) Start() error {
	defer a.Client.Close()

	ctx := context.Background()
	err := a.Client.StreamTasks(ctx, func(task *tasks.Task) error {
		result := tasks.Calculate(task)
		if err := a.Client.SetTaskResult(ctx, *task, result); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return err
	}
	return nil
}
