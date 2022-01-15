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
	iq, err := immutable.FastForward(ins)
	if err != nil {
		return nil, errors.Wrap(err, "construct *inserterQuery")
	}

	ret := iq.(*inserterQuery)
	ret.values, ret.arguments, err = ret.processValues()
	if err != nil {
		return nil, errors.Wrap(err, "process values")
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

// todo
func (iq *inserterQuery) processValues() (valueGroups *exql.ValuesGroupsFragment, args []interface{}, err error) {
	var mapOptions *MapOptions
	if len(iq.enqueuedValues) > 1 {
		mapOptions = &MapOptions{IncludeZeroed: true, IncludeNil: true} // todo
	}

	values := make([]*exql.ValuesGroupFragment, 0, len(iq.enqueuedValues))
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

// todo MapOptions represents options for the mapper.
type MapOptions struct {
	IncludeZeroed bool
	IncludeNil    bool
}

// todo
var defaultMapOptions = MapOptions{
	IncludeZeroed: false,
	IncludeNil:    false,
}

// todo
type hasIsZero interface {
	IsZero() bool
}

// todo
type fieldValue struct {
	fields []string
	values []interface{}
}

func (fv *fieldValue) Len() int {
	return len(fv.fields)
}

func (fv *fieldValue) Swap(i, j int) {
	fv.fields[i], fv.fields[j] = fv.fields[j], fv.fields[i]
	fv.values[i], fv.values[j] = fv.values[j], fv.values[i]
}

func (fv *fieldValue) Less(i, j int) bool {
	return fv.fields[i] < fv.fields[j]
}

// todo mapToColumnsAndValues receives a pointer to map or struct and maps it to columns and values.
func mapToColumnsAndValues(item interface{}, options *MapOptions) ([]string, []interface{}, error) {
	var fv fieldValue
	if options == nil {
		options = &defaultMapOptions
	}

	itemV := reflect.ValueOf(item)
	if !itemV.IsValid() {
		return nil, nil, nil
	}

	itemT := itemV.Type()

	if itemT.Kind() == reflect.Ptr {
		// Single dereference. Just in case the user passes a pointer to struct
		// instead of a struct.
		item = itemV.Elem().Interface()
		itemV = reflect.ValueOf(item)
		itemT = itemV.Type()
	}

	switch itemT.Kind() {
	case reflect.Map:
		nfields := itemV.Len()
		fv.values = make([]interface{}, nfields)
		fv.fields = make([]string, nfields)
		mkeys := itemV.MapKeys()

		for i, keyV := range mkeys {
			valv := itemV.MapIndex(keyV)
			fv.fields[i] = fmt.Sprintf("%v", keyV.Interface())
			fv.values[i] = valv.Interface()
		}

	case reflect.Struct:
		fieldMap := defaultMapper.TypeMap(itemT).Names
		nfields := len(fieldMap)

		fv.values = make([]interface{}, 0, nfields)
		fv.fields = make([]string, 0, nfields)

		for _, fi := range fieldMap {
			// Field options
			_, tagOmitEmpty := fi.Options["omitempty"]

			fld := reflectx.FieldByIndexesReadOnly(itemV, fi.Index)
			if fld.Kind() == reflect.Ptr && fld.IsNil() {
				if tagOmitEmpty && !options.IncludeNil {
					continue
				}
				fv.fields = append(fv.fields, fi.Name)
				if tagOmitEmpty {
					fv.values = append(fv.values, exql.Raw("DEFAULT"))
				} else {
					fv.values = append(fv.values, nil)
				}
				continue
			}

			value := fld.Interface()

			isZero := false
			if t, ok := fld.Interface().(hasIsZero); ok {
				if t.IsZero() {
					isZero = true
				}
			} else if fld.Kind() == reflect.Array || fld.Kind() == reflect.Slice {
				if fld.Len() == 0 {
					isZero = true
				}
			} else if reflect.DeepEqual(fi.Zero.Interface(), value) {
				isZero = true
			}

			if isZero && tagOmitEmpty && !options.IncludeZeroed {
				continue
			}

			fv.fields = append(fv.fields, fi.Name)
			if isZero && tagOmitEmpty {
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
