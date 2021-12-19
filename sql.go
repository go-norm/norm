// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package norm

import (
	"context"
	"database/sql"
)

// SQL represents a database-agnostic SQL query builder with chainable methods.
//
// Queries are immutable, so every call to any method will return a new pointer
// to the query. The variable that holds the query needs to be reassigned for
// building a query in multiple statements:
//
//   q := db.Select("bar").From("foo") // "q" is created
//
//   q.Where(...) // Does nothing as the value returned from Where is discarded.
//
//   q = q.Where(...) // "q" is reassigned and points to a different address.
type SQL interface {
	// Select creates a Selector that selects from the given columns (i.e.
	// `SELECT first_name, last_name`)
	//
	// The returned Selector does not initially point to any table, a call to From()
	// is required after Select() to construct a valid query:
	//
	//   q := db.Select("first_name", "last_name").From("users").Where(...)
	Select(columns ...interface{}) Selector
	// SelectFrom creates a Selector that selects all columns (i.e. `SELECT *`) from
	// the given table.
	//
	// Example:
	//
	//   q := db.SelectFrom("users").Where(...)
	SelectFrom(tables ...interface{}) Selector
	// InsertInto creates an Inserter targeted at the given table.
	//
	// Example:
	//
	//   q := db.InsertInto("users").Columns(...).Values(...)
	InsertInto(table string) Inserter
	// Update creates an Updater targeted at the given table.
	//
	// Example:
	//
	//   q := db.Update("users").Set(...).Where(...)
	Update(table string) Updater
	// DeleteFrom creates a Deleter targeted at the given table.
	//
	// Example:
	//
	//   q := db.DeleteFrom("users").Where(...)
	DeleteFrom(table string) Deleter
	// AlterTable() Alter
	// Create() Creator
	// Drop() Dropper
}

// SQLExecer provides methods for executing statements that do not return
// results.
type SQLExecer interface {
	// Exec executes a statement and returns sql.Result.
	Exec(context.Context) (sql.Result, error)
}

// SQLPreparer provides the Prepare and Prepare methods for creating
// prepared statements.
type SQLPreparer interface {
	// Prepare creates a prepared statement.
	Prepare(context.Context) (*sql.Stmt, error)
}

// SQLGetter provides methods for executing statements that return results.
type SQLGetter interface {
	// Query returns *sql.Rows.
	Query(context.Context) (*sql.Rows, error)
	// QueryRow returns only one row.
	QueryRow(ctx context.Context) (*sql.Row, error)
}
