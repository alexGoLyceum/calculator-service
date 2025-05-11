package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CommandTag interface {
	String() string
	RowsAffected() int64
}

type PgxCommandTag struct {
	Tag pgconn.CommandTag
}

func (t PgxCommandTag) String() string {
	return t.Tag.String()
}

func (t PgxCommandTag) RowsAffected() int64 {
	return t.Tag.RowsAffected()
}

type Row interface {
	Scan(dest ...interface{}) error
}

type PgxRow struct {
	Row pgx.Row
}

func (r PgxRow) Scan(dest ...interface{}) error {
	return r.Row.Scan(dest...)
}

type Rows interface {
	Close()
	Err() error
	Next() bool
	Scan(dest ...interface{}) error
	CommandTag() pgconn.CommandTag
	FieldDescriptions() []pgconn.FieldDescription
	Values() ([]any, error)
	RawValues() [][]byte
	Conn() *pgx.Conn
}

type PgxRows struct {
	Rows pgx.Rows
}

func (r PgxRows) Close() {
	r.Rows.Close()
}

func (r PgxRows) Err() error {
	return r.Rows.Err()
}

func (r PgxRows) Next() bool {
	return r.Rows.Next()
}

func (r PgxRows) Scan(dest ...interface{}) error {
	return r.Rows.Scan(dest...)
}

func (r PgxRows) CommandTag() pgconn.CommandTag {
	return r.Rows.CommandTag()
}

func (r PgxRows) FieldDescriptions() []pgconn.FieldDescription {
	return r.Rows.FieldDescriptions()
}

func (r PgxRows) Values() ([]any, error) {
	return r.Rows.Values()
}

func (r PgxRows) RawValues() [][]byte {
	return r.Rows.RawValues()
}

func (r PgxRows) Conn() *pgx.Conn {
	return r.Rows.Conn()
}

type Tx interface {
	Exec(ctx context.Context, query string, args ...interface{}) (CommandTag, error)
	Query(ctx context.Context, query string, args ...interface{}) (Rows, error)
	QueryRow(ctx context.Context, query string, args ...interface{}) Row
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
	Begin(ctx context.Context) (pgx.Tx, error)
	CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error)
	SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults
	LargeObjects() pgx.LargeObjects
	Prepare(ctx context.Context, name, sql string) (*pgconn.StatementDescription, error)
	Conn() *pgx.Conn
}

type PgxTx struct {
	Tx pgx.Tx
}

func (t *PgxTx) Exec(ctx context.Context, query string, args ...interface{}) (CommandTag, error) {
	tag, err := t.Tx.Exec(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return PgxCommandTag{Tag: tag}, nil
}

func (t *PgxTx) Query(ctx context.Context, query string, args ...interface{}) (Rows, error) {
	rows, err := t.Tx.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return PgxRows{Rows: rows}, nil
}

func (t *PgxTx) QueryRow(ctx context.Context, query string, args ...interface{}) Row {
	return PgxRow{Row: t.Tx.QueryRow(ctx, query, args...)}
}

func (t *PgxTx) Commit(ctx context.Context) error {
	return t.Tx.Commit(ctx)
}

func (t *PgxTx) Rollback(ctx context.Context) error {
	return t.Tx.Rollback(ctx)
}

func (t *PgxTx) Begin(ctx context.Context) (pgx.Tx, error) {
	return t.Tx.Begin(ctx)
}

func (t *PgxTx) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	return t.Tx.CopyFrom(ctx, tableName, columnNames, rowSrc)
}

func (t *PgxTx) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults {
	return t.Tx.SendBatch(ctx, b)
}

func (t *PgxTx) LargeObjects() pgx.LargeObjects {
	return t.Tx.LargeObjects()
}

func (t *PgxTx) Prepare(ctx context.Context, name, sql string) (*pgconn.StatementDescription, error) {
	return t.Tx.Prepare(ctx, name, sql)
}

func (t *PgxTx) Conn() *pgx.Conn {
	return t.Tx.Conn()
}

type Pool interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Begin(ctx context.Context) (Tx, error)
	Close()
	Ping(ctx context.Context) error
}

type PgxPoolWrapper struct {
	Pool *pgxpool.Pool
}

func (w *PgxPoolWrapper) Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error) {
	return w.Pool.Exec(ctx, sql, arguments...)
}

func (w *PgxPoolWrapper) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	return w.Pool.Query(ctx, sql, args...)
}

func (w *PgxPoolWrapper) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	return w.Pool.QueryRow(ctx, sql, args...)
}

func (w *PgxPoolWrapper) Begin(ctx context.Context) (Tx, error) {
	tx, err := w.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	return &PgxTx{Tx: tx}, nil
}

func (w *PgxPoolWrapper) Close() {
	w.Pool.Close()
}

func (w *PgxPoolWrapper) Ping(ctx context.Context) error {
	return w.Pool.Ping(ctx)
}

type DatabaseConnection interface {
	Connect(ctx context.Context) error
	Exec(ctx context.Context, query string, args ...interface{}) (CommandTag, error)
	Query(ctx context.Context, query string, args ...interface{}) (Rows, error)
	QueryRow(ctx context.Context, query string, args ...interface{}) Row
	BeginTx(ctx context.Context) (Tx, error)
	Close() error
	StartMonitoring(ctx context.Context, interval time.Duration)
	DBErrorChecker
}

type DBErrorChecker interface {
	IsUniqueViolationErr(err error) bool
	IsNoRowsErr(err error) bool
	IsForeignKeyErr(err error) bool
	IsDatabaseUnavailableErr(err error) bool
}

type Connection struct {
	ConnStr string
	Pool    Pool
}

type Config struct {
	Host     string
	Port     string
	Username string
	Password string
	Database string
}

func NewPostgresConnection(cfg *Config) *Connection {
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Database)

	return &Connection{
		ConnStr: connStr,
	}
}

func (c *Connection) Connect(ctx context.Context) error {
	config, err := pgxpool.ParseConfig(c.ConnStr)
	if err != nil {
		return fmt.Errorf("failed to parse config: %v", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}

	c.Pool = &PgxPoolWrapper{Pool: pool}

	if err := c.Pool.Ping(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %v", err)
	}
	return nil
}

func (c *Connection) Exec(ctx context.Context, query string, args ...interface{}) (CommandTag, error) {
	tag, err := c.Pool.Exec(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return PgxCommandTag{Tag: tag}, nil
}

func (c *Connection) Query(ctx context.Context, query string, args ...interface{}) (Rows, error) {
	rows, err := c.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return PgxRows{Rows: rows}, nil
}

func (c *Connection) QueryRow(ctx context.Context, query string, args ...interface{}) Row {
	return PgxRow{Row: c.Pool.QueryRow(ctx, query, args...)}
}

func (c *Connection) BeginTx(ctx context.Context) (Tx, error) {
	tx, err := c.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func (c *Connection) Close() error {
	if c.Pool != nil {
		c.Pool.Close()
	}
	return nil
}

func (c *Connection) StartMonitoring(ctx context.Context, interval time.Duration) {
	go c.monitorConnection(ctx, interval)
}

func (c *Connection) monitorConnection(ctx context.Context, reconnectInterval time.Duration) {
	ticker := time.NewTicker(reconnectInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := c.Pool.Ping(ctx); err != nil {
				_ = c.Connect(ctx)
			}
		}
	}
}

func (c *Connection) IsUniqueViolationErr(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}

func (c *Connection) IsNoRowsErr(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}

func (c *Connection) IsForeignKeyErr(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23503"
	}
	return false
}

func (c *Connection) IsDatabaseUnavailableErr(err error) bool {
	if err == nil {
		return false
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "08000", "08003", "08006", "08001", "08004", "57P01", "57P02", "57P03":
			return true
		}
	}

	errMsg := err.Error()
	if strings.Contains(errMsg, "connection refused") ||
		strings.Contains(errMsg, "no connection") ||
		strings.Contains(errMsg, "connection reset") ||
		strings.Contains(errMsg, "broken pipe") ||
		strings.Contains(errMsg, "unable to connect") ||
		strings.Contains(errMsg, "no such host") ||
		strings.Contains(errMsg, "temporary failure in name resolution") ||
		strings.Contains(errMsg, "lookup") {
		return true
	}

	return false
}
