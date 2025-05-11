package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/alexGoLyceum/calculator-service/orchestrator/internal/services"
	pb "github.com/alexGoLyceum/calculator-service/orchestrator/internal/transport/grpc/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server interface {
	pb.OrchestratorServiceServer
	Start() error
	AssignTasks(req *pb.AssignTasksRequest, stream pb.OrchestratorService_AssignTasksServer) error
}

type server struct {
	pb.UnimplementedOrchestratorServiceServer
	exprTaskService services.ExpressionTaskService
	address         string
}

func NewServer(exprTaskService services.ExpressionTaskService, host string, port int) Server {
	return &server{
		exprTaskService: exprTaskService,
		address:         fmt.Sprintf("%s:%d", host, port),
	}
}

func (s *server) AssignTasks(req *pb.AssignTasksRequest, stream pb.OrchestratorService_AssignTasksServer) error {
	ctx := stream.Context()

	for {
		select {
		case <-ctx.Done():
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				return status.Error(codes.DeadlineExceeded, ctx.Err().Error())
			}
			return nil
		default:
			task, err := s.exprTaskService.GetTask(ctx)
			if err != nil {
				if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
					return status.FromContextError(err).Err()
				}
				return status.Error(codes.Internal, "failed to get task")
			}

			if task == nil {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			if err := stream.Send(task); err != nil {
				return status.Errorf(codes.Unavailable, "failed to send task: %v", err)
			}
		}
	}
}

func (s *server) SubmitTask(ctx context.Context, req *pb.SubmitTaskRequest) (*pb.SubmitTaskResponse, error) {
	if req.Task == nil {
		return nil, status.Error(codes.InvalidArgument, "task required")
	}
	err := s.exprTaskService.SetTaskResult(ctx, req.Task, req.Result)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrDatabaseUnavailable):
			return nil, status.Error(codes.Unavailable, "server is unavailable")
		case errors.Is(err, services.ErrUnknownTaskID):
			return nil, status.Error(codes.NotFound, "task id not found")
		default:
			return nil, status.Error(codes.Internal, "failed to set result")
		}
	}
	return &pb.SubmitTaskResponse{}, nil
}

func (s *server) Start() error {
	listener, err := net.Listen("tcp", s.address)
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer()
	pb.RegisterOrchestratorServiceServer(grpcServer, s)

	return grpcServer.Serve(listener)
}
