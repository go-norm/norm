// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package sqlbuilder

import (
	"github.com/pkg/errors"

	"unknwon.dev/norm"
	"unknwon.dev/norm/expr"
	"unknwon.dev/norm/internal/exql"
	"unknwon.dev/norm/internal/immutable"
)

var _ norm.Updater = (*updater)(nil)

type updater struct {
	builder *sqlBuilder

	prev *updater
	fn   func(*updaterQuery) error
}

func (upd *updater) frame(fn func(*updaterQuery) error) *updater {
	return &updater{
		prev: upd,
		fn:   fn,
	}
}

func (upd *updater) Builder() *sqlBuilder {
	if upd.prev == nil {
		return upd.builder
	}
	return upd.prev.Builder()
}

func (upd *updater) Table(table string) *updater {
	return upd.frame(func(uq *updaterQuery) error {
		uq.table = table
		return nil
	})
}

func (upd *updater) Set(kvs ...interface{}) norm.Updater {
	if len(kvs) == 0 {
		return upd
	}
	return upd.frame(func(uq *updaterQuery) error {
		if len(kvs)%2 != 0 {
			return errors.Errorf("Set: odd number of key-value pairs: %d", len(kvs))
		}

		cvs := make([]*exql.ColumnValueFragment, 0, len(kvs)/2)
		args := make([]interface{}, 0, len(cvs))
		for i := 0; i < len(kvs); i += 2 {
			cv := exql.ColumnValue(kvs[i], upd.Builder().Layout(exql.LayoutAssignmentOperator), nil)

			switch v := kvs[i].(type) {
			case *expr.RawExpr:
				cv.Value = exql.Raw(v.Raw())
				args = append(args, v.Arguments()...)
			case *expr.FuncExpr:
				fnName, fnArgs, err := expandFuncExpr(v)
				if err != nil {
					return errors.Wrap(err, "Set: expand *expr.FuncExpr")
				}
				cv.Value = exql.Raw(fnName)
				args = append(args, fnArgs...)
			default:
				cv.Value = exql.Raw("?")
				args = append(args, v)
			}
		}

		uq.columnValues = append(uq.columnValues, cvs...)
		uq.columnValuesArgs = append(uq.columnValuesArgs, args...)
		return nil
	})
}

func (upd *updater) Where(conds ...interface{}) norm.Updater {
	if len(conds) == 0 {
		return upd
	}
	return upd.frame(func(uq *updaterQuery) error {
		uq.where, uq.whereArgs = nil, nil
		return errors.Wrap(uq.and(upd.Builder().Template, conds...), "Where")
	})
}

func (upd *updater) And(conds ...interface{}) norm.Updater {
	if len(conds) == 0 {
		return upd
	}
	return upd.frame(func(uq *updaterQuery) error {
		return errors.Wrap(uq.and(upd.Builder().Template, conds...), "And")
	})
}

func (upd *updater) Limit(n int) norm.Updater {
	if n <= 0 {
		return upd
	}
	return upd.frame(func(uq *updaterQuery) error {
		uq.limit = n
		return nil
	})
}

func (upd *updater) Amend(fn func(string) string) norm.Updater {
	return upd.frame(func(uq *updaterQuery) error {
		uq.amendFn = fn
		return nil
	})
}

func (upd *updater) String() string {
	q, err := upd.Compile()
	if err != nil {
		panic("unable to compile UPDATE query: " + err.Error())
	}
	return upd.Builder().FormatSQL(q)
}

func (upd *updater) build() (*updaterQuery, error) {
	uq, err := immutable.FastForward(upd)
	if err != nil {
		return nil, errors.Wrap(err, "construct *updaterQuery")
	}
	return uq.(*updaterQuery), nil
}

func (upd *updater) Arguments() []interface{} {
	uq, err := upd.build()
	if err != nil {
		panic("unable to build UPDATE query: " + err.Error())
	}

	args := uq.arguments()
	for i := range args {
		args[i] = upd.Builder().Typer().Valuer(args[i])
	}
	return args
}

var _ compilable = (*updater)(nil)

func (upd *updater) Compile() (string, error) {
	uq, err := upd.build()
	if err != nil {
		return "", errors.Wrap(err, "build")
	}
	return uq.statement().Compile(upd.Builder().Template)
}

var _ immutable.Immutable = (*updater)(nil)

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

type updaterQuery struct {
	table string

	columnValues     []*exql.ColumnValueFragment
	columnValuesArgs []interface{}

	where     *exql.WhereFragment
	whereArgs []interface{}

	limit int

	amendFn func(string) string
}

func (uq *updaterQuery) arguments() []interface{} {
	return flattenArguments(
		uq.columnValuesArgs,
		uq.whereArgs,
	)
}

func (uq *updaterQuery) and(t *exql.Template, conditions ...interface{}) error {
	conds, condsArgs, err := parseConditionExpressions(t, conditions)
	if err != nil {
		return errors.Wrap(err, "parse condition expressions")
	}

	if uq.where == nil {
		uq.where, uq.whereArgs = exql.Where(), []interface{}{}
	}
	uq.where.Append(conds...)
	uq.whereArgs = append(uq.whereArgs, condsArgs...)
	return nil
}

func (uq *updaterQuery) statement() *exql.Statement {
	stmt := &exql.Statement{
		Type:         exql.StatementUpdate,
		Table:        exql.Table(uq.table),
		ColumnValues: exql.ColumnValues(uq.columnValues...),
		Where:        uq.where,
		Limit:        uq.limit,
	}
	stmt.SetAmend(uq.amendFn)
	return stmt
}
