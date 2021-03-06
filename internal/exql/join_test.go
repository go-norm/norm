// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package exql

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"unknwon.dev/norm/expr"
)

func TestJoin(t *testing.T) {
	tmpl := defaultTemplate(t)

	t.Run("no table", func(t *testing.T) {
		got, err := Join(nil).Compile(tmpl)
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	tests := []struct {
		name string
		join *JoinFragment
		want string
	}{
		{
			name: "natural",
			join: Join("users"),
			want: `NATURAL JOIN "users"`,
		},
		{
			name: "natural full",
			join: JoinUsing(FullJoin, "users", nil),
			want: `NATURAL FULL JOIN "users"`,
		},
		{
			name: "full",
			join: JoinUsing(FullJoin, "users", Using(Column("users.id"))),
			want: `FULL JOIN "users" USING ("users"."id")`,
		},
		{
			name: "cross",
			join: JoinUsing(CrossJoin, "users", Using(Column("users.id"))),
			want: `CROSS JOIN "users" USING ("users"."id")`,
		},
		{
			name: "right",
			join: JoinUsing(RightJoin, "users", Using(Column("users.id"))),
			want: `RIGHT JOIN "users" USING ("users"."id")`,
		},
		{
			name: "left",
			join: JoinUsing(LeftJoin, "users", Using(Column("users.id"))),
			want: `LEFT JOIN "users" USING ("users"."id")`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := test.join.Compile(tmpl)
			require.NoError(t, err)
			assert.Equal(t, test.want, StripWhitespace(got))
		})
	}
}

func TestJoinOn(t *testing.T) {
	tmpl := defaultTemplate(t)
	join := JoinOn(
		DefaultJoin,
		"countries c",
		On(
			ColumnValue("p.country_id", expr.ComparisonEqual, Column("a.id")),
			ColumnValue("p.country_code", expr.ComparisonEqual, Column("a.code")),
		),
	)

	got, err := join.Compile(tmpl)
	require.NoError(t, err)

	want := StripWhitespace(`
JOIN "countries" AS "c" ON (
		"p"."country_id" = "a"."id"
	AND "p"."country_code" = "a"."code"
)`)
	assert.Equal(t, want, StripWhitespace(got))

	t.Run("cache hit", func(t *testing.T) {
		got, err := join.Compile(tmpl)
		assert.NoError(t, err)
		assert.Equal(t, want, StripWhitespace(got))
	})
}

func TestJoinUsing(t *testing.T) {
	tmpl := defaultTemplate(t)
	join := JoinUsing(
		DefaultJoin,
		"countries c",
		Using(
			Column("p.country_id"),
			Column("p.country_code"),
		),
	)

	got, err := join.Compile(tmpl)
	require.NoError(t, err)

	want := `JOIN "countries" AS "c" USING ("p"."country_id", "p"."country_code")`
	assert.Equal(t, want, StripWhitespace(got))

	t.Run("cache hit", func(t *testing.T) {
		got, err := join.Compile(tmpl)
		assert.NoError(t, err)
		assert.Equal(t, want, StripWhitespace(got))
	})
}

func TestOn(t *testing.T) {
	tmpl := defaultTemplate(t)

	t.Run("empty", func(t *testing.T) {
		got, err := On().Compile(tmpl)
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	on := On(
		And(
			ColumnValue("id", expr.ComparisonGreaterThan, Raw("8")),
			ColumnValue("id", expr.ComparisonLessThan, Raw("99")),
			Or(
				ColumnValue("age", expr.ComparisonLessThan, Raw("18")),
				ColumnValue("age", expr.ComparisonGreaterThan, Raw("41")),
			),
		),
		ColumnValue("name", expr.ComparisonEqual, Raw(`'John'`)),
		Or(
			ColumnValue("last_name", expr.ComparisonEqual, Raw(`'Smith'`)),
			ColumnValue("last_name", expr.ComparisonEqual, Raw(`'Reyes'`)),
		),
		Raw("city_id = 728"),
	)

	got, err := on.Compile(tmpl)
	require.NoError(t, err)

	want := StripWhitespace(`
ON (
		(
				"id" > 8
			AND "id" < 99
			AND ("age" < 18 OR "age" > 41)
		)
	AND "name" = 'John'
	AND ("last_name" = 'Smith' OR "last_name" = 'Reyes')
	AND city_id = 728
)`)
	assert.Equal(t, want, strings.TrimSpace(got))

	t.Run("cache hit", func(t *testing.T) {
		got, err := on.Compile(tmpl)
		assert.NoError(t, err)
		assert.Equal(t, want, strings.TrimSpace(got))
	})
}

func TestUsing(t *testing.T) {
	tmpl := defaultTemplate(t)

	t.Run("empty", func(t *testing.T) {
		got, err := Using().Compile(tmpl)
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	using := Using(
		Column("id"),
		Column("customer"),
		Column("service_id"),
		Column("users.name"),
		Column("users.id"),
	)

	got, err := using.Compile(tmpl)
	require.NoError(t, err)

	want := `USING ("id", "customer", "service_id", "users"."name", "users"."id")`
	assert.Equal(t, want, strings.TrimSpace(got))

	t.Run("cache hit", func(t *testing.T) {
		got, err := using.Compile(tmpl)
		assert.NoError(t, err)
		assert.Equal(t, want, strings.TrimSpace(got))
	})
}

func TestJoins(t *testing.T) {
	tmpl := defaultTemplate(t)

	t.Run("empty", func(t *testing.T) {
		got, err := Joins().Compile(tmpl)
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	js := Joins(
		Join("users"),
		JoinUsing(FullJoin, "users", Using(Column("users.id"))),
	)

	got, err := js.Compile(tmpl)
	require.NoError(t, err)

	want := `NATURAL JOIN "users" FULL JOIN "users" USING ("users"."id")`
	assert.Equal(t, want, StripWhitespace(got))

	t.Run("cache hit", func(t *testing.T) {
		got, err := js.Compile(tmpl)
		assert.NoError(t, err)
		assert.Equal(t, want, StripWhitespace(got))
	})
}
