// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package sqlbuilder

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"reflect"

	"github.com/pkg/errors"

	"unknwon.dev/norm"
	"unknwon.dev/norm/adapter"
	"unknwon.dev/norm/internal/reflectx"
)

var _ norm.Iterator = (*iterator)(nil)

type iterator struct {
	adapter adapter.Adapter
	cursor  *sql.Rows // This is the main query cursor. It starts as a nil value.
	err     error
}

func reset(data interface{}) {
	// Resetting element.
	v := reflect.ValueOf(data).Elem()
	t := v.Type()

	var z reflect.Value

	switch v.Kind() {
	case reflect.Slice:
		z = reflect.MakeSlice(t, 0, v.Cap())
	default:
		z = reflect.Zero(t)
	}

	v.Set(z)
}

var ErrExpectingMapOrStruct = errors.New("argument must be either a map or a struct")

// todo
var Mapper = reflectx.NewMapper("db")

// todo
type sessValueConverter interface {
	ConvertValue(interface{}) interface{}
}

// todo
type valueConverter interface {
	ConvertValue(in interface{}) (out interface {
		sql.Scanner
		driver.Valuer
	})
}

func fetchResult(typer adapter.Typer, iter *iterator, itemT reflect.Type, columns []string) (reflect.Value, error) {
	var item reflect.Value
	var err error
	rows := iter.cursor

	objT := itemT

	switch objT.Kind() {
	case reflect.Map:
		item = reflect.MakeMap(objT)
	case reflect.Struct:
		item = reflect.New(objT)
	case reflect.Ptr:
		objT = itemT.Elem()
		if objT.Kind() != reflect.Struct {
			return item, ErrExpectingMapOrStruct
		}
		item = reflect.New(objT)
	default:
		return item, ErrExpectingMapOrStruct
	}

	switch objT.Kind() {
	case reflect.Struct:
		values := make([]interface{}, len(columns))
		typeMap := Mapper.TypeMap(itemT)
		fieldMap := typeMap.Names

		for i, k := range columns {
			fi, ok := fieldMap[k]
			if !ok {
				values[i] = new(interface{})
				continue
			}

			f := reflectx.FieldByIndexes(item, fi.Index)

			// TODO: type switch + scanner

			if w, ok := f.Interface().(valueConverter); ok {
				wrapper := w.ConvertValue(f.Addr().Interface())
				z := reflect.ValueOf(wrapper)
				values[i] = z.Interface()
				continue
			} else {
				values[i] = f.Addr().Interface()
			}

			values[i] = typer.Scanner(values[i])

			// todo: what is this?
			if converter, ok := iter.adapter.(sessValueConverter); ok {
				values[i] = converter.ConvertValue(values[i])
				continue
			}
		}

		if err = rows.Scan(values...); err != nil {
			return item, err
		}

	case reflect.Map:
		columns, err := rows.Columns()
		if err != nil {
			return item, err
		}

		values := make([]interface{}, len(columns))
		for i := range values {
			if itemT.Elem().Kind() == reflect.Interface {
				values[i] = new(interface{})
			} else {
				values[i] = reflect.New(itemT.Elem()).Interface()
			}
		}

		if err = rows.Scan(values...); err != nil {
			return item, err
		}

		for i, column := range columns {
			item.SetMapIndex(reflect.ValueOf(column), reflect.Indirect(reflect.ValueOf(values[i])))
		}
	}

	return item, nil
}

var (
	ErrExpectingPointer        = errors.New("argument must be an address")
	ErrExpectingSlicePointer   = errors.New("argument must be a slice address")
	ErrExpectingSliceMapStruct = errors.New("argument must be a slice address of maps or structs")
)

// fetchRows receives a *sql.Rows value and tries to map all the rows into a
// slice of structs given by the pointer `dst`.
func fetchRows(typer adapter.Typer, iter *iterator, dest interface{}) error {
	rows := iter.cursor
	defer func() { _ = rows.Close() }()

	// Destination.
	destv := reflect.ValueOf(dest)

	if destv.IsNil() || destv.Kind() != reflect.Ptr {
		return ErrExpectingPointer
	}

	if destv.Elem().Kind() != reflect.Slice {
		return ErrExpectingSlicePointer
	}

	if destv.Kind() != reflect.Ptr || destv.Elem().Kind() != reflect.Slice || destv.IsNil() {
		return ErrExpectingSliceMapStruct
	}

	var err error
	var columns []string
	if columns, err = rows.Columns(); err != nil {
		return err
	}

	slicev := destv.Elem()
	itemT := slicev.Type().Elem()

	reset(dest)

	for rows.Next() {
		item, err := fetchResult(typer, iter, itemT, columns)
		if err != nil {
			return err
		}
		if itemT.Kind() == reflect.Ptr {
			slicev = reflect.Append(slicev, item)
		} else {
			slicev = reflect.Append(slicev, reflect.Indirect(item))
		}
	}

	destv.Elem().Set(slicev)

	return rows.Err()
}

func (iter *iterator) setErr(err error) error {
	iter.err = err
	return iter.err
}

func (iter *iterator) Err() (err error) {
	return iter.err
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

// todo: respect context
func (iter *iterator) All(ctx context.Context, dst interface{}) error {
	if err := iter.Err(); err != nil {
		return err
	}
	defer func() { _ = iter.Close() }()

	// Fetching all results within the cursor.
	if err := fetchRows(iter.adapter.Typer(), iter, dst); err != nil {
		return iter.setErr(err)
	}

	return nil
}

var ErrNoMoreRows = errors.New("no more rows in the result set")

// fetchRow receives a *sql.Rows value and tries to map all the rows into a
// single struct given by the pointer `dst`.
func fetchRow(typer adapter.Typer, iter *iterator, dst interface{}) error {
	var columns []string
	var err error

	rows := iter.cursor

	dstv := reflect.ValueOf(dst)

	if dstv.IsNil() || dstv.Kind() != reflect.Ptr {
		return ErrExpectingPointer
	}

	itemV := dstv.Elem()

	if columns, err = rows.Columns(); err != nil {
		return err
	}

	reset(dst)

	next := rows.Next()

	if !next {
		if err = rows.Err(); err != nil {
			return err
		}
		return ErrNoMoreRows
	}

	itemT := itemV.Type()
	item, err := fetchResult(typer, iter, itemT, columns)
	if err != nil {
		return err
	}

	if itemT.Kind() == reflect.Ptr {
		itemV.Set(item)
	} else {
		itemV.Set(reflect.Indirect(item))
	}

	return nil
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

// todo: respect context
func (iter *iterator) One(ctx context.Context, dst interface{}) error {
	if err := iter.Err(); err != nil {
		return err
	}
	defer func() { _ = iter.Close() }()
	return iter.setErr(iter.next(dst))
}
