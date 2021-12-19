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

type template struct {
	*exql.Template
}

func newTemplate(t *exql.Template) *template {
	return &template{Template: t}
}

func (t *template) PlaceholderValue(in interface{}) (exql.Fragment, []interface{}) {
	switch v := in.(type) {
	case *expr.RawExpr:
		return exql.RawValue(v.Raw()), v.Arguments()
	case *expr.FuncExpr:
		fnName := v.Name()
		fnArgs := []interface{}{}
		args, _ := toArguments(v.Arguments())
		fragments := []string{}
		for i := range args {
			frag, args := t.PlaceholderValue(args[i])
			fragment, err := frag.Compile(t.Template)
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

// toWhereClause converts the given parameters into an *exql.WhereFragment value.
func (t *template) toWhereClause(conds interface{}) (where *exql.WhereFragment, args []interface{}, err error) {
	switch v := conds.(type) {
	case []interface{}:
		if len(v) == 0 {
			return &exql.WhereFragment{}, []interface{}{}, nil
		}

		if s, ok := v[0].(string); ok {
			if strings.ContainsAny(s, "?") || len(v) == 1 {
				s, args, err = ExpandQuery(s, v[1:])
				if err != nil {
					return nil, nil, errors.Wrap(err, "preprocess []interface{}")
				}

				where = exql.WhereConditions(exql.RawValue(s))
				return where, args, nil
			}

			var val interface{}
			if len(v) > 2 {
				val = v[1:]
			} else {
				val = v[1]
			}
			cv, cvArgs, err := t.toColumnValues(expr.NewConstraint(s, val))
			if err != nil {
				return nil, nil, errors.Wrap(err, "convert []interface{} to column values")
			}

			where = exql.WhereConditions(cv.ColumnValues...)
			args = append(args, cvArgs...)
			return where, args, nil
		}

		var fragments []exql.Fragment
		for i := range v {
			w, wArgs, err := t.toWhereClause(v[i])
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
		r, rArgs, err := ExpandQuery(v.Raw(), v.Arguments())
		if err != nil {
			return nil, nil, errors.Wrap(err, "preprocess *expr.RawExpr")
		}

		where = exql.WhereConditions(exql.RawValue(r))
		args = append(args, rArgs...)
		return where, args, nil

	case expr.Constraints:
		var fragments []exql.Fragment
		for _, c := range v.Constraints() {
			w, wArgs, err := t.toWhereClause(c)
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
			w, wArgs, err := t.toWhereClause(e)
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
			return &exql.WhereFragment{}, []interface{}{}, nil
		}

		w := exql.WhereConditions(fragments...)
		if len(fragments) == 1 {
			return w, args, nil
		}

		var f exql.Fragment
		switch v.Operator() {
		case expr.LogicalNone, expr.LogicalAnd:
			q := exql.AndFragment(*w)
			f = &q
		case expr.LogicalOr:
			q := exql.OrFragment(*w)
			f = &q
		default:
			return nil, nil, errors.Errorf("unexpected logical operator %q", v.Operator())
		}
		where = exql.WhereConditions(f)
		return where, args, nil

	case expr.Constraint:
		cv, cvArgs, err := t.toColumnValues(v)
		if err != nil {
			return nil, nil, errors.Wrap(err, "convert expr.Constraint to column values")
		}

		where = exql.WhereConditions(cv.ColumnValues...)
		args = append(args, cvArgs...)
		return where, args, nil
	}
	return nil, nil, errors.Errorf("unexpected condition type %T", conds)
}

func (t *template) comparisonOperatorMapper(typ expr.ComparisonOperator) string {
	if typ == expr.ComparisonCustom {
		return ""
	}
	if t.ops != nil {
		if op, ok := t.ops[typ]; ok {
			return op
		}
	}
	if op, ok := comparisonOperators[typ]; ok {
		return op
	}
	panic(fmt.Sprintf("unsupported comparison operator %v", typ))
}

func (t *template) toColumnValues(term interface{}) (cv exql.ColumnValuesFragment, args []interface{}, err error) {
	switch v := term.(type) {
	case expr.Constraint:
		columnValue := exql.ColumnValueFragment{}

		// Getting column and operator.
		if column, ok := v.Key().(string); ok {
			chunks := strings.SplitN(strings.TrimSpace(column), " ", 2)
			columnValue.Column = exql.ColumnWithName(chunks[0])
			if len(chunks) > 1 {
				columnValue.Operator = chunks[1]
			}
		} else {
			if rawValue, ok := v.Key().(*expr.RawExpr); ok {
				columnValue.Column = exql.RawValue(rawValue.Raw())
				args = append(args, rawValue.Arguments()...)
			} else {
				columnValue.Column = exql.RawValue(fmt.Sprintf("%v", v.Key()))
			}
		}

		switch value := v.Value().(type) {
		case *expr.FuncExpr:
			fnName, fnArgs := value.Name(), value.Arguments()
			if len(fnArgs) == 0 {
				// A function with no arguments.
				fnName = fnName + "()"
			} else {
				// A function with one or more arguments.
				fnName = fnName + "(?" + strings.Repeat("?, ", len(fnArgs)-1) + ")"
			}
			fnName, fnArgs, err = ExpandQuery(fnName, fnArgs)
			if err != nil {
				return exql.ColumnValuesFragment{}, nil, errors.Wrap(err, "preprocess *expr.Constraint.FuncExpr")
			}

			columnValue.Value = exql.RawValue(fnName)
			args = append(args, fnArgs...)

		case *expr.RawExpr:
			q, a, err := ExpandQuery(value.Raw(), value.Arguments())
			if err != nil {
				return exql.ColumnValuesFragment{}, nil, errors.Wrap(err, "preprocess *expr.Constraint.RawExpr")
			}

			columnValue.Value = exql.RawValue(q)
			args = append(args, a...)

		case driver.Valuer:
			columnValue.Value = exql.RawValue("?")
			args = append(args, value)
		case *expr.Comparison:
			wrapper := &operatorWrapper{
				tu: t,
				cv: &columnValue,
				op: value,
			}

			q, a := wrapper.preprocess()
			q, a, err = ExpandQuery(q, a)
			if err != nil {
				return exql.ColumnValuesFragment{}, nil, errors.Wrap(err, "preprocess *expr.Comparison")
			}

			columnValue = exql.ColumnValueFragment{
				Column: exql.RawValue(q),
			}
			if a != nil {
				args = append(args, a...)
			}

			cv.ColumnValues = append(cv.ColumnValues, &columnValue)
			return cv, args, nil

		default:
			wrapper := &operatorWrapper{
				tu: t,
				cv: &columnValue,
				v:  value,
			}

			q, a := wrapper.preprocess()
			q, a, err = ExpandQuery(q, a)
			if err != nil {
				return exql.ColumnValuesFragment{}, nil, errors.Wrap(err, "preprocess default")
			}

			columnValue = exql.ColumnValueFragment{
				Column: exql.RawValue(q),
			}
			if a != nil {
				args = append(args, a...)
			}

			cv.ColumnValues = append(cv.ColumnValues, &columnValue)
			return cv, args, nil
		}

		if columnValue.Operator == "" {
			columnValue.Operator = t.comparisonOperatorMapper(expr.ComparisonEqual)
		}

		cv.ColumnValues = append(cv.ColumnValues, &columnValue)
		return cv, args, nil

	case *expr.RawExpr:
		columnValue := exql.ColumnValueFragment{}
		p, q, err := ExpandQuery(v.Raw(), v.Arguments())
		if err != nil {
			return exql.ColumnValuesFragment{}, nil, errors.Wrap(err, "preprocess *expr.RawExpr")
		}

		columnValue.Column = exql.RawValue(p)
		cv.ColumnValues = append(cv.ColumnValues, &columnValue)
		args = append(args, q...)
		return cv, args, nil

	case expr.Constraints:
		for _, constraint := range v.Constraints() {
			p, q, err := t.toColumnValues(constraint)
			if err != nil {
				return exql.ColumnValuesFragment{}, nil, errors.Wrap(err, "convert expr.Constraints to column values")
			}

			cv.ColumnValues = append(cv.ColumnValues, p.ColumnValues...)
			args = append(args, q...)
		}
		return cv, args, nil
	}

	panic(fmt.Sprintf("Unknown term type %T.", term))
}

func (t *template) setColumnValues(term interface{}) (cv exql.ColumnValuesFragment, args []interface{}, err error) {
	args = []interface{}{}

	switch v := term.(type) {
	case []interface{}:
		l := len(v)
		for i := 0; i < l; i++ {
			column, isString := v[i].(string)

			if !isString {
				p, q, err := t.setColumnValues(v[i])
				if err != nil {
					return exql.ColumnValuesFragment{}, nil, errors.Wrap(err, "set column values for []interface{}")
				}

				cv.ColumnValues = append(cv.ColumnValues, p.ColumnValues...)
				args = append(args, q...)
				continue
			}

			if !strings.ContainsAny(column, t.AssignmentOperator) {
				column = column + " " + t.AssignmentOperator + " ?"
			}

			chunks := strings.SplitN(column, t.AssignmentOperator, 2)

			column = chunks[0]
			format := strings.TrimSpace(chunks[1])

			columnValue := exql.ColumnValueFragment{
				Column:   exql.ColumnWithName(column),
				Operator: t.AssignmentOperator,
				Value:    exql.RawValue(format),
			}

			ps := strings.Count(format, "?")
			if i+ps < l {
				for j := 0; j < ps; j++ {
					args = append(args, v[i+j+1])
				}
				i = i + ps
			} else {
				panic(fmt.Sprintf("Format string %q has more placeholders than given arguments.", format))
			}

			cv.ColumnValues = append(cv.ColumnValues, &columnValue)
		}
		return cv, args, nil

	case *expr.RawExpr:
		columnValue := exql.ColumnValueFragment{}
		p, q, err := ExpandQuery(v.Raw(), v.Arguments())
		if err != nil {
			return exql.ColumnValuesFragment{}, nil, errors.Wrap(err, "preprocess *expr.RawExpr")
		}

		columnValue.Column = exql.RawValue(p)
		cv.ColumnValues = append(cv.ColumnValues, &columnValue)
		args = append(args, q...)
		return cv, args, nil
	}

	panic(fmt.Sprintf("Unknown term type %T.", term))
}
