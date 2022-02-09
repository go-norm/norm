// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package sqlbuilder

import (
	"context"
	"fmt"
	"reflect"
	"sort"

	"github.com/pkg/errors"

	"unknwon.dev/norm"
	"unknwon.dev/norm/internal/exql"
	"unknwon.dev/norm/internal/immutable"
	"unknwon.dev/norm/internal/reflectx"
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
		iq.queuedValues = append(iq.queuedValues, values)
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

	queuedValues [][]interface{}
	values       *exql.ValuesGroupsFragment
	arguments    []interface{}

	returning *exql.ReturningFragment

	amendFn func(string) string
}

// processValues iterate over each of queued list of values to generate value
// groups and their arguments.
func (iq *inserterQuery) processValues() error {
	includeZeroAndNil := len(iq.queuedValues) > 1
	vgs := make([]*exql.ValuesGroupFragment, 0, len(iq.queuedValues))
	args := make([]interface{}, 0)
	for _, queuedValue := range iq.queuedValues {
		if len(queuedValue) == 1 {
			cs, vs, err := mapToColumnsAndValues(queuedValue[0], includeZeroAndNil)
			if err != nil {
				return errors.Wrap(err, "map to columns and values")
			}
			if iq.columns.Empty() {
				columns := make([]*exql.ColumnFragment, 0, len(cs))
				for _, name := range cs {
					columns = append(columns, exql.Column(name))
				}
				iq.columns.Append(columns...)
			}

			vals := make([]exql.Fragment, 0, len(vs))
			for i := range vs {
				switch v := vs[i].(type) {
				case *exql.RawFragment:
					vals = append(vals, exql.Raw("DEFAULT"))
				case *exql.ValueFragment:
					vals = append(vals, v)
				default:
					vals = append(vals, exql.Raw("?"))
					args = append(args, v)
				}
			}
			vgs = append(vgs, exql.ValuesGroup(vals...))
		}

		if iq.columns.Empty() || len(queuedValue) == len(iq.columns.Columns) {
			placeholders := make([]exql.Fragment, len(queuedValue))
			for i := range queuedValue {
				placeholders[i] = exql.Raw("?")
			}
			vgs = append(vgs, exql.ValuesGroup(placeholders...))
			args = append(args, queuedValue...)
		}
	}
	iq.values = exql.ValuesGroups(vgs...)
	iq.arguments = args
	return nil
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

type hasIsZero interface {
	IsZero() bool
}

type sortableFieldsWithValues struct {
	fields []string
	values []interface{}
}

func (fv *sortableFieldsWithValues) Len() int {
	return len(fv.fields)
}

func (fv *sortableFieldsWithValues) Swap(i, j int) {
	fv.fields[i], fv.fields[j] = fv.fields[j], fv.fields[i]
	fv.values[i], fv.values[j] = fv.values[j], fv.values[i]
}

func (fv *sortableFieldsWithValues) Less(i, j int) bool {
	return fv.fields[i] < fv.fields[j]
}

// mapToColumnsAndValues receives a pointer to map or struct and maps it to
// columns and values.
func mapToColumnsAndValues(p interface{}, includeZeroAndNil bool) ([]string, []interface{}, error) {
	pv := reflect.ValueOf(p)
	if !pv.IsValid() {
		return nil, nil, errors.New("the value of map or struct is not valid")
	}

	pt := pv.Type()
	if pt.Kind() == reflect.Ptr {
		p = pv.Elem().Interface()
		pv = reflect.ValueOf(p)
		pt = pv.Type()
	}

	var fv sortableFieldsWithValues
	switch pt.Kind() {
	case reflect.Map:
		nfields := pv.Len()
		fv.values = make([]interface{}, nfields)
		fv.fields = make([]string, nfields)
		for i, key := range pv.MapKeys() {
			fv.fields[i] = fmt.Sprintf("%v", key.Interface())
			fv.values[i] = pv.MapIndex(key).Interface()
		}

	case reflect.Struct:
		fieldMap := defaultMapper.TypeMap(pt).Names
		nfields := len(fieldMap)
		fv.values = make([]interface{}, 0, nfields)
		fv.fields = make([]string, 0, nfields)
		for _, fi := range fieldMap {
			_, omitEmpty := fi.Options["omitempty"]

			field := reflectx.FieldByIndexesReadOnly(pv, fi.Index)
			if field.Kind() == reflect.Ptr && field.IsNil() {
				if omitEmpty && !includeZeroAndNil {
					continue
				}
				fv.fields = append(fv.fields, fi.Name)
				if omitEmpty {
					fv.values = append(fv.values, exql.Raw("DEFAULT"))
				} else {
					fv.values = append(fv.values, nil)
				}
				continue
			}

			value := field.Interface()

			isZero := false
			if t, ok := field.Interface().(hasIsZero); ok {
				isZero = t.IsZero()
			} else if field.Kind() == reflect.Array || field.Kind() == reflect.Slice {
				isZero = field.Len() == 0
			} else if reflect.DeepEqual(fi.Zero.Interface(), value) {
				isZero = true
			}
			if isZero && omitEmpty && !includeZeroAndNil {
				continue
			}

			fv.fields = append(fv.fields, fi.Name)
			if isZero && omitEmpty {
				value = exql.Raw("DEFAULT")
			}
			fv.values = append(fv.values, value)
		}

	default:
		return nil, nil, errors.New("the type must be a map or struct or a point to a map or struct")
	}

	sort.Sort(&fv)
	return fv.fields, fv.values, nil
}
