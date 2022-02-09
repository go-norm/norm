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
