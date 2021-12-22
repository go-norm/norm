// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package sqlbuilder

import (
	"bytes"
	"database/sql/driver"
	"reflect"
	"strings"

	"github.com/pkg/errors"

	"unknwon.dev/norm/expr"
)

// toArguments wraps the given value into an array of interfaces that can be
// used to pass as query arguments.
//
// The `isSlice` returns true when the value is a slice. The returned arguments
// list would only contain one element when the `isSlice` returns false.
func toArguments(v interface{}) (args []interface{}, isSlice bool) {
	if v == nil {
		return nil, false
	}

	if valuer, ok := v.(driver.Valuer); ok {
		return []interface{}{valuer}, false
	}

	vv := reflect.ValueOf(v)

	// Non-slice and byte slice preserves the original type
	if vv.Type().Kind() != reflect.Slice || vv.Type().Elem().Kind() == reflect.Uint8 {
		return []interface{}{v}, false
	}

	args = make([]interface{}, vv.Len())
	for i := range args {
		args[i] = vv.Index(i).Interface()
	}
	return args, true
}

// expandArgument expands the placeholder string and wrapped list of arguments
// from the given argument. It returns an empty string for the placeholder and a
// slice containing the original argument if expansion is not possible.
func expandArgument(arg interface{}) (placeholder string, args []interface{}, err error) {
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
	}

	switch v := vals[0].(type) {
	case compilable:
		q, err := v.Compile()
		if err != nil {
			return "", nil, errors.Wrap(err, "compile")
		}
		placeholder = "(" + q + ")"
		return placeholder, v.Arguments(), nil

	case *expr.FuncExpr:
		fnName, fnArgs := v.Name(), v.Arguments()
		if len(fnArgs) == 0 {
			fnName = fnName + "()"
		} else {
			fnName = fnName + "(?" + strings.Repeat("?, ", len(fnArgs)-1) + ")"
		}
		placeholder, args, err = ExpandQuery(fnName, fnArgs)
		if err != nil {
			return "", nil, errors.Wrap(err, "expand query for *expr.FuncExpr")
		}
		return placeholder, args, nil

	case *expr.RawExpr:
		placeholder, args, err = ExpandQuery(v.Raw(), v.Arguments())
		if err != nil {
			return "", nil, errors.Wrap(err, "expand query for *expr.RawExpr")
		}
		return placeholder, args, nil
	}
	return "", vals, nil
}

// ExpandQuery expands the query with given arguments with necessary
// placeholders.
func ExpandQuery(query string, args []interface{}) (string, []interface{}, error) {
	var buf bytes.Buffer
	argx := make([]interface{}, 0, len(args))
	for i := range query {
		if query[i] != '?' || len(args) == 0 {
			buf.WriteByte(query[i])
			continue
		}

		q, qArgs, err := expandArgument(args[0])
		if err != nil {
			return "", nil, errors.Wrap(err, "expand argument")
		}

		if q != "" {
			buf.WriteString(q)
		} else {
			buf.WriteByte(query[i])
		}
		argx = append(argx, qArgs...)
		args = args[1:]
	}
	return buf.String(), argx, nil
}
