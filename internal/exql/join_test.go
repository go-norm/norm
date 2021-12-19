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

}

func TestJoinOn(t *testing.T) {
	tmpl := defaultTemplate(t)

	t.Run("no table", func(t *testing.T) {
		got, err := JoinOn(DefaultJoin, nil, nil).Compile(tmpl)
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	join := JoinOn(DefaultJoin,
		Table("countries c"),
		On(
			ColumnValue(Column("p.country_id"), expr.ComparisonEqual, Column("a.id")),
			ColumnValue(Column("p.country_code"), expr.ComparisonEqual, Column("a.code")),
		),
	)

	got, err := join.Compile(tmpl)
	require.NoError(t, err)

	want := `JOIN "countries" AS "c" ON ("p"."country_id" = "a"."id" AND "p"."country_code" = "a"."code")`
	assert.Equal(t, want, stripWhitespace(got))

	t.Run("cache hit", func(t *testing.T) {
		got, err := join.Compile(tmpl)
		assert.NoError(t, err)
		assert.Equal(t, want, stripWhitespace(got))
	})
}

func TestOn(t *testing.T) {
	tmpl := defaultTemplate(t)

	t.Run("empty", func(t *testing.T) {
		got, err := On().Compile(tmpl)
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	on :=
		On(
			And(
				ColumnValue(Column("id"), expr.ComparisonGreaterThan, Raw("8")),
				ColumnValue(Column("id"), expr.ComparisonLessThan, Raw("99")),
				Or(
					ColumnValue(Column("age"), expr.ComparisonLessThan, Raw("18")),
					ColumnValue(Column("age"), expr.ComparisonGreaterThan, Raw("41")),
				),
			),
			ColumnValue(Column("name"), expr.ComparisonEqual, Raw(`'John'`)),
			Or(
				ColumnValue(Column("last_name"), expr.ComparisonEqual, Raw(`'Smith'`)),
				ColumnValue(Column("last_name"), expr.ComparisonEqual, Raw(`'Reyes'`)),
			),
			Raw("city_id = 728"),
		)

	got, err := on.Compile(tmpl)
	require.NoError(t, err)

	want := `ON (("id" > 8 AND "id" < 99 AND ("age" < 18 OR "age" > 41)) AND "name" = 'John' AND ("last_name" = 'Smith' OR "last_name" = 'Reyes') AND city_id = 728)`
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

//
// func TestInnerJoinOn(t *testing.T) {
// 	var s, e string
//
// 	join := JoinConditions(&JoinFragment{
// 		Type:  "INNER",
// 		Table: Table("countries c"),
// 		On: On(
// 			&ColumnValueFragment{
// 				Column:   &ColumnFragment{Name: "p.country_id"},
// 				Operator: "=",
// 				Value:    NewValue(ColumnWithName("a.id")),
// 			},
// 			&ColumnValueFragment{
// 				Column:   &ColumnFragment{Name: "p.country_code"},
// 				Operator: "=",
// 				Value:    NewValue(ColumnWithName("a.code")),
// 			},
// 		),
// 	})
//
// 	s = mustTrim(join.Compile(defaultTemplate))
// 	e = `INNER JOIN "countries" AS "c" ON ("p"."country_id" = "a"."id" AND "p"."country_code" = "a"."code")`
//
// 	if s != e {
// 		t.Fatalf("Got: %s, Expecting: %s", s, e)
// 	}
// }
//
// func TestLeftJoinUsing(t *testing.T) {
// 	var s, e string
//
// 	join := JoinConditions(&JoinFragment{
// 		Type:  "LEFT",
// 		Table: Table("countries"),
// 		Using: Using(ColumnWithName("name")),
// 	})
//
// 	s = mustTrim(join.Compile(defaultTemplate))
// 	e = `LEFT JOIN "countries" USING ("name")`
//
// 	if s != e {
// 		t.Fatalf("Got: %s, Expecting: %s", s, e)
// 	}
// }
//
// func TestNaturalJoinOn(t *testing.T) {
// 	var s, e string
//
// 	join := JoinConditions(&JoinFragment{
// 		Table: Table("countries"),
// 	})
//
// 	s = mustTrim(join.Compile(defaultTemplate))
// 	e = `NATURAL JOIN "countries"`
//
// 	if s != e {
// 		t.Fatalf("Got: %s, Expecting: %s", s, e)
// 	}
// }
//
// func TestNaturalInnerJoinOn(t *testing.T) {
// 	var s, e string
//
// 	join := JoinConditions(&JoinFragment{
// 		Type:  "INNER",
// 		Table: Table("countries"),
// 	})
//
// 	s = mustTrim(join.Compile(defaultTemplate))
// 	e = `NATURAL INNER JOIN "countries"`
//
// 	if s != e {
// 		t.Fatalf("Got: %s, Expecting: %s", s, e)
// 	}
// }
//
// func TestCrossJoin(t *testing.T) {
// 	var s, e string
//
// 	join := JoinConditions(&JoinFragment{
// 		Type:  "CROSS",
// 		Table: Table("countries"),
// 	})
//
// 	s = mustTrim(join.Compile(defaultTemplate))
// 	e = `CROSS JOIN "countries"`
//
// 	if s != e {
// 		t.Fatalf("Got: %s, Expecting: %s", s, e)
// 	}
// }
//
// func TestMultipleJoins(t *testing.T) {
// 	var s, e string
//
// 	join := JoinConditions(&JoinFragment{
// 		Type:  "LEFT",
// 		Table: Table("countries"),
// 	}, &JoinFragment{
// 		Table: Table("cities"),
// 	})
//
// 	s = mustTrim(join.Compile(defaultTemplate))
// 	e = `NATURAL LEFT JOIN "countries" NATURAL JOIN "cities"`
//
// 	if s != e {
// 		t.Fatalf("Got: %s, Expecting: %s", s, e)
// 	}
// }
