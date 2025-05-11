package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/alexGoLyceum/calculator-service/orchestrator/internal/database/postgres"
	"github.com/alexGoLyceum/calculator-service/orchestrator/internal/models"
	pb "github.com/alexGoLyceum/calculator-service/orchestrator/internal/transport/grpc/proto"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	ErrUserWithLoginAlreadyExists   = errors.New("user with this login already exists")
	ErrUserNotFoundByLogin          = errors.New("user with this login does not exist")
	ErrInvalidPassword              = errors.New("invalid password")
	ErrUnknownUserID                = errors.New("unknown user id")
	ErrUnknownExpressionID          = errors.New("unknown expressions id")
	ErrUnknownTaskID                = errors.New("unknown task id")
	ErrDatabaseNotAvailable         = errors.New("database is not available")
	ErrInvalidTask                  = errors.New("invalid task")
	ErrUnknownIDTasksWithDependency = errors.New("unknown ID tasks with the dependency")
)

type Repository interface {
	CreateUser(ctx context.Context, login string, password string) (uuid.UUID, error)
	AuthUser(ctx context.Context, login, password string) (uuid.UUID, error)
	GetAllExpressions(ctx context.Context, userID uuid.UUID) ([]*models.Expression, error)
	GetExpressionByID(ctx context.Context, id uuid.UUID) (*models.Expression, error)
	CreateExpressionTask(ctx context.Context, expression *models.Expression, tasks []*models.Task) error
	GetTask(ctx context.Context, getEndTime func(string) *timestamppb.Timestamp) (*pb.Task, error)
	SetTaskResult(ctx context.Context, task *pb.Task, result float64) error
	WithTransaction(ctx context.Context, fn func(ctx context.Context, tx postgres.Tx) error) error
	ResetExpiredTasks(ctx context.Context, delay time.Duration) error
}

type repository struct {
	db postgres.DatabaseConnection
}

func NewRepositoryImpl(db postgres.DatabaseConnection) Repository {
	return &repository{
		db: db,
	}
}

func (r *repository) CreateUser(ctx context.Context, login string, password string) (uuid.UUID, error) {
	hashedPassword, err := HashPassword(password)
	if err != nil {
		return uuid.Nil, err
	}

	var userId uuid.UUID
	if err = r.db.QueryRow(ctx,
		"INSERT INTO users (login, password_hash) VALUES ($1, $2) RETURNING id",
		login, hashedPassword).Scan(&userId); err != nil {
		if r.db.IsUniqueViolationErr(err) {
			return uuid.Nil, ErrUserWithLoginAlreadyExists
		}
		if r.db.IsDatabaseUnavailableErr(err) {
			return uuid.Nil, ErrDatabaseNotAvailable
		}
		return uuid.Nil, fmt.Errorf("failed to insert new user: %w", err)
	}
	return userId, nil
}

func (r *repository) AuthUser(ctx context.Context, login, password string) (uuid.UUID, error) {
	var userID uuid.UUID
	var hashedPassword string

	row := r.db.QueryRow(ctx, "SELECT id, password_hash FROM users WHERE login=$1", login)
	if err := row.Scan(&userID, &hashedPassword); err != nil {
		if r.db.IsNoRowsErr(err) {
			return uuid.Nil, ErrUserNotFoundByLogin
		}
		if r.db.IsDatabaseUnavailableErr(err) {
			return uuid.Nil, ErrDatabaseNotAvailable
		}
		return uuid.Nil, fmt.Errorf("failed to read user: %w", err)
	}

	if err := comparePasswords(hashedPassword, password); err != nil {
		return uuid.Nil, ErrInvalidPassword
	}
	return userID, nil
}

func (r *repository) CreateExpressionTask(ctx context.Context, expression *models.Expression, tasks []*models.Task) error {
	return r.WithTransaction(ctx, func(ctx context.Context, tx postgres.Tx) error {
		row := tx.QueryRow(ctx,
			`INSERT INTO expressions (id, user_id, expression, status, result) 
VALUES ($1, $2, $3, $4, $5) RETURNING id`,
			expression.ID, expression.UserID, expression.Expression, expression.Status, expression.Result)
		if err := row.Scan(&expression.ID); err != nil {
			if r.db.IsForeignKeyErr(err) {
				return ErrUnknownUserID
			}
			if r.db.IsDatabaseUnavailableErr(err) {
				return ErrDatabaseNotAvailable
			}
			return fmt.Errorf("failed to insert new expression: %w", err)
		}

		query := `INSERT INTO tasks (id, expression_id, arg1_value, arg1_task_id, arg2_value, arg2_task_id, operator, operation_time, final_task)
				  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
		for _, task := range tasks {
			if _, err := tx.Exec(ctx, query, task.ID, task.ExpressionID, task.Arg1.Value, task.Arg1.TaskID,
				task.Arg2.Value, task.Arg2.TaskID, task.Operator, task.OperationTime, task.FinalTask); err != nil {
				if r.db.IsDatabaseUnavailableErr(err) {
					return ErrDatabaseNotAvailable
				}
				return fmt.Errorf("failed to insert new task: %w", err)
			}
		}
		return nil
	})
}

func (r *repository) GetAllExpressions(ctx context.Context, userID uuid.UUID) ([]*models.Expression, error) {
	var userExists bool
	if err := r.db.QueryRow(
		ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", userID).Scan(&userExists); err != nil {
		if r.db.IsDatabaseUnavailableErr(err) {
			return nil, ErrDatabaseNotAvailable
		}
		return nil, fmt.Errorf("failed to check user existence: %v", err)
	}

	if !userExists {
		return nil, ErrUnknownUserID
	}

	rows, err := r.db.Query(ctx, "SELECT * FROM expressions WHERE user_id = $1", userID)
	if err != nil {
		if r.db.IsDatabaseUnavailableErr(err) {
			return nil, ErrDatabaseNotAvailable
		}
		return nil, fmt.Errorf("failed to execute query: %v", err)
	}
	defer rows.Close()

	expressions := make([]*models.Expression, 0)
	for rows.Next() {
		var expression models.Expression
		if err := rows.Scan(&expression.ID, &expression.UserID, &expression.Expression, &expression.Status, &expression.Result); err != nil {
			if r.db.IsDatabaseUnavailableErr(err) {
				return nil, ErrDatabaseNotAvailable
			}
			return nil, fmt.Errorf("failed to scan row: %v", err)
		}
		expressions = append(expressions, &expression)
	}

	if err := rows.Err(); err != nil {
		if r.db.IsDatabaseUnavailableErr(err) {
			return nil, ErrDatabaseNotAvailable
		}
		return nil, fmt.Errorf("rows iteration error: %v", err)
	}
	return expressions, nil
}

func (r *repository) GetExpressionByID(ctx context.Context, expressionID uuid.UUID) (*models.Expression, error) {
	var expression models.Expression
	row := r.db.QueryRow(ctx,
		"SELECT id, user_id, expression, status, result FROM expressions WHERE id = $1", expressionID)
	if err := row.Scan(&expression.ID, &expression.UserID, &expression.Expression, &expression.Status, &expression.Result); err != nil {
		if r.db.IsNoRowsErr(err) {
			return nil, ErrUnknownExpressionID
		}
		if r.db.IsDatabaseUnavailableErr(err) {
			return nil, ErrDatabaseNotAvailable
		}
		return &models.Expression{}, fmt.Errorf("failed to execute query: %v", err)
	}
	return &expression, nil
}

func (r *repository) GetTask(ctx context.Context, getEndTime func(string) *timestamppb.Timestamp) (*pb.Task, error) {
	var resultTask *pb.Task
	err := r.WithTransaction(ctx, func(ctx context.Context, tx postgres.Tx) error {
		query := `
			SELECT id, expression_id, arg1_value, arg2_value, operator, final_task
			FROM tasks
			WHERE status = 'pending'
				AND arg1_task_id IS NULL
				AND arg2_task_id IS NULL
			ORDER BY created_at
			FOR UPDATE SKIP LOCKED
			LIMIT 1
		`

		var (
			id, expressionID     uuid.UUID
			arg1Value, arg2Value sql.NullFloat64
			operator             string
			finalTask            bool
		)

		if err := tx.QueryRow(ctx, query).Scan(
			&id, &expressionID, &arg1Value, &arg2Value, &operator, &finalTask,
		); err != nil {
			if r.db.IsNoRowsErr(err) {
				return nil
			}
			if r.db.IsDatabaseUnavailableErr(err) {
				return ErrDatabaseNotAvailable
			}
			return fmt.Errorf("failed to select task: %w", err)
		}

		if _, err := tx.Exec(ctx, `
			UPDATE expressions
			SET status = 'in progress'
			WHERE id = $1 AND status != 'in progress'`, expressionID); err != nil {
			if r.db.IsDatabaseUnavailableErr(err) {
				return ErrDatabaseNotAvailable
			}
			return fmt.Errorf("failed to update expression status: %w", err)
		}

		endTimeProto := getEndTime(operator)
		if endTimeProto == nil {
			return ErrInvalidTask
		}

		endTime := endTimeProto.AsTime()
		if _, err := tx.Exec(ctx,
			`UPDATE tasks SET status = $1, operation_time = $2 WHERE id = $3`,
			models.InProgress, endTime, id,
		); err != nil {
			if r.db.IsDatabaseUnavailableErr(err) {
				return ErrDatabaseNotAvailable
			}
			return fmt.Errorf("failed to update task: %w", err)
		}

		resultTask = &pb.Task{
			Id:            id.String(),
			ExpressionId:  expressionID.String(),
			Arg1Num:       arg1Value.Float64,
			Arg2Num:       arg2Value.Float64,
			Operator:      operator,
			OperationTime: endTimeProto,
			FinalTask:     finalTask,
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return resultTask, nil
}

func (r *repository) SetTaskResult(ctx context.Context, task *pb.Task, result float64) error {
	return r.WithTransaction(ctx, func(ctx context.Context, tx postgres.Tx) error {
		if task.FinalTask {
			if res, err := tx.Exec(ctx, `DELETE FROM tasks WHERE id = $1`, task.Id); err != nil {
				if r.db.IsDatabaseUnavailableErr(err) {
					return ErrDatabaseNotAvailable
				}
				if res.RowsAffected() == 0 {
					return ErrUnknownTaskID
				}
				return err
			}

			if res, err := tx.Exec(ctx, `
			UPDATE expressions
			SET result = $1, status = $2
			WHERE id = $3
		`, result, models.Done, task.ExpressionId); err != nil {
				if r.db.IsDatabaseUnavailableErr(err) {
					return ErrDatabaseNotAvailable
				}
				if res.RowsAffected() == 0 {
					return ErrUnknownExpressionID
				}
				return err
			}
		} else {
			if res, err := tx.Exec(ctx, `
			UPDATE tasks
			SET arg1_value = CASE WHEN arg1_task_id = $1 THEN $2 ELSE arg1_value END,
			    arg1_task_id = CASE WHEN arg1_task_id = $1 THEN NULL ELSE arg1_task_id END,
			    arg2_value = CASE WHEN arg2_task_id = $1 THEN $2 ELSE arg2_value END,
			    arg2_task_id = CASE WHEN arg2_task_id = $1 THEN NULL ELSE arg2_task_id END
			WHERE arg1_task_id = $1 OR arg2_task_id = $1
		`, task.Id, result); err != nil {
				if r.db.IsDatabaseUnavailableErr(err) {
					return ErrDatabaseNotAvailable
				}
				if res.RowsAffected() == 0 {
					return ErrUnknownIDTasksWithDependency
				}
				return err
			}

			if res, err := tx.Exec(ctx, `DELETE FROM tasks WHERE id = $1`, task.Id); err != nil {
				if r.db.IsDatabaseUnavailableErr(err) {
					return ErrDatabaseNotAvailable
				}
				if res.RowsAffected() == 0 {
					return ErrUnknownTaskID
				}
				return err
			}
		}
		return nil
	})
}

func (r *repository) ResetExpiredTasks(ctx context.Context, delay time.Duration) error {
	query := `
        UPDATE tasks
        SET status = 'pending',
            result = NULL
        WHERE status != 'pending'
          AND operation_time + $1 < now()
    `
	_, err := r.db.Exec(ctx, query, delay)
	if err != nil {
		return fmt.Errorf("failed to reset expired tasks: %w", err)
	}

	return nil
}

func (r *repository) WithTransaction(ctx context.Context, fn func(ctx context.Context, tx postgres.Tx) error) error {
	tx, err := r.db.BeginTx(ctx)
	if err != nil {
		if r.db.IsDatabaseUnavailableErr(err) {
			return ErrDatabaseNotAvailable
		}
		return err
	}

	defer func() {
		if pErr := recover(); pErr != nil {
			_ = tx.Rollback(ctx)
			panic(pErr)
		} else if err != nil {
			_ = tx.Rollback(ctx)
		} else {
			err = tx.Commit(ctx)
		}
	}()

	return fn(ctx, tx)
}

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

func comparePasswords(hashedPassword, password string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
		return ErrInvalidPassword
	}
	return nil
}
