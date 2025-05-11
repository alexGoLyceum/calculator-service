package client

import (
	"context"
	"fmt"
	"time"

	"github.com/alexGoLyceum/calculator-service/agent/internal/config"
	pb "github.com/alexGoLyceum/calculator-service/agent/internal/proto"
	"github.com/alexGoLyceum/calculator-service/agent/internal/tasks"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Client interface {
	StreamTasks(ctx context.Context, handler func(task *tasks.Task) error) error
	SetTaskResult(ctx context.Context, task tasks.Task, result float64) error
	Close() error
}

type Impl struct {
	Client pb.OrchestratorServiceClient
	Conn   *grpc.ClientConn
}

type NewClientFunc func(cfg config.OrchestratorConfig) (Client, error)

var NewClient NewClientFunc = defaultNewClient

func defaultNewClient(cfg config.OrchestratorConfig) (Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(
		ctx,
		fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		grpc.WithInsecure(),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("could not connect: %w", err)
	}

	c := pb.NewOrchestratorServiceClient(conn)

	return &Impl{
		Client: c,
		Conn:   conn,
	}, nil
}

func (c *Impl) StreamTasks(ctx context.Context, handler func(task *tasks.Task) error) error {
	stream, err := c.Client.AssignTasks(ctx, &pb.AssignTasksRequest{})
	if err != nil {
		return fmt.Errorf("failed to start stream: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			t, err := stream.Recv()
			if err != nil {
				return fmt.Errorf("stream receive failed: %w", err)
			}
			exprID, _ := uuid.Parse(t.ExpressionId)
			taskID, _ := uuid.Parse(t.Id)

			task := &tasks.Task{
				ID:            taskID,
				ExpressionID:  exprID,
				Arg1:          tasks.Operand{Value: t.Arg1Num},
				Arg2:          tasks.Operand{Value: t.Arg2Num},
				Operator:      t.Operator,
				OperationTime: t.OperationTime.AsTime(),
				FinalTask:     t.FinalTask,
			}

			if err := handler(task); err != nil {
				return err
			}
		}
	}
}

func (c *Impl) SetTaskResult(ctx context.Context, task tasks.Task, result float64) error {
	req := &pb.SubmitTaskRequest{
		Task: &pb.Task{
			Id:            task.ID.String(),
			ExpressionId:  task.ExpressionID.String(),
			Arg1Num:       task.Arg1.Value,
			Arg2Num:       task.Arg2.Value,
			Operator:      task.Operator,
			OperationTime: timestamppb.New(task.OperationTime),
			FinalTask:     task.FinalTask,
		},
		Result: result,
	}

	if _, err := c.Client.SubmitTask(ctx, req); err != nil {
		return fmt.Errorf("failed to submit task result: %w", err)
	}
	return nil
}

func (c *Impl) Close() error {
	return c.Conn.Close()
}
