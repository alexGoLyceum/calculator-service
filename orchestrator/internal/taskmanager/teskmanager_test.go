package taskmanager_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/alexGoLyceum/calculator-service/orchestrator/internal/config"
	"github.com/alexGoLyceum/calculator-service/orchestrator/internal/taskmanager"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestValidateExpression(t *testing.T) {
	tm := &taskmanager.TaskManager{}

	tests := []struct {
		expression string
		hasError   bool
	}{
		{"", true},
		{"123", true},
		{"1+2", false},
		{"1++2", true},
		{"1+2-", true},
		{"+1+2", true},
		{"(1+2", true},
		{"1+2)", true},
		{"(1+2))", true},
		{"((1+2)", true},
		{"(1+2)-(*3)", true},
		{"5+()", true},
		{"5+(-)", true},
		{"(+)5", true},
		{"10/0", true},
		{"3.14.15+2", true},
		{".5+1", true},
		{"5.+1", true},
		{"012+5", true},
		{"", true},
		{"+", true},
		{"----", true},
		{"324.54-", true},
		{"*.234", true},
		{"413+435-234*", true},
		{"", true},
		{"32456", true},
		{".324", true},
		{"...2435", true},
		{"324.2345.2345", true},
		{"2345.4325......345", true},
		{"234.......34536", true},
		{".2343-2134", true},
		{"324.*234", true},
		{"345-2345.2345+2345./234", true},
		{"23425.234324*.324-324", true},
		{"6+(83247-234)", false},
		{"5+((6-)*2)*3", true},
		{"()()()()", true},
		{")()()(", true},
		{".(32-234)+23", true},
		{"(234-34).+324", true},
		{"(324-34).(324-2432)", true},
		{"5+()-1234", true},
		{"()", true},
		{"(324)", true},
		{"(435.345)", true},
		{"34095-2384*02384+32948.345", true},
		{")()()()(+213", true},
		{"word", true},
		{"34.f43r-324", true},
		{"34/0", true},
		{"2345-5.+1", true},
		{"345.345-2134.*3245", true},
		{"((((53454-3245))))", false},
		{"5+(-5.345)", true},
		{"132(23-234)", true},
		{"132(23-234+2134", true},
		{"234+(23-321)(234+2331)", true},
		{"1++2", true},
		{"1--2", true},
		{"1+-2", true},
		{"1**2", true},
		{"1//2", true},
		{"1+--2", true},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("expression: %s", test.expression), func(t *testing.T) {
			err := tm.ValidateExpression(test.expression)
			if test.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func newTestTaskManager() *taskmanager.TaskManager {
	cfg := &config.TaskManagerConfig{
		TimeAdditionMS:        10,
		TimeSubtractionMS:     10,
		TimeMultiplicationsMS: 10,
		TimeDivisionsMS:       10,
	}
	return taskmanager.NewTaskManager(cfg)
}

func TestInfixToPostfix(t *testing.T) {
	result := taskmanager.InfixToPostfix("3+5*2")
	expected := []string{"3", "5", "2", "*", "+"}

	if len(result) != len(expected) {
		t.Fatalf("expected length %d, got %d", len(expected), len(result))
	}

	for i := range result {
		if result[i] != expected[i] {
			t.Fatalf("expected %q at index %d, got %q", expected[i], i, result[i])
		}
	}
}

func TestPrecedence(t *testing.T) {
	if taskmanager.Precedence("+") != 1 {
		t.Fatal("expected Precedence of '+' to be 1")
	}
	if taskmanager.Precedence("*") != 2 {
		t.Fatal("expected Precedence of '*' to be 2")
	}
	if taskmanager.Precedence("") != 0 {
		t.Fatalf("expected Precedence to be 0")
	}
}

func TestIsNumber(t *testing.T) {
	if !taskmanager.IsNumber("3.14") {
		t.Fatal("expected '3.14' to be a number")
	}
	if taskmanager.IsNumber("abc") {
		t.Fatal("expected 'abc' not to be a number")
	}
}

func TestTaskManager_CreateAndStoreTasksFromExpression(t *testing.T) {
	tm := newTestTaskManager()

	exprID, err := tm.CreateAndStoreTasksFromExpression("3+5*2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expr, err := tm.GetExpressionByID(exprID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if expr.Status != taskmanager.Pending {
		t.Fatalf("expected status %q, got %q", taskmanager.Pending, expr.Status)
	}
}

func TestTaskManager_GetExpressions(t *testing.T) {
	tm := newTestTaskManager()

	expr1ID := uuid.New()
	expr2ID := uuid.New()
	tm.Expressions.Store(expr1ID, &taskmanager.Expression{ID: expr1ID, Status: taskmanager.Pending})
	tm.Expressions.Store(expr2ID, &taskmanager.Expression{ID: expr2ID, Status: taskmanager.Completed})

	expressions := tm.GetExpressions()

	if len(expressions) != 2 {
		t.Fatalf("expected 2 expressions, got %d", len(expressions))
	}

	statuses := map[taskmanager.Status]bool{}
	for _, expr := range expressions {
		statuses[expr.Status] = true
	}

	if !statuses[taskmanager.Pending] {
		t.Fatalf("expected status %q, but it was not found", taskmanager.Pending)
	}

	if !statuses[taskmanager.Completed] {
		t.Fatalf("expected status %q, but it was not found", taskmanager.Completed)
	}
}

func TestTaskManager_GetTask(t *testing.T) {
	tm := newTestTaskManager()

	_, err := tm.CreateAndStoreTasksFromExpression("3+5*2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tm.TaskResults.Store(tm.Tasks[0].ID, 3.0)
	tm.TaskResults.Store(tm.Tasks[1].ID, 5.0)

	task := tm.GetTask()
	if task == nil {
		t.Fatal("expected a task to be returned, but got nil")
	}

	expr, err := tm.GetExpressionByID(task.ExpressionID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if expr.Status != taskmanager.InProgress {
		t.Fatalf("expected status %q, got %q", taskmanager.InProgress, expr.Status)
	}
}

func TestTaskManager_GetOperationEndTime(t *testing.T) {
	tm := newTestTaskManager()

	addEndTime := tm.GetOperationEndTime("+")
	subEndTime := tm.GetOperationEndTime("-")
	mulEndTime := tm.GetOperationEndTime("*")
	divEndTime := tm.GetOperationEndTime("/")
	unknownEndTime := tm.GetOperationEndTime("?")

	if addEndTime.Equal(time.Now()) {
		t.Fatalf("expected non-zero time for addition operation")
	}
	if subEndTime.Equal(time.Now()) {
		t.Fatalf("expected non-zero time for subtraction operation")
	}
	if mulEndTime.Equal(time.Now()) {
		t.Fatalf("expected non-zero time for multiplication operation")
	}
	if divEndTime.Equal(time.Now()) {
		t.Fatalf("expected non-zero time for division operation")
	}
	if unknownEndTime.Equal(time.Now()) {
		t.Fatalf("expected non-zero time for division operation")
	}
}

func TestCheckExpiredTasks(t *testing.T) {
	cfg := &config.TaskManagerConfig{
		TimeAdditionMS:        100,
		TimeSubtractionMS:     100,
		TimeMultiplicationsMS: 100,
		TimeDivisionsMS:       100,
	}

	tm := &taskmanager.TaskManager{
		Config: cfg,
		Cond:   sync.NewCond(&sync.Mutex{}),
	}

	taskID := uuid.New()
	exprID := uuid.New()
	task := taskmanager.Task{
		ID:            taskID,
		ExpressionID:  exprID,
		Arg1:          taskmanager.Operand{Value: 5},
		Arg2:          taskmanager.Operand{Value: 3},
		Operator:      "+",
		OperationTime: time.Now().Add(-15 * time.Second),
		FinalTask:     true,
	}

	tm.InProcessTasks = append(tm.InProcessTasks, task)

	go tm.CheckExpiredTasks()

	time.Sleep(6 * time.Second)

	if len(tm.InProcessTasks) != 0 {
		t.Errorf("expected InProcessTasks to be empty, but got %d tasks", len(tm.InProcessTasks))
	}

	if len(tm.Tasks) == 0 {
		t.Errorf("expected task to be moved to Tasks, but no tasks found")
	}
}

func TestCheckNoExpiredTasks(t *testing.T) {
	cfg := &config.TaskManagerConfig{
		TimeAdditionMS:        100,
		TimeSubtractionMS:     100,
		TimeMultiplicationsMS: 100,
		TimeDivisionsMS:       100,
	}

	tm := &taskmanager.TaskManager{
		Config:      cfg,
		Cond:        sync.NewCond(&sync.Mutex{}),
		Expressions: sync.Map{},
		TaskResults: sync.Map{},
		Tasks:       []taskmanager.Task{},
		InProcessTasks: []taskmanager.Task{
			{
				ID:            uuid.New(),
				ExpressionID:  uuid.New(),
				Arg1:          taskmanager.Operand{Value: 10},
				Arg2:          taskmanager.Operand{Value: 5},
				Operator:      "+",
				OperationTime: time.Now().Add(-5 * time.Second),
				FinalTask:     true,
			},
			{
				ID:            uuid.New(),
				ExpressionID:  uuid.New(),
				Arg1:          taskmanager.Operand{Value: 5},
				Arg2:          taskmanager.Operand{Value: 3},
				Operator:      "+",
				OperationTime: time.Now().Add(5 * time.Second),
				FinalTask:     true,
			},
		},
	}

	go tm.CheckExpiredTasks()

	time.Sleep(6 * time.Second)

	assert.Len(t, tm.InProcessTasks, 1)

	assert.Len(t, tm.Tasks, 1)
}

func TestCheckEmptyInProcessTasks(t *testing.T) {
	cfg := &config.TaskManagerConfig{
		TimeAdditionMS:        100,
		TimeSubtractionMS:     100,
		TimeMultiplicationsMS: 100,
		TimeDivisionsMS:       100,
	}

	tm := &taskmanager.TaskManager{
		Config:         cfg,
		Cond:           sync.NewCond(&sync.Mutex{}),
		Expressions:    sync.Map{},
		TaskResults:    sync.Map{},
		Tasks:          []taskmanager.Task{},
		InProcessTasks: []taskmanager.Task{},
	}

	go tm.CheckExpiredTasks()

	time.Sleep(6 * time.Second)

	if len(tm.InProcessTasks) != 0 {
		t.Errorf("expected InProcessTasks to be empty, but got %d tasks", len(tm.InProcessTasks))
	}

	if len(tm.Tasks) != 0 {
		t.Errorf("expected no tasks to be moved to Tasks, but found %d tasks", len(tm.Tasks))
	}
}

func TestSetResult_TaskNotFound(t *testing.T) {
	tm := &taskmanager.TaskManager{
		Cond: sync.NewCond(&sync.Mutex{}),
	}

	taskID := uuid.New()
	expressionID := uuid.New()

	task := taskmanager.Task{
		ID:            taskID,
		ExpressionID:  expressionID,
		Arg1:          taskmanager.Operand{Value: 1},
		Arg2:          taskmanager.Operand{Value: 2},
		Operator:      "+",
		OperationTime: time.Now(),
		FinalTask:     false,
	}

	err := tm.SetResult(task, 3)

	expectedError := fmt.Sprintf("task with ID %s not found", taskID)
	assert.EqualError(t, err, expectedError)
}

func TestSetResult_ExpressionNotFound(t *testing.T) {
	tm := &taskmanager.TaskManager{
		Cond: sync.NewCond(&sync.Mutex{}),
	}

	taskID := uuid.New()
	expressionID := uuid.New()

	task := taskmanager.Task{
		ID:            taskID,
		ExpressionID:  expressionID,
		Arg1:          taskmanager.Operand{Value: 1},
		Arg2:          taskmanager.Operand{Value: 2},
		Operator:      "+",
		OperationTime: time.Now(),
		FinalTask:     false,
	}

	tm.InProcessTasks = append(tm.InProcessTasks, task)

	err := tm.SetResult(task, 3)

	expectedError := fmt.Sprintf("expression with id %s not found", expressionID)
	assert.EqualError(t, err, expectedError)
}

func TestSetResult_SuccessFinalTask(t *testing.T) {
	tm := &taskmanager.TaskManager{
		Cond: sync.NewCond(&sync.Mutex{}),
	}

	taskID := uuid.New()
	expressionID := uuid.New()

	expression := &taskmanager.Expression{
		ID:     expressionID,
		Status: taskmanager.Pending,
	}

	tm.Expressions.Store(expressionID, expression)

	task := taskmanager.Task{
		ID:            taskID,
		ExpressionID:  expressionID,
		Arg1:          taskmanager.Operand{Value: 1},
		Arg2:          taskmanager.Operand{Value: 2},
		Operator:      "+",
		OperationTime: time.Now(),
		FinalTask:     true,
	}

	tm.InProcessTasks = append(tm.InProcessTasks, task)

	err := tm.SetResult(task, 3)

	assert.NoError(t, err)

	assert.Equal(t, taskmanager.Completed, expression.Status)

	assert.Equal(t, 3.0, expression.Result)
}

func TestSetResult_SuccessNonFinalTask(t *testing.T) {
	tm := &taskmanager.TaskManager{
		Cond: sync.NewCond(&sync.Mutex{}),
	}

	taskID := uuid.New()
	expressionID := uuid.New()

	expression := &taskmanager.Expression{
		ID:     expressionID,
		Status: taskmanager.Pending,
	}

	tm.Expressions.Store(expressionID, expression)

	task := taskmanager.Task{
		ID:            taskID,
		ExpressionID:  expressionID,
		Arg1:          taskmanager.Operand{Value: 1},
		Arg2:          taskmanager.Operand{Value: 2},
		Operator:      "+",
		OperationTime: time.Now(),
		FinalTask:     false,
	}

	tm.InProcessTasks = append(tm.InProcessTasks, task)

	err := tm.SetResult(task, 3)

	assert.NoError(t, err)

	result, ok := tm.TaskResults.Load(taskID)
	assert.True(t, ok)
	assert.Equal(t, 3.0, result)
}
