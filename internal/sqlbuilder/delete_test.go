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

func TestDeleter(t *testing.T) {
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
		deleter   norm.Deleter
		wantQuery string
		wantArgs  []interface{}
	}{
		{
			name: "all",
			deleter: sql.
				DeleteFrom("users"),
			wantQuery: `DELETE FROM "users"`,
			wantArgs:  []interface{}(nil),
		},
		{
			name: "where",
			deleter: sql.
				DeleteFrom("users").
				Where("id = ?", 1),
			wantQuery: `DELETE FROM "users" WHERE id = ?`,
			wantArgs:  []interface{}{1},
		},
		{
			name: "and",
			deleter: sql.
				DeleteFrom("users").
				Where("id = ?", 1).
				And(expr.NewConstraint("delete_at", expr.IsNull())),
			wantQuery: `DELETE FROM "users" WHERE id = ? AND "delete_at" IS NULL`,
			wantArgs:  []interface{}{1},
		},
		{
			name: "returning",
			deleter: sql.
				DeleteFrom("users").
				Where("id = ?", 1).
				Returning("first_name", "age"),
			wantQuery: `DELETE FROM "users" WHERE id = ? RETURNING "first_name", "age"`,
			wantArgs:  []interface{}{1},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.wantQuery, test.deleter.String())
			assert.Equal(t, test.wantArgs, test.deleter.Arguments())
		})
	}
}

func TestDeleter_Amend(t *testing.T) {
	adapter := NewMockAdapter()
	adapter.FormatSQLFunc.SetDefaultHook(func(sql string) string {
		return exql.StripWhitespace(sql)
	})

	tmpl := defaultTemplate(t)
	got := New(adapter, tmpl).
		DeleteFrom("users").
		Amend(func(query string) string {
			return query + " RETURNING id"
		}).
		String()
	want := `DELETE FROM "users" RETURNING id`
	assert.Equal(t, want, got)
}

func TestDeleter_Iterate(t *testing.T) {
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
	err := sqlb.DeleteFrom("users").All(ctx, &dest)
	assert.NoError(t, err)
	mockrequire.Called(t, cursor.ScanFunc)

	err = sqlb.DeleteFrom("users").One(ctx, &dest)
	assert.EqualError(t, err, sql.ErrNoRows.Error())
}
