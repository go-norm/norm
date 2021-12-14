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

// toInterfaceArguments converts the given value into an array of interfaces.
func toInterfaceArguments(value interface{}) (args []interface{}, isSlice bool) {
	if value == nil {
		return nil, false
	}

	switch t := value.(type) {
	case driver.Valuer:
		return []interface{}{t}, false
	}

	v := reflect.ValueOf(value)
	if v.Type().Kind() == reflect.Slice {
		// Byte slice gets transformed into a string.
		if v.Type().Elem().Kind() == reflect.Uint8 {
			return []interface{}{string(value.([]byte))}, false
		}

		total := v.Len()
		args = make([]interface{}, total)
		for i := 0; i < total; i++ {
			args[i] = v.Index(i).Interface()
		}
		return args, true
	}
	return []interface{}{value}, false
}

// todo
func expandPlaceholders(arg interface{}) (placeholder string, args []interface{}, err error) {
	vals, isSlice := toInterfaceArguments(arg)
	if isSlice {
		if len(vals) == 0 {
			return `(NULL)`, nil, nil
		}
		placeholder = `(?` + strings.Repeat(`, ?`, len(vals)-1) + `)`
		return placeholder, vals, nil
	}

	if len(vals) == 0 {
		return `NULL`, nil, nil
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
			placeholder = `(` + q + `)`
			return placeholder, t.Arguments(), nil
		}
	}
	return "", []interface{}{arg}, nil
}

// todo: ExpandQuery expands arguments that need to be expanded and compiles a query
// into a single string.
func ExpandQuery(query string, args []interface{}) (string, []interface{}, error) {
	argn := 0
	argx := make([]interface{}, 0, len(args))
	for i := 0; i < len(query); i++ {
		if query[i] != '?' {
			continue
		}
		if len(args) > argn {
			k, vals, err := expandPlaceholders(args[argn])
			if err != nil {
				return "", nil, errors.Wrap(err, "expand argument")
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
