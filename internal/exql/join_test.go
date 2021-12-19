// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package exql

import (
	"testing"
)

func TestJoin(t *testing.T) {

}

// func TestOnAndRawOrAnd(t *testing.T) {
// 	var s, e string
//
// 	on := On(
// 		And(
// 			&ColumnValueFragment{Column: &ColumnFragment{Name: "id"}, Operator: ">", Value: NewValue(&RawFragment{Value: "8"})},
// 			&ColumnValueFragment{Column: &ColumnFragment{Name: "id"}, Operator: "<", Value: NewValue(&RawFragment{Value: "99"})},
// 		),
// 		&ColumnValueFragment{Column: &ColumnFragment{Name: "name"}, Operator: "=", Value: NewValue("John")},
// 		&RawFragment{Value: "city_id = 728"},
// 		Or(
// 			&ColumnValueFragment{Column: &ColumnFragment{Name: "last_name"}, Operator: "=", Value: NewValue("Smith")},
// 			&ColumnValueFragment{Column: &ColumnFragment{Name: "last_name"}, Operator: "=", Value: NewValue("Reyes")},
// 		),
// 		And(
// 			&ColumnValueFragment{Column: &ColumnFragment{Name: "age"}, Operator: ">", Value: NewValue(&RawFragment{Value: "18"})},
// 			&ColumnValueFragment{Column: &ColumnFragment{Name: "age"}, Operator: "<", Value: NewValue(&RawFragment{Value: "41"})},
// 		),
// 	)
//
// 	s = mustTrim(on.Compile(defaultTemplate))
// 	e = `ON (("id" > 8 AND "id" < 99) AND "name" = 'John' AND city_id = 728 AND ("last_name" = 'Smith' OR "last_name" = 'Reyes') AND ("age" > 18 AND "age" < 41))`
//
// 	if s != e {
// 		t.Fatalf("Got: %s, Expecting: %s", s, e)
// 	}
// }
//
// func TestUsing(t *testing.T) {
// 	var s, e string
//
// 	using := Using(
// 		&ColumnFragment{Name: "country"},
// 		&ColumnFragment{Name: "state"},
// 	)
//
// 	s = mustTrim(using.Compile(defaultTemplate))
// 	e = `USING ("country", "state")`
//
// 	if s != e {
// 		t.Fatalf("Got: %s, Expecting: %s", s, e)
// 	}
// }
//
// func TestJoinOn(t *testing.T) {
// 	var s, e string
//
// 	join := JoinConditions(
// 		&JoinFragment{
// 			Table: Table("countries c"),
// 			On: On(
// 				&ColumnValueFragment{
// 					Column:   &ColumnFragment{Name: "p.country_id"},
// 					Operator: "=",
// 					Value:    NewValue(&ColumnFragment{Name: "a.id"}),
// 				},
// 				&ColumnValueFragment{
// 					Column:   &ColumnFragment{Name: "p.country_code"},
// 					Operator: "=",
// 					Value:    NewValue(&ColumnFragment{Name: "a.code"}),
// 				},
// 			),
// 		},
// 	)
//
// 	s = mustTrim(join.Compile(defaultTemplate))
// 	e = `JOIN "countries" AS "c" ON ("p"."country_id" = "a"."id" AND "p"."country_code" = "a"."code")`
//
// 	if s != e {
// 		t.Fatalf("Got: %s, Expecting: %s", s, e)
// 	}
// }
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
