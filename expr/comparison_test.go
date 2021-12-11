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
	assert.Equal(t, ComparisonCustom, c.Operator())
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
			want: newComparisonOperator(ComparisonEqual, 6),
			got:  Eq(6),
		},
		{
			name: "NotEq",
			want: newComparisonOperator(ComparisonNotEqual, 67),
			got:  NotEq(67),
		},

		{
			name: "Lt",
			want: newComparisonOperator(ComparisonLessThan, 47),
			got:  Lt(47),
		},
		{
			name: "Gt",
			want: newComparisonOperator(ComparisonGreaterThan, 4),
			got:  Gt(4),
		},

		{
			name: "Lte",
			want: newComparisonOperator(ComparisonLessThanOrEqualTo, 22),
			got:  Lte(22),
		},
		{
			name: "Gte",
			want: newComparisonOperator(ComparisonGreaterThanOrEqualTo, 1),
			got:  Gte(1),
		},

		{
			name: "Between",
			want: newComparisonOperator(ComparisonBetween, []interface{}{11, 35}),
			got:  Between(11, 35),
		},
		{
			name: "NotBetween",
			want: newComparisonOperator(ComparisonNotBetween, []interface{}{11, 35}),
			got:  NotBetween(11, 35),
		},

		{
			name: "In",
			want: newComparisonOperator(ComparisonIn, []interface{}{1, 22, 34}),
			got:  In(1, 22, 34),
		},
		{
			name: "NotIn",
			want: newComparisonOperator(ComparisonNotIn, []interface{}{1, 22, 34}),
			got:  NotIn(1, 22, 34),
		},
		{
			name: "AnyOf",
			want: newComparisonOperator(ComparisonIn, []interface{}{1, 22, 34}),
			got:  AnyOf([]interface{}{1, 22, 34}),
		},
		{
			name: "AnyOf",
			want: newComparisonOperator(ComparisonIn, []interface{}{1}),
			got:  AnyOf(1),
		},
		{
			name: "AnyOf",
			want: newComparisonOperator(ComparisonIn, []interface{}{1}),
			got:  AnyOf(1),
		},
		{
			name: "NotAnyOf",
			want: newComparisonOperator(ComparisonNotIn, []interface{}{1, 22, 34}),
			got:  NotAnyOf([]interface{}{1, 22, 34}),
		},
		{
			name: "NotAnyOf",
			want: newComparisonOperator(ComparisonNotIn, []interface{}{now}),
			got:  NotAnyOf(&now),
		},

		{
			name: "Is",
			want: newComparisonOperator(ComparisonIs, 178),
			got:  Is(178),
		},
		{
			name: "IsNot",
			want: newComparisonOperator(ComparisonIsNot, 32),
			got:  IsNot(32),
		},

		{
			name: "Like",
			want: newComparisonOperator(ComparisonLike, "%a%"),
			got:  Like("%a%"),
		},
		{
			name: "NotLike",
			want: newComparisonOperator(ComparisonNotLike, "%z%"),
			got:  NotLike("%z%"),
		},

		{
			name: "Regexp",
			want: newComparisonOperator(ComparisonRegexp, ".*"),
			got:  Regexp(".*"),
		},
		{
			name: "NotRegexp",
			want: newComparisonOperator(ComparisonNotRegexp, ".*"),
			got:  NotRegexp(".*"),
		},

		{
			name: "After",
			want: newComparisonOperator(ComparisonGreaterThan, now),
			got:  After(now),
		},
		{
			name: "Before",
			want: newComparisonOperator(ComparisonLessThan, now),
			got:  Before(now),
		},
		{
			name: "OnOrAfter",
			want: newComparisonOperator(ComparisonGreaterThanOrEqualTo, now),
			got:  OnOrAfter(now),
		},
		{
			name: "OnOrBefore",
			want: newComparisonOperator(ComparisonLessThanOrEqualTo, now),
			got:  OnOrBefore(now),
		},

		{
			name: "IsNull",
			got:  newComparisonOperator(ComparisonIs, nil),
			want: IsNull(),
		},
		{
			name: "IsNotNull",
			got:  newComparisonOperator(ComparisonIsNot, nil),
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
