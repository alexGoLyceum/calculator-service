package taskmanager

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/alexGoLyceum/calculator-service/orchestrator/internal/config"

	"github.com/google/uuid"
)

type Status string

const (
	Pending    Status = "pending"
	InProgress Status = "in progress"
	Completed  Status = "completed"
)

type Expression struct {
	ID         uuid.UUID `json:"id"`
	Expression string    `json:"expression"`
	Status     Status    `json:"status"`
	Result     float64   `json:"result,omitempty"`
}

type Task struct {
	ID            uuid.UUID `json:"id"`
	ExpressionID  uuid.UUID `json:"expression_id"`
	Arg1          Operand   `json:"arg1"`
	Arg2          Operand   `json:"arg2"`
	Operator      string    `json:"operator"`
	OperationTime time.Time `json:"operation_time"`
	FinalTask     bool      `json:"final_task"`
}

type Operand struct {
	Value  float64   `json:"value"`
	TaskID uuid.UUID `json:"task_id"`
}

type TaskManager struct {
	Config         *config.TaskManagerConfig
	Expressions    sync.Map
	TaskResults    sync.Map
	InProcessTasks []Task
	Tasks          []Task
	Cond           *sync.Cond
}

func (tm *TaskManager) CheckExpiredTasks() {
	for {
		time.Sleep(5 * time.Second)
		if len(tm.InProcessTasks) == 0 {
			continue
		}

		now := time.Now()
		tm.Cond.L.Lock()
		for i := 0; i < len(tm.InProcessTasks); i++ {
			task := tm.InProcessTasks[i]
			if task.OperationTime.Add(10 * time.Second).Before(now) {
				tm.InProcessTasks = append(tm.InProcessTasks[:i], tm.InProcessTasks[i+1:]...)
				tm.Tasks = append(tm.Tasks, task)
				i--
			}
		}
		tm.Cond.L.Unlock()
	}
}

func NewTaskManager(config *config.TaskManagerConfig) *TaskManager {
	return &TaskManager{
		Cond:   sync.NewCond(&sync.Mutex{}),
		Config: config,
	}
}

func (tm *TaskManager) GetExpressions() []Expression {
	expressions := make([]Expression, 0)
	tm.Expressions.Range(func(key, value interface{}) bool {
		expressions = append(expressions, *value.(*Expression))
		return true
	})
	return expressions
}

func (tm *TaskManager) GetExpressionByID(id uuid.UUID) (*Expression, error) {
	value, ok := tm.Expressions.Load(id)
	if !ok {
		return nil, fmt.Errorf("expression with id %s not found", id)
	}
	return value.(*Expression), nil
}

func (tm *TaskManager) SetResult(task Task, result float64) error {
	tm.Cond.L.Lock()
	defer tm.Cond.L.Unlock()
	found := true
	for i, t := range tm.InProcessTasks {
		if t.ID == task.ID {
			found = false
			tm.InProcessTasks = append(tm.InProcessTasks[:i], tm.InProcessTasks[i+1:]...)
			break
		}
	}
	if found {
		return fmt.Errorf("task with ID %s not found", task.ID)
	}

	if expr, ok := tm.Expressions.Load(task.ExpressionID); ok {
		if task.FinalTask {
			expr.(*Expression).Result = result
			expr.(*Expression).Status = Completed
		} else {
			tm.TaskResults.Store(task.ID, result)
		}
	} else {
		return fmt.Errorf("expression with id %s not found", task.ExpressionID)
	}
	tm.Cond.Broadcast()
	return nil
}

func (tm *TaskManager) GetTask() *Task {
	tm.Cond.L.Lock()
	defer tm.Cond.L.Unlock()

	for i := 0; i < len(tm.Tasks); i++ {
		task := &tm.Tasks[i]

		if task.Arg1.TaskID != uuid.Nil {
			if result, ok := tm.TaskResults.Load(task.Arg1.TaskID); ok {
				task.Arg1.Value = result.(float64)
				task.Arg1.TaskID = uuid.Nil
				tm.TaskResults.Delete(task.Arg1.TaskID)
			}
		}
		if task.Arg2.TaskID != uuid.Nil {
			if result, ok := tm.TaskResults.Load(task.Arg2.TaskID); ok {
				task.Arg2.Value = result.(float64)
				task.Arg2.TaskID = uuid.Nil
				tm.TaskResults.Delete(task.Arg2.TaskID)
			}
		}
		if task.Arg1.TaskID == uuid.Nil && task.Arg2.TaskID == uuid.Nil {
			if expr, ok := tm.Expressions.Load(task.ExpressionID); ok {
				expr.(*Expression).Status = InProgress
			}
			selectedTask := *task
			selectedTask.OperationTime = tm.GetOperationEndTime(selectedTask.Operator)
			tm.Tasks = append(tm.Tasks[:i], tm.Tasks[i+1:]...)
			tm.InProcessTasks = append(tm.InProcessTasks, selectedTask)
			return &selectedTask
		}
	}
	return nil
}

func (tm *TaskManager) CreateAndStoreTasksFromExpression(expression string) (uuid.UUID, error) {
	exprID := uuid.New()
	expression = strings.ReplaceAll(expression, " ", "")
	tm.Expressions.Store(exprID, &Expression{
		ID:         exprID,
		Expression: expression,
		Status:     Pending},
	)

	postfix := InfixToPostfix(expression)

	tm.Cond.L.Lock()
	defer tm.Cond.L.Unlock()
	var taskStack []Operand
	for i, token := range postfix {
		if IsNumber(token) {
			val, err := strconv.ParseFloat(token, 64)
			if err != nil {
				return uuid.Nil, fmt.Errorf("invalid expresseion")
			}
			taskStack = append(taskStack, Operand{Value: val})
		} else if isOperator(token) {
			right := taskStack[len(taskStack)-1]
			left := taskStack[len(taskStack)-2]
			taskStack = taskStack[:len(taskStack)-2]
			isFinalTask := (i == len(postfix)-1)
			task := Task{
				ID:            uuid.New(),
				ExpressionID:  exprID,
				Arg1:          left,
				Arg2:          right,
				Operator:      token,
				OperationTime: time.Time{},
				FinalTask:     isFinalTask,
			}
			tm.Tasks = append(tm.Tasks, task)
			taskStack = append(taskStack, Operand{Value: math.NaN(), TaskID: task.ID})
		}
	}
	tm.Cond.Broadcast()
	return exprID, nil
}

func InfixToPostfix(expression string) []string {
	var stack []string
	var output []string
	for i := 0; i < len(expression); i++ {
		ch := string(expression[i])
		if IsNumber(ch) {
			num := ch
			for i+1 < len(expression) && (IsNumber(string(expression[i+1])) || string(expression[i+1]) == ".") {
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
			for len(stack) > 0 && Precedence(stack[len(stack)-1]) >= Precedence(ch) {
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

func (tm *TaskManager) GetOperationEndTime(operator string) time.Time {
	switch operator {
	case "+":
		return time.Now().Add(time.Duration(tm.Config.TimeAdditionMS) * time.Millisecond)
	case "-":
		return time.Now().Add(time.Duration(tm.Config.TimeSubtractionMS) * time.Millisecond)
	case "*":
		return time.Now().Add(time.Duration(tm.Config.TimeMultiplicationsMS) * time.Millisecond)
	case "/":
		return time.Now().Add(time.Duration(tm.Config.TimeDivisionsMS) * time.Millisecond)
	default:
		return time.Now()
	}
}

func Precedence(op string) int {
	switch op {
	case "+", "-":
		return 1
	case "*", "/":
		return 2
	}
	return 0
}

func IsNumber(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

func isOperator(s string) bool {
	return s == "+" || s == "-" || s == "*" || s == "/"
}

func (tm *TaskManager) ValidateExpression(expression string) error {
	expression = strings.ReplaceAll(expression, " ", "")
	if len(expression) == 0 {
		return ErrEmptyExpression
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

	if strings.Contains(expression, "/0") {
		return ErrDivisionByZero
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
			return ErrOperatorInsideParenthesis
		}
	}

	return nil
}
