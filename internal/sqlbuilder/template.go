// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package sqlbuilder

import (
	"database/sql/driver"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"unknwon.dev/norm/expr"
	"unknwon.dev/norm/internal/exql"
)

type templateWithUtils struct {
	*exql.Template
}

func newTemplateWithUtils(template *exql.Template) *templateWithUtils {
	return &templateWithUtils{template}
}

func (tu *templateWithUtils) PlaceholderValue(in interface{}) (exql.Fragment, []interface{}) {
	switch t := in.(type) {
	case *expr.RawExpr:
		return exql.RawValue(t.Raw()), t.Arguments()
	case *expr.FuncExpr:
		fnName := t.Name()
		fnArgs := []interface{}{}
		args, _ := toInterfaceArguments(t.Arguments())
		fragments := []string{}
		for i := range args {
			frag, args := tu.PlaceholderValue(args[i])
			fragment, err := frag.Compile(tu.Template)
			if err == nil {
				fragments = append(fragments, fragment)
				fnArgs = append(fnArgs, args...)
			}
		}
		return exql.RawValue(fnName + `(` + strings.Join(fragments, `, `) + `)`), fnArgs
	default:
		return sqlPlaceholder, []interface{}{in}
	}
}

// toWhereClause converts the given parameters into an *exql.Where value.
func (tu *templateWithUtils) toWhereClause(conds interface{}) (where *exql.Where, args []interface{}, err error) {
	switch v := conds.(type) {
	case []interface{}:
		if len(v) == 0 {
			return &exql.Where{}, []interface{}{}, nil
		}

		if s, ok := v[0].(string); ok {
			if strings.ContainsAny(s, "?") || len(v) == 1 {
				s, args = Preprocess(s, v[1:])
				where = exql.WhereConditions(exql.RawValue(s))
				return where, args, nil
			}

			var val interface{}
			if len(v) > 2 {
				val = v[1:]
			} else {
				val = v[1]
			}
			cv, cvArgs := tu.toColumnValues(expr.NewConstraint(s, val))
			where = exql.WhereConditions(cv.ColumnValues...)
			args = append(args, cvArgs...)
			return where, args, nil
		}

		var fragments []exql.Fragment
		for i := range v {
			w, wArgs, err := tu.toWhereClause(v[i])
			if err != nil {
				return nil, nil, err
			}
			if len(w.Conditions) == 0 {
				continue
			}
			fragments = append(fragments, w.Conditions...)
			args = append(args, wArgs...)
		}
		where = exql.WhereConditions(fragments...)
		return where, args, nil

	case *expr.RawExpr:
		r, rArgs := Preprocess(v.Raw(), v.Arguments())
		where = exql.WhereConditions(exql.RawValue(r))
		args = append(args, rArgs...)
		return where, args, nil

	case expr.Constraints:
		var fragments []exql.Fragment
		for _, c := range v.Constraints() {
			w, wArgs, err := tu.toWhereClause(c)
			if err != nil {
				return nil, nil, err
			}
			if len(w.Conditions) == 0 {
				continue
			}
			fragments = append(fragments, w.Conditions...)
			args = append(args, wArgs...)
		}
		where = exql.WhereConditions(fragments...)
		return where, args, nil

	case expr.LogicalExpr:
		var fragments []exql.Fragment
		for _, e := range v.Expressions() {
			w, wArgs, err := tu.toWhereClause(e)
			if err != nil {
				return nil, nil, err
			}
			if len(w.Conditions) == 0 {
				continue
			}
			fragments = append(fragments, w.Conditions...)
			args = append(args, wArgs...)
		}
		if len(fragments) == 0 {
			return &exql.Where{}, []interface{}{}, nil
		}

		w := exql.WhereConditions(fragments...)
		if len(fragments) == 1 {
			return w, args, nil
		}

		var f exql.Fragment
		switch v.Operator() {
		case expr.LogicalNone, expr.LogicalAnd:
			q := exql.And(*w)
			f = &q
		case expr.LogicalOr:
			q := exql.Or(*w)
			f = &q
		default:
			return nil, nil, errors.Errorf("unexpected logical operator %q", v.Operator())
		}
		where = exql.WhereConditions(f)
		return where, args, nil

	case expr.Constraint:
		cv, cvArgs := tu.toColumnValues(v)
		where = exql.WhereConditions(cv.ColumnValues...)
		args = append(args, cvArgs...)
		return where, args, nil
	}
	return nil, nil, errors.Errorf("unexpected condition type %T", conds)
}

func (tu *templateWithUtils) comparisonOperatorMapper(t expr.ComparisonOperator) string {
	if t == expr.ComparisonCustom {
		return ""
	}
	if tu.ComparisonOperator != nil {
		if op, ok := tu.ComparisonOperator[t]; ok {
			return op
		}
	}
	if op, ok := comparisonOperators[t]; ok {
		return op
	}
	panic(fmt.Sprintf("unsupported comparison operator %v", t))
}

func (tu *templateWithUtils) toColumnValues(term interface{}) (cv exql.ColumnValues, args []interface{}) {
	args = []interface{}{}

	switch t := term.(type) {
	case expr.Constraint:
		columnValue := exql.ColumnValue{}

		// Getting column and operator.
		if column, ok := t.Key().(string); ok {
			chunks := strings.SplitN(strings.TrimSpace(column), " ", 2)
			columnValue.Column = exql.ColumnWithName(chunks[0])
			if len(chunks) > 1 {
				columnValue.Operator = chunks[1]
			}
		} else {
			if rawValue, ok := t.Key().(*expr.RawExpr); ok {
				columnValue.Column = exql.RawValue(rawValue.Raw())
				args = append(args, rawValue.Arguments()...)
			} else {
				columnValue.Column = exql.RawValue(fmt.Sprintf("%v", t.Key()))
			}
		}

		switch value := t.Value().(type) {
		case *expr.FuncExpr:
			fnName, fnArgs := value.Name(), value.Arguments()
			if len(fnArgs) == 0 {
				// A function with no arguments.
				fnName = fnName + "()"
			} else {
				// A function with one or more arguments.
				fnName = fnName + "(?" + strings.Repeat("?, ", len(fnArgs)-1) + ")"
			}
			fnName, fnArgs = Preprocess(fnName, fnArgs)
			columnValue.Value = exql.RawValue(fnName)
			args = append(args, fnArgs...)
		case *expr.RawExpr:
			q, a := Preprocess(value.Raw(), value.Arguments())
			columnValue.Value = exql.RawValue(q)
			args = append(args, a...)
		case driver.Valuer:
			columnValue.Value = exql.RawValue("?")
			args = append(args, value)
		case *expr.Comparison:
			wrapper := &operatorWrapper{
				tu: tu,
				cv: &columnValue,
				op: value,
			}

			q, a := wrapper.preprocess()
			q, a = Preprocess(q, a)

			columnValue = exql.ColumnValue{
				Column: exql.RawValue(q),
			}
			if a != nil {
				args = append(args, a...)
			}

			cv.ColumnValues = append(cv.ColumnValues, &columnValue)
			return cv, args
		default:
			wrapper := &operatorWrapper{
				tu: tu,
				cv: &columnValue,
				v:  value,
			}

			q, a := wrapper.preprocess()
			q, a = Preprocess(q, a)

			columnValue = exql.ColumnValue{
				Column: exql.RawValue(q),
			}
			if a != nil {
				args = append(args, a...)
			}

			cv.ColumnValues = append(cv.ColumnValues, &columnValue)
			return cv, args
		}

		if columnValue.Operator == "" {
			columnValue.Operator = tu.comparisonOperatorMapper(expr.ComparisonEqual)
		}

		cv.ColumnValues = append(cv.ColumnValues, &columnValue)
		return cv, args

	case *expr.RawExpr:
		columnValue := exql.ColumnValue{}
		p, q := Preprocess(t.Raw(), t.Arguments())
		columnValue.Column = exql.RawValue(p)
		cv.ColumnValues = append(cv.ColumnValues, &columnValue)
		args = append(args, q...)
		return cv, args

	case expr.Constraints:
		for _, constraint := range t.Constraints() {
			p, q := tu.toColumnValues(constraint)
			cv.ColumnValues = append(cv.ColumnValues, p.ColumnValues...)
			args = append(args, q...)
		}
		return cv, args
	}

	panic(fmt.Sprintf("Unknown term type %T.", term))
}

func (tu *templateWithUtils) setColumnValues(term interface{}) (cv exql.ColumnValues, args []interface{}) {
	args = []interface{}{}

	switch t := term.(type) {
	case []interface{}:
		l := len(t)
		for i := 0; i < l; i++ {
			column, isString := t[i].(string)

			if !isString {
				p, q := tu.setColumnValues(t[i])
				cv.ColumnValues = append(cv.ColumnValues, p.ColumnValues...)
				args = append(args, q...)
				continue
			}

			if !strings.ContainsAny(column, tu.AssignmentOperator) {
				column = column + " " + tu.AssignmentOperator + " ?"
			}

			chunks := strings.SplitN(column, tu.AssignmentOperator, 2)

			column = chunks[0]
			format := strings.TrimSpace(chunks[1])

			columnValue := exql.ColumnValue{
				Column:   exql.ColumnWithName(column),
				Operator: tu.AssignmentOperator,
				Value:    exql.RawValue(format),
			}

			ps := strings.Count(format, "?")
			if i+ps < l {
				for j := 0; j < ps; j++ {
					args = append(args, t[i+j+1])
				}
				i = i + ps
			} else {
				panic(fmt.Sprintf("Format string %q has more placeholders than given arguments.", format))
			}

			cv.ColumnValues = append(cv.ColumnValues, &columnValue)
		}
		return cv, args
	case *expr.RawExpr:
		columnValue := exql.ColumnValue{}
		p, q := Preprocess(t.Raw(), t.Arguments())
		columnValue.Column = exql.RawValue(p)
		cv.ColumnValues = append(cv.ColumnValues, &columnValue)
		args = append(args, q...)
		return cv, args
	}

	panic(fmt.Sprintf("Unknown term type %T.", term))
}
