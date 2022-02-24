// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package sqlbuilder

import (
	"context"

	"github.com/pkg/errors"

	"unknwon.dev/norm"
	"unknwon.dev/norm/internal/exql"
	"unknwon.dev/norm/internal/immutable"
)

var _ norm.Deleter = (*deleter)(nil)

type deleter struct {
	builder *sqlBuilder

	prev *deleter
	fn   func(*deleterQuery) error
}

func (del *deleter) frame(fn func(*deleterQuery) error) *deleter {
	return &deleter{
		prev: del,
		fn:   fn,
	}
}

func (del *deleter) Builder() *sqlBuilder {
	if del.prev == nil {
		return del.builder
	}
	return del.prev.Builder()
}

func (del *deleter) Table(table string) *deleter {
	return del.frame(func(dq *deleterQuery) error {
		dq.table = table
		return nil
	})
}

func (del *deleter) Where(conds ...interface{}) norm.Deleter {
	if len(conds) == 0 {
		return del
	}
	return del.frame(func(dq *deleterQuery) error {
		dq.where, dq.whereArgs = nil, nil
		return errors.Wrap(dq.and(del.Builder().Template, conds...), "Where")
	})
}

func (del *deleter) And(conds ...interface{}) norm.Deleter {
	if len(conds) == 0 {
		return del
	}
	return del.frame(func(dq *deleterQuery) error {
		return errors.Wrap(dq.and(del.Builder().Template, conds...), "And")
	})
}

func (del *deleter) Returning(columns ...interface{}) norm.Deleter {
	if len(columns) == 0 {
		return del
	}
	return del.frame(func(dq *deleterQuery) error {
		cs := make([]*exql.ColumnFragment, len(columns))
		for i := range columns {
			cs[i] = exql.Column(columns[i])
		}
		dq.returning = exql.Returning(cs...)
		return nil
	})
}

func (del *deleter) Amend(fn func(string) string) norm.Deleter {
	return del.frame(func(dq *deleterQuery) error {
		dq.amendFn = fn
		return nil
	})
}

func (del *deleter) Iterate(ctx context.Context) norm.Iterator {
	iq, err := del.build()
	if err != nil {
		return &iterator{err: errors.Wrap(err, "build query")}
	}

	adapter := del.Builder().Adapter
	rows, err := adapter.Executor().Query(ctx, iq.statement(), del.Arguments()...) //nolint:rowserrcheck
	return &iterator{
		adapter: adapter,
		cursor:  rows,
		err:     errors.Wrap(err, "execute query"),
	}
}

func (del *deleter) All(ctx context.Context, destSlice interface{}) error {
	return del.Iterate(ctx).All(ctx, destSlice)
}

func (del *deleter) One(ctx context.Context, dest interface{}) error {
	return del.Iterate(ctx).One(ctx, dest)
}

func (del *deleter) String() string {
	q, err := del.Compile()
	if err != nil {
		panic("unable to compile DELETE query: " + err.Error())
	}
	return del.Builder().FormatSQL(q)
}

func (del *deleter) build() (*deleterQuery, error) {
	dq, err := immutable.FastForward(del)
	if err != nil {
		return nil, errors.Wrap(err, "construct *deleterQuery")
	}
	return dq.(*deleterQuery), nil
}

func (del *deleter) Arguments() []interface{} {
	dq, err := del.build()
	if err != nil {
		panic("unable to build DELETE query: " + err.Error())
	}

	args := dq.arguments()
	for i := range args {
		args[i] = del.Builder().Typer().Valuer(args[i])
	}
	return args
}

var _ compilable = (*deleter)(nil)

func (del *deleter) Compile() (string, error) {
	dq, err := del.build()
	if err != nil {
		return "", errors.Wrap(err, "build")
	}
	return dq.statement().Compile(del.Builder().Template)
}

var _ immutable.Immutable = (*deleter)(nil)

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

type deleterQuery struct {
	table string

	where     *exql.WhereFragment
	whereArgs []interface{}

	returning *exql.ReturningFragment

	amendFn func(string) string
}

func (dq *deleterQuery) arguments() []interface{} {
	return flattenArguments(dq.whereArgs)
}

func (dq *deleterQuery) and(t *exql.Template, conditions ...interface{}) error {
	conds, condsArgs, err := parseConditionExpressions(t, conditions)
	if err != nil {
		return errors.Wrap(err, "parse condition expressions")
	}

	if dq.where == nil {
		dq.where, dq.whereArgs = exql.Where(), []interface{}{}
	}
	dq.where.Append(conds...)
	dq.whereArgs = append(dq.whereArgs, condsArgs...)
	return nil
}

func (dq *deleterQuery) statement() *exql.Statement {
	stmt := &exql.Statement{
		Type:      exql.StatementDelete,
		Table:     exql.Table(dq.table),
		Where:     dq.where,
		Returning: dq.returning,
	}
	stmt.SetAmend(dq.amendFn)
	return stmt
}
