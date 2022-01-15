// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package adapter

import (
	"context"
	"database/sql"

	"github.com/pkg/errors"

	"unknwon.dev/norm/adapter"
	"unknwon.dev/norm/internal/exql"
	"unknwon.dev/norm/internal/sqlbuilder"
)

type BaseDBExecutor struct {
	db      *sql.DB // todo: Factor out 4 methods so we can have an interface here
	t       *exql.Template
	adapter adapter.Adapter
}

func NewBaseDBExecutor(db *sql.DB, t *exql.Template, adapter adapter.Adapter) *BaseDBExecutor {
	return &BaseDBExecutor{
		db:      db,
		t:       t,
		adapter: adapter,
	}
}

func compileStatement(t *exql.Template, adapter adapter.Adapter, stmt *exql.Statement, args []interface{}) (string, []interface{}, error) {
	q, err := stmt.Compile(t)
	if err != nil {
		return "", nil, err
	}

	for i := range args {
		args[i] = adapter.Typer().Valuer(args[i])
	}

	q, args, err = sqlbuilder.ExpandQuery(q, args)
	if err != nil {
		return "", nil, errors.Wrap(err, "expand query")
	}

	q = adapter.FormatSQL(q)
	return q, args, nil
}

func (e *BaseDBExecutor) Exec(ctx context.Context, stmt *exql.Statement, args ...interface{}) (sql.Result, error) {
	s, args, err := compileStatement(e.t, e.adapter, stmt, args)
	if err != nil {
		return nil, err
	}
	return e.db.ExecContext(ctx, s, args...)
}

func (e *BaseDBExecutor) Prepare(ctx context.Context, stmt *exql.Statement) (*sql.Stmt, error) {
	s, _, err := compileStatement(e.t, e.adapter, stmt, nil)
	if err != nil {
		return nil, err
	}
	return e.db.PrepareContext(ctx, s)
}

func (e *BaseDBExecutor) Query(ctx context.Context, stmt *exql.Statement, args ...interface{}) (adapter.Rows, error) {
	s, args, err := compileStatement(e.t, e.adapter, stmt, args)
	if err != nil {
		return nil, err
	}
	return e.db.QueryContext(ctx, s, args...) //nolint:rowserrcheck
}

func (e *BaseDBExecutor) QueryRow(ctx context.Context, stmt *exql.Statement, args ...interface{}) (*sql.Row, error) {
	s, args, err := compileStatement(e.t, e.adapter, stmt, args)
	if err != nil {
		return nil, err
	}
	return e.db.QueryRowContext(ctx, s, args...), nil
}

type BaseTxExecutor struct {
	tx      *sql.Tx
	t       *exql.Template
	adapter adapter.Adapter
}

func NewBaseTxExecutor(tx *sql.Tx, t *exql.Template, adapter adapter.Adapter) *BaseTxExecutor {
	return &BaseTxExecutor{
		tx:      tx,
		t:       t,
		adapter: adapter,
	}
}

func (e *BaseTxExecutor) Exec(ctx context.Context, stmt *exql.Statement, args ...interface{}) (sql.Result, error) {
	s, args, err := compileStatement(e.t, e.adapter, stmt, args)
	if err != nil {
		return nil, err
	}
	return e.tx.ExecContext(ctx, s, args...)
}

func (e *BaseTxExecutor) Prepare(ctx context.Context, stmt *exql.Statement) (*sql.Stmt, error) {
	s, _, err := compileStatement(e.t, e.adapter, stmt, nil)
	if err != nil {
		return nil, err
	}
	return e.tx.PrepareContext(ctx, s)
}

func (e *BaseTxExecutor) Query(ctx context.Context, stmt *exql.Statement, args ...interface{}) (adapter.Rows, error) {
	s, args, err := compileStatement(e.t, e.adapter, stmt, args)
	if err != nil {
		return nil, err
	}
	return e.tx.QueryContext(ctx, s, args...) //nolint:rowserrcheck
}

func (e *BaseTxExecutor) QueryRow(ctx context.Context, stmt *exql.Statement, args ...interface{}) (*sql.Row, error) {
	s, args, err := compileStatement(e.t, e.adapter, stmt, args)
	if err != nil {
		return nil, err
	}
	return e.tx.QueryRowContext(ctx, s, args...), nil
}
