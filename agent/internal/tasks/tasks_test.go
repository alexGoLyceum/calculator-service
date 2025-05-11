package tasks_test

import (
	"math"
	"testing"
	"time"

	"github.com/alexGoLyceum/calculator-service/agent/internal/tasks"

	"github.com/stretchr/testify/assert"
)

func TestCalculate_AllOperators(t *testing.T) {
	baseTime := time.Now()
	futureTime := baseTime.Add(10 * time.Millisecond)

	tests := []struct {
		name     string
		task     tasks.Task
		expected float64
	}{
		{
			name: "Addition",
			task: tasks.Task{
				Arg1:     tasks.Operand{Value: 2},
				Arg2:     tasks.Operand{Value: 3},
				Operator: "+",
			},
			expected: 5,
		},
		{
			name: "Subtraction",
			task: tasks.Task{
				Arg1:     tasks.Operand{Value: 5},
				Arg2:     tasks.Operand{Value: 2},
				Operator: "-",
			},
			expected: 3,
		},
		{
			name: "Multiplication",
			task: tasks.Task{
				Arg1:     tasks.Operand{Value: 3},
				Arg2:     tasks.Operand{Value: 4},
				Operator: "*",
			},
			expected: 12,
		},
		{
			name: "Division",
			task: tasks.Task{
				Arg1:     tasks.Operand{Value: 10},
				Arg2:     tasks.Operand{Value: 2},
				Operator: "/",
			},
			expected: 5,
		},
		{
			name: "Division by zero",
			task: tasks.Task{
				Arg1:     tasks.Operand{Value: 10},
				Arg2:     tasks.Operand{Value: 0},
				Operator: "/",
			},
			expected: math.NaN(),
		},
		{
			name: "Unknown operator",
			task: tasks.Task{
				Arg1:     tasks.Operand{Value: 10},
				Arg2:     tasks.Operand{Value: 2},
				Operator: "^",
			},
			expected: math.NaN(),
		},
		{
			name: "Future operation time sleeps",
			task: tasks.Task{
				Arg1:          tasks.Operand{Value: 1},
				Arg2:          tasks.Operand{Value: 2},
				Operator:      "+",
				OperationTime: futureTime,
			},
			expected: 3,
		},
		{
			name: "Past operation time does not sleep",
			task: tasks.Task{
				Arg1:          tasks.Operand{Value: 1},
				Arg2:          tasks.Operand{Value: 2},
				Operator:      "+",
				OperationTime: baseTime.Add(-1 * time.Minute),
			},
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start := time.Now()
			result := tasks.Calculate(&tt.task)
			duration := time.Since(start)

			if math.IsNaN(tt.expected) {
				assert.True(t, math.IsNaN(result), "expected NaN")
			} else {
				assert.Equal(t, tt.expected, result)
			}

			if tt.task.OperationTime.After(time.Now()) {
				assert.GreaterOrEqual(t, duration, 10*time.Millisecond, "should sleep if future time")
			}
		})
	}
}
