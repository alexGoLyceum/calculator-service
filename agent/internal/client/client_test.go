package client_test

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/alexGoLyceum/calculator-service/agent/internal/client"
	"github.com/alexGoLyceum/calculator-service/agent/internal/config"
	pb "github.com/alexGoLyceum/calculator-service/agent/internal/proto"
	"github.com/alexGoLyceum/calculator-service/agent/internal/tasks"
	"github.com/alexGoLyceum/calculator-service/agent/mocks"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestSetTaskResult_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockOrchestratorServiceClient(ctrl)

	task := tasks.Task{
		ID:            uuid.New(),
		ExpressionID:  uuid.New(),
		Arg1:          tasks.Operand{Value: 5},
		Arg2:          tasks.Operand{Value: 3},
		Operator:      "+",
		OperationTime: time.Now(),
		FinalTask:     false,
	}

	mockClient.EXPECT().
		SubmitTask(gomock.Any(), gomock.Any()).
		Return(&pb.SubmitTaskResponse{}, nil)

	c := &client.Impl{Client: mockClient}
	err := c.SetTaskResult(context.Background(), task, 8)
	require.NoError(t, err)
}

func TestSetTaskResult_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockOrchestratorServiceClient(ctrl)
	task := tasks.Task{
		ID:            uuid.New(),
		ExpressionID:  uuid.New(),
		Arg1:          tasks.Operand{Value: 2},
		Arg2:          tasks.Operand{Value: 3},
		Operator:      "*",
		OperationTime: time.Now(),
		FinalTask:     false,
	}

	mockClient.EXPECT().
		SubmitTask(gomock.Any(), gomock.Any()).
		Return(nil, errors.New("submit failed"))

	c := &client.Impl{Client: mockClient}
	err := c.SetTaskResult(context.Background(), task, 6)
	require.Error(t, err)
	require.Contains(t, err.Error(), "submit failed")
}

type mockStream struct {
	pb.OrchestratorService_AssignTasksClient
	tasks []*pb.Task
	index int
}

func (m *mockStream) Recv() (*pb.Task, error) {
	if m.index >= len(m.tasks) {
		return nil, errors.New("EOF")
	}
	t := m.tasks[m.index]
	m.index++
	return t, nil
}

func TestStreamTasks_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockOrchestratorServiceClient(ctrl)

	task := &pb.Task{
		Id:            uuid.New().String(),
		ExpressionId:  uuid.New().String(),
		Arg1Num:       1,
		Arg2Num:       2,
		Operator:      "+",
		OperationTime: timestamppb.Now(),
		FinalTask:     false,
	}

	stream := &mockStream{
		tasks: []*pb.Task{task},
	}

	mockClient.EXPECT().
		AssignTasks(gomock.Any(), gomock.Any()).
		Return(stream, nil)

	c := &client.Impl{Client: mockClient}
	err := c.StreamTasks(context.Background(), func(task *tasks.Task) error {
		require.Equal(t, 1.0, task.Arg1.Value)
		require.Equal(t, 2.0, task.Arg2.Value)
		require.Equal(t, "+", task.Operator)
		return nil
	})
	require.ErrorContains(t, err, "EOF")
}

func TestStreamTasks_AssignError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockOrchestratorServiceClient(ctrl)

	mockClient.EXPECT().
		AssignTasks(gomock.Any(), gomock.Any()).
		Return(nil, errors.New("assign failed"))

	c := &client.Impl{Client: mockClient}
	err := c.StreamTasks(context.Background(), func(t *tasks.Task) error { return nil })
	require.ErrorContains(t, err, "assign failed")
}

func TestNewClient_Success(t *testing.T) {
	lis, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)

	s := grpc.NewServer()
	go func() {
		_ = s.Serve(lis)
	}()
	defer s.Stop()

	cfg := config.OrchestratorConfig{
		Host: "localhost",
		Port: lis.Addr().(*net.TCPAddr).Port,
	}

	cl, err := client.NewClient(cfg)
	require.NoError(t, err)
	require.NotNil(t, cl)

	require.NoError(t, cl.Close())
}

func TestStreamTasks_HandlerError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockOrchestratorServiceClient(ctrl)

	task := &pb.Task{
		Id:            uuid.New().String(),
		ExpressionId:  uuid.New().String(),
		Arg1Num:       1,
		Arg2Num:       2,
		Operator:      "+",
		OperationTime: timestamppb.Now(),
		FinalTask:     false,
	}

	stream := &mockStream{
		tasks: []*pb.Task{task},
	}

	mockClient.EXPECT().
		AssignTasks(gomock.Any(), gomock.Any()).
		Return(stream, nil)

	c := &client.Impl{Client: mockClient}

	handlerErr := errors.New("handler failed")
	err := c.StreamTasks(context.Background(), func(task *tasks.Task) error {
		return handlerErr
	})

	require.Error(t, err)
	require.Equal(t, handlerErr, err)
}

func TestNewClient_Error(t *testing.T) {
	cfg := config.OrchestratorConfig{
		Host: "invalid-host",
		Port: 12345,
	}

	cl, err := client.NewClient(cfg)
	require.Error(t, err)
	require.Nil(t, cl)
	require.Contains(t, err.Error(), "could not connect")
}

func TestStreamTasks_ContextCancel(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockOrchestratorServiceClient(ctrl)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	stream := &mockStream{
		tasks: []*pb.Task{},
	}

	mockClient.EXPECT().
		AssignTasks(gomock.Any(), gomock.Any()).
		Return(stream, nil)

	c := &client.Impl{Client: mockClient}
	err := c.StreamTasks(ctx, func(task *tasks.Task) error {
		return nil
	})
	require.Error(t, err)
	require.Equal(t, context.Canceled, err)
}
