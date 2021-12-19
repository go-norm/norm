// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package sqlbuilder

import (
	"fmt"
	"strings"

	"unknwon.dev/norm/expr"
	"unknwon.dev/norm/internal/exql"
)

var comparisonOperators = map[expr.ComparisonOperator]string{
	expr.ComparisonEqual:    "=",
	expr.ComparisonNotEqual: "!=",

	expr.ComparisonLessThan:    "<",
	expr.ComparisonGreaterThan: ">",

	expr.ComparisonLessThanOrEqualTo:    "<=",
	expr.ComparisonGreaterThanOrEqualTo: ">=",

	expr.ComparisonBetween:    "BETWEEN",
	expr.ComparisonNotBetween: "NOT BETWEEN",

	expr.ComparisonIn:    "IN",
	expr.ComparisonNotIn: "NOT IN",

	expr.ComparisonIs:    "IS",
	expr.ComparisonIsNot: "IS NOT",

	expr.ComparisonLike:    "LIKE",
	expr.ComparisonNotLike: "NOT LIKE",

	expr.ComparisonRegexp:    "REGEXP",
	expr.ComparisonNotRegexp: "NOT REGEXP",
}

type operatorWrapper struct {
	tu *template
	cv *exql.ColumnValueFragment

	op *expr.Comparison
	v  interface{}
}

func (ow *operatorWrapper) cmp() *expr.Comparison {
	if ow.op != nil {
		return ow.op
	}

	if ow.cv.Operator != "" {
		return expr.Op(ow.cv.Operator, ow.v)
	}

	if ow.v == nil {
		return expr.Is(nil)
	}

	args, isSlice := toArguments(ow.v)
	if isSlice {
		return expr.In(args...)
	}

	return expr.Eq(ow.v)
}

func (ow *operatorWrapper) preprocess() (string, []interface{}) {
	placeholder := "?"

	column, err := ow.cv.Column.Compile(ow.tu.Template)
	if err != nil {
		panic(fmt.Sprintf("could not compile column: %v", err.Error()))
	}

	c := ow.cmp()

	op := ow.tu.comparisonOperatorMapper(c.Operator())

	var args []interface{}

	switch c.Operator() {
	case expr.ComparisonNone:
		panic("no operator given")
	case expr.ComparisonCustom:
		op = c.CustomOperator()
	case expr.ComparisonIn, expr.ComparisonNotIn:
		values := c.Value().([]interface{})
		if len(values) < 1 {
			placeholder, args = "(NULL)", []interface{}{}
			break
		}
		placeholder, args = "(?"+strings.Repeat(", ?", len(values)-1)+")", values
	case expr.ComparisonIs, expr.ComparisonIsNot:
		switch c.Value() {
		case nil:
			placeholder, args = "NULL", []interface{}{}
		case false:
			placeholder, args = "FALSE", []interface{}{}
		case true:
			placeholder, args = "TRUE", []interface{}{}
		}
	case expr.ComparisonBetween, expr.ComparisonNotBetween:
		values := c.Value().([]interface{})
		placeholder, args = "? AND ?", []interface{}{values[0], values[1]}
	case expr.ComparisonEqual:
		v := c.Value()
		if b, ok := v.([]byte); ok {
			v = string(b)
		}
		args = []interface{}{v}
	}

	if args == nil {
		args = []interface{}{c.Value()}
	}

	if strings.Contains(op, ":column") {
		return strings.Replace(op, ":column", column, -1), args
	}

	return column + " " + op + " " + placeholder, args
}
