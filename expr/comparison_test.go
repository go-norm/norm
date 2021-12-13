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
			want: newComparison(ComparisonEqual, 6),
			got:  Eq(6),
		},
		{
			name: "NotEq",
			want: newComparison(ComparisonNotEqual, 67),
			got:  NotEq(67),
		},

		{
			name: "Lt",
			want: newComparison(ComparisonLessThan, 47),
			got:  Lt(47),
		},
		{
			name: "Gt",
			want: newComparison(ComparisonGreaterThan, 4),
			got:  Gt(4),
		},

		{
			name: "Lte",
			want: newComparison(ComparisonLessThanOrEqualTo, 22),
			got:  Lte(22),
		},
		{
			name: "Gte",
			want: newComparison(ComparisonGreaterThanOrEqualTo, 1),
			got:  Gte(1),
		},

		{
			name: "Between",
			want: newComparison(ComparisonBetween, []interface{}{11, 35}),
			got:  Between(11, 35),
		},
		{
			name: "NotBetween",
			want: newComparison(ComparisonNotBetween, []interface{}{11, 35}),
			got:  NotBetween(11, 35),
		},

		{
			name: "In",
			want: newComparison(ComparisonIn, []interface{}{1, 22, 34}),
			got:  In(1, 22, 34),
		},
		{
			name: "NotIn",
			want: newComparison(ComparisonNotIn, []interface{}{1, 22, 34}),
			got:  NotIn(1, 22, 34),
		},
		{
			name: "AnyOf",
			want: newComparison(ComparisonIn, []interface{}{1, 22, 34}),
			got:  AnyOf([]interface{}{1, 22, 34}),
		},
		{
			name: "AnyOf",
			want: newComparison(ComparisonIn, []interface{}{1}),
			got:  AnyOf(1),
		},
		{
			name: "AnyOf",
			want: newComparison(ComparisonIn, []interface{}{1}),
			got:  AnyOf(1),
		},
		{
			name: "NotAnyOf",
			want: newComparison(ComparisonNotIn, []interface{}{1, 22, 34}),
			got:  NotAnyOf([]interface{}{1, 22, 34}),
		},
		{
			name: "NotAnyOf",
			want: newComparison(ComparisonNotIn, []interface{}{now}),
			got:  NotAnyOf(&now),
		},

		{
			name: "Is",
			want: newComparison(ComparisonIs, 178),
			got:  Is(178),
		},
		{
			name: "IsNot",
			want: newComparison(ComparisonIsNot, 32),
			got:  IsNot(32),
		},

		{
			name: "Like",
			want: newComparison(ComparisonLike, "%a%"),
			got:  Like("%a%"),
		},
		{
			name: "NotLike",
			want: newComparison(ComparisonNotLike, "%z%"),
			got:  NotLike("%z%"),
		},

		{
			name: "Regexp",
			want: newComparison(ComparisonRegexp, ".*"),
			got:  Regexp(".*"),
		},
		{
			name: "NotRegexp",
			want: newComparison(ComparisonNotRegexp, ".*"),
			got:  NotRegexp(".*"),
		},

		{
			name: "After",
			want: newComparison(ComparisonGreaterThan, now),
			got:  After(now),
		},
		{
			name: "Before",
			want: newComparison(ComparisonLessThan, now),
			got:  Before(now),
		},
		{
			name: "OnOrAfter",
			want: newComparison(ComparisonGreaterThanOrEqualTo, now),
			got:  OnOrAfter(now),
		},
		{
			name: "OnOrBefore",
			want: newComparison(ComparisonLessThanOrEqualTo, now),
			got:  OnOrBefore(now),
		},

		{
			name: "IsNull",
			got:  newComparison(ComparisonIs, nil),
			want: IsNull(),
		},
		{
			name: "IsNotNull",
			got:  newComparison(ComparisonIsNot, nil),
			want: IsNotNull(),
		},

		{
			name: "Op",
			got:  newCustomComparison("~", 56),
			want: Op("~", 56),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.want, test.got)
		})
	}
}
