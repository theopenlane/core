package entdb

import (
	"errors"
)

var (
	// ErrInvalidTypeArgs is returned when the type of args is invalid
	ErrInvalidTypeArgs = errors.New("dialect/sql: invalid type %T. expect []any for args")
	// ErrInvalidTypeResult is returned when the type of result is invalid
	ErrInvalidTypeResult = errors.New("dialect/sql: invalid type %T. expect *sql.Result")
	// ErrInvalidTypeRows is returned when the type of rows is invalid
	ErrInvalidTypeRows = errors.New("dialect/sql: invalid type %T. expect *sql.Rows")
	// ErrInvalidTypeIsolation is returned when the type of isolation level is invalid
	ErrInvalidTypeIsolation = errors.New("unsupported isolation level: %v")
	// ErrDBKeyNotFound is returned when the db key is not found in the context
	ErrDBKeyNotFound = errors.New("db key not found in context")
	// ErrTxKeyNotFound is returned when the tx key is not found in the context
	ErrTxKeyNotFound = errors.New("tx key not found in context")
	// ErrFailedToParseConnectionString is returned when the connection string is invalid
	ErrFailedToParseConnectionString = errors.New("failed to parse connection string")
	// ErrFfailedToConnectToDatabase is returned when the connection to the database fails
	ErrFfailedToConnectToDatabase = errors.New("failed to connect to database")
	// ErrFailedToStartDatabaseTransaction is returned when the database transaction fails to start
	ErrFailedToStartDatabaseTransaction = errors.New("failed to start database transaction")
	// ErrFailedToCommitDatabaseTransaction is returned when the database transaction fails to commit
	ErrFailedToCommitDatabaseTransaction = errors.New("failed to commit database transaction")
)
