package models

import (
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	Pending    Status = "pending"
	InProgress Status = "in progress"
	Done       Status = "done"
)

type Expression struct {
	ID         uuid.UUID `json:"id"`
	UserID     uuid.UUID `json:"user_id,omitempty"`
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
	Value  float64    `json:"value"`
	TaskID *uuid.UUID `json:"task_id"`
}
