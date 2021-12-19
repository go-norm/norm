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

type inserterQuery struct {
	table          string
	enqueuedValues [][]interface{}
	returning      []exql.Fragment
	columns        []exql.Fragment
	values         []*exql.Values
	arguments      []interface{}
	amendFn        func(string) string
}

var sqlDefault = exql.RawValue("DEFAULT")

// toColumnsValuesAndArguments maps the given columnNames and columnValues into
// expr's Columns and Values, it also extracts and returns query arguments.
func toColumnsValuesAndArguments(columnNames []string, columnValues []interface{}) (*exql.ColumnsFragment, *exql.Values, []interface{}, error) {
	var arguments []interface{}

	columns := new(exql.ColumnsFragment)

	columns.Columns = make([]exql.Fragment, 0, len(columnNames))
	for i := range columnNames {
		columns.Columns = append(columns.Columns, exql.ColumnWithName(columnNames[i]))
	}

	values := new(exql.Values)

	arguments = make([]interface{}, 0, len(columnValues))
	values.Values = make([]exql.Fragment, 0, len(columnValues))

	for i := range columnValues {
		switch v := columnValues[i].(type) {
		case *exql.RawFragment, exql.RawFragment:
			values.Values = append(values.Values, sqlDefault)
		case *exql.Value:
			// Adding value.
			values.Values = append(values.Values, v)
		case exql.Value:
			// Adding value.
			values.Values = append(values.Values, &v)
		default:
			// Adding both value and placeholder.
			values.Values = append(values.Values, sqlPlaceholder)
			arguments = append(arguments, v)
		}
	}

	return columns, values, arguments, nil
}

func (iq *inserterQuery) processValues() ([]*exql.Values, []interface{}, error) {
	var values []*exql.Values
	var arguments []interface{}

	var mapOptions *MapOptions
	if len(iq.enqueuedValues) > 1 {
		mapOptions = &MapOptions{IncludeZeroed: true, IncludeNil: true}
	}

	for _, enqueuedValue := range iq.enqueuedValues {
		if len(enqueuedValue) == 1 {
			// If and only if we passed one argument to Values.
			ff, vv, err := Map(enqueuedValue[0], mapOptions)

			if err == nil {
				// If we didn't have any problem with mapping we can convert it into
				// columns and values.
				columns, vals, args, _ := toColumnsValuesAndArguments(ff, vv)

				values, arguments = append(values, vals), append(arguments, args...)

				if len(iq.columns) == 0 {
					iq.columns = append(iq.columns, columns.Columns...)
				}
				continue
			}

			// The only error we can expect without exiting is this argument not
			// being a map or struct, in which case we can continue.
			if !errors.Is(err, ErrExpectingPointerToEitherMapOrStruct) {
				return nil, nil, err
			}
		}

		if len(iq.columns) == 0 || len(enqueuedValue) == len(iq.columns) {
			arguments = append(arguments, enqueuedValue...)

			l := len(enqueuedValue)
			placeholders := make([]exql.Fragment, l)
			for i := 0; i < l; i++ {
				placeholders[i] = exql.RawValue(`?`)
			}
			values = append(values, exql.NewValueGroup(placeholders...))
		}
	}

	return values, arguments, nil
}

func (iq *inserterQuery) statement() *exql.Statement {
	stmt := &exql.Statement{
		Type:  exql.Insert,
		Table: exql.Table(iq.table),
	}

	if len(iq.values) > 0 {
		stmt.Values = exql.JoinValueGroups(iq.values...)
	}

	if len(iq.columns) > 0 {
		stmt.Columns = exql.Columns(iq.columns...)
	}

	if len(iq.returning) > 0 {
		stmt.Returning = exql.ReturningColumns(iq.returning...)
	}

	stmt.SetAmendment(iq.amendFn)

	return stmt
}

type inserter struct {
	builder *sqlBuilder

	fn   func(*inserterQuery) error
	prev *inserter
}

var _ = immutable.Immutable(&inserter{})

func (ins *inserter) Builder() *sqlBuilder {
	if ins.prev == nil {
		return ins.builder
	}
	return ins.prev.Builder()
}

func (ins *inserter) template() *exql.Template {
	return ins.Builder().t.Template
}

func (ins *inserter) String() string {
	q, err := ins.Compile()
	if err != nil {
		panic(err.Error())
	}
	return ins.Builder().t.FormatSQL(q)
}

func (ins *inserter) frame(fn func(*inserterQuery) error) *inserter {
	return &inserter{prev: ins, fn: fn}
}

func (ins *inserter) Batch(n int) norm.BatchInserter {
	return newBatchInserter(ins, n)
}

func (ins *inserter) Amend(fn func(string) string) norm.Inserter {
	return ins.frame(func(iq *inserterQuery) error {
		iq.amendFn = fn
		return nil
	})
}

func (ins *inserter) Arguments() []interface{} {
	iq, err := ins.build()
	if err != nil {
		return nil
	}

	args := iq.arguments
	for i := range args {
		args[i] = ins.Builder().Typer().Valuer(args[i])
	}
	return args
}

func (ins *inserter) Returning(columns ...string) norm.Inserter {
	return ins.frame(func(iq *inserterQuery) error {
		columnsToFragments(&iq.returning, columns)
		return nil
	})
}

func (ins *inserter) Exec(ctx context.Context) (sql.Result, error) {
	iq, err := ins.build()
	if err != nil {
		return nil, err
	}
	return ins.Builder().Executor().Exec(ctx, iq.statement(), iq.arguments...)
}

func (ins *inserter) Prepare(ctx context.Context) (*sql.Stmt, error) {
	iq, err := ins.build()
	if err != nil {
		return nil, err
	}
	return ins.Builder().Executor().Prepare(ctx, iq.statement())
}

func (ins *inserter) Query(ctx context.Context) (*sql.Rows, error) {
	iq, err := ins.build()
	if err != nil {
		return nil, err
	}
	return ins.Builder().Executor().Query(ctx, iq.statement(), iq.arguments...)
}

func (ins *inserter) QueryRow(ctx context.Context) (*sql.Row, error) {
	iq, err := ins.build()
	if err != nil {
		return nil, err
	}
	return ins.Builder().Executor().QueryRow(ctx, iq.statement(), iq.arguments...)
}

func (ins *inserter) Iterator(ctx context.Context) norm.Iterator {
	rows, err := ins.Query(ctx) //nolint:rowserrcheck
	return &iterator{ins.Builder().Adapter, rows, err}
}

func (ins *inserter) Into(table string) norm.Inserter {
	return ins.frame(func(iq *inserterQuery) error {
		iq.table = table
		return nil
	})
}

func (ins *inserter) Columns(columns ...string) norm.Inserter {
	return ins.frame(func(iq *inserterQuery) error {
		columnsToFragments(&iq.columns, columns)
		return nil
	})
}

func (ins *inserter) Values(values ...interface{}) norm.Inserter {
	return ins.frame(func(iq *inserterQuery) error {
		iq.enqueuedValues = append(iq.enqueuedValues, values)
		return nil
	})
}

func (ins *inserter) statement() (*exql.Statement, error) {
	iq, err := ins.build()
	if err != nil {
		return nil, err
	}
	return iq.statement(), nil
}

func (ins *inserter) build() (*inserterQuery, error) {
	iq, err := immutable.FastForward(ins)
	if err != nil {
		return nil, err
	}
	ret := iq.(*inserterQuery)
	ret.values, ret.arguments, err = ret.processValues()
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (ins *inserter) Compile() (string, error) {
	s, err := ins.statement()
	if err != nil {
		return "", err
	}
	return s.Compile(ins.template())
}

func (ins *inserter) Prev() immutable.Immutable {
	if ins == nil {
		return nil
	}
	return ins.prev
}

func (ins *inserter) Fn(in interface{}) error {
	if ins.fn == nil {
		return nil
	}
	return ins.fn(in.(*inserterQuery))
}

func (ins *inserter) Base() interface{} {
	return &inserterQuery{}
}

func columnsToFragments(dst *[]exql.Fragment, columns []string) {
	l := len(columns)
	f := make([]exql.Fragment, l)
	for i := 0; i < l; i++ {
		f[i] = exql.ColumnWithName(columns[i])
	}
	*dst = append(*dst, f...)
}
