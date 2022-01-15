// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package sqlbuilder

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"unknwon.dev/norm/expr"
	"unknwon.dev/norm/internal/exql"
)

//go:generate go-mockgen --force database/sql/driver -i Valuer -o mock_driver_valuer_test.go
func TestExpandQuery(t *testing.T) {
	mockCompilable := NewMockCompilable()
	mockCompilable.CompileFunc.SetDefaultReturn("DISTINCT(id)", nil)

	mockValuer := NewMockValuer()
	mockValuer.ValueFunc.SetDefaultReturn("bar", nil)

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
		query     string
		args      []interface{}
		wantQuery string
		wantArgs  []interface{}
	}{
		{
			name:      "one",
			query:     "?",
			args:      []interface{}{1},
			wantQuery: "?",
			wantArgs:  []interface{}{1},
		},
		{
			name:      "no args",
			query:     "?",
			args:      nil,
			wantQuery: "?",
			wantArgs:  []interface{}(nil),
		},
		{
			name:      "nil args",
			query:     "?",
			args:      []interface{}{nil},
			wantQuery: "NULL",
			wantArgs:  []interface{}{},
		},

		{
			name:      "many",
			query:     "?, ?, ?",
			args:      []interface{}{1, 2, 3},
			wantQuery: "?, ?, ?",
			wantArgs:  []interface{}{1, 2, 3},
		},

		{
			name:  "array",
			query: "?, ?, ?",
			args: []interface{}{
				1,
				2,
				[]interface{}{3, 4, 5},
			},
			wantQuery: "?, ?, (?, ?, ?)",
			wantArgs:  []interface{}{1, 2, 3, 4, 5},
		},
		{
			name:  "array",
			query: "?, ?, ?",
			args: []interface{}{
				[]interface{}{1, 2, 3},
				4,
				5,
			},
			wantQuery: "(?, ?, ?), ?, ?",
			wantArgs:  []interface{}{1, 2, 3, 4, 5},
		},
		{
			name:  "array",
			query: "?, ?, ?",
			args: []interface{}{
				1,
				[]interface{}{2, 3, 4},
				5,
			},
			wantQuery: "?, (?, ?, ?), ?",
			wantArgs:  []interface{}{1, 2, 3, 4, 5},
		},
		{
			name:  "array",
			query: "?, ?",
			args: []interface{}{
				[]interface{}{1, 2, 3},
				[]interface{}{4, 5},
			},
			wantQuery: "(?, ?, ?), (?, ?)",
			wantArgs:  []interface{}{1, 2, 3, 4, 5},
		},
		{
			name:  "array",
			query: "???",
			args: []interface{}{
				1,
				[]interface{}{2, 3, 4},
				5,
			},
			wantQuery: "?(?, ?, ?)?",
			wantArgs:  []interface{}{1, 2, 3, 4, 5},
		},
		{
			name:  "array",
			query: "??",
			args: []interface{}{
				[]interface{}{1, 2, 3},
				[]interface{}{},
				[]interface{}{4, 5},
				[]interface{}{},
			},
			wantQuery: "(?, ?, ?)(NULL)",
			wantArgs: []interface{}{
				1,
				2,
				3,
			},
		},

		{
			name:      "raw",
			query:     "?, ?, ?",
			args:      []interface{}{1, expr.Raw("foo"), 3},
			wantQuery: "?, foo, ?",
			wantArgs:  []interface{}{1, 3},
		},

		{
			name:      "compilable",
			query:     "?, ?, ?",
			args:      []interface{}{1, mockCompilable, 3},
			wantQuery: "?, (DISTINCT(id)), ?",
			wantArgs:  []interface{}{1, 3},
		},
		{
			name:  "compilable - selector",
			query: "EXISTS ?",
			args: []interface{}{
				sql.Select(expr.Raw("1")).
					From("users").
					Where("name = ?", "alice"),
			},
			wantQuery: `EXISTS (SELECT 1 FROM "users" WHERE name = ?)`,
			wantArgs:  []interface{}{"alice"},
		},

		{
			name:      "driver.Valuer",
			query:     "?, ?, ?",
			args:      []interface{}{1, mockValuer, 3},
			wantQuery: "?, ?, ?",
			wantArgs:  []interface{}{1, mockValuer, 3},
		},

		{
			name:      "byte slice",
			query:     "?, ?, ?",
			args:      []interface{}{1, []byte("baz"), 3},
			wantQuery: "?, ?, ?",
			wantArgs:  []interface{}{1, []byte("baz"), 3},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gotQuery, gotArgs, err := ExpandQuery(test.query, test.args)
			assert.NoError(t, err)
			assert.Equal(t, test.wantQuery, exql.StripWhitespace(gotQuery))
			assert.Equal(t, test.wantArgs, gotArgs)
		})
	}
}

func defaultTemplate(t testing.TB) *exql.Template {
	tmpl, err := exql.DefaultTemplate()
	if err != nil {
		t.Fatalf("Failed to get default template: %v", err)
	}
	return tmpl
}
