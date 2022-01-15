// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package sqlbuilder

import (
	"database/sql/driver"
	"strings"

	"github.com/pkg/errors"

	"unknwon.dev/norm/expr"
	"unknwon.dev/norm/internal/exql"
)

// todo
// toWhere converts the given conditions into an *exql.WhereFragment value.
func toWhere(t *exql.Template, conditions interface{}) (where *exql.WhereFragment, args []interface{}, err error) {
	switch v := conditions.(type) {
	case []interface{}:
		if len(v) == 0 {
			return exql.Where(), []interface{}{}, nil
		}

		if s, ok := v[0].(string); ok {
			if strings.ContainsAny(s, "?") || len(v) == 1 {
				s, args, err = ExpandQuery(s, v[1:])
				if err != nil {
					return nil, nil, errors.Wrap(err, "expand query for []interface{}")
				}

				where = exql.Where(exql.Raw(s))
				return where, args, nil
			}

			var val interface{}
			if len(v) > 2 {
				val = v[1:]
			} else {
				val = v[1]
			}
			conds, condsArgs, err := toConditions(t, expr.NewConstraint(s, val))
			if err != nil {
				return nil, nil, errors.Wrap(err, "convert []interface{} to conditions")
			}

			where = exql.Where(conds...)
			args = append(args, condsArgs...)
			return where, args, nil
		}

		var fragments []exql.Fragment
		for i := range v {
			w, wArgs, err := toWhere(t, v[i])
			if err != nil {
				return nil, nil, err
			}
			if len(w.Conditions) == 0 {
				continue
			}
			fragments = append(fragments, w.Conditions...)
			args = append(args, wArgs...)
		}
		where = exql.Where(fragments...)
		return where, args, nil

	case compilable:
		q, qArgs, err := expandCompilable(v)
		if err != nil {
			return nil, nil, errors.Wrap(err, "expand compilable")
		}

		where = exql.Where(exql.Raw(q))
		args = append(args, qArgs...)
		return where, args, nil

	case *expr.FuncExpr:
		fnName, fnArgs, err := expandFuncExpr(v)
		if err != nil {
			return nil, nil, errors.Wrap(err, "expand *expr.FuncExpr")
		}

		where = exql.Where(exql.Raw(fnName))
		args = append(args, fnArgs...)
		return where, args, nil

	case *expr.RawExpr:
		r, rArgs, err := ExpandQuery(v.Raw(), v.Arguments())
		if err != nil {
			return nil, nil, errors.Wrap(err, "expand query for *expr.RawExpr")
		}

		where = exql.Where(exql.Raw(r))
		args = append(args, rArgs...)
		return where, args, nil

	case expr.Constraints:
		var fragments []exql.Fragment
		for _, c := range v.Constraints() {
			w, wArgs, err := toWhere(t, c)
			if err != nil {
				return nil, nil, err
			}
			if len(w.Conditions) == 0 {
				continue
			}
			fragments = append(fragments, w.Conditions...)
			args = append(args, wArgs...)
		}
		where = exql.Where(fragments...)
		return where, args, nil

	case expr.Constraint:
		conds, condsArgs, err := toConditions(t, v)
		if err != nil {
			return nil, nil, errors.Wrap(err, "convert expr.Constraint to conditions")
		}

		where = exql.Where(conds...)
		args = append(args, condsArgs...)
		return where, args, nil

	case expr.LogicalExpr:
		// CAUTION: This case must be after "expr.Constraints" to avoid infinite loop on
		// `expr.Cond` which satisfies both.

		var fragments []exql.Fragment
		for _, e := range v.Expressions() {
			w, wArgs, err := toWhere(t, e)
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
			return exql.Where(), nil, nil
		}

		w := exql.Where(fragments...)
		if len(fragments) == 1 {
			return w, args, nil
		}

		var f exql.Fragment
		switch v.Operator() {
		case expr.LogicalNone, expr.LogicalAnd:
			f = exql.And(w.Conditions...)
		case expr.LogicalOr:
			f = exql.Or(w.Conditions...)
		default:
			return nil, nil, errors.Errorf("unexpected logical operator %q", v.Operator())
		}
		where = exql.Where(f)
		return where, args, nil

	case exql.Fragment:
		where = exql.Where(v)
		return where, args, nil

	case string, int:
		where = exql.Where(exql.Value(v))
		return where, args, nil
	}
	return nil, nil, errors.Errorf("unsupported condition type %T", conditions)
}

// todo
func toConditions(t *exql.Template, expression interface{}) (conditions []exql.Fragment, args []interface{}, err error) {
	switch v := expression.(type) {
	case *expr.RawExpr:
		q, qArgs, err := ExpandQuery(v.Raw(), v.Arguments())
		if err != nil {
			return nil, nil, errors.Wrap(err, "expand query for *expr.RawExpr")
		}

		cv := exql.ColumnValue(q, expr.ComparisonCustom, nil)
		conditions = append(conditions, cv)
		args = append(args, qArgs...)
		return conditions, args, nil

	case expr.Constraints:
		for _, constraint := range v.Constraints() {
			conds, condsArgs, err := toConditions(t, constraint)
			if err != nil {
				return nil, nil, errors.Wrap(err, "convert expr.Constraints to column values")
			}

			conditions = append(conditions, conds...)
			args = append(args, condsArgs...)
		}
		return conditions, args, nil

	case expr.Constraint:
		var column string
		var operator interface{} = expr.ComparisonEqual
		var value exql.Fragment

		switch key := v.Key().(type) {
		case string:
			chunks := strings.SplitN(strings.TrimSpace(key), " ", 2)
			column = chunks[0]
			if len(chunks) > 1 {
				operator = chunks[1]
			}

		case *expr.RawExpr:
			column = key.Raw()
			args = append(args, key.Arguments()...)

		default:
			return nil, nil, errors.Errorf("unsupported expr.Constraint.Key() type %T", key)
		}

		switch val := v.Value().(type) {
		case *expr.FuncExpr:
			fnName, fnArgs, err := expandFuncExpr(val)
			if err != nil {
				return nil, nil, errors.Wrap(err, "expand *expr.FuncExpr")
			}

			value = exql.Raw(fnName)
			args = append(args, fnArgs...)

		case *expr.RawExpr:
			r, rArgs, err := ExpandQuery(val.Raw(), val.Arguments())
			if err != nil {
				return nil, nil, errors.Wrap(err, "expand query for *expr.RawExpr")
			}

			value = exql.Raw(r)
			args = append(args, rArgs...)

		case driver.Valuer:
			value = exql.Raw("?")
			args = append(args, val)

		case *expr.Comparison:
			cmpOp, cmpPlaceholder, cmpArgs, err := expandComparison(t, val)
			if err != nil {
				return nil, nil, errors.Wrap(err, "expand comparison")
			}
			operator = cmpOp
			value = exql.Raw(cmpPlaceholder)
			args = append(args, cmpArgs...)

		default:
			return nil, nil, errors.Errorf("unsupported expr.Constraint.Value() type %T", val)
		}

		conditions = append(conditions, exql.ColumnValue(column, operator, value))
		return conditions, args, nil
	}
	return nil, nil, errors.Errorf("unsupported expression type %T", expression)
}

// todo
func expandComparison(t *exql.Template, cmp *expr.Comparison) (operator, placeholder string, args []interface{}, err error) {
	op := cmp.Operator()
	operator = t.Operator(op)
	placeholder = "?"
	switch op {
	case expr.ComparisonCustom:
		operator = cmp.CustomOperator()
		args = []interface{}{cmp.Value()}

	case
		expr.ComparisonEqual,
		expr.ComparisonNotEqual,
		expr.ComparisonLessThan,
		expr.ComparisonGreaterThan,
		expr.ComparisonLessThanOrEqualTo,
		expr.ComparisonGreaterThanOrEqualTo,
		expr.ComparisonLike,
		expr.ComparisonNotLike,
		expr.ComparisonRegexp,
		expr.ComparisonNotRegexp:
		args = []interface{}{cmp.Value()}

	case expr.ComparisonBetween, expr.ComparisonNotBetween:
		values := cmp.Value().([]interface{})
		placeholder = "? AND ?"
		args = []interface{}{values[0], values[1]}

	case expr.ComparisonIn, expr.ComparisonNotIn:
		values := cmp.Value().([]interface{})
		if len(values) < 1 {
			placeholder = "(NULL)"
			break
		}
		placeholder = "(?" + strings.Repeat(", ?", len(values)-1) + ")"
		args = values

	case expr.ComparisonIs, expr.ComparisonIsNot:
		v := cmp.Value()
		switch v {
		case nil:
			placeholder = "NULL"
		case false:
			placeholder = "FALSE"
		case true:
			placeholder = "TRUE"
		default:
			return "", "", nil, errors.Errorf("unsupported value type %T for expr.ComparisonIs", v)
		}

	default:
		return "", "", nil, errors.Errorf("unexpected operator %v", op)
	}
	return operator, placeholder, args, nil
}
