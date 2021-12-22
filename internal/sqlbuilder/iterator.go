// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package sqlbuilder

import (
	"context"
	"database/sql"
	"io"
	"reflect"

	"github.com/pkg/errors"

	"unknwon.dev/norm"
	"unknwon.dev/norm/adapter"
	"unknwon.dev/norm/internal/reflectx"
)

var _ norm.Iterator = (*iterator)(nil)

type iterator struct {
	adapter adapter.Adapter
	cursor  cursor
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

type cursor interface {
	io.Closer
	Columns() ([]string, error)
	Err() error
	Next() bool
	Scan(dest ...interface{}) error
}

func reset(v interface{}) {
	elem := reflect.ValueOf(v).Elem()
	typ := elem.Type()

	var zero reflect.Value
	switch elem.Kind() {
	case reflect.Slice:
		zero = reflect.MakeSlice(typ, 0, elem.Cap())
	default:
		zero = reflect.Zero(typ)
	}

	elem.Set(zero)
}

var defaultMapper = reflectx.NewMapper("db")

func scanResult(typer adapter.Typer, rows cursor, typ reflect.Type, columns []string) (result reflect.Value, err error) {
	switch typ.Kind() {
	case reflect.Map:
		result = reflect.MakeMap(typ)
	case reflect.Struct:
		result = reflect.New(typ)
	case reflect.Ptr:
		elem := typ.Elem()
		if elem.Kind() != reflect.Struct {
			return reflect.Value{}, errors.New("the type must be a map or struct or a pointer to a map or struct")
		}
		result = reflect.New(elem)
	default:
		return reflect.Value{}, errors.New("the type must be a map or struct or a pointer to a map or struct")
	}

	switch typ.Kind() {
	case reflect.Struct:
		values := make([]interface{}, len(columns))
		typeMap := defaultMapper.TypeMap(typ)
		fieldMap := typeMap.Names
		for i, k := range columns {
			fi, ok := fieldMap[k]
			if !ok {
				values[i] = new(interface{})
				continue
			}

			f := reflectx.FieldByIndexes(result, fi.Index)
			values[i] = typer.Scanner(f.Addr().Interface())
		}

		if err = rows.Scan(values...); err != nil {
			return reflect.Value{}, errors.Wrap(err, "scan")
		}
		return result, nil

	case reflect.Map:
		values := make([]interface{}, len(columns))
		for i := range values {
			if typ.Elem().Kind() == reflect.Interface {
				values[i] = new(interface{})
			} else {
				values[i] = reflect.New(typ.Elem()).Interface()
			}
		}

		if err = rows.Scan(values...); err != nil {
			return reflect.Value{}, errors.Wrap(err, "scan")
		}

		for i, column := range columns {
			result.SetMapIndex(reflect.ValueOf(column), reflect.Indirect(reflect.ValueOf(values[i])))
		}
		return result, nil
	}
	panic("unreachable")
}

// fetchRows maps all the rows coming from the *sql.Rows into the given
// destination. The typer is used to wrap custom types to satisfy sql.Scanner.
func fetchRows(ctx context.Context, typer adapter.Typer, rows cursor, dest interface{}) error {
	defer func() { _ = rows.Close() }()

	destv := reflect.ValueOf(dest)
	if destv.IsNil() || destv.Kind() != reflect.Ptr {
		return errors.New("the destination must be an pointer and cannot be nil")
	} else if destv.Elem().Kind() != reflect.Slice {
		return errors.New("the destination must be a slice")
	}
	reset(dest)

	columns, err := rows.Columns()
	if err != nil {
		return errors.Wrap(err, "get columns")
	}

	elem := destv.Elem()
	typ := elem.Type() // .Elem()
	var item reflect.Value
	for rows.Next() {
		select {
		case err = <-ctx.Done():
			return err
		default:
		}

		item, err = scanResult(typer, rows, typ, columns)
		if err != nil {
			return errors.Wrap(err, "scan result")
		}

		if typ.Kind() == reflect.Ptr {
			elem = reflect.Append(elem, item)
		} else {
			elem = reflect.Append(elem, reflect.Indirect(item))
		}
	}
	return rows.Err()
}

// fetchRow maps the next row coming from the *sql.Rows into the given
// destination. The typer is used to wrap custom types to satisfy sql.Scanner.
func fetchRow(typer adapter.Typer, rows cursor, dest interface{}) error {
	destv := reflect.ValueOf(dest)
	if destv.IsNil() || destv.Kind() != reflect.Ptr {
		return errors.New("the destination must be an pointer and cannot be nil")
	}
	reset(dest)

	columns, err := rows.Columns()
	if err != nil {
		return errors.Wrap(err, "get columns")
	}

	if !rows.Next() {
		if err = rows.Err(); err != nil {
			return err
		}
		return sql.ErrNoRows
	}

	elem := destv.Elem()
	typ := elem.Type() // .Elem()
	item, err := scanResult(typer, rows, typ, columns)
	if err != nil {
		return errors.Wrap(err, "scan result")
	}

	if typ.Kind() == reflect.Ptr {
		elem.Set(item)
	} else {
		elem.Set(reflect.Indirect(item))
	}
	// destv.Elem().Set(elem)
	return nil
}
