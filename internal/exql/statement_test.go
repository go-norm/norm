// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package exql

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"unknwon.dev/norm/expr"
)

func TestStatement(t *testing.T) {
	tmpl := defaultTemplate(t)

	t.Run("unexpected type", func(t *testing.T) {
		_, err := (&Statement{}).Compile(tmpl)
		assert.Error(t, err)
	})

	tests := []struct {
		name      string
		statement *Statement
		want      string
	}{
		{
			name: "delete",
			statement: &Statement{
				Type:  StatementDelete,
				Table: Table("users"),
				Where: Where(
					ColumnValue(Column("id"), expr.ComparisonEqual, Raw("99")),
				),
			},
			want: `DELETE FROM "users" WHERE ("id" = 99)`,
		},
		{
			name: "drop database",
			statement: &Statement{
				Type:     StatementDropDatabase,
				Database: Database("norm"),
			},
			want: `DROP DATABASE "norm"`,
		},
		{
			name: "drop table",
			statement: &Statement{
				Type:  StatementDropTable,
				Table: Table("users"),
			},
			want: `DROP TABLE "users"`,
		},
		{
			name: "select",
			statement: &Statement{
				Type:  StatementSelect,
				Table: Table("users"),
				Columns: Columns(
					Column("name"),
					Column("email"),
					Column("created_at"),
				),
			},
			want: `SELECT "name", "email", "created_at" FROM "users"`,
		},
		{
			name: "truncate table",
			statement: &Statement{
				Type:  StatementTruncate,
				Table: Table("users"),
			},
			want: `TRUNCATE TABLE "users"`,
		},
		{
			name: "update",
			statement: &Statement{
				Type:  StatementUpdate,
				Table: Table("users"),
				ColumnValues: ColumnValues(
					ColumnValue(Column("email"), expr.ComparisonEqual, Value("alice@example.com")),
				),
				Where: Where(
					ColumnValue(Column("name"), expr.ComparisonEqual, Value("alice")),
				),
			},
			want: `UPDATE "users" SET "email" = 'alice@example.com' WHERE ("name" = 'alice')`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := test.statement.Compile(tmpl)
			require.NoError(t, err)
			assert.Equal(t, test.want, stripWhitespace(got))
		})
	}

	t.Run("cache hit", func(t *testing.T) {
		s := &Statement{
			Type:  StatementTruncate,
			Table: Table("users"),
		}
		got, err := s.Compile(tmpl)
		assert.NoError(t, err)
		assert.Equal(t, `TRUNCATE TABLE "users"`, got)
	})
}

func TestStatement_Count(t *testing.T) {
	tmpl := defaultTemplate(t)
	tests := []struct {
		name      string
		statement *Statement
		want      string
	}{
		{
			name: "normal",
			statement: &Statement{
				Type:  StatementCount,
				Table: Table("users"),
			},
			want: `SELECT COUNT(*) FROM "users"`,
		},
		{
			name: "relation",
			statement: &Statement{
				Type:  StatementCount,
				Table: Table("information_schema.tables"),
			},
			want: `SELECT COUNT(*) FROM "information_schema"."tables"`,
		},
		{
			name: "where",
			statement: &Statement{
				Type:  StatementCount,
				Table: Table("users"),
				Where: Where(
					ColumnValue(Column("created_at"), expr.ComparisonGreaterThan, Raw("NOW()")),
				),
			},
			want: `SELECT COUNT(*) FROM "users" WHERE ("created_at" > NOW())`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := test.statement.Compile(tmpl)
			require.NoError(t, err)
			assert.Equal(t, test.want, stripWhitespace(got))
		})
	}
}

func TestStatement_Insert(t *testing.T) {
	tmpl := defaultTemplate(t)

	t.Run("unexpected type", func(t *testing.T) {
		_, err := (&Statement{}).Compile(tmpl)
		assert.Error(t, err)
	})

	tests := []struct {
		name      string
		statement *Statement
		want      string
	}{
		{
			name: "single",
			statement: &Statement{
				Type:  StatementInsert,
				Table: Table("users"),
				Columns: Columns(
					Column("name"),
					Column("email"),
					Column("created_at"),
				),
				Values: ValuesGroups(
					ValuesGroup(Value("alice"), Value("alice@example.com"), Raw("NOW()")),
				),
			},
			want: `INSERT INTO "users" ("name", "email", "created_at") VALUES ('alice', 'alice@example.com', NOW())`,
		},
		{
			name: "multiple",
			statement: &Statement{
				Type:  StatementInsert,
				Table: Table("users"),
				Columns: Columns(
					Column("name"),
					Column("email"),
					Column("created_at"),
				),
				Values: ValuesGroups(
					ValuesGroup(Value("alice"), Value("alice@example.com"), Raw("NOW()")),
					ValuesGroup(Value("bob"), Value("bob@example.com"), Raw("NOW()")),
				),
			},
			want: `INSERT INTO "users" ("name", "email", "created_at") VALUES ('alice', 'alice@example.com', NOW()), ('bob', 'bob@example.com', NOW())`,
		},
		{
			name: "returning",
			statement: &Statement{
				Type:  StatementInsert,
				Table: Table("users"),
				Columns: Columns(
					Column("name"),
					Column("email"),
					Column("created_at"),
				),
				Values: ValuesGroups(
					ValuesGroup(Value("alice"), Value("alice@example.com"), Raw("NOW()")),
				),
				Returning: Returning(Column("id")),
			},
			want: `INSERT INTO "users" ("name", "email", "created_at") VALUES ('alice', 'alice@example.com', NOW()) RETURNING "id"`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := test.statement.Compile(tmpl)
			require.NoError(t, err)
			assert.Equal(t, test.want, stripWhitespace(got))
		})
	}
}

//
// func TestSelectStarFrom(t *testing.T) {
// 	var s, e string
//
// 	stmt := Statement{
// 		Type:  StatementSelect,
// 		Table: Table("table_name"),
// 	}
//
// 	s = mustTrim(stmt.Compile(defaultTemplate))
// 	e = `SELECT * FROM "table_name"`
//
// 	if s != e {
// 		t.Fatalf("Got: %s, Expecting: %s", s, e)
// 	}
// }
//
// func TestSelectStarFromAlias(t *testing.T) {
// 	var s, e string
//
// 	stmt := Statement{
// 		Type:  StatementSelect,
// 		Table: Table("table.name AS foo"),
// 	}
//
// 	s = mustTrim(stmt.Compile(defaultTemplate))
// 	e = `SELECT * FROM "table"."name" AS "foo"`
//
// 	if s != e {
// 		t.Fatalf("Got: %s, Expecting: %s", s, e)
// 	}
// }
//
// func TestSelectStarFromRawWhere(t *testing.T) {
// 	var s, e string
// 	var stmt Statement
//
// 	stmt = Statement{
// 		Type:  StatementSelect,
// 		Table: Table("table.name AS foo"),
// 		Where: WhereConditions(
// 			&RawFragment{Value: "foo.id = bar.foo_id"},
// 		),
// 	}
//
// 	s = mustTrim(stmt.Compile(defaultTemplate))
// 	e = `SELECT * FROM "table"."name" AS "foo" WHERE (foo.id = bar.foo_id)`
//
// 	if s != e {
// 		t.Fatalf("Got: %s, Expecting: %s", s, e)
// 	}
//
// 	stmt = Statement{
// 		Type:  StatementSelect,
// 		Table: Table("table.name AS foo"),
// 		Where: WhereConditions(
// 			&RawFragment{Value: "foo.id = bar.foo_id"},
// 			&RawFragment{Value: "baz.id = exp.baz_id"},
// 		),
// 	}
//
// 	s = mustTrim(stmt.Compile(defaultTemplate))
// 	e = `SELECT * FROM "table"."name" AS "foo" WHERE (foo.id = bar.foo_id AND baz.id = exp.baz_id)`
//
// 	if s != e {
// 		t.Fatalf("Got: %s, Expecting: %s", s, e)
// 	}
// }
//
// func TestSelectStarFromMany(t *testing.T) {
// 	var s, e string
//
// 	stmt := Statement{
// 		Type:  StatementSelect,
// 		Table: Table("first.table AS foo, second.table as BAR, third.table aS baz"),
// 	}
//
// 	s = mustTrim(stmt.Compile(defaultTemplate))
// 	e = `SELECT * FROM "first"."table" AS "foo", "second"."table" AS "BAR", "third"."table" AS "baz"`
//
// 	if s != e {
// 		t.Fatalf("Got: %s, Expecting: %s", s, e)
// 	}
// }
//
// func TestSelectTableStarFromMany(t *testing.T) {
// 	var s, e string
//
// 	stmt := Statement{
// 		Type: StatementSelect,
// 		Columns: Columns(
// 			&ColumnFragment{Name: "foo.name"},
// 			&ColumnFragment{Name: "BAR.*"},
// 			&ColumnFragment{Name: "baz.last_name"},
// 		),
// 		Table: Table("first.table AS foo, second.table as BAR, third.table aS baz"),
// 	}
//
// 	s = mustTrim(stmt.Compile(defaultTemplate))
// 	e = `SELECT "foo"."name", "BAR".*, "baz"."last_name" FROM "first"."table" AS "foo", "second"."table" AS "BAR", "third"."table" AS "baz"`
//
// 	if s != e {
// 		t.Fatalf("Got: %s, Expecting: %s", s, e)
// 	}
// }
//
// func TestSelectArtistNameFrom(t *testing.T) {
// 	var s, e string
//
// 	stmt := Statement{
// 		Type:  StatementSelect,
// 		Table: Table("artist"),
// 		Columns: Columns(
// 			&ColumnFragment{Name: "artist.name"},
// 		),
// 	}
//
// 	s = mustTrim(stmt.Compile(defaultTemplate))
// 	e = `SELECT "artist"."name" FROM "artist"`
//
// 	if s != e {
// 		t.Fatalf("Got: %s, Expecting: %s", s, e)
// 	}
// }
//
// func TestSelectJoin(t *testing.T) {
// 	var s, e string
//
// 	stmt := Statement{
// 		Type:  StatementSelect,
// 		Table: Table("artist a"),
// 		Columns: Columns(
// 			&ColumnFragment{Name: "a.name"},
// 		),
// 		Joins: JoinConditions(&JoinFragment{
// 			Table: Table("books b"),
// 			On: OnConditions(
// 				&ColumnValueFragment{
// 					Column:   ColumnWithName("b.author_id"),
// 					Operator: `=`,
// 					Value:    NewValue(ColumnWithName("a.id")),
// 				},
// 			),
// 		}),
// 	}
//
// 	s = mustTrim(stmt.Compile(defaultTemplate))
// 	e = `SELECT "a"."name" FROM "artist" AS "a" JOIN "books" AS "b" ON ("b"."author_id" = "a"."id")`
//
// 	if s != e {
// 		t.Fatalf("Got: %s, Expecting: %s", s, e)
// 	}
// }
//
// func TestSelectJoinUsing(t *testing.T) {
// 	var s, e string
//
// 	stmt := Statement{
// 		Type:  StatementSelect,
// 		Table: Table("artist a"),
// 		Columns: Columns(
// 			&ColumnFragment{Name: "a.name"},
// 		),
// 		Joins: JoinConditions(&JoinFragment{
// 			Table: Table("books b"),
// 			Using: UsingColumns(
// 				ColumnWithName("artist_id"),
// 				ColumnWithName("country"),
// 			),
// 		}),
// 	}
//
// 	s = mustTrim(stmt.Compile(defaultTemplate))
// 	e = `SELECT "a"."name" FROM "artist" AS "a" JOIN "books" AS "b" USING ("artist_id", "country")`
//
// 	if s != e {
// 		t.Fatalf("Got: %s, Expecting: %s", s, e)
// 	}
// }
//
// func TestSelectUnfinishedJoin(t *testing.T) {
// 	stmt := Statement{
// 		Type:  StatementSelect,
// 		Table: Table("artist a"),
// 		Columns: Columns(
// 			&ColumnFragment{Name: "a.name"},
// 		),
// 		Joins: JoinConditions(&JoinFragment{}),
// 	}
//
// 	s := mustTrim(stmt.Compile(defaultTemplate))
// 	e := `SELECT "a"."name" FROM "artist" AS "a"`
// 	if s != e {
// 		t.Fatalf("Got: %s, Expecting: %s", s, e)
// 	}
// }
//
// func TestSelectNaturalJoin(t *testing.T) {
// 	var s, e string
//
// 	stmt := Statement{
// 		Type:  StatementSelect,
// 		Table: Table("artist"),
// 		Joins: JoinConditions(&JoinFragment{
// 			Table: Table("books"),
// 		}),
// 	}
//
// 	s = mustTrim(stmt.Compile(defaultTemplate))
// 	e = `SELECT * FROM "artist" NATURAL JOIN "books"`
//
// 	if s != e {
// 		t.Fatalf("Got: %s, Expecting: %s", s, e)
// 	}
// }
//
// func TestSelectRawFrom(t *testing.T) {
// 	var s, e string
//
// 	stmt := Statement{
// 		Type:  StatementSelect,
// 		Table: Table(`artist`),
// 		Columns: Columns(
// 			&ColumnFragment{Name: `artist.name`},
// 			&ColumnFragment{Name: RawFragment{Value: `CONCAT(artist.name, " ", artist.last_name)`}},
// 		),
// 	}
//
// 	s = mustTrim(stmt.Compile(defaultTemplate))
// 	e = `SELECT "artist"."name", CONCAT(artist.name, " ", artist.last_name) FROM "artist"`
//
// 	if s != e {
// 		t.Fatalf("Got: %s, Expecting: %s", s, e)
// 	}
// }
//
// func TestSelectFieldsFromWithLimitOffset(t *testing.T) {
// 	var s, e string
// 	var stmt Statement
//
// 	// LIMIT only.
// 	stmt = Statement{
// 		Type: StatementSelect,
// 		Columns: Columns(
// 			&ColumnFragment{Name: "foo"},
// 			&ColumnFragment{Name: "bar"},
// 			&ColumnFragment{Name: "baz"},
// 		),
// 		Limit: 42,
// 		Table: Table("table_name"),
// 	}
//
// 	s = mustTrim(stmt.Compile(defaultTemplate))
// 	e = `SELECT "foo", "bar", "baz" FROM "table_name" LIMIT 42`
//
// 	if s != e {
// 		t.Fatalf("Got: %s, Expecting: %s", s, e)
// 	}
//
// 	// OFFSET only.
// 	stmt = Statement{
// 		Type: StatementSelect,
// 		Columns: Columns(
// 			&ColumnFragment{Name: "foo"},
// 			&ColumnFragment{Name: "bar"},
// 			&ColumnFragment{Name: "baz"},
// 		),
// 		Offset: 17,
// 		Table:  Table("table_name"),
// 	}
//
// 	s = mustTrim(stmt.Compile(defaultTemplate))
// 	e = `SELECT "foo", "bar", "baz" FROM "table_name" OFFSET 17`
//
// 	if s != e {
// 		t.Fatalf("Got: %s, Expecting: %s", s, e)
// 	}
//
// 	// LIMIT AND OFFSET.
// 	stmt = Statement{
// 		Type: StatementSelect,
// 		Columns: Columns(
// 			&ColumnFragment{Name: "foo"},
// 			&ColumnFragment{Name: "bar"},
// 			&ColumnFragment{Name: "baz"},
// 		),
// 		Limit:  42,
// 		Offset: 17,
// 		Table:  Table("table_name"),
// 	}
//
// 	s = mustTrim(stmt.Compile(defaultTemplate))
// 	e = `SELECT "foo", "bar", "baz" FROM "table_name" LIMIT 42 OFFSET 17`
//
// 	if s != e {
// 		t.Fatalf("Got: %s, Expecting: %s", s, e)
// 	}
// }
//
// func TestStatementGroupBy(t *testing.T) {
// 	var s, e string
// 	var stmt Statement
//
// 	// Simple GROUP BY
// 	stmt = Statement{
// 		Type: StatementSelect,
// 		Columns: Columns(
// 			&ColumnFragment{Name: "foo"},
// 			&ColumnFragment{Name: "bar"},
// 			&ColumnFragment{Name: "baz"},
// 		),
// 		GroupBy: GroupBy(
// 			&ColumnFragment{Name: "foo"},
// 		),
// 		Table: Table("table_name"),
// 	}
//
// 	s = mustTrim(stmt.Compile(defaultTemplate))
// 	e = `SELECT "foo", "bar", "baz" FROM "table_name" GROUP BY "foo"`
//
// 	if s != e {
// 		t.Fatalf("Got: %s, Expecting: %s", s, e)
// 	}
//
// 	stmt = Statement{
// 		Type: StatementSelect,
// 		Columns: Columns(
// 			&ColumnFragment{Name: "foo"},
// 			&ColumnFragment{Name: "bar"},
// 			&ColumnFragment{Name: "baz"},
// 		),
// 		GroupBy: GroupBy(
// 			&ColumnFragment{Name: "foo"},
// 			&ColumnFragment{Name: "bar"},
// 		),
// 		Table: Table("table_name"),
// 	}
//
// 	s = mustTrim(stmt.Compile(defaultTemplate))
// 	e = `SELECT "foo", "bar", "baz" FROM "table_name" GROUP BY "foo", "bar"`
//
// 	if s != e {
// 		t.Fatalf("Got: %s, Expecting: %s", s, e)
// 	}
// }
//
// func TestSelectFieldsFromWithOrderBy(t *testing.T) {
// 	var s, e string
// 	var stmt Statement
//
// 	// Simple ORDER BY
// 	stmt = Statement{
// 		Type: StatementSelect,
// 		Columns: Columns(
// 			&ColumnFragment{Name: "foo"},
// 			&ColumnFragment{Name: "bar"},
// 			&ColumnFragment{Name: "baz"},
// 		),
// 		OrderBy: OrderBy(
// 			JoinSortColumns(
// 				&SortColumnFragment{Column: &ColumnFragment{Name: "foo"}},
// 			),
// 		),
// 		Table: Table("table_name"),
// 	}
//
// 	s = mustTrim(stmt.Compile(defaultTemplate))
// 	e = `SELECT "foo", "bar", "baz" FROM "table_name" ORDER BY "foo"`
//
// 	if s != e {
// 		t.Fatalf("Got: %s, Expecting: %s", s, e)
// 	}
//
// 	// ORDER BY field ASC
// 	stmt = Statement{
// 		Type: StatementSelect,
// 		Columns: Columns(
// 			&ColumnFragment{Name: "foo"},
// 			&ColumnFragment{Name: "bar"},
// 			&ColumnFragment{Name: "baz"},
// 		),
// 		OrderBy: OrderBy(
// 			JoinSortColumns(
// 				&SortColumnFragment{Column: &ColumnFragment{Name: "foo"}, SortOrder: SortAscendant},
// 			),
// 		),
// 		Table: Table("table_name"),
// 	}
//
// 	s = mustTrim(stmt.Compile(defaultTemplate))
// 	e = `SELECT "foo", "bar", "baz" FROM "table_name" ORDER BY "foo" ASC`
//
// 	if s != e {
// 		t.Fatalf("Got: %s, Expecting: %s", s, e)
// 	}
//
// 	// ORDER BY field DESC
// 	stmt = Statement{
// 		Type: StatementSelect,
// 		Columns: Columns(
// 			&ColumnFragment{Name: "foo"},
// 			&ColumnFragment{Name: "bar"},
// 			&ColumnFragment{Name: "baz"},
// 		),
// 		OrderBy: OrderBy(
// 			JoinSortColumns(
// 				&SortColumnFragment{Column: &ColumnFragment{Name: "foo"}, SortOrder: SortDescendent},
// 			),
// 		),
// 		Table: Table("table_name"),
// 	}
//
// 	s = mustTrim(stmt.Compile(defaultTemplate))
// 	e = `SELECT "foo", "bar", "baz" FROM "table_name" ORDER BY "foo" DESC`
//
// 	if s != e {
// 		t.Fatalf("Got: %s, Expecting: %s", s, e)
// 	}
//
// 	// ORDER BY many fields
// 	stmt = Statement{
// 		Type: StatementSelect,
// 		Columns: Columns(
// 			&ColumnFragment{Name: "foo"},
// 			&ColumnFragment{Name: "bar"},
// 			&ColumnFragment{Name: "baz"},
// 		),
// 		OrderBy: OrderBy(
// 			JoinSortColumns(
// 				&SortColumnFragment{Column: &ColumnFragment{Name: "foo"}, SortOrder: SortDescendent},
// 				&SortColumnFragment{Column: &ColumnFragment{Name: "bar"}, SortOrder: SortAscendant},
// 				&SortColumnFragment{Column: &ColumnFragment{Name: "baz"}, SortOrder: SortDescendent},
// 			),
// 		),
// 		Table: Table("table_name"),
// 	}
//
// 	s = mustTrim(stmt.Compile(defaultTemplate))
// 	e = `SELECT "foo", "bar", "baz" FROM "table_name" ORDER BY "foo" DESC, "bar" ASC, "baz" DESC`
//
// 	if s != e {
// 		t.Fatalf("Got: %s, Expecting: %s", s, e)
// 	}
//
// 	// ORDER BY function
// 	stmt = Statement{
// 		Type: StatementSelect,
// 		Columns: Columns(
// 			&ColumnFragment{Name: "foo"},
// 			&ColumnFragment{Name: "bar"},
// 			&ColumnFragment{Name: "baz"},
// 		),
// 		OrderBy: OrderBy(
// 			JoinSortColumns(
// 				&SortColumnFragment{Column: &ColumnFragment{Name: RawFragment{Value: "FOO()"}}, SortOrder: SortDescendent},
// 				&SortColumnFragment{Column: &ColumnFragment{Name: RawFragment{Value: "BAR()"}}, SortOrder: SortAscendant},
// 			),
// 		),
// 		Table: Table("table_name"),
// 	}
//
// 	s = mustTrim(stmt.Compile(defaultTemplate))
// 	e = `SELECT "foo", "bar", "baz" FROM "table_name" ORDER BY FOO() DESC, BAR() ASC`
//
// 	if s != e {
// 		t.Fatalf("Got: %s, Expecting: %s", s, e)
// 	}
// }
//
// func TestSelectFieldsFromWhere(t *testing.T) {
// 	var s, e string
//
// 	stmt := Statement{
// 		Type: StatementSelect,
// 		Columns: Columns(
// 			&ColumnFragment{Name: "foo"},
// 			&ColumnFragment{Name: "bar"},
// 			&ColumnFragment{Name: "baz"},
// 		),
// 		Table: Table("table_name"),
// 		Where: WhereConditions(
// 			&ColumnValueFragment{Column: &ColumnFragment{Name: "baz"}, Operator: "=", Value: NewValue(99)},
// 		),
// 	}
//
// 	s = mustTrim(stmt.Compile(defaultTemplate))
// 	e = `SELECT "foo", "bar", "baz" FROM "table_name" WHERE ("baz" = '99')`
//
// 	if s != e {
// 		t.Fatalf("Got: %s, Expecting: %s", s, e)
// 	}
// }
//
// func TestSelectFieldsFromWhereLimitOffset(t *testing.T) {
// 	var s, e string
//
// 	stmt := Statement{
// 		Type: StatementSelect,
// 		Columns: Columns(
// 			&ColumnFragment{Name: "foo"},
// 			&ColumnFragment{Name: "bar"},
// 			&ColumnFragment{Name: "baz"},
// 		),
// 		Table: Table("table_name"),
// 		Where: WhereConditions(
// 			&ColumnValueFragment{Column: &ColumnFragment{Name: "baz"}, Operator: "=", Value: NewValue(99)},
// 		),
// 		Limit:  10,
// 		Offset: 23,
// 	}
//
// 	s = mustTrim(stmt.Compile(defaultTemplate))
// 	e = `SELECT "foo", "bar", "baz" FROM "table_name" WHERE ("baz" = '99') LIMIT 10 OFFSET 23`
//
// 	if s != e {
// 		t.Fatalf("Got: %s, Expecting: %s", s, e)
// 	}
// }
//
//
// func TestUpdate(t *testing.T) {
//
// 	stmt = Statement{
// 		Type:  StatementUpdate,
// 		Table: Table("table_name"),
// 		ColumnValues: JoinColumnValues(
// 			&ColumnValueFragment{Column: &ColumnFragment{Name: "foo"}, Operator: "=", Value: NewValue(76)},
// 			&ColumnValueFragment{Column: &ColumnFragment{Name: "bar"}, Operator: "=", Value: NewValue(RawFragment{Value: "88"})},
// 		),
// 		Where: WhereConditions(
// 			&ColumnValueFragment{Column: &ColumnFragment{Name: "baz"}, Operator: "=", Value: NewValue(99)},
// 		),
// 	}
//
// 	s = mustTrim(stmt.Compile(defaultTemplate))
// 	e = `UPDATE "table_name" SET "foo" = '76', "bar" = 88 WHERE ("baz" = '99')`
//
// 	if s != e {
// 		t.Fatalf("Got: %s, Expecting: %s", s, e)
// 	}
// }

//
// func TestRawSQLStatement(t *testing.T) {
// 	stmt := RawSQL(`SELECT * FROM "foo" ORDER BY "bar"`)
//
// 	s := mustTrim(stmt.Compile(defaultTemplate))
// 	e := `SELECT * FROM "foo" ORDER BY "bar"`
//
// 	if s != e {
// 		t.Fatalf("Got: %s, Expecting: %s", s, e)
// 	}
// }
