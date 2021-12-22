// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package adapter

import (
	"context"
	"database/sql"

	"unknwon.dev/norm/internal/exql"
)

// Name is the name of the database adapter.
type Name string

const (
	PostgreSQL Name = "postgres"
	MySQL      Name = "mysql"
	SQLite3    Name = "sqlite3"
)

// Adapter represents a database adapter that interacts with the corresponding
// database backend. It is responsible for executing the SQL queries using the
// underlying database driver and transparently wrapping types for scanning and
// storing values from and to the database.
type Adapter interface {
	// Name returns the name of the adapter.
	Name() Name
	// Executor returns the executor of the adapter.
	Executor() Executor
	// Typer returns the typer of the adapter.
	Typer() Typer
	// FormatSQL returns formatted SQL from the given string.
	FormatSQL(sql string) string
}

// Executor compiles query statements into actual SQL queries and executes them
// using the underlying database driver.
type Executor interface {
	// Exec compiles the statement to a query and executes it without returning any
	// rows. The args are for any placeholder parameters in the query.
	Exec(ctx context.Context, stmt *exql.Statement, args ...interface{}) (sql.Result, error)
	// Prepare compiles and prepares the statement for later queries or executions.
	//
	// Multiple queries or executions may be run concurrently from the returned
	// statement. The caller must call the statement's Close method when the
	// statement is no longer needed.
	Prepare(ctx context.Context, stmt *exql.Statement) (*sql.Stmt, error)
	// Query compiles the statement to a query and executes it for returning rows.
	// The args are for any placeholder parameters in the query.
	Query(ctx context.Context, stmt *exql.Statement, args ...interface{}) (*sql.Rows, error)
	// QueryRow compiles the statement to a query and executes it for returning at
	// most one row.
	//
	// This method always returns a non-nil value, and errors are deferred until
	// `(*sql.Row).Scan` method is called. If the query selects no rows, the
	// `(*sql.Row).Scan` will return `sql.ErrNoRows`. Otherwise, the
	// `(*sql.Row).Scan` scans the first selected row and discards the rest.
	QueryRow(ctx context.Context, stmt *exql.Statement, args ...interface{}) (*sql.Row, error)
}

// Typer transparently wraps types for scanning and storing values from and to
// the database. This allows type definitions in user structs to be
// database-agnostic and let the typer handle the marshalling and unmarshalling.
type Typer interface {
	// Scanner tries to wrap the given type to be a `sql.Scanner`. It is a no-op
	// (returns the original type) when the type is unrecognizable.
	Scanner(v interface{}) interface{}
	// Valuer tries to wrap the given type to be a `driver.Valuer`. It is a no-op
	// (returns the original type) when the type is unrecognizable.
	Valuer(v interface{}) interface{}
}
