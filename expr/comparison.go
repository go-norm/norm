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
	ComparisonNone ComparisonOperator = iota
	ComparisonCustom

	ComparisonEqual
	ComparisonNotEqual

	ComparisonLessThan
	ComparisonGreaterThan

	ComparisonLessThanOrEqualTo
	ComparisonGreaterThanOrEqualTo

	ComparisonBetween
	ComparisonNotBetween

	ComparisonIn
	ComparisonNotIn

	ComparisonIs
	ComparisonIsNot

	ComparisonLike
	ComparisonNotLike

	ComparisonRegexp
	ComparisonNotRegexp
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
		op:     ComparisonCustom,
		custom: op,
		value:  v,
	}
}

// Gte is a comparison that means: is greater than or equal to the value.
func Gte(value interface{}) *Comparison {
	return newComparisonOperator(ComparisonGreaterThanOrEqualTo, value)
}

// Lte is a comparison that means: is less than or equal to the value.
func Lte(value interface{}) *Comparison {
	return newComparisonOperator(ComparisonLessThanOrEqualTo, value)
}

// Eq is a comparison that means: is equal to the value.
func Eq(value interface{}) *Comparison {
	return newComparisonOperator(ComparisonEqual, value)
}

// NotEq is a comparison that means: is not equal to the value.
func NotEq(value interface{}) *Comparison {
	return newComparisonOperator(ComparisonNotEqual, value)
}

// Gt is a comparison that means: is greater than the value.
func Gt(value interface{}) *Comparison {
	return newComparisonOperator(ComparisonGreaterThan, value)
}

// Lt is a comparison that means: is less than the value.
func Lt(value interface{}) *Comparison {
	return newComparisonOperator(ComparisonLessThan, value)
}

// In is a comparison that means: is any of the values.
func In(value ...interface{}) *Comparison {
	return newComparisonOperator(ComparisonIn, toInterfaceArray(value))
}

// NotIn is a comparison that means: is none of the values.
func NotIn(value ...interface{}) *Comparison {
	return newComparisonOperator(ComparisonNotIn, toInterfaceArray(value))
}

// AnyOf is a comparison that means: is any of the values of the slice.
func AnyOf(value interface{}) *Comparison {
	return newComparisonOperator(ComparisonIn, toInterfaceArray(value))
}

// NotAnyOf is a comparison that means: is none of the values of the slice.
func NotAnyOf(value interface{}) *Comparison {
	return newComparisonOperator(ComparisonNotIn, toInterfaceArray(value))
}

// After is a comparison that means: is after the time.
func After(value time.Time) *Comparison {
	return newComparisonOperator(ComparisonGreaterThan, value)
}

// Before is a comparison that means: is before the time.
func Before(value time.Time) *Comparison {
	return newComparisonOperator(ComparisonLessThan, value)
}

// OnOrAfter is a comparison that means: is on or after the time.
func OnOrAfter(value time.Time) *Comparison {
	return newComparisonOperator(ComparisonGreaterThanOrEqualTo, value)
}

// OnOrBefore is a comparison that means: is on or before the time.
func OnOrBefore(value time.Time) *Comparison {
	return newComparisonOperator(ComparisonLessThanOrEqualTo, value)
}

// Between is a comparison that means: is between lower and upper bound.
func Between(lower, upper interface{}) *Comparison {
	return newComparisonOperator(ComparisonBetween, []interface{}{lower, upper})
}

// NotBetween is a comparison that means: is not between lower and upper bound.
func NotBetween(lower, upper interface{}) *Comparison {
	return newComparisonOperator(ComparisonNotBetween, []interface{}{lower, upper})
}

// Is is a comparison that means: is equivalent to nil, true or false.
func Is(value interface{}) *Comparison {
	return newComparisonOperator(ComparisonIs, value)
}

// IsNot is a comparison that means: is not equivalent to nil, true nor false.
func IsNot(value interface{}) *Comparison {
	return newComparisonOperator(ComparisonIsNot, value)
}

// IsNull is a comparison that means: is equivalent to nil.
func IsNull() *Comparison {
	return newComparisonOperator(ComparisonIs, nil)
}

// IsNotNull is a comparison that means: is not equivalent to nil.
func IsNotNull() *Comparison {
	return newComparisonOperator(ComparisonIsNot, nil)
}

// Like is a comparison that checks whether the reference matches the wildcard
// of the value.
func Like(value string) *Comparison {
	return newComparisonOperator(ComparisonLike, value)
}

// NotLike is a comparison that checks whether the reference does not match the
// wildcard of the value.
func NotLike(value string) *Comparison {
	return newComparisonOperator(ComparisonNotLike, value)
}

// Regexp is a comparison that checks whether the reference matches the regular
// expression.
func Regexp(value string) *Comparison {
	return newComparisonOperator(ComparisonRegexp, value)
}

// NotRegexp is a comparison that checks whether the reference does not match
// the regular expression.
func NotRegexp(value string) *Comparison {
	return newComparisonOperator(ComparisonNotRegexp, value)
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
