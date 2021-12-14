// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package sqlbuilder

import (
	"database/sql/driver"
	"reflect"
	"strings"

	"github.com/pkg/errors"

	"unknwon.dev/norm/expr"
)

// toArguments wraps the given value into an array of interfaces that can be
// used to pass as query arguments. The `isSlice` returns true when the value is
// a slice.
func toArguments(v interface{}) (args []interface{}, isSlice bool) {
	if v == nil {
		return nil, false
	}

	if valuer, ok := v.(driver.Valuer); ok {
		return []interface{}{valuer}, false
	}

	vv := reflect.ValueOf(v)
	if vv.Type().Kind() != reflect.Slice {
		return []interface{}{v}, false
	}

	// Byte slice gets transformed into a string.
	if vv.Type().Elem().Kind() == reflect.Uint8 {
		return []interface{}{string(v.([]byte))}, false
	}

	args = make([]interface{}, vv.Len())
	for i := range args {
		args[i] = vv.Index(i).Interface()
	}
	return args, true

}

// expandPlaceholder expands the placeholder string and wrapped list of
// arguments from the given argument. It returns an empty string for the
// placeholder and a slice containing the original argument if expansion is not
// possible.
func expandPlaceholder(arg interface{}) (placeholder string, args []interface{}, err error) {
	vals, isSlice := toArguments(arg)
	if isSlice {
		if len(vals) == 0 {
			return "(NULL)", nil, nil
		}
		placeholder = "(?" + strings.Repeat(", ?", len(vals)-1) + ")"
		return placeholder, vals, nil
	}

	if len(vals) == 0 {
		return "NULL", nil, nil
	} else if len(vals) == 1 {
		switch t := arg.(type) {
		case *expr.RawExpr:
			placeholder, args, err = ExpandQuery(t.Raw(), t.Arguments())
			if err != nil {
				return "", nil, errors.Wrap(err, "expand query for *expr.RawExpr")
			}
			return placeholder, args, nil

		case compilable:
			q, err := t.Compile()
			if err != nil {
				return "", nil, errors.Wrap(err, "compile")
			}
			placeholder = "(" + q + ")"
			return placeholder, t.Arguments(), nil
		}
	}
	return "", []interface{}{arg}, nil
}

// ExpandQuery expands the query with given arguments with necessary
// placeholders.
func ExpandQuery(query string, args []interface{}) (string, []interface{}, error) {
	argn := 0
	argx := make([]interface{}, 0, len(args))
	for i := 0; i < len(query); i++ {
		if query[i] != '?' {
			continue
		}
		if len(args) > argn {
			k, vals, err := expandPlaceholder(args[argn])
			if err != nil {
				return "", nil, errors.Wrap(err, "expand placeholder")
			}

			k, vals, err = ExpandQuery(k, vals)
			if err != nil {
				return "", nil, errors.Wrap(err, "expand query")
			}

			if k != "" {
				query = query[:i] + k + query[i+1:]
				i += len(k) - 1
			}
			if len(vals) > 0 {
				argx = append(argx, vals...)
			}
			argn++
		}
	}
	if len(argx) < len(args) {
		argx = append(argx, args[argn:]...)
	}
	return query, argx, nil
}
