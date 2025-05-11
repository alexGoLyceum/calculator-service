package server_test

import (
	"context"
	"errors"
	"testing"

	"github.com/alexGoLyceum/calculator-service/orchestrator/internal/services"
	"github.com/alexGoLyceum/calculator-service/orchestrator/internal/transport/grpc/proto"
	"github.com/alexGoLyceum/calculator-service/orchestrator/internal/transport/grpc/server"
	"github.com/alexGoLyceum/calculator-service/orchestrator/mocks"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestSubmitTask(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockExpressionTaskService(ctrl)
	s := server.NewServer(mockService, "localhost", 0)

	t.Run("nil task", func(t *testing.T) {
		resp, err := s.SubmitTask(context.Background(), &proto.SubmitTaskRequest{})
		require.Nil(t, resp)
		require.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("database unavailable", func(t *testing.T) {
		req := &proto.SubmitTaskRequest{Task: &proto.Task{}}
		mockService.EXPECT().
			SetTaskResult(gomock.Any(), req.Task, req.Result).
			Return(services.ErrDatabaseUnavailable)

		resp, err := s.SubmitTask(context.Background(), req)
		require.Nil(t, resp)
		require.Equal(t, codes.Unavailable, status.Code(err))
	})

	t.Run("unknown task id", func(t *testing.T) {
		req := &proto.SubmitTaskRequest{Task: &proto.Task{}}
		mockService.EXPECT().
			SetTaskResult(gomock.Any(), req.Task, req.Result).
			Return(services.ErrUnknownTaskID)

		resp, err := s.SubmitTask(context.Background(), req)
		require.Nil(t, resp)
		require.Equal(t, codes.NotFound, status.Code(err))
	})

	t.Run("internal error", func(t *testing.T) {
		req := &proto.SubmitTaskRequest{Task: &proto.Task{}}
		mockService.EXPECT().
			SetTaskResult(gomock.Any(), req.Task, req.Result).
			Return(errors.New("unexpected"))

		resp, err := s.SubmitTask(context.Background(), req)
		require.Nil(t, resp)
		require.Equal(t, codes.Internal, status.Code(err))
	})

	t.Run("success", func(t *testing.T) {
		req := &proto.SubmitTaskRequest{Task: &proto.Task{}}
		mockService.EXPECT().
			SetTaskResult(gomock.Any(), req.Task, req.Result).
			Return(nil)

		resp, err := s.SubmitTask(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, resp)
	})
}

func TestStart_ListenError(t *testing.T) {
	mockService := mocks.NewMockExpressionTaskService(gomock.NewController(t))
	s := server.NewServer(mockService, "invalid_host", -1)

	err := s.Start()
	require.Error(t, err)
}

func TestAssignTasks(t *testing.T) {
	tests := []struct {
		name       string
		setupMock  func(*mocks.MockExpressionTaskService, *mocks.MockOrchestratorService_AssignTasksServer[*proto.Task])
		wantErr    bool
		errContain string
		code       codes.Code
	}{
		{
			name: "successful stream",
			setupMock: func(ets *mocks.MockExpressionTaskService, stream *mocks.MockOrchestratorService_AssignTasksServer[*proto.Task]) {
				ctx, cancel := context.WithCancel(context.Background())
				stream.EXPECT().Context().Return(ctx).AnyTimes()

				task := &proto.Task{Id: "123"}
				ets.EXPECT().GetTask(gomock.Any()).Return(task, nil).Times(1)
				stream.EXPECT().Send(task).Return(nil).Times(1)

				ets.EXPECT().GetTask(gomock.Any()).DoAndReturn(func(_ context.Context) (*proto.Task, error) {
					cancel()
					return nil, nil
				}).Times(1)
			},
			wantErr: false,
		},
		{
			name: "GetTask returns error",
			setupMock: func(ets *mocks.MockExpressionTaskService, stream *mocks.MockOrchestratorService_AssignTasksServer[*proto.Task]) {
				stream.EXPECT().Context().Return(context.Background()).AnyTimes()
				ets.EXPECT().GetTask(gomock.Any()).Return(nil, errors.New("internal error")).Times(1)
			},
			wantErr: true,
			code:    codes.Internal,
		},
		{
			name: "Send returns error",
			setupMock: func(ets *mocks.MockExpressionTaskService, stream *mocks.MockOrchestratorService_AssignTasksServer[*proto.Task]) {
				stream.EXPECT().Context().Return(context.Background()).AnyTimes()
				task := &proto.Task{Id: "123"}
				ets.EXPECT().GetTask(gomock.Any()).Return(task, nil).Times(1)
				stream.EXPECT().Send(task).Return(errors.New("unavailable")).Times(1)
			},
			wantErr: true,
			code:    codes.Unavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockETS := mocks.NewMockExpressionTaskService(ctrl)
			mockStream := mocks.NewMockOrchestratorService_AssignTasksServer[*proto.Task](ctrl)

			tt.setupMock(mockETS, mockStream)

			srv := server.NewServer(mockETS, "localhost", 50051)

			err := srv.AssignTasks(&proto.AssignTasksRequest{}, mockStream)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContain != "" {
					require.Contains(t, err.Error(), tt.errContain)
				}
				if tt.code != 0 {
					st, ok := status.FromError(err)
					require.True(t, ok, "error should be a gRPC status error")
					require.Equal(t, tt.code, st.Code())
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}
