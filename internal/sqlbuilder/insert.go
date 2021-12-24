// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package sqlbuilder

import (
	"github.com/pkg/errors"

	"unknwon.dev/norm"
	"unknwon.dev/norm/internal/exql"
	"unknwon.dev/norm/internal/immutable"
)

var _ norm.Inserter = (*inserter)(nil)

type inserter struct {
	builder *sqlBuilder

	prev *inserter
	fn   func(*inserterQuery) error
}

func (ins *inserter) frame(fn func(*inserterQuery) error) *inserter {
	return &inserter{
		prev: ins,
		fn:   fn,
	}
}

func (ins *inserter) Builder() *sqlBuilder {
	if ins.prev == nil {
		return ins.builder
	}
	return ins.prev.Builder()
}

func (ins *inserter) Into(table string) norm.Inserter {
	return ins.frame(func(iq *inserterQuery) error {
		iq.table = table
		return nil
	})
}

func (ins *inserter) Columns(columns ...interface{}) norm.Inserter {
	if len(columns) == 0 {
		return ins
	}
	return ins.frame(func(iq *inserterQuery) error {
		cs := make([]*exql.ColumnFragment, len(columns))
		for i := range columns {
			cs[i] = exql.Column(columns[i])
		}
		iq.columns = exql.Columns(cs...)
		return nil
	})
}

func (ins *inserter) Values(values ...interface{}) norm.Inserter {
	if len(values) == 0 {
		return ins
	}
	return ins.frame(func(iq *inserterQuery) error {
		iq.enqueuedValues = append(iq.enqueuedValues, values)
		return nil
	})
}

func (ins *inserter) Returning(columns ...interface{}) norm.Inserter {
	if len(columns) == 0 {
		return ins
	}
	return ins.frame(func(iq *inserterQuery) error {
		cs := make([]*exql.ColumnFragment, len(columns))
		for i := range columns {
			cs[i] = exql.Column(columns[i])
		}
		iq.returning = exql.Returning(cs...)
		return nil
	})
}

func (ins *inserter) String() string {
	q, err := ins.Compile()
	if err != nil {
		panic("unable to compile INSERT query: " + err.Error())
	}
	return ins.Builder().FormatSQL(q)
}

func (ins *inserter) build() (*inserterQuery, error) {
	iq, err := immutable.FastForward(ins)
	if err != nil {
		return nil, errors.Wrap(err, "construct *inserterQuery")
	}

	ret := iq.(*inserterQuery)
	ret.values, ret.arguments, err = ret.processValues()
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (ins *inserter) Arguments() []interface{} {
	iq, err := ins.build()
	if err != nil {
		panic("unable to build INSERT query: " + err.Error())
	}

	args := iq.arguments
	for i := range args {
		args[i] = ins.Builder().Typer().Valuer(args[i])
	}
	return args
}

var _ compilable = (*inserter)(nil)

func (ins *inserter) Compile() (string, error) {
	iq, err := ins.build()
	if err != nil {
		return "", errors.Wrap(err, "build")
	}
	return iq.statement().Compile(ins.Builder().Template)
}

var _ immutable.Immutable = (*inserter)(nil)

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

type inserterQuery struct {
	table   string
	columns *exql.ColumnsFragment

	enqueuedValues [][]interface{}
	values         *exql.ValuesGroupsFragment

	returning *exql.ReturningFragment

	arguments []interface{}

	amendFn func(string) string
}

func (iq *inserterQuery) processValues() (*exql.ValuesGroupsFragment, []interface{}, error) {
	var mapOptions *MapOptions
	if len(iq.enqueuedValues) > 1 {
		mapOptions = &MapOptions{IncludeZeroed: true, IncludeNil: true} // todo
	}

	values := make([]*exql.ValuesGroupFragment, 0, len(iq.enqueuedValues))
	args := []interface{}{}
	for _, enqueuedValue := range iq.enqueuedValues {
		if len(enqueuedValue) == 1 {
			// If and only if we passed one argument to ValuesGroup.
			ff, vv, err := mapToColumnsAndValues(enqueuedValue[0], mapOptions)
			if err != nil {
				// // The only error we can expect without exiting is this argument not
				// // being a map or struct, in which case we can continue.
				// if !errors.Is(err, ErrExpectingPointerToEitherMapOrStruct) {
				// 	return nil, nil, err
				// }
				return nil, nil, errors.Wrap(err, "TODO") // todo
			}

			columns, vals, cvArgs, err := toColumnsValuesAndArguments(ff, vv)
			if err != nil {
				return nil, nil, errors.Wrap(err, "TODO") // todo
			}

			values = append(values, vals)
			args = append(args, cvArgs...)
			if iq.columns.Empty() {
				iq.columns.Append(columns.Columns...)
			}
		}

		if iq.columns.Empty() || len(enqueuedValue) == len(iq.columns.Columns) {
			placeholders := make([]exql.Fragment, len(enqueuedValue))
			for i := range enqueuedValue {
				placeholders[i] = exql.Raw("?")
			}
			values = append(values, exql.ValuesGroup(placeholders...))
			args = append(args, enqueuedValue...)
		}
	}
	return exql.ValuesGroups(values...), args, nil
}

func (iq *inserterQuery) statement() *exql.Statement {
	stmt := &exql.Statement{
		Type:      exql.StatementInsert,
		Table:     exql.Table(iq.table),
		Columns:   iq.columns,
		Values:    iq.values,
		Returning: iq.returning,
	}
	stmt.SetAmend(iq.amendFn)
	return stmt
}
