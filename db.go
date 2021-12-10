// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package norm

import (
	"context"
	"database/sql"
	"time"

	"unknwon.dev/norm/adapter"
)

// DB represents a database handle representing a pool of zero or more
// underlying connections. It's safe for concurrent use by multiple goroutines.
type DB interface {
	// Now returns the current time in the perspective of the database handle.
	Now() time.Time
	// Driver returns the driver of the database handle.
	Driver() Driver
	// Adapter returns the adapter of the database handle.
	Adapter() adapter.Adapter
	// Close closes the database and prevents new queries from starting. It waits
	// for all queries that have started processing on the database to finish.
	Close() error

	Transactor
	SQL
}

// Driver represents the underlying database driver for managing database
// connections.
type Driver interface {
	// SetConnMaxLifetime sets the maximum amount of time a connection may be
	// reused. Expired connections may be closed lazily before reuse.
	//
	// If d <= 0, connections are not closed due to a connection's age.
	SetConnMaxLifetime(d time.Duration)
	// SetConnMaxIdleTime sets the maximum amount of time a connection may be idle.
	// Expired connections may be closed lazily before reuse.
	//
	// If d <= 0, connections are not closed due to a connection's idle time.
	SetConnMaxIdleTime(d time.Duration)
	// SetMaxIdleConns sets the maximum number of connections in the idle connection
	// pool.
	//
	// If MaxOpenConns is greater than 0 but less than the new MaxIdleConns, then
	// the new MaxIdleConns will be reduced to match the MaxOpenConns limit.
	//
	// If n <= 0, no idle connections are retained.
	//
	// The default max idle connections is currently 2 as of Go 1.17.3. This may
	// change in a future release.
	SetMaxIdleConns(n int)
	// SetMaxOpenConns sets the maximum number of open connections to the database.
	//
	// If MaxIdleConns is greater than 0 and the new MaxOpenConns is less than
	// MaxIdleConns, then MaxIdleConns will be reduced to match the new MaxOpenConns
	// limit.
	//
	// If n <= 0, then there is no limit on the number of open connections. The
	// default is 0 (unlimited).
	SetMaxOpenConns(n int)
}

// TxOptions contains the options to be used to start a new transaction.
type TxOptions struct {
	// Isolation is the transaction isolation level. If zero, the driver or
	// database's default level is used.
	Isolation sql.IsolationLevel
	// ReadOnly indicates whether the transaction will be read-only.
	ReadOnly bool
}

// Transactor defines a collection of methods to be used with database
// transactions.
type Transactor interface {
	// Transaction runs the given function in a transaction. It starts and closes
	// the transaction along with the function execution. The transaction will be
	// rolled back automatically when the function returns an error.
	Transaction(ctx context.Context, fn func(tx DB) error, opts ...*TxOptions) error
}
