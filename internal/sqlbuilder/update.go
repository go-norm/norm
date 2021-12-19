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

type updaterQuery struct {
	table string

	columnValues     *exql.ColumnValuesFragment
	columnValuesArgs []interface{}

	limit int

	where     *exql.WhereFragment
	whereArgs []interface{}

	amendFn func(string) string
}

func (uq *updaterQuery) and(b *sqlBuilder, conds ...interface{}) error {
	where, whereArgs, err := b.t.toWhereClause(conds)
	if err != nil {
		return errors.Wrap(err, "convert to WHERE clause")
	}

	if uq.where == nil {
		uq.where, uq.whereArgs = &exql.WhereFragment{}, []interface{}{}
	}
	uq.where.Append(where)
	uq.whereArgs = append(uq.whereArgs, whereArgs...)

	return nil
}

func (uq *updaterQuery) statement() *exql.Statement {
	stmt := &exql.Statement{
		Type:         exql.Update,
		Table:        exql.Table(uq.table),
		ColumnValues: uq.columnValues,
	}

	if uq.where != nil {
		stmt.Where = uq.where
	}

	if uq.limit != 0 {
		stmt.Limit = exql.Limit(uq.limit)
	}

	stmt.SetAmendment(uq.amendFn)

	return stmt
}

func (uq *updaterQuery) arguments() []interface{} {
	return flattenArguments(
		uq.columnValuesArgs,
		uq.whereArgs,
	)
}

type updater struct {
	builder *sqlBuilder

	fn   func(*updaterQuery) error
	prev *updater
}

var _ = immutable.Immutable(&updater{})

func (upd *updater) Builder() *sqlBuilder {
	if upd.prev == nil {
		return upd.builder
	}
	return upd.prev.Builder()
}

func (upd *updater) template() *exql.Template {
	return upd.Builder().t.Template
}

func (upd *updater) String() string {
	q, err := upd.Compile()
	if err != nil {
		panic(err.Error())
	}
	return upd.Builder().t.FormatSQL(q)
}

func (upd *updater) setTable(table string) *updater {
	return upd.frame(func(uq *updaterQuery) error {
		uq.table = table
		return nil
	})
}

func (upd *updater) frame(fn func(*updaterQuery) error) *updater {
	return &updater{prev: upd, fn: fn}
}

func (upd *updater) Set(terms ...interface{}) norm.Updater {
	return upd.frame(func(uq *updaterQuery) error {
		if uq.columnValues == nil {
			uq.columnValues = &exql.ColumnValuesFragment{}
		}

		if len(terms) == 1 {
			ff, vv, err := Map(terms[0], nil)
			if err == nil && len(ff) > 0 {
				cvs := make([]exql.Fragment, 0, len(ff))
				args := make([]interface{}, 0, len(vv))

				for i := range ff {
					cv := &exql.ColumnValueFragment{
						Column:   exql.ColumnWithName(ff[i]),
						Operator: upd.Builder().t.AssignmentOperator,
					}

					var localArgs []interface{}
					cv.Value, localArgs = upd.Builder().t.PlaceholderValue(vv[i])

					args = append(args, localArgs...)
					cvs = append(cvs, cv)
				}

				uq.columnValues.Append(cvs...)
				uq.columnValuesArgs = append(uq.columnValuesArgs, args...)

				return nil
			}
		}

		cv, args, err := upd.Builder().t.setColumnValues(terms)
		if err != nil {
			return errors.Wrap(err, "Set: set column values")
		}

		uq.columnValues.Append(cv.ColumnValues...)
		uq.columnValuesArgs = append(uq.columnValuesArgs, args...)
		return nil
	})
}

func (upd *updater) Amend(fn func(string) string) norm.Updater {
	return upd.frame(func(uq *updaterQuery) error {
		uq.amendFn = fn
		return nil
	})
}

func (upd *updater) Arguments() []interface{} {
	uq, err := upd.build()
	if err != nil {
		return nil
	}

	args := uq.arguments()
	for i := range args {
		args[i] = upd.Builder().Typer().Valuer(args[i])
	}
	return args
}

func (upd *updater) Where(terms ...interface{}) norm.Updater {
	return upd.frame(func(uq *updaterQuery) error {
		uq.where, uq.whereArgs = &exql.WhereFragment{}, []interface{}{}
		return uq.and(upd.Builder(), terms...)
	})
}

func (upd *updater) And(terms ...interface{}) norm.Updater {
	return upd.frame(func(uq *updaterQuery) error {
		return uq.and(upd.Builder(), terms...)
	})
}

func (upd *updater) Prepare(ctx context.Context) (*sql.Stmt, error) {
	uq, err := upd.build()
	if err != nil {
		return nil, err
	}
	return upd.Builder().Executor().Prepare(ctx, uq.statement())
}

func (upd *updater) Exec(ctx context.Context) (sql.Result, error) {
	uq, err := upd.build()
	if err != nil {
		return nil, err
	}
	return upd.Builder().Executor().Exec(ctx, uq.statement(), uq.arguments()...)
}

func (upd *updater) Limit(limit int) norm.Updater {
	return upd.frame(func(uq *updaterQuery) error {
		uq.limit = limit
		return nil
	})
}

func (upd *updater) statement() (*exql.Statement, error) {
	iq, err := upd.build()
	if err != nil {
		return nil, err
	}
	return iq.statement(), nil
}

func (upd *updater) build() (*updaterQuery, error) {
	uq, err := immutable.FastForward(upd)
	if err != nil {
		return nil, err
	}
	return uq.(*updaterQuery), nil
}

func (upd *updater) Compile() (string, error) {
	s, err := upd.statement()
	if err != nil {
		return "", err
	}
	return s.Compile(upd.template())
}

func (upd *updater) Prev() immutable.Immutable {
	if upd == nil {
		return nil
	}
	return upd.prev
}

func (upd *updater) Fn(in interface{}) error {
	if upd.fn == nil {
		return nil
	}
	return upd.fn(in.(*updaterQuery))
}

func (upd *updater) Base() interface{} {
	return &updaterQuery{}
}
