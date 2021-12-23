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
					ColumnValue("id", expr.ComparisonEqual, Raw("99")),
				),
			},
			want: `DELETE FROM "users" WHERE "id" = 99`,
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
					ColumnValue("email", expr.ComparisonEqual, Value("alice@example.com")),
				),
				Where: Where(
					ColumnValue("name", expr.ComparisonEqual, Value("alice")),
				),
			},
			want: `UPDATE "users" SET "email" = 'alice@example.com' WHERE "name" = 'alice'`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := test.statement.Compile(tmpl)
			require.NoError(t, err)
			assert.Equal(t, test.want, StripWhitespace(got))
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
					ColumnValue("created_at", expr.ComparisonGreaterThan, Raw("NOW()")),
				),
			},
			want: `SELECT COUNT(*) FROM "users" WHERE "created_at" > NOW()`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := test.statement.Compile(tmpl)
			require.NoError(t, err)
			assert.Equal(t, test.want, StripWhitespace(got))
		})
	}
}

func TestStatement_Insert(t *testing.T) {
	tmpl := defaultTemplate(t)
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
				Values: ValuesGroup(Value("alice"), Value("alice@example.com"), Raw("NOW()")),
			},
			want: StripWhitespace(`
INSERT INTO "users"
	("name", "email", "created_at")
VALUES
	('alice', 'alice@example.com', NOW())
`),
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
			want: StripWhitespace(`
INSERT INTO "users"
	("name", "email", "created_at")
VALUES
	('alice', 'alice@example.com', NOW()),
	('bob', 'bob@example.com', NOW())
`),
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
				Values:    ValuesGroup(Value("alice"), Value("alice@example.com"), Raw("NOW()")),
				Returning: Returning(Column("id")),
			},
			want: StripWhitespace(`
INSERT INTO "users"
	("name", "email", "created_at")
VALUES
	('alice', 'alice@example.com', NOW())
RETURNING "id"
`),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := test.statement.Compile(tmpl)
			require.NoError(t, err)
			assert.Equal(t, test.want, StripWhitespace(got))
		})
	}
}

func TestStatement_Select(t *testing.T) {
	tmpl := defaultTemplate(t)
	tests := []struct {
		name      string
		statement *Statement
		want      string
	}{
		{
			name: "normal",
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
			name: "alias",
			statement: &Statement{
				Type:  StatementSelect,
				Table: Table("users.name AS foo"),
			},
			want: `SELECT * FROM "users"."name" AS "foo"`,
		},
		{
			name: "where",
			statement: &Statement{
				Type:  StatementSelect,
				Table: Table("users"),
				Columns: Columns(
					Column("name"),
					Column("email"),
					Column("created_at"),
				),
				Where: Where(
					ColumnValue("id", expr.ComparisonEqual, Raw("1")),
					Raw("deleted_at IS NULL"),
				),
			},
			want: StripWhitespace(`
SELECT
	"name",
	"email",
	"created_at"
FROM "users"
WHERE
	"id" = 1
AND deleted_at IS NULL
`),
		},
		{
			name: "from many",
			statement: &Statement{
				Type: StatementSelect,
				Table: Tables(
					Table("users"),
					Table("customers AS c"),
					Table("sellers.id AS sid"),
					Table("sellers.name as sname"),
				),
				Columns: Columns(
					Column("users.*"),
					Column("c.id"),
					Column("sid"),
				),
			},
			want: StripWhitespace(`
SELECT
	"users".*,
	"c"."id",
	"sid"
FROM
	"users",
	"customers" AS "c",
	"sellers"."id" AS "sid",
	"sellers"."name" AS "sname"
`),
		},
		{
			name: "join on",
			statement: &Statement{
				Type:    StatementSelect,
				Table:   Table("users"),
				Columns: Column("user_emails.email"),
				Joins: JoinOn(
					DefaultJoin,
					"user_emails",
					On(ColumnValue("users.id", expr.ComparisonEqual, Column("user_emails.id"))),
				),
			},
			want: StripWhitespace(`
SELECT
	"user_emails"."email"
FROM "users"
JOIN "user_emails" ON ("users"."id" = "user_emails"."id")
`),
		},
		{
			name: "join using",
			statement: &Statement{
				Type:    StatementSelect,
				Table:   Table("users"),
				Columns: Column("user_emails.email"),
				Joins: JoinUsing(
					DefaultJoin,
					"user_emails",
					Using(Column("user_id")),
				),
			},
			want: StripWhitespace(`
SELECT
	"user_emails"."email"
FROM "users"
JOIN "user_emails" USING ("user_id")
`),
		},
		{
			name: "natural join",
			statement: &Statement{
				Type:    StatementSelect,
				Table:   Table("users"),
				Columns: Column("user_emails.email"),
				Joins:   Join("user_emails"),
			},
			want: StripWhitespace(`
SELECT
	"user_emails"."email"
FROM "users"
NATURAL JOIN "user_emails"
`),
		},
		{
			name: "multiple joins",
			statement: &Statement{
				Type:  StatementSelect,
				Table: Table("users"),
				Columns: Columns(
					Column("user_emails.email"),
					Column("user_invites.id"),
				),
				Joins: Joins(
					JoinOn(
						DefaultJoin,
						"user_emails",
						On(ColumnValue("users.id", expr.ComparisonEqual, Column("user_emails.id"))),
					),
					JoinOn(
						DefaultJoin,
						"user_invites",
						On(ColumnValue("users.id", expr.ComparisonEqual, Column("user_invites.id"))),
					),
				),
			},
			want: StripWhitespace(`
SELECT
	"user_emails"."email",
	"user_invites"."id"
FROM "users"
JOIN "user_emails" ON ("users"."id" = "user_emails"."id")
JOIN "user_invites" ON ("users"."id" = "user_invites"."id")
`),
		},
		{
			name: "raw",
			statement: &Statement{
				Type:  StatementSelect,
				Table: Table("users"),
				Columns: Columns(
					Column("users.name"),
					Column(Raw(`CONCAT(users.name, " ", users.last_name)`)),
				),
			},
			want: `SELECT "users"."name", CONCAT(users.name, " ", users.last_name) FROM "users"`,
		},
		{
			name: "limit",
			statement: &Statement{
				Type:    StatementSelect,
				Table:   Table("users"),
				Columns: Column("users.name"),
				Limit:   10,
			},
			want: `SELECT "users"."name" FROM "users" LIMIT 10`,
		},
		{
			name: "offset",
			statement: &Statement{
				Type:    StatementSelect,
				Table:   Table("users"),
				Columns: Column("users.name"),
				Offset:  10,
			},
			want: `SELECT "users"."name" FROM "users" OFFSET 10`,
		},
		{
			name: "limit and offset",
			statement: &Statement{
				Type:    StatementSelect,
				Table:   Table("users"),
				Columns: Column("users.name"),
				Limit:   10,
				Offset:  10,
			},
			want: `SELECT "users"."name" FROM "users" LIMIT 10 OFFSET 10`,
		},
		{
			name: "group by",
			statement: &Statement{
				Type:    StatementSelect,
				Table:   Table("users"),
				Columns: Column("*"),
				GroupBy: GroupBy(Column("users.country")),
			},
			want: `SELECT * FROM "users" GROUP BY "users"."country"`,
		},
		{
			name: "multiple group bys",
			statement: &Statement{
				Type:    StatementSelect,
				Table:   Table("users"),
				Columns: Column("*"),
				GroupBy: GroupBy(
					Column("users.country"),
					Column("users.gender"),
				),
			},
			want: `SELECT * FROM "users" GROUP BY "users"."country", "users"."gender"`,
		},
		{
			name: "order by",
			statement: &Statement{
				Type:    StatementSelect,
				Table:   Table("users"),
				Columns: Column("*"),
				OrderBy: OrderBy(
					SortColumn("users.country"),
				),
			},
			want: `SELECT * FROM "users" ORDER BY "users"."country"`,
		},
		{
			name: "multiple order bys",
			statement: &Statement{
				Type:    StatementSelect,
				Table:   Table("users"),
				Columns: Column("*"),
				OrderBy: OrderBy(
					SortColumn("users.country"),
					SortColumn("users.gender", SortDescendent),
				),
			},
			want: `SELECT * FROM "users" ORDER BY "users"."country", "users"."gender" DESC`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := test.statement.Compile(tmpl)
			require.NoError(t, err)
			assert.Equal(t, test.want, StripWhitespace(got))
		})
	}
}

func TestStatement_Amend(t *testing.T) {
	s := &Statement{
		Type:    StatementSelect,
		Table:   Table("users"),
		Columns: Column("name"),
	}
	s.SetAmend(func(s string) string {
		return s + " FOR UPDATE"
	})

	got, err := s.Compile(defaultTemplate(t))
	require.NoError(t, err)

	want := `SELECT "name" FROM "users" FOR UPDATE`
	assert.Equal(t, want, StripWhitespace(got))
}

func TestRawSQL(t *testing.T) {
	const sql = `SELECT * FROM "foo" ORDER BY "bar"`
	s := RawSQL(sql)

	got, err := s.Compile(defaultTemplate(t))
	require.NoError(t, err)
	assert.Equal(t, sql, StripWhitespace(got))
}
