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

// ExpandQuery flattens elements of the arguments list to their indivisible
// forms (e.g. primitive types, []byte, driver.Valuer) and expands placeholders
// ("?") in the query accordingly.
func ExpandQuery(query string, args []interface{}) (string, []interface{}, error) {
	// Quick path for queries without arguments
	if len(args) == 0 {
		return query, args, nil
	}

	var expandedQuery bytes.Buffer
	expandedArgs := make([]interface{}, 0, len(args))
	for i := range query {
		// Write out the rest of the query if no argument remaining
		if len(args) == 0 {
			expandedQuery.WriteString(query[i:])
			break
		}

		// Only look for expanding placeholders and ignore others
		if query[i] != '?' {
			expandedQuery.WriteByte(query[i])
			continue
		}

		p, pArgs, err := expandArgument(args[0])
		if err != nil {
			return "", nil, errors.Wrap(err, "expand argument")
		}

		expandedQuery.WriteString(p)
		expandedArgs = append(expandedArgs, pArgs...)
		args = args[1:]
	}
	return expandedQuery.String(), expandedArgs, nil
}

// expandArgument derives the placeholder and wrapped list of arguments from the
// given argument.
func expandArgument(arg interface{}) (placeholder string, args []interface{}, err error) {
	if arg == nil {
		return "NULL", nil, nil
	}

	if valuer, ok := arg.(driver.Valuer); ok {
		return "?", []interface{}{valuer}, nil
	}

	varg := reflect.ValueOf(arg)
	if varg.Type().Kind() == reflect.Slice {
		// Byte slice preserves the original type
		if varg.Type().Elem().Kind() == reflect.Uint8 {
			return "?", []interface{}{arg}, nil
		}
		if varg.Len() == 0 {
			return "(NULL)", nil, nil
		}

		vals := make([]interface{}, varg.Len())
		for i := range vals {
			vals[i] = varg.Index(i).Interface()
		}
		placeholder = "(?" + strings.Repeat(", ?", len(vals)-1) + ")"
		return placeholder, vals, nil
	}

	switch v := arg.(type) {
	case compilable:
		placeholder, args, err = expandCompilable(v)
		if err != nil {
			return "", nil, errors.Wrap(err, "expand compilable")
		}

	case *expr.FuncExpr:
		placeholder, args, err = expandFuncExpr(v)
		if err != nil {
			return "", nil, errors.Wrap(err, "expand *expr.FuncExpr")
		}

	case *expr.RawExpr:
		placeholder, args, err = ExpandQuery(v.Raw(), v.Arguments())
		if err != nil {
			return "", nil, errors.Wrap(err, "expand query for *expr.RawExpr")
		}

	default:
		placeholder = "?"
		args = []interface{}{arg}
	}
	return placeholder, args, nil
}

// expandCompilable derives the placeholder and its arguments from the
// compilable.
func expandCompilable(c compilable) (placeholder string, args []interface{}, err error) {
	placeholder, err = c.Compile()
	if err != nil {
		return "", nil, errors.Wrap(err, "compile")
	}

	placeholder, args, err = ExpandQuery(placeholder, c.Arguments())
	if err != nil {
		return "", nil, errors.Wrap(err, "expand query")
	}
	return "(" + placeholder + ")", args, nil
}

// expandFuncExpr derives the placeholder and its arguments from the
// expr.FuncExpr.
func expandFuncExpr(e *expr.FuncExpr) (placeholder string, args []interface{}, err error) {
	fnName, fnArgs := e.Name(), e.Arguments()
	if len(fnArgs) == 0 {
		fnName = fnName + "()"
	} else {
		fnName = fnName + "(?" + strings.Repeat("?, ", len(fnArgs)-1) + ")"
	}
	placeholder, args, err = ExpandQuery(fnName, fnArgs)
	if err != nil {
		return "", nil, errors.Wrap(err, "expand query")
	}
	return placeholder, args, nil
}
