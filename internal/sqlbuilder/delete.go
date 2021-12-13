// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package sqlbuilder

import (
	"context"
	"database/sql"

	"github.com/pkg/errors"
	"unknwon.dev/norm"
	"unknwon.dev/norm/internal/exql"
	"unknwon.dev/norm/internal/immutable"
)

type deleterQuery struct {
	table string
	limit int

	where     *exql.Where
	whereArgs []interface{}

	amendFn func(string) string
}

func (dq *deleterQuery) and(b *sqlBuilder, terms ...interface{}) error {
	where, whereArgs, err := b.t.toWhereClause(terms)
	if err != nil {
		return errors.Wrap(err, "convert to WHERE clause")
	}

	if dq.where == nil {
		dq.where, dq.whereArgs = &exql.Where{}, []interface{}{}
	}
	dq.where.Append(where)
	dq.whereArgs = append(dq.whereArgs, whereArgs...)

	return nil
}

func (dq *deleterQuery) statement() *exql.Statement {
	stmt := &exql.Statement{
		Type:  exql.Delete,
		Table: exql.TableWithName(dq.table),
	}

	if dq.where != nil {
		stmt.Where = dq.where
	}

	if dq.limit != 0 {
		stmt.Limit = exql.Limit(dq.limit)
	}

	stmt.SetAmendment(dq.amendFn)

	return stmt
}

type deleter struct {
	builder *sqlBuilder

	fn   func(*deleterQuery) error
	prev *deleter
}

var _ = immutable.Immutable(&deleter{})

func (del *deleter) Builder() *sqlBuilder {
	if del.prev == nil {
		return del.builder
	}
	return del.prev.Builder()
}

func (del *deleter) template() *exql.Template {
	return del.Builder().t.Template
}

func (del *deleter) String() string {
	q, err := del.Compile()
	if err != nil {
		panic(err.Error())
	}
	return del.Builder().t.FormatSQL(q)
}

func (del *deleter) setTable(table string) *deleter {
	return del.frame(func(uq *deleterQuery) error {
		uq.table = table
		return nil
	})
}

func (del *deleter) frame(fn func(*deleterQuery) error) *deleter {
	return &deleter{prev: del, fn: fn}
}

func (del *deleter) Where(terms ...interface{}) norm.Deleter {
	return del.frame(func(dq *deleterQuery) error {
		dq.where, dq.whereArgs = &exql.Where{}, []interface{}{}
		return dq.and(del.Builder(), terms...)
	})
}

func (del *deleter) And(terms ...interface{}) norm.Deleter {
	return del.frame(func(dq *deleterQuery) error {
		return dq.and(del.Builder(), terms...)
	})
}

func (del *deleter) Limit(limit int) norm.Deleter {
	return del.frame(func(dq *deleterQuery) error {
		dq.limit = limit
		return nil
	})
}

func (del *deleter) Amend(fn func(string) string) norm.Deleter {
	return del.frame(func(dq *deleterQuery) error {
		dq.amendFn = fn
		return nil
	})
}

func (dq *deleterQuery) arguments() []interface{} {
	return joinArguments(dq.whereArgs)
}

func (del *deleter) Arguments() []interface{} {
	dq, err := del.build()
	if err != nil {
		return nil
	}

	args := dq.arguments()
	for i := range args {
		args[i] = del.Builder().Typer().Valuer(args[i])
	}
	return args
}

func (del *deleter) Prepare(ctx context.Context) (*sql.Stmt, error) {
	dq, err := del.build()
	if err != nil {
		return nil, err
	}
	return del.Builder().Executor().Prepare(ctx, dq.statement())
}

func (del *deleter) Exec(ctx context.Context) (sql.Result, error) {
	dq, err := del.build()
	if err != nil {
		return nil, err
	}
	return del.Builder().Executor().Exec(ctx, dq.statement(), dq.arguments()...)
}

func (del *deleter) statement() (*exql.Statement, error) {
	iq, err := del.build()
	if err != nil {
		return nil, err
	}
	return iq.statement(), nil
}

func (del *deleter) build() (*deleterQuery, error) {
	dq, err := immutable.FastForward(del)
	if err != nil {
		return nil, err
	}
	return dq.(*deleterQuery), nil
}

func (del *deleter) Compile() (string, error) {
	s, err := del.statement()
	if err != nil {
		return "", err
	}
	return s.Compile(del.template())
}

func (del *deleter) Prev() immutable.Immutable {
	if del == nil {
		return nil
	}
	return del.prev
}

func (del *deleter) Fn(in interface{}) error {
	if del.fn == nil {
		return nil
	}
	return del.fn(in.(*deleterQuery))
}

func (del *deleter) Base() interface{} {
	return &deleterQuery{}
}
