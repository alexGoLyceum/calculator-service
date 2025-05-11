package integration_test

import (
	"context"
	"github.com/alexGoLyceum/calculator-service/agent/internal/agent"
	"github.com/alexGoLyceum/calculator-service/agent/internal/config"
	"github.com/alexGoLyceum/calculator-service/agent/internal/tasks"
	"github.com/alexGoLyceum/calculator-service/pkg/logging/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/mock/gomock"
)

type mockClient struct {
	mock.Mock
}

func (m *mockClient) StreamTasks(ctx context.Context, handler func(task *tasks.Task) error) error {
	args := m.Called(ctx, handler)
	return args.Error(0)
}

func (m *mockClient) SetTaskResult(ctx context.Context, task tasks.Task, result float64) error {
	args := m.Called(ctx, task, result)
	return args.Error(0)
}

func (m *mockClient) Close() error {
	return nil
}

func TestAgent_Start(t *testing.T) {
	mockGrpcClient := new(mockClient)
	mockGrpcClient.On("StreamTasks", mock.Anything, mock.Anything).
		Return(nil).
		Run(func(args mock.Arguments) {
			handler := args.Get(1).(func(task *tasks.Task) error)
			task := &tasks.Task{
				Arg1:     tasks.Operand{Value: 2},
				Arg2:     tasks.Operand{Value: 3},
				Operator: "+",
			}
			_ = handler(task)
		})

	mockGrpcClient.On("SetTaskResult", mock.Anything, mock.Anything, float64(5)).
		Return(nil)

	cfg := &config.Config{}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	logger := mocks.NewMockLogger(ctrl)
	a := &agent.Impl{
		Config: cfg,
		Logger: logger,
		Client: mockGrpcClient,
	}
	err := a.Start()
	assert.NoError(t, err)
	mockGrpcClient.AssertExpectations(t)
}
