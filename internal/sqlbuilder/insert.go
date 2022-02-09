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
		iq.rawValuesGroups = append(iq.rawValuesGroups, values)
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

func (ins *inserter) Amend(fn func(query string) string) norm.Inserter {
	return ins.frame(func(iq *inserterQuery) error {
		iq.amendFn = fn
		return nil
	})
}

func (ins *inserter) Iterate(ctx context.Context) norm.Iterator {
	iq, err := ins.build()
	if err != nil {
		return &iterator{err: errors.Wrap(err, "build query")}
	}

	adapter := ins.Builder().Adapter
	rows, err := adapter.Executor().Query(ctx, iq.statement(), iq.arguments...) //nolint:rowserrcheck
	return &iterator{
		adapter: adapter,
		cursor:  rows,
		err:     errors.Wrap(err, "execute query"),
	}
}

func (ins *inserter) All(ctx context.Context, destSlice interface{}) error {
	return ins.Iterate(ctx).All(ctx, destSlice)
}

func (ins *inserter) One(ctx context.Context, dest interface{}) error {
	return ins.Iterate(ctx).One(ctx, dest)
}

func (ins *inserter) String() string {
	q, err := ins.Compile()
	if err != nil {
		panic("unable to compile INSERT query: " + err.Error())
	}
	return ins.Builder().FormatSQL(q)
}

func (ins *inserter) build() (*inserterQuery, error) {
	v, err := immutable.FastForward(ins)
	if err != nil {
		return nil, errors.Wrap(err, "construct *inserterQuery")
	}

	iq := v.(*inserterQuery)
	err = iq.processValues()
	if err != nil {
		return nil, errors.Wrap(err, "process values")
	}
	return iq, nil
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

	rawValuesGroups [][]interface{}
	valuesGroups    *exql.ValuesGroupsFragment
	arguments       []interface{}

	returning *exql.ReturningFragment

	amendFn func(string) string
}

// processValues iterate over each of queued list of values to generate value
// groups and their arguments.
func (iq *inserterQuery) processValues() error {
	vgs := make([]*exql.ValuesGroupFragment, 0, len(iq.rawValuesGroups))
	args := make([]interface{}, 0)
	for _, valuesGroup := range iq.rawValuesGroups {
		vs := make([]exql.Fragment, 0, len(valuesGroup))
		for i := range valuesGroup {
			switch v := valuesGroup[i].(type) {
			case exql.Fragment:
				vs = append(vs, v)
			default:
				vs = append(vs, exql.Raw("?"))
				args = append(args, v)
			}
		}
		vgs = append(vgs, exql.ValuesGroup(vs...))
	}
	iq.valuesGroups = exql.ValuesGroups(vgs...)
	iq.arguments = args
	return nil
}

func (iq *inserterQuery) statement() *exql.Statement {
	stmt := &exql.Statement{
		Type:      exql.StatementInsert,
		Table:     exql.Table(iq.table),
		Columns:   iq.columns,
		Values:    iq.valuesGroups,
		Returning: iq.returning,
	}
	stmt.SetAmend(iq.amendFn)
	return stmt
}
