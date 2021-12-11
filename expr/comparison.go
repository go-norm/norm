// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package expr

import (
	"reflect"
	"time"
)

// ComparisonOperator is a comparison operator.
type ComparisonOperator uint8

const (
	ComparisonOperatorNone ComparisonOperator = iota
	ComparisonOperatorCustom

	ComparisonOperatorEqual
	ComparisonOperatorNotEqual

	ComparisonOperatorLessThan
	ComparisonOperatorGreaterThan

	ComparisonOperatorLessThanOrEqualTo
	ComparisonOperatorGreaterThanOrEqualTo

	ComparisonOperatorBetween
	ComparisonOperatorNotBetween

	ComparisonOperatorIn
	ComparisonOperatorNotIn

	ComparisonOperatorIs
	ComparisonOperatorIsNot

	ComparisonOperatorLike
	ComparisonOperatorNotLike

	ComparisonOperatorRegexp
	ComparisonOperatorNotRegexp
)

// Comparison represents the relationship between values.
type Comparison struct {
	op     ComparisonOperator
	custom string // The custom operator when not empty.
	value  interface{}
}

// CustomOperator returns the custom operator of the comparison.
func (c *Comparison) CustomOperator() string {
	return c.custom
}

// Operator returns the ComparisonOperator.
func (c *Comparison) Operator() ComparisonOperator {
	return c.op
}

// Value returns the value of the comparison.
func (c *Comparison) Value() interface{} {
	return c.value
}

func newComparisonOperator(op ComparisonOperator, v interface{}) *Comparison {
	return &Comparison{
		op:    op,
		value: v,
	}
}

func newCustomComparisonOperator(op string, v interface{}) *Comparison {
	return &Comparison{
		op:     ComparisonOperatorCustom,
		custom: op,
		value:  v,
	}
}

// Gte is a comparison that means: is greater than or equal to the value.
func Gte(value interface{}) *Comparison {
	return newComparisonOperator(ComparisonOperatorGreaterThanOrEqualTo, value)
}

// Lte is a comparison that means: is less than or equal to the value.
func Lte(value interface{}) *Comparison {
	return newComparisonOperator(ComparisonOperatorLessThanOrEqualTo, value)
}

// Eq is a comparison that means: is equal to the value.
func Eq(value interface{}) *Comparison {
	return newComparisonOperator(ComparisonOperatorEqual, value)
}

// NotEq is a comparison that means: is not equal to the value.
func NotEq(value interface{}) *Comparison {
	return newComparisonOperator(ComparisonOperatorNotEqual, value)
}

// Gt is a comparison that means: is greater than the value.
func Gt(value interface{}) *Comparison {
	return newComparisonOperator(ComparisonOperatorGreaterThan, value)
}

// Lt is a comparison that means: is less than the value.
func Lt(value interface{}) *Comparison {
	return newComparisonOperator(ComparisonOperatorLessThan, value)
}

// In is a comparison that means: is any of the values.
func In(value ...interface{}) *Comparison {
	return newComparisonOperator(ComparisonOperatorIn, toInterfaceArray(value))
}

// NotIn is a comparison that means: is none of the values.
func NotIn(value ...interface{}) *Comparison {
	return newComparisonOperator(ComparisonOperatorNotIn, toInterfaceArray(value))
}

// AnyOf is a comparison that means: is any of the values of the slice.
func AnyOf(value interface{}) *Comparison {
	return newComparisonOperator(ComparisonOperatorIn, toInterfaceArray(value))
}

// NotAnyOf is a comparison that means: is none of the values of the slice.
func NotAnyOf(value interface{}) *Comparison {
	return newComparisonOperator(ComparisonOperatorNotIn, toInterfaceArray(value))
}

// After is a comparison that means: is after the time.
func After(value time.Time) *Comparison {
	return newComparisonOperator(ComparisonOperatorGreaterThan, value)
}

// Before is a comparison that means: is before the time.
func Before(value time.Time) *Comparison {
	return newComparisonOperator(ComparisonOperatorLessThan, value)
}

// OnOrAfter is a comparison that means: is on or after the time.
func OnOrAfter(value time.Time) *Comparison {
	return newComparisonOperator(ComparisonOperatorGreaterThanOrEqualTo, value)
}

// OnOrBefore is a comparison that means: is on or before the time.
func OnOrBefore(value time.Time) *Comparison {
	return newComparisonOperator(ComparisonOperatorLessThanOrEqualTo, value)
}

// Between is a comparison that means: is between lower and upper bound.
func Between(lower, upper interface{}) *Comparison {
	return newComparisonOperator(ComparisonOperatorBetween, []interface{}{lower, upper})
}

// NotBetween is a comparison that means: is not between lower and upper bound.
func NotBetween(lower, upper interface{}) *Comparison {
	return newComparisonOperator(ComparisonOperatorNotBetween, []interface{}{lower, upper})
}

// Is is a comparison that means: is equivalent to nil, true or false.
func Is(value interface{}) *Comparison {
	return newComparisonOperator(ComparisonOperatorIs, value)
}

// IsNot is a comparison that means: is not equivalent to nil, true nor false.
func IsNot(value interface{}) *Comparison {
	return newComparisonOperator(ComparisonOperatorIsNot, value)
}

// IsNull is a comparison that means: is equivalent to nil.
func IsNull() *Comparison {
	return newComparisonOperator(ComparisonOperatorIs, nil)
}

// IsNotNull is a comparison that means: is not equivalent to nil.
func IsNotNull() *Comparison {
	return newComparisonOperator(ComparisonOperatorIsNot, nil)
}

// Like is a comparison that checks whether the reference matches the wildcard
// of the value.
func Like(value string) *Comparison {
	return newComparisonOperator(ComparisonOperatorLike, value)
}

// NotLike is a comparison that checks whether the reference does not match the
// wildcard of the value.
func NotLike(value string) *Comparison {
	return newComparisonOperator(ComparisonOperatorNotLike, value)
}

// Regexp is a comparison that checks whether the reference matches the regular
// expression.
func Regexp(value string) *Comparison {
	return newComparisonOperator(ComparisonOperatorRegexp, value)
}

// NotRegexp is a comparison that checks whether the reference does not match
// the regular expression.
func NotRegexp(value string) *Comparison {
	return newComparisonOperator(ComparisonOperatorNotRegexp, value)
}

// Op returns a comparison with the custom operator.
func Op(op string, value interface{}) *Comparison {
	return newCustomComparisonOperator(op, value)
}

func toInterfaceArray(v interface{}) []interface{} {
	vv := reflect.ValueOf(v)
	switch vv.Type().Kind() {
	case reflect.Ptr:
		return toInterfaceArray(vv.Elem().Interface())
	case reflect.Slice:
		elems := vv.Len()
		args := make([]interface{}, elems)
		for i := 0; i < elems; i++ {
			args[i] = vv.Index(i).Interface()
		}
		return args
	}
	return []interface{}{v}
}
