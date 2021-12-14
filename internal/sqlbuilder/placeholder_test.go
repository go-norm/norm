// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package sqlbuilder

import (
	"database/sql/driver"
	"testing"

	"github.com/stretchr/testify/assert"

	"unknwon.dev/norm/expr"
)

var _ driver.Valuer = (*mockDriverValuer)(nil)

type mockDriverValuer struct {
	value string
}

func (m mockDriverValuer) Value() (driver.Value, error) {
	return m.value, nil
}

func TestExpandQuery(t *testing.T) {
	mockCompilable := NewMockCompilable()
	mockCompilable.CompileFunc.SetDefaultReturn("DISTINCT(id)", nil)

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
			wantArgs:  []interface{}{},
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
				[]interface{}{4, 5},
				[]interface{}{},
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
			name:      "driver.Valuer",
			query:     "?, ?, ?",
			args:      []interface{}{1, mockDriverValuer{value: "bar"}, 3},
			wantQuery: "?, ?, ?",
			wantArgs:  []interface{}{1, mockDriverValuer{value: "bar"}, 3},
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
			assert.Equal(t, test.wantQuery, gotQuery)
			assert.Equal(t, test.wantArgs, gotArgs)
		})
	}
}
