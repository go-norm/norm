// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package exql

import (
	"regexp"
	"strings"
	"testing"
)

var (
	reInvisible = regexp.MustCompile(`[\t\n\r]`)
	reSpace     = regexp.MustCompile(`\s+`)
)

func mustTrim(a string, err error) string {
	if err != nil {
		panic(err.Error())
	}
	a = reInvisible.ReplaceAllString(strings.TrimSpace(a), " ")
	a = reSpace.ReplaceAllString(strings.TrimSpace(a), " ")
	return a
}

func TestTruncateTable(t *testing.T) {
	var s, e string

	stmt := Statement{
		Type:  Truncate,
		Table: Table("table_name"),
	}

	s = mustTrim(stmt.Compile(defaultTemplate))
	e = `TRUNCATE TABLE "table_name"`

	if s != e {
		t.Fatalf("Got: %s, Expecting: %s", s, e)
	}
}

func TestDropTable(t *testing.T) {
	var s, e string

	stmt := Statement{
		Type:  DropTable,
		Table: Table("table_name"),
	}

	s = mustTrim(stmt.Compile(defaultTemplate))
	e = `DROP TABLE "table_name"`

	if s != e {
		t.Fatalf("Got: %s, Expecting: %s", s, e)
	}
}

func TestDropDatabase(t *testing.T) {
	var s, e string

	stmt := Statement{
		Type:     DropDatabase,
		Database: &Database{Name: "table_name"},
	}

	s = mustTrim(stmt.Compile(defaultTemplate))
	e = `DROP DATABASE "table_name"`

	if s != e {
		t.Fatalf("Got: %s, Expecting: %s", s, e)
	}
}

func TestCount(t *testing.T) {
	var s, e string

	stmt := Statement{
		Type:  Count,
		Table: Table("table_name"),
	}

	s = mustTrim(stmt.Compile(defaultTemplate))
	e = `SELECT COUNT(1) AS _t FROM "table_name"`

	if s != e {
		t.Fatalf("Got: %s, Expecting: %s", s, e)
	}
}

func TestCountRelation(t *testing.T) {
	var s, e string

	stmt := Statement{
		Type:  Count,
		Table: Table("information_schema.tables"),
	}

	s = mustTrim(stmt.Compile(defaultTemplate))
	e = `SELECT COUNT(1) AS _t FROM "information_schema"."tables"`

	if s != e {
		t.Fatalf("Got: %s, Expecting: %s", s, e)
	}
}

func TestCountWhere(t *testing.T) {
	var s, e string

	stmt := Statement{
		Type:  Count,
		Table: Table("table_name"),
		Where: WhereConditions(
			&ColumnValueFragment{Column: &ColumnFragment{Name: "a"}, Operator: "=", Value: NewValue(RawValue("7"))},
		),
	}

	s = mustTrim(stmt.Compile(defaultTemplate))
	e = `SELECT COUNT(1) AS _t FROM "table_name" WHERE ("a" = 7)`

	if s != e {
		t.Fatalf("Got: %s, Expecting: %s", s, e)
	}
}

func TestSelectStarFrom(t *testing.T) {
	var s, e string

	stmt := Statement{
		Type:  Select,
		Table: Table("table_name"),
	}

	s = mustTrim(stmt.Compile(defaultTemplate))
	e = `SELECT * FROM "table_name"`

	if s != e {
		t.Fatalf("Got: %s, Expecting: %s", s, e)
	}
}

func TestSelectStarFromAlias(t *testing.T) {
	var s, e string

	stmt := Statement{
		Type:  Select,
		Table: Table("table.name AS foo"),
	}

	s = mustTrim(stmt.Compile(defaultTemplate))
	e = `SELECT * FROM "table"."name" AS "foo"`

	if s != e {
		t.Fatalf("Got: %s, Expecting: %s", s, e)
	}
}

func TestSelectStarFromRawWhere(t *testing.T) {
	var s, e string
	var stmt Statement

	stmt = Statement{
		Type:  Select,
		Table: Table("table.name AS foo"),
		Where: WhereConditions(
			&RawFragment{Value: "foo.id = bar.foo_id"},
		),
	}

	s = mustTrim(stmt.Compile(defaultTemplate))
	e = `SELECT * FROM "table"."name" AS "foo" WHERE (foo.id = bar.foo_id)`

	if s != e {
		t.Fatalf("Got: %s, Expecting: %s", s, e)
	}

	stmt = Statement{
		Type:  Select,
		Table: Table("table.name AS foo"),
		Where: WhereConditions(
			&RawFragment{Value: "foo.id = bar.foo_id"},
			&RawFragment{Value: "baz.id = exp.baz_id"},
		),
	}

	s = mustTrim(stmt.Compile(defaultTemplate))
	e = `SELECT * FROM "table"."name" AS "foo" WHERE (foo.id = bar.foo_id AND baz.id = exp.baz_id)`

	if s != e {
		t.Fatalf("Got: %s, Expecting: %s", s, e)
	}
}

func TestSelectStarFromMany(t *testing.T) {
	var s, e string

	stmt := Statement{
		Type:  Select,
		Table: Table("first.table AS foo, second.table as BAR, third.table aS baz"),
	}

	s = mustTrim(stmt.Compile(defaultTemplate))
	e = `SELECT * FROM "first"."table" AS "foo", "second"."table" AS "BAR", "third"."table" AS "baz"`

	if s != e {
		t.Fatalf("Got: %s, Expecting: %s", s, e)
	}
}

func TestSelectTableStarFromMany(t *testing.T) {
	var s, e string

	stmt := Statement{
		Type: Select,
		Columns: Columns(
			&ColumnFragment{Name: "foo.name"},
			&ColumnFragment{Name: "BAR.*"},
			&ColumnFragment{Name: "baz.last_name"},
		),
		Table: Table("first.table AS foo, second.table as BAR, third.table aS baz"),
	}

	s = mustTrim(stmt.Compile(defaultTemplate))
	e = `SELECT "foo"."name", "BAR".*, "baz"."last_name" FROM "first"."table" AS "foo", "second"."table" AS "BAR", "third"."table" AS "baz"`

	if s != e {
		t.Fatalf("Got: %s, Expecting: %s", s, e)
	}
}

func TestSelectArtistNameFrom(t *testing.T) {
	var s, e string

	stmt := Statement{
		Type:  Select,
		Table: Table("artist"),
		Columns: Columns(
			&ColumnFragment{Name: "artist.name"},
		),
	}

	s = mustTrim(stmt.Compile(defaultTemplate))
	e = `SELECT "artist"."name" FROM "artist"`

	if s != e {
		t.Fatalf("Got: %s, Expecting: %s", s, e)
	}
}

func TestSelectJoin(t *testing.T) {
	var s, e string

	stmt := Statement{
		Type:  Select,
		Table: Table("artist a"),
		Columns: Columns(
			&ColumnFragment{Name: "a.name"},
		),
		Joins: JoinConditions(&JoinFragment{
			Table: Table("books b"),
			On: OnConditions(
				&ColumnValueFragment{
					Column:   ColumnWithName("b.author_id"),
					Operator: `=`,
					Value:    NewValue(ColumnWithName("a.id")),
				},
			),
		}),
	}

	s = mustTrim(stmt.Compile(defaultTemplate))
	e = `SELECT "a"."name" FROM "artist" AS "a" JOIN "books" AS "b" ON ("b"."author_id" = "a"."id")`

	if s != e {
		t.Fatalf("Got: %s, Expecting: %s", s, e)
	}
}

func TestSelectJoinUsing(t *testing.T) {
	var s, e string

	stmt := Statement{
		Type:  Select,
		Table: Table("artist a"),
		Columns: Columns(
			&ColumnFragment{Name: "a.name"},
		),
		Joins: JoinConditions(&JoinFragment{
			Table: Table("books b"),
			Using: UsingColumns(
				ColumnWithName("artist_id"),
				ColumnWithName("country"),
			),
		}),
	}

	s = mustTrim(stmt.Compile(defaultTemplate))
	e = `SELECT "a"."name" FROM "artist" AS "a" JOIN "books" AS "b" USING ("artist_id", "country")`

	if s != e {
		t.Fatalf("Got: %s, Expecting: %s", s, e)
	}
}

func TestSelectUnfinishedJoin(t *testing.T) {
	stmt := Statement{
		Type:  Select,
		Table: Table("artist a"),
		Columns: Columns(
			&ColumnFragment{Name: "a.name"},
		),
		Joins: JoinConditions(&JoinFragment{}),
	}

	s := mustTrim(stmt.Compile(defaultTemplate))
	e := `SELECT "a"."name" FROM "artist" AS "a"`
	if s != e {
		t.Fatalf("Got: %s, Expecting: %s", s, e)
	}
}

func TestSelectNaturalJoin(t *testing.T) {
	var s, e string

	stmt := Statement{
		Type:  Select,
		Table: Table("artist"),
		Joins: JoinConditions(&JoinFragment{
			Table: Table("books"),
		}),
	}

	s = mustTrim(stmt.Compile(defaultTemplate))
	e = `SELECT * FROM "artist" NATURAL JOIN "books"`

	if s != e {
		t.Fatalf("Got: %s, Expecting: %s", s, e)
	}
}

func TestSelectRawFrom(t *testing.T) {
	var s, e string

	stmt := Statement{
		Type:  Select,
		Table: Table(`artist`),
		Columns: Columns(
			&ColumnFragment{Name: `artist.name`},
			&ColumnFragment{Name: RawFragment{Value: `CONCAT(artist.name, " ", artist.last_name)`}},
		),
	}

	s = mustTrim(stmt.Compile(defaultTemplate))
	e = `SELECT "artist"."name", CONCAT(artist.name, " ", artist.last_name) FROM "artist"`

	if s != e {
		t.Fatalf("Got: %s, Expecting: %s", s, e)
	}
}

func TestSelectFieldsFrom(t *testing.T) {
	var s, e string

	stmt := Statement{
		Type: Select,
		Columns: Columns(
			&ColumnFragment{Name: "foo"},
			&ColumnFragment{Name: "bar"},
			&ColumnFragment{Name: "baz"},
		),
		Table: Table("table_name"),
	}

	s = mustTrim(stmt.Compile(defaultTemplate))
	e = `SELECT "foo", "bar", "baz" FROM "table_name"`

	if s != e {
		t.Fatalf("Got: %s, Expecting: %s", s, e)
	}
}

func TestSelectFieldsFromWithLimitOffset(t *testing.T) {
	var s, e string
	var stmt Statement

	// LIMIT only.
	stmt = Statement{
		Type: Select,
		Columns: Columns(
			&ColumnFragment{Name: "foo"},
			&ColumnFragment{Name: "bar"},
			&ColumnFragment{Name: "baz"},
		),
		Limit: 42,
		Table: Table("table_name"),
	}

	s = mustTrim(stmt.Compile(defaultTemplate))
	e = `SELECT "foo", "bar", "baz" FROM "table_name" LIMIT 42`

	if s != e {
		t.Fatalf("Got: %s, Expecting: %s", s, e)
	}

	// OFFSET only.
	stmt = Statement{
		Type: Select,
		Columns: Columns(
			&ColumnFragment{Name: "foo"},
			&ColumnFragment{Name: "bar"},
			&ColumnFragment{Name: "baz"},
		),
		Offset: 17,
		Table:  Table("table_name"),
	}

	s = mustTrim(stmt.Compile(defaultTemplate))
	e = `SELECT "foo", "bar", "baz" FROM "table_name" OFFSET 17`

	if s != e {
		t.Fatalf("Got: %s, Expecting: %s", s, e)
	}

	// LIMIT AND OFFSET.
	stmt = Statement{
		Type: Select,
		Columns: Columns(
			&ColumnFragment{Name: "foo"},
			&ColumnFragment{Name: "bar"},
			&ColumnFragment{Name: "baz"},
		),
		Limit:  42,
		Offset: 17,
		Table:  Table("table_name"),
	}

	s = mustTrim(stmt.Compile(defaultTemplate))
	e = `SELECT "foo", "bar", "baz" FROM "table_name" LIMIT 42 OFFSET 17`

	if s != e {
		t.Fatalf("Got: %s, Expecting: %s", s, e)
	}
}

func TestStatementGroupBy(t *testing.T) {
	var s, e string
	var stmt Statement

	// Simple GROUP BY
	stmt = Statement{
		Type: Select,
		Columns: Columns(
			&ColumnFragment{Name: "foo"},
			&ColumnFragment{Name: "bar"},
			&ColumnFragment{Name: "baz"},
		),
		GroupBy: GroupByColumns(
			&ColumnFragment{Name: "foo"},
		),
		Table: Table("table_name"),
	}

	s = mustTrim(stmt.Compile(defaultTemplate))
	e = `SELECT "foo", "bar", "baz" FROM "table_name" GROUP BY "foo"`

	if s != e {
		t.Fatalf("Got: %s, Expecting: %s", s, e)
	}

	stmt = Statement{
		Type: Select,
		Columns: Columns(
			&ColumnFragment{Name: "foo"},
			&ColumnFragment{Name: "bar"},
			&ColumnFragment{Name: "baz"},
		),
		GroupBy: GroupByColumns(
			&ColumnFragment{Name: "foo"},
			&ColumnFragment{Name: "bar"},
		),
		Table: Table("table_name"),
	}

	s = mustTrim(stmt.Compile(defaultTemplate))
	e = `SELECT "foo", "bar", "baz" FROM "table_name" GROUP BY "foo", "bar"`

	if s != e {
		t.Fatalf("Got: %s, Expecting: %s", s, e)
	}
}

func TestSelectFieldsFromWithOrderBy(t *testing.T) {
	var s, e string
	var stmt Statement

	// Simple ORDER BY
	stmt = Statement{
		Type: Select,
		Columns: Columns(
			&ColumnFragment{Name: "foo"},
			&ColumnFragment{Name: "bar"},
			&ColumnFragment{Name: "baz"},
		),
		OrderBy: JoinWithOrderBy(
			JoinSortColumns(
				&SortColumn{Column: &ColumnFragment{Name: "foo"}},
			),
		),
		Table: Table("table_name"),
	}

	s = mustTrim(stmt.Compile(defaultTemplate))
	e = `SELECT "foo", "bar", "baz" FROM "table_name" ORDER BY "foo"`

	if s != e {
		t.Fatalf("Got: %s, Expecting: %s", s, e)
	}

	// ORDER BY field ASC
	stmt = Statement{
		Type: Select,
		Columns: Columns(
			&ColumnFragment{Name: "foo"},
			&ColumnFragment{Name: "bar"},
			&ColumnFragment{Name: "baz"},
		),
		OrderBy: JoinWithOrderBy(
			JoinSortColumns(
				&SortColumn{Column: &ColumnFragment{Name: "foo"}, Order: Ascendant},
			),
		),
		Table: Table("table_name"),
	}

	s = mustTrim(stmt.Compile(defaultTemplate))
	e = `SELECT "foo", "bar", "baz" FROM "table_name" ORDER BY "foo" ASC`

	if s != e {
		t.Fatalf("Got: %s, Expecting: %s", s, e)
	}

	// ORDER BY field DESC
	stmt = Statement{
		Type: Select,
		Columns: Columns(
			&ColumnFragment{Name: "foo"},
			&ColumnFragment{Name: "bar"},
			&ColumnFragment{Name: "baz"},
		),
		OrderBy: JoinWithOrderBy(
			JoinSortColumns(
				&SortColumn{Column: &ColumnFragment{Name: "foo"}, Order: Descendent},
			),
		),
		Table: Table("table_name"),
	}

	s = mustTrim(stmt.Compile(defaultTemplate))
	e = `SELECT "foo", "bar", "baz" FROM "table_name" ORDER BY "foo" DESC`

	if s != e {
		t.Fatalf("Got: %s, Expecting: %s", s, e)
	}

	// ORDER BY many fields
	stmt = Statement{
		Type: Select,
		Columns: Columns(
			&ColumnFragment{Name: "foo"},
			&ColumnFragment{Name: "bar"},
			&ColumnFragment{Name: "baz"},
		),
		OrderBy: JoinWithOrderBy(
			JoinSortColumns(
				&SortColumn{Column: &ColumnFragment{Name: "foo"}, Order: Descendent},
				&SortColumn{Column: &ColumnFragment{Name: "bar"}, Order: Ascendant},
				&SortColumn{Column: &ColumnFragment{Name: "baz"}, Order: Descendent},
			),
		),
		Table: Table("table_name"),
	}

	s = mustTrim(stmt.Compile(defaultTemplate))
	e = `SELECT "foo", "bar", "baz" FROM "table_name" ORDER BY "foo" DESC, "bar" ASC, "baz" DESC`

	if s != e {
		t.Fatalf("Got: %s, Expecting: %s", s, e)
	}

	// ORDER BY function
	stmt = Statement{
		Type: Select,
		Columns: Columns(
			&ColumnFragment{Name: "foo"},
			&ColumnFragment{Name: "bar"},
			&ColumnFragment{Name: "baz"},
		),
		OrderBy: JoinWithOrderBy(
			JoinSortColumns(
				&SortColumn{Column: &ColumnFragment{Name: RawFragment{Value: "FOO()"}}, Order: Descendent},
				&SortColumn{Column: &ColumnFragment{Name: RawFragment{Value: "BAR()"}}, Order: Ascendant},
			),
		),
		Table: Table("table_name"),
	}

	s = mustTrim(stmt.Compile(defaultTemplate))
	e = `SELECT "foo", "bar", "baz" FROM "table_name" ORDER BY FOO() DESC, BAR() ASC`

	if s != e {
		t.Fatalf("Got: %s, Expecting: %s", s, e)
	}
}

func TestSelectFieldsFromWhere(t *testing.T) {
	var s, e string

	stmt := Statement{
		Type: Select,
		Columns: Columns(
			&ColumnFragment{Name: "foo"},
			&ColumnFragment{Name: "bar"},
			&ColumnFragment{Name: "baz"},
		),
		Table: Table("table_name"),
		Where: WhereConditions(
			&ColumnValueFragment{Column: &ColumnFragment{Name: "baz"}, Operator: "=", Value: NewValue(99)},
		),
	}

	s = mustTrim(stmt.Compile(defaultTemplate))
	e = `SELECT "foo", "bar", "baz" FROM "table_name" WHERE ("baz" = '99')`

	if s != e {
		t.Fatalf("Got: %s, Expecting: %s", s, e)
	}
}

func TestSelectFieldsFromWhereLimitOffset(t *testing.T) {
	var s, e string

	stmt := Statement{
		Type: Select,
		Columns: Columns(
			&ColumnFragment{Name: "foo"},
			&ColumnFragment{Name: "bar"},
			&ColumnFragment{Name: "baz"},
		),
		Table: Table("table_name"),
		Where: WhereConditions(
			&ColumnValueFragment{Column: &ColumnFragment{Name: "baz"}, Operator: "=", Value: NewValue(99)},
		),
		Limit:  10,
		Offset: 23,
	}

	s = mustTrim(stmt.Compile(defaultTemplate))
	e = `SELECT "foo", "bar", "baz" FROM "table_name" WHERE ("baz" = '99') LIMIT 10 OFFSET 23`

	if s != e {
		t.Fatalf("Got: %s, Expecting: %s", s, e)
	}
}

func TestDelete(t *testing.T) {
	var s, e string

	stmt := Statement{
		Type:  Delete,
		Table: Table("table_name"),
		Where: WhereConditions(
			&ColumnValueFragment{Column: &ColumnFragment{Name: "baz"}, Operator: "=", Value: NewValue(99)},
		),
	}

	s = mustTrim(stmt.Compile(defaultTemplate))
	e = `DELETE FROM "table_name" WHERE ("baz" = '99')`

	if s != e {
		t.Fatalf("Got: %s, Expecting: %s", s, e)
	}
}

func TestUpdate(t *testing.T) {
	var s, e string
	var stmt Statement

	stmt = Statement{
		Type:  Update,
		Table: Table("table_name"),
		ColumnValues: JoinColumnValues(
			&ColumnValueFragment{Column: &ColumnFragment{Name: "foo"}, Operator: "=", Value: NewValue(76)},
		),
		Where: WhereConditions(
			&ColumnValueFragment{Column: &ColumnFragment{Name: "baz"}, Operator: "=", Value: NewValue(99)},
		),
	}

	s = mustTrim(stmt.Compile(defaultTemplate))
	e = `UPDATE "table_name" SET "foo" = '76' WHERE ("baz" = '99')`

	if s != e {
		t.Fatalf("Got: %s, Expecting: %s", s, e)
	}

	stmt = Statement{
		Type:  Update,
		Table: Table("table_name"),
		ColumnValues: JoinColumnValues(
			&ColumnValueFragment{Column: &ColumnFragment{Name: "foo"}, Operator: "=", Value: NewValue(76)},
			&ColumnValueFragment{Column: &ColumnFragment{Name: "bar"}, Operator: "=", Value: NewValue(RawFragment{Value: "88"})},
		),
		Where: WhereConditions(
			&ColumnValueFragment{Column: &ColumnFragment{Name: "baz"}, Operator: "=", Value: NewValue(99)},
		),
	}

	s = mustTrim(stmt.Compile(defaultTemplate))
	e = `UPDATE "table_name" SET "foo" = '76', "bar" = 88 WHERE ("baz" = '99')`

	if s != e {
		t.Fatalf("Got: %s, Expecting: %s", s, e)
	}
}

func TestInsert(t *testing.T) {
	var s, e string

	stmt := Statement{
		Type:  Insert,
		Table: Table("table_name"),
		Columns: Columns(
			&ColumnFragment{Name: "foo"},
			&ColumnFragment{Name: "bar"},
			&ColumnFragment{Name: "baz"},
		),
		Values: NewValueGroup(
			&Value{V: "1"},
			&Value{V: 2},
			&Value{V: RawFragment{Value: "3"}},
		),
	}

	s = mustTrim(stmt.Compile(defaultTemplate))
	e = `INSERT INTO "table_name" ("foo", "bar", "baz") VALUES ('1', '2', 3)`

	if s != e {
		t.Fatalf("Got: %s, Expecting: %s", s, e)
	}
}

func TestInsertMultiple(t *testing.T) {
	var s, e string

	stmt := Statement{
		Type:  Insert,
		Table: Table("table_name"),
		Columns: Columns(
			&ColumnFragment{Name: "foo"},
			&ColumnFragment{Name: "bar"},
			&ColumnFragment{Name: "baz"},
		),
		Values: JoinValueGroups(
			NewValueGroup(
				NewValue("1"),
				NewValue("2"),
				NewValue(RawValue("3")),
			),
			NewValueGroup(
				NewValue(RawValue("4")),
				NewValue(RawValue("5")),
				NewValue(RawValue("6")),
			),
		),
	}

	s = mustTrim(stmt.Compile(defaultTemplate))
	e = `INSERT INTO "table_name" ("foo", "bar", "baz") VALUES ('1', '2', 3), (4, 5, 6)`

	if s != e {
		t.Fatalf("Got: %s, Expecting: %s", s, e)
	}
}

func TestInsertReturning(t *testing.T) {
	var s, e string

	stmt := Statement{
		Type:  Insert,
		Table: Table("table_name"),
		Returning: ReturningColumns(
			ColumnWithName("id"),
		),
		Columns: Columns(
			&ColumnFragment{Name: "foo"},
			&ColumnFragment{Name: "bar"},
			&ColumnFragment{Name: "baz"},
		),
		Values: NewValueGroup(
			&Value{V: "1"},
			&Value{V: 2},
			&Value{V: RawFragment{Value: "3"}},
		),
	}

	s = mustTrim(stmt.Compile(defaultTemplate))
	e = `INSERT INTO "table_name" ("foo", "bar", "baz") VALUES ('1', '2', 3) RETURNING "id"`

	if s != e {
		t.Fatalf("Got: %s, Expecting: %s", s, e)
	}
}

func TestRawSQLStatement(t *testing.T) {
	stmt := RawSQL(`SELECT * FROM "foo" ORDER BY "bar"`)

	s := mustTrim(stmt.Compile(defaultTemplate))
	e := `SELECT * FROM "foo" ORDER BY "bar"`

	if s != e {
		t.Fatalf("Got: %s, Expecting: %s", s, e)
	}
}

func BenchmarkStatementSimpleQuery(b *testing.B) {
	stmt := Statement{
		Type:  Count,
		Table: Table("table_name"),
		Where: WhereConditions(
			&ColumnValueFragment{Column: &ColumnFragment{Name: "a"}, Operator: "=", Value: NewValue(RawFragment{Value: "7"})},
		),
	}

	for i := 0; i < b.N; i++ {
		_, _ = stmt.Compile(defaultTemplate)
	}
}

func BenchmarkStatementSimpleQueryHash(b *testing.B) {
	stmt := Statement{
		Type:  Count,
		Table: Table("table_name"),
		Where: WhereConditions(
			&ColumnValueFragment{Column: &ColumnFragment{Name: "a"}, Operator: "=", Value: NewValue(RawFragment{Value: "7"})},
		),
	}

	for i := 0; i < b.N; i++ {
		_ = stmt.Hash()
	}
}

func BenchmarkStatementSimpleQueryNoCache(b *testing.B) {
	for i := 0; i < b.N; i++ {
		stmt := Statement{
			Type:  Count,
			Table: Table("table_name"),
			Where: WhereConditions(
				&ColumnValueFragment{Column: &ColumnFragment{Name: "a"}, Operator: "=", Value: NewValue(RawFragment{Value: "7"})},
			),
		}
		_, _ = stmt.Compile(defaultTemplate)
	}
}

func BenchmarkStatementComplexQuery(b *testing.B) {
	stmt := Statement{
		Type:  Insert,
		Table: Table("table_name"),
		Columns: Columns(
			&ColumnFragment{Name: "foo"},
			&ColumnFragment{Name: "bar"},
			&ColumnFragment{Name: "baz"},
		),
		Values: NewValueGroup(
			&Value{V: "1"},
			&Value{V: 2},
			&Value{V: RawFragment{Value: "3"}},
		),
	}

	for i := 0; i < b.N; i++ {
		_, _ = stmt.Compile(defaultTemplate)
	}
}

func BenchmarkStatementComplexQueryNoCache(b *testing.B) {
	for i := 0; i < b.N; i++ {
		stmt := Statement{
			Type:  Insert,
			Table: Table("table_name"),
			Columns: Columns(
				&ColumnFragment{Name: "foo"},
				&ColumnFragment{Name: "bar"},
				&ColumnFragment{Name: "baz"},
			),
			Values: NewValueGroup(
				&Value{V: "1"},
				&Value{V: 2},
				&Value{V: RawFragment{Value: "3"}},
			),
		}
		_, _ = stmt.Compile(defaultTemplate)
	}
}
