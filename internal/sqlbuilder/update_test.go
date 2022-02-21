// Copyright 2022 Joe Chen. All rights reserved.
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

func TestUpdater(t *testing.T) {
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
		updater   norm.Updater
		wantQuery string
		wantArgs  []interface{}
	}{
		{
			name: "single column",
			updater: sql.
				Update("users").
				Set("first_name", "john"),
			wantQuery: `UPDATE "users" SET "first_name" = ?`,
			wantArgs:  []interface{}{"john"},
		},
		{
			name: "multiple columns",
			updater: sql.
				Update("users").
				Set(
					"first_name", "john",
					"last_name", "smith",
				),
			wantQuery: `UPDATE "users" SET "first_name" = ?, "last_name" = ?`,
			wantArgs:  []interface{}{"john", "smith"},
		},
		{
			name: "multiple sets",
			updater: sql.
				Update("users").
				Set("first_name", "john").
				Set("last_name", "smith"),
			wantQuery: `UPDATE "users" SET "first_name" = ?, "last_name" = ?`,
			wantArgs:  []interface{}{"john", "smith"},
		},
		{
			name: "where",
			updater: sql.
				Update("users").
				Set("first_name", "john").
				Set("age", 18).
				Where("id = ?", 1),
			wantQuery: `UPDATE "users" SET "first_name" = ?, "age" = ? WHERE id = ?`,
			wantArgs:  []interface{}{"john", 18, 1},
		},
		{
			name: "and",
			updater: sql.
				Update("users").
				Set("first_name", "john").
				Set("age", 18).
				Where("id = ?", 1).
				And(expr.NewConstraint("delete_at", expr.IsNull())),
			wantQuery: `UPDATE "users" SET "first_name" = ?, "age" = ? WHERE id = ? AND "delete_at" IS NULL`,
			wantArgs:  []interface{}{"john", 18, 1},
		},
		{
			name: "raw",
			updater: sql.
				Update("users").
				Set("first_name", "john").
				Set("age", expr.Raw("18")).
				Where("id = ?", 1),
			wantQuery: `UPDATE "users" SET "first_name" = ?, "age" = 18 WHERE id = ?`,
			wantArgs:  []interface{}{"john", 1},
		},
		{
			name: "func",
			updater: sql.
				Update("users").
				Set("first_name", "john").
				Set("age", expr.Func("MAX", expr.Raw("age"))).
				Where("id = ?", 1),
			wantQuery: `UPDATE "users" SET "first_name" = ?, "age" = MAX(age) WHERE id = ?`,
			wantArgs:  []interface{}{"john", 1},
		},
		{
			name: "value",
			updater: sql.
				Update("users").
				Set("first_name", exql.Value("john")).
				Set("age", 18).
				Where("id = ?", 1),
			wantQuery: `UPDATE "users" SET "first_name" = 'john', "age" = ? WHERE id = ?`,
			wantArgs:  []interface{}{18, 1},
		},
		{
			name: "returning",
			updater: sql.
				Update("users").
				Set("first_name", exql.Value("john")).
				Set("age", 18).
				Where("id = ?", 1).
				Returning("first_name", "age"),
			wantQuery: `UPDATE "users" SET "first_name" = 'john', "age" = ? WHERE id = ? RETURNING "first_name", "age"`,
			wantArgs:  []interface{}{18, 1},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.wantQuery, test.updater.String())
			assert.Equal(t, test.wantArgs, test.updater.Arguments())
		})
	}
}

func TestUpdater_Amend(t *testing.T) {
	adapter := NewMockAdapter()
	adapter.FormatSQLFunc.SetDefaultHook(func(sql string) string {
		return exql.StripWhitespace(sql)
	})

	tmpl := defaultTemplate(t)
	got := New(adapter, tmpl).
		Update("users").
		Set("name", "alice").
		Amend(func(query string) string {
			return query + " RETURNING id"
		}).
		String()
	want := `UPDATE "users" SET "name" = ? RETURNING id`
	assert.Equal(t, want, got)
}

func TestUpdater_Iterate(t *testing.T) {
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
	err := sqlb.Update("users").All(ctx, &dest)
	assert.NoError(t, err)
	mockrequire.Called(t, cursor.ScanFunc)

	err = sqlb.Update("users").One(ctx, &dest)
	assert.EqualError(t, err, sql.ErrNoRows.Error())
}
