// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package sqlbuilder

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"sort"

	"github.com/pkg/errors"

	"unknwon.dev/norm"
	"unknwon.dev/norm/adapter"
	"unknwon.dev/norm/expr"
	"unknwon.dev/norm/internal/exql"
	"unknwon.dev/norm/internal/reflectx"
)

// MapOptions represents options for the mapper.
type MapOptions struct {
	IncludeZeroed bool
	IncludeNil    bool
}

var defaultMapOptions = MapOptions{
	IncludeZeroed: false,
	IncludeNil:    false,
}

type hasIsZero interface {
	IsZero() bool
}

type iterator struct {
	adapter adapter.Adapter
	cursor  *sql.Rows // This is the main query cursor. It starts as a nil value.
	err     error
}

type fieldValue struct {
	fields []string
	values []interface{}
}

var (
	sqlPlaceholder = exql.RawValue(`?`)
)

type sqlBuilder struct {
	adapter.Adapter
	t *templateWithUtils
}

// New returns a query builder that is bound to the given database session.
func New(adapter adapter.Adapter, t *exql.Template) norm.SQL {
	return &sqlBuilder{
		Adapter: adapter,
		t:       newTemplateWithUtils(t),
	}
}

// WithTemplate returns a builder that is based on the given template.
// todo: delete me
func WithTemplate(t *exql.Template) norm.SQL {
	return &sqlBuilder{
		t: newTemplateWithUtils(t),
	}
}

func (b *sqlBuilder) NewIterator(rows *sql.Rows) norm.Iterator {
	return &iterator{b.Adapter, rows, nil}
}

func (b *sqlBuilder) Iterator(ctx context.Context, query interface{}, args ...interface{}) norm.Iterator {
	rows, err := b.Query(ctx, query, args...) //nolint:rowserrcheck
	return &iterator{b.Adapter, rows, err}
}

func (b *sqlBuilder) Prepare(ctx context.Context, query interface{}) (*sql.Stmt, error) {
	switch q := query.(type) {
	case *exql.Statement:
		return b.Executor().Prepare(ctx, q)
	case string:
		return b.Executor().Prepare(ctx, exql.RawSQL(q))
	case *expr.RawExpr:
		return b.Prepare(ctx, q.Raw())
	default:
		return nil, fmt.Errorf("unsupported query type %T", query)
	}
}

func (b *sqlBuilder) Exec(ctx context.Context, query interface{}, args ...interface{}) (sql.Result, error) {
	switch q := query.(type) {
	case *exql.Statement:
		return b.Executor().Exec(ctx, q, args...)
	case string:
		return b.Executor().Exec(ctx, exql.RawSQL(q), args...)
	case *expr.RawExpr:
		return b.Exec(ctx, q.Raw(), q.Arguments()...)
	default:
		return nil, fmt.Errorf("unsupported query type %T", query)
	}
}

func (b *sqlBuilder) Query(ctx context.Context, query interface{}, args ...interface{}) (*sql.Rows, error) {
	switch q := query.(type) {
	case *exql.Statement:
		return b.Executor().Query(ctx, q, args...)
	case string:
		return b.Executor().Query(ctx, exql.RawSQL(q), args...)
	case *expr.RawExpr:
		return b.Query(ctx, q.Raw(), q.Arguments()...)
	default:
		return nil, fmt.Errorf("unsupported query type %T", query)
	}
}

func (b *sqlBuilder) QueryRow(ctx context.Context, query interface{}, args ...interface{}) (*sql.Row, error) {
	switch q := query.(type) {
	case *exql.Statement:
		return b.Executor().QueryRow(ctx, q, args...)
	case string:
		return b.Executor().QueryRow(ctx, exql.RawSQL(q), args...)
	case *expr.RawExpr:
		return b.QueryRow(ctx, q.Raw(), q.Arguments()...)
	default:
		return nil, fmt.Errorf("unsupported query type %T", query)
	}
}

func (b *sqlBuilder) SelectFrom(tables ...interface{}) norm.Selector {
	qs := &selector{
		builder: b,
	}
	return qs.From(tables...)
}

func (b *sqlBuilder) Select(columns ...interface{}) norm.Selector {
	qs := &selector{
		builder: b,
	}
	return qs.Columns(columns...)
}

func (b *sqlBuilder) InsertInto(table string) norm.Inserter {
	qi := &inserter{
		builder: b,
	}
	return qi.Into(table)
}

func (b *sqlBuilder) DeleteFrom(table string) norm.Deleter {
	qd := &deleter{
		builder: b,
	}
	return qd.setTable(table)
}

func (b *sqlBuilder) Update(table string) norm.Updater {
	qu := &updater{
		builder: b,
	}
	return qu.setTable(table)
}

// Map receives a pointer to map or struct and maps it to columns and values.
func Map(item interface{}, options *MapOptions) ([]string, []interface{}, error) {
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
	case reflect.Struct:
		fieldMap := Mapper.TypeMap(itemT).Names
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
					fv.values = append(fv.values, sqlDefault)
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
				value = sqlDefault
			}
			fv.values = append(fv.values, value)
		}

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
	default:
		return nil, nil, ErrExpectingPointerToEitherMapOrStruct
	}

	sort.Sort(&fv)

	return fv.fields, fv.values, nil
}

func (iter *iterator) NextScan(dst ...interface{}) error {
	if ok := iter.Next(); ok {
		return iter.Scan(dst...)
	}
	if err := iter.Err(); err != nil {
		return err
	}
	return ErrNoMoreRows
}

func (iter *iterator) ScanOne(dst ...interface{}) error {
	defer iter.Close()
	return iter.NextScan(dst...)
}

func (iter *iterator) Scan(dst ...interface{}) error {
	if err := iter.Err(); err != nil {
		return err
	}
	return iter.cursor.Scan(dst...)
}

func (iter *iterator) setErr(err error) error {
	iter.err = err
	return iter.err
}

func (iter *iterator) One(ctx context.Context, dst interface{}) error {
	if err := iter.Err(); err != nil {
		return err
	}
	defer iter.Close()
	return iter.setErr(iter.next(dst))
}

func (iter *iterator) All(ctx context.Context, dst interface{}) error {
	if err := iter.Err(); err != nil {
		return err
	}
	defer iter.Close()

	// Fetching all results within the cursor.
	if err := fetchRows(iter.adapter.Typer(), iter, dst); err != nil {
		return iter.setErr(err)
	}

	return nil
}

func (iter *iterator) Err() (err error) {
	return iter.err
}

func (iter *iterator) Next(dst ...interface{}) bool {
	if err := iter.Err(); err != nil {
		return false
	}

	if err := iter.next(dst...); err != nil {
		// ignore ErrNoMoreRows, just break.
		if !errors.Is(err, ErrNoMoreRows) {
			_ = iter.setErr(err)
		}
		return false
	}

	return true
}

func (iter *iterator) next(dst ...interface{}) error {
	if iter.cursor == nil {
		return iter.setErr(ErrNoMoreRows)
	}

	switch len(dst) {
	case 0:
		if ok := iter.cursor.Next(); !ok {
			defer iter.Close()
			err := iter.cursor.Err()
			if err == nil {
				err = ErrNoMoreRows
			}
			return err
		}
		return nil
	case 1:
		if err := fetchRow(iter.adapter.Typer(), iter, dst[0]); err != nil {
			defer iter.Close()
			return err
		}
		return nil
	}

	return errors.New("Next does not currently supports more than one parameters")
}

func (iter *iterator) Close() error {
	if iter.cursor == nil {
		return nil
	}
	defer func() {
		iter.cursor = nil
	}()

	err := iter.cursor.Close()
	if err != nil {
		return err
	}
	return iter.cursor.Err()
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

var (
	_ = norm.SQL(&sqlBuilder{})
)

func joinArguments(args ...[]interface{}) []interface{} {
	total := 0
	for i := range args {
		total += len(args[i])
	}
	if total == 0 {
		return nil
	}

	flatten := make([]interface{}, 0, total)
	for i := range args {
		flatten = append(flatten, args[i]...)
	}
	return flatten
}
