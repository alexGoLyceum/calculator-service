package tasks

import (
	"math"
	"time"

	"github.com/google/uuid"
)

type Status string

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

func Calculate(task *Task) float64 {
	if !task.OperationTime.IsZero() {
		duration := task.OperationTime.Sub(time.Now())
		if duration > 0 {
			time.Sleep(duration)
		}
	}

	switch task.Operator {
	case "+":
		return task.Arg1.Value + task.Arg2.Value
	case "-":
		return task.Arg1.Value - task.Arg2.Value
	case "*":
		return task.Arg1.Value * task.Arg2.Value
	case "/":
		if task.Arg2.Value != 0 {
			return task.Arg1.Value / task.Arg2.Value
		}
	}
	return math.NaN()
}
