// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package sqlbuilder

import (
	"database/sql"
	"database/sql/driver"
	"reflect"

	"unknwon.dev/norm/adapter"
	"unknwon.dev/norm/internal/reflectx"
)

type sessValueConverter interface {
	ConvertValue(interface{}) interface{}
}

type valueConverter interface {
	ConvertValue(in interface{}) (out interface {
		sql.Scanner
		driver.Valuer
	})
}

var Mapper = reflectx.NewMapper("db")

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

// fetchRows receives a *sql.Rows value and tries to map all the rows into a
// slice of structs given by the pointer `dst`.
func fetchRows(typer adapter.Typer, iter *iterator, dst interface{}) error {
	var err error
	rows := iter.cursor
	defer rows.Close()

	// Destination.
	dstv := reflect.ValueOf(dst)

	if dstv.IsNil() || dstv.Kind() != reflect.Ptr {
		return ErrExpectingPointer
	}

	if dstv.Elem().Kind() != reflect.Slice {
		return ErrExpectingSlicePointer
	}

	if dstv.Kind() != reflect.Ptr || dstv.Elem().Kind() != reflect.Slice || dstv.IsNil() {
		return ErrExpectingSliceMapStruct
	}

	var columns []string
	if columns, err = rows.Columns(); err != nil {
		return err
	}

	slicev := dstv.Elem()
	itemT := slicev.Type().Elem()

	reset(dst)

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

	dstv.Elem().Set(slicev)

	return rows.Err()
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
