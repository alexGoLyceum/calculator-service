package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alexGoLyceum/calculator-service/orchestrator/internal/database/postgres"
	"github.com/alexGoLyceum/calculator-service/orchestrator/internal/models"
	"github.com/alexGoLyceum/calculator-service/orchestrator/internal/repository"
	"github.com/alexGoLyceum/calculator-service/orchestrator/internal/services"
	pb "github.com/alexGoLyceum/calculator-service/orchestrator/internal/transport/grpc/proto"
	"github.com/alexGoLyceum/calculator-service/orchestrator/mocks"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestExpressionTaskService_CreateExpressionTask(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	opTimes := &services.OperationTimesMS{
		Addition:       100 * time.Millisecond,
		Subtraction:    100 * time.Millisecond,
		Multiplication: 200 * time.Millisecond,
		Division:       200 * time.Millisecond,
	}
	service := services.NewExpressionTaskService(mockRepo, opTimes)

	userID := uuid.New()
	expressionID := uuid.New()

	tests := []struct {
		name        string
		expression  string
		mockSetup   func()
		expectedID  uuid.UUID
		expectedErr error
	}{
		{
			name:       "valid expression",
			expression: "2+2",
			mockSetup: func() {
				mockRepo.EXPECT().CreateExpressionTask(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
			expectedID:  expressionID,
			expectedErr: nil,
		},
		{
			name:        "empty expression",
			expression:  "",
			mockSetup:   func() {},
			expectedID:  uuid.Nil,
			expectedErr: services.ErrEmptyExpression,
		},
		{
			name:        "invalid characters",
			expression:  "2+a",
			mockSetup:   func() {},
			expectedID:  uuid.Nil,
			expectedErr: services.ErrInvalidCharacter,
		},
		{
			name:        "division by zero",
			expression:  "2/0",
			mockSetup:   func() {},
			expectedID:  uuid.Nil,
			expectedErr: services.ErrDivisionByZero,
		},
		{
			name:       "database unavailable",
			expression: "2+2",
			mockSetup: func() {
				mockRepo.EXPECT().CreateExpressionTask(gomock.Any(), gomock.Any(), gomock.Any()).Return(repository.ErrDatabaseNotAvailable)
			},
			expectedID:  uuid.Nil,
			expectedErr: services.ErrDatabaseUnavailable,
		},
		{
			name:       "unknown user ID",
			expression: "2+2",
			mockSetup: func() {
				mockRepo.EXPECT().CreateExpressionTask(gomock.Any(), gomock.Any(), gomock.Any()).Return(repository.ErrUnknownUserID)
			},
			expectedID:  uuid.Nil,
			expectedErr: services.ErrUnknownUserID,
		},
		{
			name:       "unexpected error",
			expression: "2+2",
			mockSetup: func() {
				mockRepo.EXPECT().CreateExpressionTask(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("unexpected error"))
			},
			expectedID:  uuid.Nil,
			expectedErr: errors.New("unexpected error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			id, err := service.CreateExpressionTask(context.Background(), userID, tt.expression)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr, err)
			} else {
				assert.NoError(t, err)
				assert.NotEqual(t, uuid.Nil, id)
			}
		})
	}
}

func TestExpressionTaskService_GetAllExpressions(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	opTimes := &services.OperationTimesMS{}
	service := services.NewExpressionTaskService(mockRepo, opTimes)

	userID := uuid.New()
	expressions := []*models.Expression{
		{ID: uuid.New(), UserID: userID, Expression: "2+2", Status: models.Done, Result: 4},
		{ID: uuid.New(), UserID: userID, Expression: "3*3", Status: models.InProgress},
	}

	tests := []struct {
		name           string
		mockSetup      func()
		expectedResult []*models.Expression
		expectedErr    error
	}{
		{
			name: "success",
			mockSetup: func() {
				mockRepo.EXPECT().GetAllExpressions(gomock.Any(), userID).Return(expressions, nil)
			},
			expectedResult: expressions,
			expectedErr:    nil,
		},
		{
			name: "unknown user ID",
			mockSetup: func() {
				mockRepo.EXPECT().GetAllExpressions(gomock.Any(), userID).Return(nil, repository.ErrUnknownUserID)
			},
			expectedResult: []*models.Expression{},
			expectedErr:    services.ErrUnknownUserID,
		},
		{
			name: "database unavailable",
			mockSetup: func() {
				mockRepo.EXPECT().GetAllExpressions(gomock.Any(), userID).Return(nil, repository.ErrDatabaseNotAvailable)
			},
			expectedResult: []*models.Expression{},
			expectedErr:    services.ErrDatabaseUnavailable,
		},
		{
			name: "unexpected error",
			mockSetup: func() {
				mockRepo.EXPECT().GetAllExpressions(gomock.Any(), userID).Return(nil, errors.New("unexpected error"))
			},
			expectedResult: nil,
			expectedErr:    errors.New("unexpected error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			result, err := service.GetAllExpressions(context.Background(), userID)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestExpressionTaskService_GetExpressionById(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	opTimes := &services.OperationTimesMS{}
	service := services.NewExpressionTaskService(mockRepo, opTimes)

	expressionID := uuid.New()
	userID := uuid.New()
	expression := &models.Expression{
		ID:         expressionID,
		UserID:     userID,
		Expression: "2+2",
		Status:     models.Done,
		Result:     4,
	}

	tests := []struct {
		name           string
		ctx            context.Context
		mockSetup      func()
		expectedResult *models.Expression
		expectedErr    error
	}{
		{
			name: "success",
			ctx:  context.WithValue(context.Background(), "user_id", userID),
			mockSetup: func() {
				mockRepo.EXPECT().GetExpressionByID(gomock.Any(), expressionID).Return(expression, nil)
			},
			expectedResult: expression,
			expectedErr:    nil,
		},
		{
			name: "expression not found",
			ctx:  context.WithValue(context.Background(), "user_id", userID),
			mockSetup: func() {
				mockRepo.EXPECT().GetExpressionByID(gomock.Any(), expressionID).Return(nil, repository.ErrUnknownExpressionID)
			},
			expectedResult: nil,
			expectedErr:    services.ErrUnknownExpressionsID,
		},
		{
			name: "database unavailable",
			ctx:  context.WithValue(context.Background(), "user_id", userID),
			mockSetup: func() {
				mockRepo.EXPECT().GetExpressionByID(gomock.Any(), expressionID).Return(nil, repository.ErrDatabaseNotAvailable)
			},
			expectedResult: nil,
			expectedErr:    services.ErrDatabaseUnavailable,
		},
		{
			name: "unexpected error",
			ctx:  context.WithValue(context.Background(), "user_id", userID),
			mockSetup: func() {
				mockRepo.EXPECT().GetExpressionByID(gomock.Any(), expressionID).Return(nil, errors.New("unexpected error"))
			},
			expectedResult: nil,
			expectedErr:    errors.New("unexpected error"),
		},
		{
			name: "forbidden - different user",
			ctx:  context.WithValue(context.Background(), "user_id", uuid.New()),
			mockSetup: func() {
				mockRepo.EXPECT().GetExpressionByID(gomock.Any(), expressionID).Return(expression, nil)
			},
			expectedResult: nil,
			expectedErr:    services.ErrForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			result, err := service.GetExpressionById(tt.ctx, expressionID)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestExpressionTaskService_GetTask(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	opTimes := &services.OperationTimesMS{
		Addition:       100 * time.Millisecond,
		Subtraction:    100 * time.Millisecond,
		Multiplication: 200 * time.Millisecond,
		Division:       200 * time.Millisecond,
	}
	service := services.NewExpressionTaskService(mockRepo, opTimes)

	expressionID := uuid.New()
	taskID := uuid.New()
	expectedTask := &pb.Task{
		Id:            taskID.String(),
		ExpressionId:  expressionID.String(),
		Arg1Num:       2,
		Arg2Num:       3,
		Operator:      "+",
		OperationTime: timestamppb.New(time.Now().Add(opTimes.Addition)),
		FinalTask:     false,
	}

	tests := []struct {
		name           string
		mockSetup      func()
		expectedResult *pb.Task
		expectedErr    error
	}{
		{
			name: "success",
			mockSetup: func() {
				mockRepo.EXPECT().GetTask(gomock.Any(), gomock.Any()).Return(expectedTask, nil)
			},
			expectedResult: expectedTask,
			expectedErr:    nil,
		},
		{
			name: "no tasks available",
			mockSetup: func() {
				mockRepo.EXPECT().GetTask(gomock.Any(), gomock.Any()).Return(nil, nil)
			},
			expectedResult: nil,
			expectedErr:    nil,
		},
		{
			name: "database unavailable",
			mockSetup: func() {
				mockRepo.EXPECT().GetTask(gomock.Any(), gomock.Any()).Return(nil, repository.ErrDatabaseNotAvailable)
			},
			expectedResult: nil,
			expectedErr:    services.ErrDatabaseUnavailable,
		},
		{
			name: "unexpected error",
			mockSetup: func() {
				mockRepo.EXPECT().GetTask(gomock.Any(), gomock.Any()).Return(nil, errors.New("unexpected error"))
			},
			expectedResult: nil,
			expectedErr:    errors.New("unexpected error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			result, err := service.GetTask(context.Background())

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestExpressionTaskService_SetTaskResult(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	opTimes := &services.OperationTimesMS{}
	service := services.NewExpressionTaskService(mockRepo, opTimes)

	task := &pb.Task{
		Id:           uuid.New().String(),
		ExpressionId: uuid.New().String(),
		Arg1Num:      2,
		Arg2Num:      3,
		Operator:     "+",
		FinalTask:    false,
	}

	tests := []struct {
		name        string
		mockSetup   func()
		result      float64
		expectedErr error
	}{
		{
			name: "success",
			mockSetup: func() {
				mockRepo.EXPECT().SetTaskResult(gomock.Any(), task, 5.0).Return(nil)
			},
			result:      5.0,
			expectedErr: nil,
		},
		{
			name: "task not found",
			mockSetup: func() {
				mockRepo.EXPECT().SetTaskResult(gomock.Any(), task, 5.0).Return(repository.ErrUnknownTaskID)
			},
			result:      5.0,
			expectedErr: services.ErrUnknownTaskID,
		},
		{
			name: "database unavailable",
			mockSetup: func() {
				mockRepo.EXPECT().SetTaskResult(gomock.Any(), task, 5.0).Return(repository.ErrDatabaseNotAvailable)
			},
			result:      5.0,
			expectedErr: services.ErrDatabaseUnavailable,
		},
		{
			name: "unexpected error",
			mockSetup: func() {
				mockRepo.EXPECT().SetTaskResult(gomock.Any(), task, 5.0).Return(errors.New("unexpected error"))
			},
			result:      5.0,
			expectedErr: errors.New("unexpected error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			err := service.SetTaskResult(context.Background(), task, tt.result)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestExpressionTaskService_StartExpiredTaskReset(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	opTimes := &services.OperationTimesMS{}
	service := services.NewExpressionTaskService(mockRepo, opTimes)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	interval := 100 * time.Millisecond
	delay := 10 * time.Second

	t.Run("start and stop", func(t *testing.T) {
		mockRepo.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		mockRepo.EXPECT().ResetExpiredTasks(gomock.Any(), delay).Return(nil).AnyTimes()

		service.StartExpiredTaskReset(ctx, interval, delay)
		time.Sleep(interval * 2)
		cancel()
	})

	t.Run("database unavailable", func(t *testing.T) {
		mockRepo.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).Return(repository.ErrDatabaseNotAvailable).AnyTimes()
		mockRepo.EXPECT().ResetExpiredTasks(gomock.Any(), delay).Return(repository.ErrDatabaseNotAvailable).AnyTimes()

		service.StartExpiredTaskReset(ctx, interval, delay)
		time.Sleep(interval * 2)
		cancel()
	})
}

func TestValidateExpression(t *testing.T) {
	tests := []struct {
		name        string
		expression  string
		expectedErr error
	}{
		{"valid expression", "2+2", nil},
		{"empty expression", "", services.ErrEmptyExpression},
		{"invalid characters", "2+a", services.ErrInvalidCharacter},
		{"division by zero", "2/0", services.ErrDivisionByZero},
		{"missing operator", "22", services.ErrMissingOperator},
		{"invalid parentheses", "(2+2", services.ErrParenthesisIssue},
		{"invalid number format", "2.2.2+3", services.ErrNumberFormatIssue},
		{"consecutive operators", "2++2", services.ErrOperatorIssue},
		{"invalid start with operator", "+2+2", services.ErrInvalidExpressionStartEnd},
		{"invalid end with operator", "2+2+", services.ErrInvalidExpressionStartEnd},
		{"unary operator not supported", "(-2)+3", services.ErrUnaryOperatorNotSupported},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := services.ValidateExpression(tt.expression)
			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestInfixToPostfix(t *testing.T) {
	tests := []struct {
		name       string
		expression string
		expected   []string
	}{
		{"simple addition", "2+3", []string{"2", "3", "+"}},
		{"with parentheses", "(2+3)*4", []string{"2", "3", "+", "4", "*"}},
		{"operator precedence", "2+3*4", []string{"2", "3", "4", "*", "+"}},
		{"complex expression", "3+4*2/(1-5)", []string{"3", "4", "2", "*", "1", "5", "-", "/", "+"}},
		{"decimal numbers", "2.5+3.7", []string{"2.5", "3.7", "+"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := services.InfixToPostfix(tt.expression)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetOperationEndTime(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	opTimes := &services.OperationTimesMS{
		Addition:       100 * time.Millisecond,
		Subtraction:    150 * time.Millisecond,
		Multiplication: 200 * time.Millisecond,
		Division:       250 * time.Millisecond,
	}
	service := services.NewExpressionTaskService(mockRepo, opTimes)

	tests := []struct {
		name     string
		operator string
	}{
		{"addition", "+"},
		{"subtraction", "-"},
		{"multiplication", "*"},
		{"division", "/"},
		{"invalid operator", "?"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.GetOperationEndTime(tt.operator)
			if tt.operator == "?" {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
			}
		})
	}
}

func TestIsExpressionError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"empty expression", services.ErrEmptyExpression, true},
		{"missing operator", services.ErrMissingOperator, true},
		{"invalid character", services.ErrInvalidCharacter, true},
		{"parenthesis issue", services.ErrParenthesisIssue, true},
		{"number format issue", services.ErrNumberFormatIssue, true},
		{"operator issue", services.ErrOperatorIssue, true},
		{"division by zero", services.ErrDivisionByZero, true},
		{"invalid start/end", services.ErrInvalidExpressionStartEnd, true},
		{"unary operator", services.ErrUnaryOperatorNotSupported, true},
		{"invalid expression", services.ErrInvalidExpression, true},
		{"other error", errors.New("other error"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := services.IsExpressionError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestStartExpiredTaskReset_ErrorHandling(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	opTimes := &services.OperationTimesMS{}
	service := services.NewExpressionTaskService(mockRepo, opTimes)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	interval := 10 * time.Millisecond
	delay := 1 * time.Second

	t.Run("continue on ErrDatabaseNotAvailable", func(t *testing.T) {
		mockRepo.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).DoAndReturn(
			func(ctx context.Context, fn func(context.Context, postgres.Tx) error) error {
				return fn(ctx, nil)
			},
		).AnyTimes()
		mockRepo.EXPECT().ResetExpiredTasks(gomock.Any(), delay).Return(repository.ErrDatabaseNotAvailable).AnyTimes()

		service.StartExpiredTaskReset(ctx, interval, delay)
		time.Sleep(interval * 2)
	})

	t.Run("return other error", func(t *testing.T) {
		expectedErr := errors.New("other error")
		mockRepo.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).DoAndReturn(
			func(ctx context.Context, fn func(context.Context, postgres.Tx) error) error {
				return fn(ctx, nil)
			},
		).AnyTimes()
		mockRepo.EXPECT().ResetExpiredTasks(gomock.Any(), delay).Return(expectedErr).AnyTimes()

		service.StartExpiredTaskReset(ctx, interval, delay)
		time.Sleep(interval * 2)
	})
}

func TestCreateExpressionTask_InvalidNumberParsing(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	opTimes := &services.OperationTimesMS{}
	service := services.NewExpressionTaskService(mockRepo, opTimes)

	userID := uuid.New()

	_, err := service.CreateExpressionTask(context.Background(), userID, "2..3+4")
	assert.Error(t, err)
	assert.Equal(t, services.ErrNumberFormatIssue, err)
}

func TestValidateExpression_ParenthesisAfterOperator(t *testing.T) {
	tests := []struct {
		name        string
		expression  string
		expectedErr error
	}{
		{
			name:        "operator before opening parenthesis",
			expression:  "2+(",
			expectedErr: services.ErrParenthesisIssue,
		},
		{
			name:        "number before opening parenthesis",
			expression:  "2(",
			expectedErr: services.ErrMissingOperator,
		},
		{
			name:        "valid operator before opening parenthesis",
			expression:  "2+(",
			expectedErr: services.ErrParenthesisIssue,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := services.ValidateExpression(tt.expression)
			assert.Error(t, err)
			assert.Equal(t, tt.expectedErr, err)
		})
	}
}

func TestValidateExpression_BalanceAndAfterParenthesis(t *testing.T) {
	tests := []struct {
		name        string
		expression  string
		expectedErr error
	}{
		{
			name:        "negative balance",
			expression:  "2+)3+4(",
			expectedErr: services.ErrParenthesisIssue,
		},
		{
			name:        "number after closing parenthesis",
			expression:  "(2+3)4",
			expectedErr: services.ErrParenthesisIssue,
		},
		{
			name:        "operator after closing parenthesis",
			expression:  "(2+3)+4",
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := services.ValidateExpression(tt.expression)
			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateExpression_NumberFormat(t *testing.T) {
	tests := []struct {
		name        string
		expression  string
		expectedErr error
	}{
		{
			name:        "leading dot",
			expression:  ".2+3",
			expectedErr: services.ErrNumberFormatIssue,
		},
		{
			name:        "trailing dot",
			expression:  "2.+3",
			expectedErr: services.ErrNumberFormatIssue,
		},
		{
			name:        "leading zero",
			expression:  "02+3",
			expectedErr: services.ErrNumberFormatIssue,
		},
		{
			name:        "valid decimal",
			expression:  "0.2+3",
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := services.ValidateExpression(tt.expression)
			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateExpression_ConsecutiveOperators(t *testing.T) {
	tests := []struct {
		name        string
		expression  string
		expectedErr error
	}{
		{
			name:        "consecutive operators",
			expression:  "2++3",
			expectedErr: services.ErrOperatorIssue,
		},
		{
			name:        "operator at start",
			expression:  "+2+3",
			expectedErr: services.ErrInvalidExpressionStartEnd,
		},
		{
			name:        "operator at end",
			expression:  "2+3+",
			expectedErr: services.ErrInvalidExpressionStartEnd,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := services.ValidateExpression(tt.expression)
			assert.Error(t, err)
			assert.True(t, errors.Is(err, tt.expectedErr) || errors.Is(err, services.ErrInvalidExpressionStartEnd))
		})
	}
}

func TestValidateExpression_ClosingParenthesisAfterOperator(t *testing.T) {
	tests := []struct {
		name        string
		expression  string
		expectedErr error
	}{
		{
			name:        "closing parenthesis after operator",
			expression:  "(+)",
			expectedErr: services.ErrUnaryOperatorNotSupported,
		},
		{
			name:        "empty parentheses",
			expression:  "2+()",
			expectedErr: services.ErrParenthesisIssue,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := services.ValidateExpression(tt.expression)
			assert.Error(t, err)
			assert.Equal(t, tt.expectedErr, err)
		})
	}
}
