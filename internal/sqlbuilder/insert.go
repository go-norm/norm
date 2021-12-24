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

// todo toColumnsValuesAndArguments maps the given columnNames and columnValues into
// expr's Columns and ValuesGroup, it also extracts and returns query arguments.
func toColumnsValuesAndArguments(columnNames []string, columnValues []interface{}) (*exql.ColumnsFragment, *exql.ValuesGroupFragment, []interface{}, error) {
	var arguments []interface{}

	columns := new(exql.ColumnsFragment)

	columns.Columns = make([]*exql.ColumnFragment, 0, len(columnNames))
	for i := range columnNames {
		columns.Columns = append(columns.Columns, exql.Column(columnNames[i]))
	}

	values := new(exql.ValuesGroupFragment)

	arguments = make([]interface{}, 0, len(columnValues))
	values.Values = make([]exql.Fragment, 0, len(columnValues))

	for i := range columnValues {
		switch v := columnValues[i].(type) {
		case *exql.RawFragment, exql.RawFragment:
			values.Values = append(values.Values, exql.Raw("DEFAULT"))
		case *exql.ValueFragment:
			// Adding value.
			values.Values = append(values.Values, v)
		case exql.ValueFragment:
			// Adding value.
			values.Values = append(values.Values, &v)
		default:
			// Adding both value and placeholder.
			values.Values = append(values.Values, exql.Raw("?"))
			arguments = append(arguments, v)
		}
	}

	return columns, values, arguments, nil
}

// todo
func (iq *inserterQuery) processValues() (*exql.ValuesGroupsFragment, []interface{}, error) {
	var values []*exql.ValuesGroupFragment
	var arguments []interface{}

	var mapOptions *MapOptions
	if len(iq.enqueuedValues) > 1 {
		mapOptions = &MapOptions{IncludeZeroed: true, IncludeNil: true}
	}

	for _, enqueuedValue := range iq.enqueuedValues {
		if len(enqueuedValue) == 1 {
			// If and only if we passed one argument to ValuesGroup.
			ff, vv, err := Map(enqueuedValue[0], mapOptions)

			if err == nil {
				// If we didn't have any problem with mapping we can convert it into
				// columns and values.
				columns, vals, args, _ := toColumnsValuesAndArguments(ff, vv)

				values, arguments = append(values, vals), append(arguments, args...)

				if iq.columns.Empty() {
					iq.columns.Append(columns.Columns...)
				}
				continue
			}

			// The only error we can expect without exiting is this argument not
			// being a map or struct, in which case we can continue.
			if !errors.Is(err, ErrExpectingPointerToEitherMapOrStruct) {
				return nil, nil, err
			}
		}

		if iq.columns.Empty() || len(enqueuedValue) == len(iq.columns.Columns) {
			arguments = append(arguments, enqueuedValue...)

			l := len(enqueuedValue)
			placeholders := make([]exql.Fragment, l)
			for i := 0; i < l; i++ {
				placeholders[i] = exql.Raw("?")
			}
			values = append(values, exql.ValuesGroup(placeholders...))
		}
	}

	return exql.ValuesGroups(values...), arguments, nil
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
