package dao

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/Terry-Mao/goim/internal/chatapi/conf"
)

// MySQL DAO
type MySQL struct {
	db *sql.DB
}

// NewMySQL creates a new MySQL DAO
func NewMySQL(c *conf.MySQL) (*MySQL, error) {
	db, err := sql.Open("mysql", c.DSN)
	if err != nil {
		return nil, fmt.Errorf("open mysql error: %w", err)
	}

	db.SetMaxIdleConns(c.MaxIdleConn)
	db.SetMaxOpenConns(c.MaxOpenConn)
	db.SetConnMaxLifetime(time.Duration(c.MaxLifetime) * time.Second)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping mysql error: %w", err)
	}

	return &MySQL{db: db}, nil
}

// Close closes the database connection
func (m *MySQL) Close() error {
	return m.db.Close()
}

// Exec executes a query without returning any rows
func (m *MySQL) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return m.db.ExecContext(ctx, query, args...)
}

// Query executes a query that returns rows
func (m *MySQL) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return m.db.QueryContext(ctx, query, args...)
}

// QueryRow executes a query that returns at most one row
func (m *MySQL) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return m.db.QueryRowContext(ctx, query, args...)
}

// BeginTx starts a transaction
func (m *MySQL) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return m.db.BeginTx(ctx, opts)
}
