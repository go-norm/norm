// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package sqlbuilder

import (
	"context"
	"database/sql"
	"testing"

	mockrequire "github.com/derision-test/go-mockgen/testutil/require"
	"github.com/stretchr/testify/assert"

	"unknwon.dev/norm"
	"unknwon.dev/norm/expr"
	"unknwon.dev/norm/internal/exql"
)

func TestSelector(t *testing.T) {
	typer := NewMockTyper()
	typer.ValuerFunc.SetDefaultHook(func(v interface{}) interface{} {
		return v
	})

	adapter := NewMockAdapter()
	adapter.TyperFunc.SetDefaultReturn(typer)
	adapter.FormatSQLFunc.SetDefaultHook(func(sql string) string {
		return exql.StripWhitespace(sql)
	})

	tmpl := defaultTemplate(t)
	sql := New(adapter, tmpl)
	tests := []struct {
		name      string
		selector  norm.Selector
		wantQuery string
		wantArgs  []interface{}
	}{
		{
			name:      "normal",
			selector:  sql.Select("name", "email", "created_at").From("users"),
			wantQuery: `SELECT "name", "email", "created_at" FROM "users"`,
			wantArgs:  nil,
		},
		{
			name:      "alias",
			selector:  sql.SelectFrom("users.name AS foo"),
			wantQuery: `SELECT * FROM "users"."name" AS "foo"`,
			wantArgs:  nil,
		},
		{
			name: "where",
			selector: sql.
				Select("name", "email", "created_at").
				From("users").
				Where(
					expr.NewConstraint("id", expr.Eq(1)),
					expr.NewConstraint("deleted_at", expr.IsNull()),
				),
			wantQuery: exql.StripWhitespace(`
SELECT
	"name",
	"email",
	"created_at"
FROM "users"
WHERE
	"id" = ?
AND "deleted_at" IS NULL
`),
			wantArgs: []interface{}{1},
		},
		{
			name: "from many",
			selector: sql.
				Select("users.*", "c.id", "sid").
				From("users", "customers AS c", "sellers.id AS sid", "sellers.name as sname"),
			wantQuery: exql.StripWhitespace(`
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
			wantArgs: nil,
		},
		{
			name: "join on",
			selector: sql.
				Select("user_emails.email").
				From("users").
				Join("user_emails").Using("user_id"),
			wantQuery: exql.StripWhitespace(`
SELECT
	"user_emails"."email"
FROM "users"
JOIN "user_emails" USING ("user_id")
`),
			wantArgs: nil,
		},
		{
			name: "natural join",
			selector: sql.
				Select("user_emails.email").
				From("users").
				Join("user_emails"),
			wantQuery: exql.StripWhitespace(`
SELECT
	"user_emails"."email"
FROM "users"
NATURAL JOIN "user_emails"
`),
			wantArgs: nil,
		},
		{
			name: "multiple joins",
			selector: sql.
				Select("user_emails.email", "user_invites.id").
				From("users").
				Join("user_emails").On("users.id = user_emails.id").
				Join("user_invites").On("users.id = user_invites.id"),
			wantQuery: exql.StripWhitespace(`
SELECT
	"user_emails"."email",
	"user_invites"."id"
FROM "users"
JOIN "user_emails" ON ("users"."id" = "user_emails"."id")
JOIN "user_invites" ON ("users"."id" = "user_invites"."id")
`),
			wantArgs: nil,
		},
		{
			name: "raw",
			selector: sql.
				Select(
					"users.name",
					expr.Raw(`CONCAT(users.name, " ", users.last_name)`),
				).
				From("users"),
			wantQuery: `SELECT "users"."name", CONCAT(users.name, " ", users.last_name) FROM "users"`,
			wantArgs:  nil,
		},
		{
			name:      "limit",
			selector:  sql.Select("users.name").From("users").Limit(10),
			wantQuery: `SELECT "users"."name" FROM "users" LIMIT 10`,
			wantArgs:  nil,
		},
		{
			name:      "offset",
			selector:  sql.Select("users.name").From("users").Offset(10),
			wantQuery: `SELECT "users"."name" FROM "users" OFFSET 10`,
			wantArgs:  nil,
		},
		{
			name:      "limit and offset",
			selector:  sql.Select("users.name").From("users").Limit(10).Offset(10),
			wantQuery: `SELECT "users"."name" FROM "users" LIMIT 10 OFFSET 10`,
			wantArgs:  nil,
		},
		{
			name:      "group by",
			selector:  sql.SelectFrom("users").GroupBy("users.country"),
			wantQuery: `SELECT * FROM "users" GROUP BY "users"."country"`,
			wantArgs:  nil,
		},
		{
			name:      "multiple group bys",
			selector:  sql.SelectFrom("users").GroupBy("users.country", "users.gender"),
			wantQuery: `SELECT * FROM "users" GROUP BY "users"."country", "users"."gender"`,
			wantArgs:  nil,
		},
		{
			name:      "order by",
			selector:  sql.SelectFrom("users").OrderBy("users.country"),
			wantQuery: `SELECT * FROM "users" ORDER BY "users"."country" ASC`,
			wantArgs:  nil,
		},
		{
			name:      "multiple order bys",
			selector:  sql.SelectFrom("users").OrderBy("users.country", "users.gender DESC"),
			wantQuery: `SELECT * FROM "users" ORDER BY "users"."country" ASC, "users"."gender" DESC`,
			wantArgs:  nil,
		},

		{
			name: "nested queries",
			selector: sql.SelectFrom("actions").
				Where(
					expr.Cond{
						"user_id": expr.Eq(12),
					},
					expr.Or(
						expr.Bool(false),
						expr.Cond{"id": expr.Lt(88)},
					),
				).
				And("repo_id IN ?",
					sql.Select("repository.id").
						From("repositories").
						Join("team_repos").On("repository.id = team_repo.repo_id").
						Where("team_repos.team_id IN ?",
							sql.Select("team_id").
								From("team_users").
								Where(
									expr.Or(
										expr.Cond{
											"team_users.org_id": expr.Eq(11),
											"uid":               expr.Eq(13),
										},
										expr.Cond{
											"repositories.is_private":  expr.Eq(false),
											"repositories.is_unlisted": expr.Eq(false),
										},
									),
								),
						),
				).OrderBy("id DESC").
				Limit(10),
			wantQuery: exql.StripWhitespace(`
SELECT
	*
FROM "actions"
WHERE
	"user_id" = ?
AND (FALSE OR ("id" < ?))
AND repo_id IN (
	SELECT
		"repository"."id"
	FROM "repositories"
	JOIN "team_repos" ON ("repository"."id" = "team_repo"."repo_id")
	WHERE
		team_repos.team_id IN (
			SELECT "team_id" FROM "team_users" WHERE (
					("team_users"."org_id" = ? AND "uid" = ?)
				OR  ("repositories"."is_private" = ? AND "repositories"."is_unlisted" = ?)
			)
	)
)
ORDER BY "id" DESC
LIMIT 10
`),
			wantArgs: []interface{}{12, 88, 11, 13, false, false},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.wantQuery, test.selector.String())
			assert.Equal(t, test.wantArgs, test.selector.Arguments())
		})
	}
}

func TestSelector_Columns(t *testing.T) {
	adapter := NewMockAdapter()
	adapter.FormatSQLFunc.SetDefaultHook(func(sql string) string {
		return exql.StripWhitespace(sql)
	})

	tmpl := defaultTemplate(t)
	sql := New(adapter, tmpl)
	sel := sql.Select().
		Columns(
			sql.SelectFrom("users_emails"),
			expr.Func("version"),
			expr.Raw("NOW()"),
			"users.name",
			857,
		)

	want := `SELECT (SELECT * FROM "users_emails"), version(), NOW(), "users"."name", 857`
	assert.Equal(t, want, sel.String())
}

func TestSelector_From(t *testing.T) {
	adapter := NewMockAdapter()
	adapter.FormatSQLFunc.SetDefaultHook(func(sql string) string {
		return exql.StripWhitespace(sql)
	})

	tmpl := defaultTemplate(t)
	sql := New(adapter, tmpl)
	sel := sql.SelectFrom().
		From(
			sql.SelectFrom("users_emails"),
			expr.Func("version"),
			expr.Raw("NOW()"),
			"users.name",
			857,
		)

	want := `SELECT * FROM (SELECT * FROM "users_emails"), version(), NOW(), "users"."name", 857`
	assert.Equal(t, want, sel.String())
}

func TestSelector_Distinct(t *testing.T) {
	adapter := NewMockAdapter()
	adapter.FormatSQLFunc.SetDefaultHook(func(sql string) string {
		return exql.StripWhitespace(sql)
	})

	tmpl := defaultTemplate(t)
	sql := New(adapter, tmpl)
	tests := []struct {
		name      string
		selector  norm.Selector
		wantQuery string
	}{
		{
			name:      "all",
			selector:  sql.Select("name").Distinct().Columns("email"),
			wantQuery: `SELECT DISTINCT "name", "email"`,
		},
		{
			name:      "some",
			selector:  sql.Select("name").Distinct("email", "gender"),
			wantQuery: `SELECT "name", DISTINCT("email", "gender")`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.wantQuery, test.selector.String())
		})
	}
}

func TestSelector_As(t *testing.T) {
	adapter := NewMockAdapter()
	adapter.FormatSQLFunc.SetDefaultHook(func(sql string) string {
		return exql.StripWhitespace(sql)
	})

	tmpl := defaultTemplate(t)
	sql := New(adapter, tmpl)

	t.Run("no table", func(t *testing.T) {
		assert.Panics(t, func() {
			_ = sql.SelectFrom().As("u").String()
		})
	})

	sel := sql.SelectFrom("users").As("u")
	want := `SELECT * FROM "users" AS "u"`
	assert.Equal(t, want, sel.String())
}

func TestSelector_Where(t *testing.T) {
	typer := NewMockTyper()
	typer.ValuerFunc.SetDefaultHook(func(v interface{}) interface{} {
		return v
	})

	adapter := NewMockAdapter()
	adapter.TyperFunc.SetDefaultReturn(typer)
	adapter.FormatSQLFunc.SetDefaultHook(func(sql string) string {
		return exql.StripWhitespace(sql)
	})

	tmpl := defaultTemplate(t)
	sql := New(adapter, tmpl)
	tests := []struct {
		name      string
		selector  norm.Selector
		wantQuery string
		wantArgs  []interface{}
	}{
		{
			name: "random",
			selector: sql.Select().Where().
				Where(
					sql.SelectFrom("users_emails"),
					expr.Func("version"),
					expr.Raw("NOW()"),
					"users.name",
					857,
				),
			wantQuery: `SELECT * WHERE (SELECT * FROM "users_emails") AND version() AND NOW() AND 'users.name' AND '857'`,
			wantArgs:  nil,
		},
		{
			name: "subquery",
			selector: sql.SelectFrom("users").
				Where("EXISTS ?",
					sql.Select(1).
						From("users").
						Where("name = ?", "alice"),
				),
			wantQuery: `SELECT * FROM "users" WHERE EXISTS (SELECT 1 FROM "users" WHERE name = ?)`,
			wantArgs:  []interface{}{"alice"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.wantQuery, test.selector.String())
			assert.Equal(t, test.wantArgs, test.selector.Arguments())
		})
	}
}

func TestSelector_And(t *testing.T) {
	adapter := NewMockAdapter()
	adapter.FormatSQLFunc.SetDefaultHook(func(sql string) string {
		return exql.StripWhitespace(sql)
	})

	tmpl := defaultTemplate(t)
	sql := New(adapter, tmpl)
	sel := sql.SelectFrom().And().
		And(
			sql.SelectFrom("users_emails"),
			expr.Func("version"),
			expr.Raw("NOW()"),
			"users.name",
			857,
		)

	want := `SELECT * WHERE (SELECT * FROM "users_emails") AND version() AND NOW() AND 'users.name' AND '857'`
	assert.Equal(t, want, sel.String())
}

func TestSelector_GroupBy(t *testing.T) {
	adapter := NewMockAdapter()
	adapter.FormatSQLFunc.SetDefaultHook(func(sql string) string {
		return exql.StripWhitespace(sql)
	})

	tmpl := defaultTemplate(t)
	sql := New(adapter, tmpl)
	sel := sql.SelectFrom().GroupBy().
		GroupBy(
			sql.SelectFrom("users_emails"),
			expr.Func("version"),
			expr.Raw("NOW()"),
			"users.name",
			857,
		)

	want := `SELECT * GROUP BY (SELECT * FROM "users_emails"), version(), NOW(), "users"."name", 857`
	assert.Equal(t, want, sel.String())
}

func TestSelector_OrderBy(t *testing.T) {
	adapter := NewMockAdapter()
	adapter.FormatSQLFunc.SetDefaultHook(func(sql string) string {
		return exql.StripWhitespace(sql)
	})

	tmpl := defaultTemplate(t)
	sql := New(adapter, tmpl)
	sel := sql.SelectFrom().OrderBy().
		OrderBy(
			expr.Func("version"),
			expr.Func("AVG", 9.6),
			expr.Raw("NOW()"),
			"users.name",
			"-users.gender",
		)

	want := `SELECT * ORDER BY version(), AVG(?), NOW(), "users"."name" ASC, "users"."gender" DESC`
	assert.Equal(t, want, sel.String())
}

func TestSelector_Join(t *testing.T) {
	adapter := NewMockAdapter()
	adapter.FormatSQLFunc.SetDefaultHook(func(sql string) string {
		return exql.StripWhitespace(sql)
	})

	tmpl := defaultTemplate(t)
	sql := New(adapter, tmpl)
	sel := sql.SelectFrom("users").
		Join("user_emails").On("users.id = user_emails.id").
		FullJoin("user_addresses").On("users.id = user_addresses.id").
		CrossJoin("user_permissions").On("users.id = user_permissions.id").
		RightJoin("customers").On("users.id = customers.id").
		LeftJoin("sellers").On("users.id = sellers.id")

	want := exql.StripWhitespace(`
SELECT
	*
FROM "users"
JOIN "user_emails" ON ("users"."id" = "user_emails"."id")
FULL JOIN "user_addresses" ON ("users"."id" = "user_addresses"."id")
CROSS JOIN "user_permissions" ON ("users"."id" = "user_permissions"."id")
RIGHT JOIN "customers" ON ("users"."id" = "customers"."id")
LEFT JOIN "sellers" ON ("users"."id" = "sellers"."id")
`)
	assert.Equal(t, want, sel.String())
}

func TestSelector_On(t *testing.T) {
	adapter := NewMockAdapter()
	adapter.FormatSQLFunc.SetDefaultHook(func(sql string) string {
		return exql.StripWhitespace(sql)
	})

	tmpl := defaultTemplate(t)
	sql := New(adapter, tmpl)

	t.Run("no join", func(t *testing.T) {
		assert.Panics(t, func() {
			_ = sql.Select().On("a.id = b.id").String()
		})
	})

	t.Run("multiple ons", func(t *testing.T) {
		assert.Panics(t, func() {
			_ = sql.Select().Join("b").On("a.id = b.id").On("").String()
		})
	})

	sel := sql.SelectFrom("users").
		Join("user_emails").On().On("users.id = user_emails.id")

	want := exql.StripWhitespace(`SELECT * FROM "users" JOIN "user_emails" ON ("users"."id" = "user_emails"."id")`)
	assert.Equal(t, want, sel.String())
}

func TestSelector_Using(t *testing.T) {
	adapter := NewMockAdapter()
	adapter.FormatSQLFunc.SetDefaultHook(func(sql string) string {
		return exql.StripWhitespace(sql)
	})

	tmpl := defaultTemplate(t)
	sql := New(adapter, tmpl)

	t.Run("no join", func(t *testing.T) {
		assert.Panics(t, func() {
			_ = sql.Select().Using("a.id").String()
		})
	})

	t.Run("multiple ons", func(t *testing.T) {
		assert.Panics(t, func() {
			_ = sql.Select().Join("b").Using("a.id").Using("").String()
		})
	})

	sel := sql.SelectFrom("users").
		Join("user_emails").Using().Using("users.id")

	want := exql.StripWhitespace(`SELECT * FROM "users" JOIN "user_emails" USING ("users"."id")`)
	assert.Equal(t, want, sel.String())
}

func TestSelector_Limit(t *testing.T) {
	adapter := NewMockAdapter()
	adapter.FormatSQLFunc.SetDefaultHook(func(sql string) string {
		return exql.StripWhitespace(sql)
	})

	tmpl := defaultTemplate(t)
	sql := New(adapter, tmpl)
	tests := []struct {
		name      string
		selector  norm.Selector
		wantQuery string
	}{
		{
			name:      "good",
			selector:  sql.Select().Limit(10),
			wantQuery: `SELECT * LIMIT 10`,
		},
		{
			name:      "bad",
			selector:  sql.Select().Limit(-1),
			wantQuery: `SELECT *`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.wantQuery, test.selector.String())
		})
	}
}

func TestSelector_Offset(t *testing.T) {
	adapter := NewMockAdapter()
	adapter.FormatSQLFunc.SetDefaultHook(func(sql string) string {
		return exql.StripWhitespace(sql)
	})

	tmpl := defaultTemplate(t)
	sql := New(adapter, tmpl)
	tests := []struct {
		name      string
		selector  norm.Selector
		wantQuery string
	}{
		{
			name:      "good",
			selector:  sql.Select().Offset(10),
			wantQuery: `SELECT * OFFSET 10`,
		},
		{
			name:      "bad",
			selector:  sql.Select().Offset(-1),
			wantQuery: `SELECT *`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.wantQuery, test.selector.String())
		})
	}
}

func TestSelector_Iterate(t *testing.T) {
	ctx := context.Background()

	// Mock two results
	cursor := NewMockCursor()
	cursor.ColumnsFunc.SetDefaultReturn([]string{"name", "email"}, nil)
	cursor.NextFunc.PushReturn(true)
	cursor.NextFunc.PushReturn(true)
	cursor.ScanFunc.PushHook(func(dest ...interface{}) error {
		assert.Len(t, dest, 2)
		return nil
	})

	executor := NewMockExecutor()
	executor.QueryFunc.SetDefaultReturn(cursor, nil)

	typer := NewMockTyper()
	typer.ValuerFunc.SetDefaultHook(func(v interface{}) interface{} {
		return v
	})

	adapter := NewMockAdapter()
	adapter.ExecutorFunc.SetDefaultReturn(executor)
	adapter.TyperFunc.SetDefaultReturn(typer)
	adapter.FormatSQLFunc.SetDefaultHook(func(sql string) string {
		return exql.StripWhitespace(sql)
	})

	tmpl := defaultTemplate(t)
	sqlb := New(adapter, tmpl)

	dest := make([]map[string]interface{}, 0)
	err := sqlb.Select().All(ctx, &dest)
	assert.NoError(t, err)
	mockrequire.Called(t, cursor.ScanFunc)

	err = sqlb.Select().One(ctx, &dest)
	assert.EqualError(t, err, sql.ErrNoRows.Error())
}
