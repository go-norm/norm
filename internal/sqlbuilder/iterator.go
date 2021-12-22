// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package sqlbuilder

import (
	"context"
	"database/sql"
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
	defer func() { iter.cursor = nil }()

	err := iter.cursor.Close()
	if err != nil {
		return err
	}
	return iter.cursor.Err()
}

func (iter *iterator) All(ctx context.Context, dest interface{}) (err error) {
	if err = iter.Err(); err != nil {
		return err
	}
	defer func() { _ = iter.Close() }()

	if err = fetchRows(ctx, iter.adapter.Typer(), iter.cursor, dest); err != nil {
		return iter.setErr(err)
	}
	return nil
}

func (iter *iterator) next(dest interface{}) error {
	if iter.cursor == nil {
		return iter.setErr(sql.ErrNoRows)
	}

	if err := fetchRow(iter.adapter.Typer(), iter.cursor, dest); err != nil {
		defer func() { _ = iter.Close() }()
		return iter.setErr(err)
	}
	return nil
}

func (iter *iterator) One(_ context.Context, dest interface{}) (err error) {
	if err = iter.Err(); err != nil {
		return err
	}
	defer func() { _ = iter.Close() }()

	if err = iter.next(dest); err != nil {
		return iter.setErr(err)
	}
	return nil
}

// fetchRows maps all the rows coming from the *sql.Rows into the given
// destination. The typer is used to wrap custom types to satisfy sql.Scanner.
func fetchRows(ctx context.Context, typer adapter.Typer, rows *sql.Rows, dest interface{}) error {
	defer func() { _ = rows.Close() }()

	destv := reflect.ValueOf(dest)
	if destv.IsNil() || destv.Kind() != reflect.Ptr {
		return errors.New("the destination must be an pointer and cannot be nil")
	} else if destv.Elem().Kind() != reflect.Slice {
		return errors.New("the destination must be a slice")
	}

	columns, err := rows.Columns()
	if err != nil {
		return errors.Wrap(err, "get columns")
	}

	slicev := destv.Elem()
	itemT := slicev.Type().Elem()

	reset(dest)

	for rows.Next() {
		item, err := fetchResult(typer, rows, itemT, columns)
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

// todo

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

func fetchResult(typer adapter.Typer, rows *sql.Rows, itemT reflect.Type, columns []string) (reflect.Value, error) {
	var item reflect.Value
	var err error

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
			values[i] = typer.Scanner(f.Addr().Interface())
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

// fetchRow receives a *sql.Rows value and tries to map all the rows into a
// single struct given by the pointer `dst`.
func fetchRow(typer adapter.Typer, rows *sql.Rows, dest interface{}) error {
	var columns []string
	var err error

	dstv := reflect.ValueOf(dest)

	if dstv.IsNil() || dstv.Kind() != reflect.Ptr {
		return errors.New("the destination must be an pointer and cannot be nil")
	}

	itemV := dstv.Elem()

	if columns, err = rows.Columns(); err != nil {
		return err
	}

	reset(dest)

	next := rows.Next()

	if !next {
		if err = rows.Err(); err != nil {
			return err
		}
		return sql.ErrNoRows
	}

	itemT := itemV.Type()
	item, err := fetchResult(typer, rows, itemT, columns)
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
