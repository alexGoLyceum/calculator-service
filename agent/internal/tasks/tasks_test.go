package tasks_test

import (
	"math"
	"testing"
	"time"

	"github.com/alexGoLyceum/calculator-service/agent/internal/tasks"

	"github.com/stretchr/testify/assert"
)

func TestCalculate_Addition(t *testing.T) {
	task := &tasks.Task{
		Arg1:     tasks.Operand{Value: 3},
		Arg2:     tasks.Operand{Value: 5},
		Operator: "+",
	}

	result := tasks.Calculate(task)

	assert.Equal(t, 8.0, result, "Expected result to be 8 for addition")
}

func TestCalculate_Subtraction(t *testing.T) {
	task := &tasks.Task{
		Arg1:     tasks.Operand{Value: 10},
		Arg2:     tasks.Operand{Value: 4},
		Operator: "-",
	}

	result := tasks.Calculate(task)

	assert.Equal(t, 6.0, result, "Expected result to be 6 for subtraction")
}

func TestCalculate_Multiplication(t *testing.T) {
	task := &tasks.Task{
		Arg1:     tasks.Operand{Value: 3},
		Arg2:     tasks.Operand{Value: 5},
		Operator: "*",
	}

	result := tasks.Calculate(task)

	assert.Equal(t, 15.0, result, "Expected result to be 15 for multiplication")
}

func TestCalculate_Division(t *testing.T) {
	task := &tasks.Task{
		Arg1:     tasks.Operand{Value: 10},
		Arg2:     tasks.Operand{Value: 2},
		Operator: "/",
	}

	result := tasks.Calculate(task)

	assert.Equal(t, 5.0, result, "Expected result to be 5 for division")
}

func TestCalculate_DivisionByZero(t *testing.T) {
	task := &tasks.Task{
		Arg1:     tasks.Operand{Value: 10},
		Arg2:     tasks.Operand{Value: 0},
		Operator: "/",
	}

	result := tasks.Calculate(task)

	assert.True(t, math.IsNaN(result), "Expected result to be NaN when dividing by zero")
}

func TestCalculate_NaNForUnknownOperator(t *testing.T) {
	task := &tasks.Task{
		Arg1:     tasks.Operand{Value: 5},
		Arg2:     tasks.Operand{Value: 3},
		Operator: "?",
	}

	result := tasks.Calculate(task)

	assert.True(t, math.IsNaN(result), "Expected result to be NaN for an unknown operator")
}

func TestCalculate_WithFutureOperationTime(t *testing.T) {
	task := &tasks.Task{
		Arg1:          tasks.Operand{Value: 3},
		Arg2:          tasks.Operand{Value: 5},
		Operator:      "+",
		OperationTime: time.Now().Add(2 * time.Second),
	}

	done := make(chan bool)

	go func() {
		tasks.Calculate(task)
		done <- true
	}()

	select {
	case <-done:
		result := tasks.Calculate(task)
		assert.Equal(t, 8.0, result, "Expected result to be 8 for addition")
	case <-time.After(3 * time.Second):
		t.Fatal("Test timed out, expected calculation to finish after 2 seconds")
	}
}

func TestCalculate_WithPastOperationTime(t *testing.T) {
	task := &tasks.Task{
		Arg1:          tasks.Operand{Value: 3},
		Arg2:          tasks.Operand{Value: 5},
		Operator:      "+",
		OperationTime: time.Now().Add(-2 * time.Second),
	}

	result := tasks.Calculate(task)

	assert.Equal(t, 8.0, result, "Expected result to be 8 for addition")
}
