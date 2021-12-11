// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package expr

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestComparison(t *testing.T) {
	c := Op("test", 1)
	assert.Equal(t, "test", c.CustomOperator())
	assert.Equal(t, ComparisonOperatorCustom, c.Operator())
	assert.Equal(t, 1, c.Value())
}

func TestComparisonOperator(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name string
		want *Comparison
		got  *Comparison
	}{
		{
			name: "Eq",
			want: newComparisonOperator(ComparisonOperatorEqual, 6),
			got:  Eq(6),
		},
		{
			name: "NotEq",
			want: newComparisonOperator(ComparisonOperatorNotEqual, 67),
			got:  NotEq(67),
		},

		{
			name: "Lt",
			want: newComparisonOperator(ComparisonOperatorLessThan, 47),
			got:  Lt(47),
		},
		{
			name: "Gt",
			want: newComparisonOperator(ComparisonOperatorGreaterThan, 4),
			got:  Gt(4),
		},

		{
			name: "Lte",
			want: newComparisonOperator(ComparisonOperatorLessThanOrEqualTo, 22),
			got:  Lte(22),
		},
		{
			name: "Gte",
			want: newComparisonOperator(ComparisonOperatorGreaterThanOrEqualTo, 1),
			got:  Gte(1),
		},

		{
			name: "Between",
			want: newComparisonOperator(ComparisonOperatorBetween, []interface{}{11, 35}),
			got:  Between(11, 35),
		},
		{
			name: "NotBetween",
			want: newComparisonOperator(ComparisonOperatorNotBetween, []interface{}{11, 35}),
			got:  NotBetween(11, 35),
		},

		{
			name: "In",
			want: newComparisonOperator(ComparisonOperatorIn, []interface{}{1, 22, 34}),
			got:  In(1, 22, 34),
		},
		{
			name: "NotIn",
			want: newComparisonOperator(ComparisonOperatorNotIn, []interface{}{1, 22, 34}),
			got:  NotIn(1, 22, 34),
		},
		{
			name: "AnyOf",
			want: newComparisonOperator(ComparisonOperatorIn, []interface{}{1, 22, 34}),
			got:  AnyOf([]interface{}{1, 22, 34}),
		},
		{
			name: "AnyOf",
			want: newComparisonOperator(ComparisonOperatorIn, []interface{}{1}),
			got:  AnyOf(1),
		},
		{
			name: "AnyOf",
			want: newComparisonOperator(ComparisonOperatorIn, []interface{}{1}),
			got:  AnyOf(1),
		},
		{
			name: "NotAnyOf",
			want: newComparisonOperator(ComparisonOperatorNotIn, []interface{}{1, 22, 34}),
			got:  NotAnyOf([]interface{}{1, 22, 34}),
		},
		{
			name: "NotAnyOf",
			want: newComparisonOperator(ComparisonOperatorNotIn, []interface{}{now}),
			got:  NotAnyOf(&now),
		},

		{
			name: "Is",
			want: newComparisonOperator(ComparisonOperatorIs, 178),
			got:  Is(178),
		},
		{
			name: "IsNot",
			want: newComparisonOperator(ComparisonOperatorIsNot, 32),
			got:  IsNot(32),
		},

		{
			name: "Like",
			want: newComparisonOperator(ComparisonOperatorLike, "%a%"),
			got:  Like("%a%"),
		},
		{
			name: "NotLike",
			want: newComparisonOperator(ComparisonOperatorNotLike, "%z%"),
			got:  NotLike("%z%"),
		},

		{
			name: "Regexp",
			want: newComparisonOperator(ComparisonOperatorRegexp, ".*"),
			got:  Regexp(".*"),
		},
		{
			name: "NotRegexp",
			want: newComparisonOperator(ComparisonOperatorNotRegexp, ".*"),
			got:  NotRegexp(".*"),
		},

		{
			name: "After",
			want: newComparisonOperator(ComparisonOperatorGreaterThan, now),
			got:  After(now),
		},
		{
			name: "Before",
			want: newComparisonOperator(ComparisonOperatorLessThan, now),
			got:  Before(now),
		},
		{
			name: "OnOrAfter",
			want: newComparisonOperator(ComparisonOperatorGreaterThanOrEqualTo, now),
			got:  OnOrAfter(now),
		},
		{
			name: "OnOrBefore",
			want: newComparisonOperator(ComparisonOperatorLessThanOrEqualTo, now),
			got:  OnOrBefore(now),
		},

		{
			name: "IsNull",
			got:  newComparisonOperator(ComparisonOperatorIs, nil),
			want: IsNull(),
		},
		{
			name: "IsNotNull",
			got:  newComparisonOperator(ComparisonOperatorIsNot, nil),
			want: IsNotNull(),
		},

		{
			name: "Op",
			got:  newCustomComparisonOperator("~", 56),
			want: Op("~", 56),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.want, test.got)
		})
	}
}
