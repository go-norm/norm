// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package adapter

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
}

// Executor compiles query statements into actual SQL queries and executes them
// using the underlying database driver.
type Executor interface {
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
