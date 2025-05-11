package services

import (
	"context"
	"errors"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/alexGoLyceum/calculator-service/orchestrator/internal/database/postgres"
	"github.com/alexGoLyceum/calculator-service/orchestrator/internal/models"
	"github.com/alexGoLyceum/calculator-service/orchestrator/internal/repository"
	pb "github.com/alexGoLyceum/calculator-service/orchestrator/internal/transport/grpc/proto"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ExpressionTaskService interface {
	CreateExpressionTask(ctx context.Context, userID uuid.UUID, expression string) (uuid.UUID, error)
	GetAllExpressions(ctx context.Context, userID uuid.UUID) ([]*models.Expression, error)
	GetExpressionById(ctx context.Context, expression uuid.UUID) (*models.Expression, error)
	GetTask(ctx context.Context) (*pb.Task, error)
	SetTaskResult(ctx context.Context, task *pb.Task, result float64) error
	StartExpiredTaskReset(ctx context.Context, interval, delay time.Duration)
	GetOperationEndTime(operator string) *timestamppb.Timestamp
}

type OperationTimesMS struct {
	Addition       time.Duration
	Subtraction    time.Duration
	Multiplication time.Duration
	Division       time.Duration
}

type expressionTaskService struct {
	cfg  *OperationTimesMS
	repo repository.Repository
}

func NewExpressionTaskService(repo repository.Repository, cfg *OperationTimesMS) ExpressionTaskService {
	return &expressionTaskService{
		repo: repo,
		cfg:  cfg,
	}
}

func (s *expressionTaskService) StartExpiredTaskReset(ctx context.Context, interval, delay time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := s.repo.WithTransaction(ctx, func(txCtx context.Context, tx postgres.Tx) error {
					return s.repo.ResetExpiredTasks(txCtx, delay)
				}); err != nil {
					if errors.Is(err, repository.ErrDatabaseNotAvailable) {
						continue
					}
				}
			}
		}
	}()
}

func (s *expressionTaskService) CreateExpressionTask(ctx context.Context, userID uuid.UUID, expression string) (uuid.UUID, error) {
	if err := ValidateExpression(expression); err != nil {
		return uuid.Nil, err
	}

	exprID := uuid.New()
	expression = strings.ReplaceAll(expression, " ", "")

	var tasks []*models.Task
	expressionToSave := &models.Expression{
		ID:         exprID,
		UserID:     userID,
		Expression: expression,
		Status:     models.Pending,
	}

	postfix := InfixToPostfix(expression)
	var stack []*models.Operand

	for i, token := range postfix {
		if isNumber(token) {
			val, err := strconv.ParseFloat(token, 64)
			if err != nil {
				return uuid.Nil, ErrInvalidExpression
			}
			stack = append(stack, &models.Operand{Value: val})
		} else if isOperator(token) {
			right := stack[len(stack)-1]
			left := stack[len(stack)-2]
			stack = stack[:len(stack)-2]
			isFinalTask := (i == len(postfix)-1)
			task := models.Task{
				ID:            uuid.New(),
				ExpressionID:  exprID,
				Arg1:          *left,
				Arg2:          *right,
				Operator:      token,
				OperationTime: time.Time{},
				FinalTask:     isFinalTask,
			}
			tasks = append(tasks, &task)
			stack = append(stack, &models.Operand{Value: math.NaN(), TaskID: &task.ID})
		}
	}

	if err := s.repo.CreateExpressionTask(ctx, expressionToSave, tasks); err != nil {
		if errors.Is(err, repository.ErrUnknownUserID) {
			return uuid.Nil, ErrUnknownUserID
		}
		if errors.Is(err, repository.ErrDatabaseNotAvailable) {
			return uuid.Nil, ErrDatabaseUnavailable
		}
		return uuid.Nil, err
	}

	return exprID, nil
}

func (s *expressionTaskService) GetAllExpressions(ctx context.Context, userID uuid.UUID) ([]*models.Expression, error) {
	expressions, err := s.repo.GetAllExpressions(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrUnknownUserID) {
			return []*models.Expression{}, ErrUnknownUserID
		}
		if errors.Is(err, repository.ErrDatabaseNotAvailable) {
			return []*models.Expression{}, ErrDatabaseUnavailable
		}
		return nil, err
	}
	return expressions, nil
}

func (s *expressionTaskService) GetExpressionById(ctx context.Context, expressionID uuid.UUID) (*models.Expression, error) {
	userID, ok := ctx.Value("user_id").(uuid.UUID)
	if !ok {
		return nil, ErrForbidden
	}

	expression, err := s.repo.GetExpressionByID(ctx, expressionID)
	if err != nil {
		if errors.Is(err, repository.ErrUnknownExpressionID) {
			return nil, ErrUnknownExpressionsID
		}
		if errors.Is(err, repository.ErrDatabaseNotAvailable) {
			return nil, ErrDatabaseUnavailable
		}
		return nil, err
	}
	if expression.UserID != userID {
		return nil, ErrForbidden
	}
	return expression, nil
}

func (s *expressionTaskService) GetTask(ctx context.Context) (*pb.Task, error) {
	task, err := s.repo.GetTask(ctx, s.GetOperationEndTime)
	if err != nil {
		if errors.Is(err, repository.ErrDatabaseNotAvailable) {
			return nil, ErrDatabaseUnavailable
		}
		return nil, err
	}
	if task == nil {
		return nil, nil
	}
	task.OperationTime = s.GetOperationEndTime(task.Operator)
	return task, nil
}

func (s *expressionTaskService) SetTaskResult(ctx context.Context, task *pb.Task, result float64) error {
	if err := s.repo.SetTaskResult(ctx, task, result); err != nil {
		if errors.Is(err, repository.ErrDatabaseNotAvailable) {
			return ErrDatabaseUnavailable
		}
		if errors.Is(err, repository.ErrUnknownTaskID) {
			return ErrUnknownTaskID
		}
		return err
	}
	return nil
}

func InfixToPostfix(expression string) []string {
	var stack []string
	var output []string
	for i := 0; i < len(expression); i++ {
		ch := string(expression[i])
		if isNumber(ch) {
			num := ch
			for i+1 < len(expression) && (isNumber(string(expression[i+1])) || string(expression[i+1]) == ".") {
				i++
				num += string(expression[i])
			}
			output = append(output, num)
		} else if ch == "(" {
			stack = append(stack, ch)
		} else if ch == ")" {
			for len(stack) > 0 && stack[len(stack)-1] != "(" {
				output = append(output, stack[len(stack)-1])
				stack = stack[:len(stack)-1]
			}
			stack = stack[:len(stack)-1]
		} else if isOperator(ch) {
			for len(stack) > 0 && precedence(stack[len(stack)-1]) >= precedence(ch) {
				output = append(output, stack[len(stack)-1])
				stack = stack[:len(stack)-1]
			}
			stack = append(stack, ch)
		}
	}
	for len(stack) > 0 {
		output = append(output, stack[len(stack)-1])
		stack = stack[:len(stack)-1]
	}
	return output
}

func (s *expressionTaskService) GetOperationEndTime(operator string) *timestamppb.Timestamp {
	var endTime time.Time
	switch operator {
	case "+":
		endTime = time.Now().Add(s.cfg.Addition)
	case "-":
		endTime = time.Now().Add(s.cfg.Subtraction)
	case "*":
		endTime = time.Now().Add(s.cfg.Multiplication)
	case "/":
		endTime = time.Now().Add(s.cfg.Division)
	default:
		return nil
	}
	return timestamppb.New(endTime)
}

func precedence(op string) int {
	switch op {
	case "+", "-":
		return 1
	case "*", "/":
		return 2
	}
	return 0
}

func isNumber(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

func isOperator(s string) bool {
	return s == "+" || s == "-" || s == "*" || s == "/"
}

func ValidateExpression(expression string) error {
	expression = strings.ReplaceAll(expression, " ", "")
	if len(expression) == 0 {
		return ErrEmptyExpression
	}

	if err := checkDivisionByZero(expression); err != nil {
		return err
	}

	if !containsOperator(expression) {
		return ErrMissingOperator
	}

	if err := validateCharacters(expression); err != nil {
		return err
	}

	if err := validateParentheses(expression); err != nil {
		return err
	}

	if err := validateNumbers(expression); err != nil {
		return err
	}

	if err := validateOperators(expression); err != nil {
		return err
	}

	if isOperator(string(expression[0])) || isOperator(string(expression[len(expression)-1])) {
		return ErrInvalidExpressionStartEnd
	}

	return nil
}

func containsOperator(expression string) bool {
	for _, ch := range expression {
		if isOperator(string(ch)) {
			return true
		}
	}
	return false
}

func validateCharacters(expression string) error {
	for _, ch := range expression {
		if !(unicode.IsDigit(ch) || ch == '.' || isOperator(string(ch)) || ch == '(' || ch == ')') {
			return ErrInvalidCharacter
		}
	}
	return nil
}

func validateParentheses(expression string) error {
	balance := 0
	lastChar := ' '

	for i, ch := range expression {
		switch ch {
		case '(':
			balance++
			if i > 0 && !isOperator(string(lastChar)) && lastChar != '(' {
				return ErrParenthesisIssue
			}
		case ')':
			balance--
			if balance < 0 {
				return ErrParenthesisIssue
			}
			if i+1 < len(expression) && !isOperator(string(expression[i+1])) && expression[i+1] != ')' {
				return ErrParenthesisIssue
			}
		}
		lastChar = ch
	}

	if balance != 0 {
		return ErrParenthesisIssue
	}
	return nil
}

func validateNumbers(expression string) error {
	parts := strings.FieldsFunc(expression, func(c rune) bool {
		return isOperator(string(c)) || c == '(' || c == ')'
	})

	for _, part := range parts {
		if strings.Count(part, ".") > 1 {
			return ErrNumberFormatIssue
		}
		if strings.HasPrefix(part, ".") || strings.HasSuffix(part, ".") {
			return ErrNumberFormatIssue
		}
		if len(part) > 1 && part[0] == '0' && part[1] != '.' {
			return ErrNumberFormatIssue
		}
	}
	return nil
}

func validateOperators(expression string) error {
	lastWasOperator := false

	for i, ch := range expression {
		if isOperator(string(ch)) {
			if lastWasOperator {
				return ErrOperatorIssue
			}
			if i+1 < len(expression) && (expression[i+1] == ')' || isOperator(string(expression[i+1]))) {
				return ErrOperatorIssue
			}
			lastWasOperator = true
		} else {
			lastWasOperator = false
		}

		if ch == '(' {
			if i+1 < len(expression) && expression[i+1] == ')' {
				return ErrParenthesisIssue
			}
		}

		if ch == '(' && i+1 < len(expression) && isOperator(string(expression[i+1])) {
			return ErrUnaryOperatorNotSupported
		}
	}

	return nil
}

var divisionByZeroRegex = regexp.MustCompile(`/\s*[+-]?\s*0+(\.0*)?([^0-9.]|$)`)

func checkDivisionByZero(expression string) error {
	if divisionByZeroRegex.MatchString(expression) {
		return ErrDivisionByZero
	}
	return nil
}

func IsExpressionError(err error) bool {
	return errors.Is(err, ErrEmptyExpression) ||
		errors.Is(err, ErrMissingOperator) ||
		errors.Is(err, ErrInvalidCharacter) ||
		errors.Is(err, ErrParenthesisIssue) ||
		errors.Is(err, ErrNumberFormatIssue) ||
		errors.Is(err, ErrOperatorIssue) ||
		errors.Is(err, ErrDivisionByZero) ||
		errors.Is(err, ErrInvalidExpressionStartEnd) ||
		errors.Is(err, ErrUnaryOperatorNotSupported) ||
		errors.Is(err, ErrInvalidExpression)
}
