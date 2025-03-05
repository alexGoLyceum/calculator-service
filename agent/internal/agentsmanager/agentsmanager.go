package agentsmanager

import (
	"sync"
	"time"

	"github.com/alexGoLyceum/calculator-service/agent/internal/client"
	"github.com/alexGoLyceum/calculator-service/agent/internal/config"
	"github.com/alexGoLyceum/calculator-service/agent/internal/tasks"

	"go.uber.org/zap"
)

type AgentsManager struct {
	Config *config.AgentConfig
	Logger *zap.Logger
}

func NewAgentsManager(cfg *config.AgentConfig, logger *zap.Logger) *AgentsManager {
	return &AgentsManager{
		Config: cfg,
		Logger: logger,
	}
}

func (a *AgentsManager) Agent() {
	for {
		task, err := client.GetTask()
		if err != nil {
			a.Logger.Error("Failed to get task", zap.Error(err))
			time.Sleep(500 * time.Millisecond)
			continue
		}
		if task == nil {
			time.Sleep(500 * time.Millisecond)
			continue
		}
		result := tasks.Calculate(task)
		if err := client.SetTaskResult(*task, result); err != nil {
			a.Logger.Error("Failed to set task result", zap.Error(err))
		}
	}
}

func (a *AgentsManager) Start() {
	var wg sync.WaitGroup
	for range a.Config.ComputingPower {
		wg.Add(1)
		go func() {
			defer wg.Done()
			a.Agent()
		}()
	}
	wg.Wait()
}
