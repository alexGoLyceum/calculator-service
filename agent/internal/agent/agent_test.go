package agent_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/alexGoLyceum/calculator-service/agent/internal/agent"
	"github.com/alexGoLyceum/calculator-service/agent/internal/client"
	"github.com/alexGoLyceum/calculator-service/agent/internal/config"
	"github.com/alexGoLyceum/calculator-service/agent/internal/tasks"
	"github.com/alexGoLyceum/calculator-service/agent/mocks"
	logmock "github.com/alexGoLyceum/calculator-service/pkg/logging/mocks"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestAgent_Start_SuccessfulFlow(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockClient(ctrl)
	mockLogger := logmock.NewMockLogger(ctrl)

	testTask := &tasks.Task{
		ID:           uuid.New(),
		ExpressionID: uuid.Nil,
		Arg1: tasks.Operand{
			Value: 2.0,
		},
		Arg2: tasks.Operand{
			Value: 2.0,
		},
		Operator:      "+",
		OperationTime: time.Now(),
	}

	mockClient.EXPECT().StreamTasks(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, handler func(*tasks.Task) error) error {
			err := handler(testTask)
			require.NoError(t, err)
			return nil
		})

	expectedResult := tasks.Calculate(testTask)
	mockClient.EXPECT().SetTaskResult(gomock.Any(), *testTask, expectedResult).Return(nil)
	mockClient.EXPECT().Close().Return(nil)

	a := &agent.Impl{
		Config: &config.Config{},
		Logger: mockLogger,
		Client: mockClient,
	}

	a.Start()
}

func TestAgent_Start_SetTaskResultError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockClient(ctrl)
	mockLogger := logmock.NewMockLogger(ctrl)

	testTask := &tasks.Task{
		ID:           uuid.New(),
		ExpressionID: uuid.Nil,
		Arg1: tasks.Operand{
			Value: 2.0,
		},
		Arg2: tasks.Operand{
			Value: 2.0,
		},
		Operator:      "+",
		OperationTime: time.Now(),
	}

	mockClient.EXPECT().StreamTasks(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, handler func(*tasks.Task) error) error {
			_ = handler(testTask)
			return nil
		})

	mockClient.EXPECT().SetTaskResult(gomock.Any(), *testTask, tasks.Calculate(testTask)).
		Return(errors.New("set task result failed"))

	mockClient.EXPECT().Close().Return(nil)

	a := &agent.Impl{
		Config: &config.Config{},
		Logger: mockLogger,
		Client: mockClient,
	}

	a.Start()
}

func TestAgent_Start_StreamTasksError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockClient(ctrl)
	mockLogger := logmock.NewMockLogger(ctrl)

	expectedErr := errors.New("stream error")
	mockClient.EXPECT().StreamTasks(gomock.Any(), gomock.Any()).Return(expectedErr)
	mockClient.EXPECT().Close().Return(nil)

	a := &agent.Impl{
		Config: &config.Config{},
		Logger: mockLogger,
		Client: mockClient,
	}

	err := a.Start()
	if !errors.Is(err, expectedErr) {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}
}

func TestNewAgent_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	originalNewClient := client.NewClient
	defer func() {
		client.NewClient = originalNewClient
	}()

	mockClient := mocks.NewMockClient(ctrl)
	client.NewClient = func(cfg config.OrchestratorConfig) (client.Client, error) {
		return mockClient, nil
	}

	mockLogger := logmock.NewMockLogger(ctrl)

	cfg := &config.Config{
		Orchestrator: config.OrchestratorConfig{
			Host: "localhost",
			Port: 0,
		},
	}

	agentInstance, err := agent.NewAgent(cfg, mockLogger)

	require.NoError(t, err)
	require.NotNil(t, agentInstance)
	require.Equal(t, cfg, agentInstance.Config)
	require.Equal(t, mockLogger, agentInstance.Logger)
	require.Equal(t, mockClient, agentInstance.Client)
}

func TestNewAgent_ClientCreationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	originalNewClient := client.NewClient
	defer func() {
		client.NewClient = originalNewClient
	}()

	expectedErr := errors.New("connection error")
	client.NewClient = func(cfg config.OrchestratorConfig) (client.Client, error) {
		return nil, expectedErr
	}

	mockLogger := logmock.NewMockLogger(ctrl)

	cfg := &config.Config{
		Orchestrator: config.OrchestratorConfig{
			Host: "localhost",
			Port: 0,
		},
	}

	agentInstance, err := agent.NewAgent(cfg, mockLogger)

	require.Error(t, err)
	require.Nil(t, agentInstance)
	require.EqualError(t, err, fmt.Sprintf("failed to create grpc client: %v", expectedErr))
}
