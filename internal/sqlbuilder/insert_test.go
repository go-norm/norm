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
	"unknwon.dev/norm/internal/exql"
)

func TestInserter(t *testing.T) {
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
		inserter  norm.Inserter
		wantQuery string
		wantArgs  []interface{}
	}{
		{
			name: "single",
			inserter: sql.
				InsertInto("users").
				Columns("first_name", "last_name", "age").
				Values("alice", "john", 9),
			wantQuery: `INSERT INTO "users" ("first_name", "last_name", "age") VALUES (?, ?, ?)`,
			wantArgs:  []interface{}{"alice", "john", 9},
		},
		{
			name: "multiple",
			inserter: sql.
				InsertInto("users").
				Columns("first_name", "last_name", "age").
				Values("alice", "john", 9).
				Values("bob", "youth", 10),
			wantQuery: `INSERT INTO "users" ("first_name", "last_name", "age") VALUES (?, ?, ?), (?, ?, ?)`,
			wantArgs:  []interface{}{"alice", "john", 9, "bob", "youth", 10},
		},
		{
			name: "raw",
			inserter: sql.
				InsertInto("users").
				Columns("first_name", "last_name", "age").
				Values("alice", exql.Raw("john"), 9),
			wantQuery: `INSERT INTO "users" ("first_name", "last_name", "age") VALUES (?, john, ?)`,
			wantArgs:  []interface{}{"alice", 9},
		},
		{
			name: "value",
			inserter: sql.
				InsertInto("users").
				Columns("first_name", "last_name", "age").
				Values("alice", exql.Value("john"), 9),
			wantQuery: `INSERT INTO "users" ("first_name", "last_name", "age") VALUES (?, 'john', ?)`,
			wantArgs:  []interface{}{"alice", 9},
		},
		{
			name: "returning",
			inserter: sql.
				InsertInto("users").
				Columns("first_name", "last_name", "age").
				Values("alice", "john", 9).
				Returning("id"),
			wantQuery: `INSERT INTO "users" ("first_name", "last_name", "age") VALUES (?, ?, ?) RETURNING "id"`,
			wantArgs:  []interface{}{"alice", "john", 9},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.wantQuery, test.inserter.String())
			assert.Equal(t, test.wantArgs, test.inserter.Arguments())
		})
	}
}

func TestInserter_Amend(t *testing.T) {
	adapter := NewMockAdapter()
	adapter.FormatSQLFunc.SetDefaultHook(func(sql string) string {
		return exql.StripWhitespace(sql)
	})

	tmpl := defaultTemplate(t)
	got := New(adapter, tmpl).
		InsertInto("users").
		Columns("name").
		Values("alice").
		Amend(func(query string) string {
			return query + " RETURNING id"
		}).
		String()
	want := `INSERT INTO "users" ("name") VALUES (?) RETURNING id`
	assert.Equal(t, want, got)
}

func TestInserter_Iterate(t *testing.T) {
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
	err := sqlb.InsertInto("users").All(ctx, &dest)
	assert.NoError(t, err)
	mockrequire.Called(t, cursor.ScanFunc)

	err = sqlb.InsertInto("users").One(ctx, &dest)
	assert.EqualError(t, err, sql.ErrNoRows.Error())
}
