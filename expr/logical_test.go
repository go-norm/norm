// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package expr

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogical_Empty(t *testing.T) {
	empty := Logical(LogicalNone)
	assert.True(t, empty.Empty(), "should be empty")
}

func TestAnd_Or_Raw(t *testing.T) {
	expr := And(
		Raw("id = 1"),
		And(
			Raw("name = 'Joe'"),
			Or(
				Raw("year = 2021"),
				Cond{
					"year =": "2022",
				},
			).Or(
				Raw("year = 2023"),
			),
		),
	).And(
		Logical(LogicalNone),
	)

	assert.Equal(t, LogicalAnd, expr.Operator())
	assert.False(t, expr.Empty(), "should not be empty")

	want := "(AND id = 1 (AND name = 'Joe' (OR year = 2021 (AND year = 2022) year = 2023)) NONE)"
	assert.Equal(t, want, expr.String())
}

func TestRaw(t *testing.T) {
	t.Run("no argument", func(t *testing.T) {
		raw := Raw("DISTINCT(id)")
		assert.Equal(t, "DISTINCT(id)", raw.Raw())
		assert.Equal(t, []interface{}(nil), raw.Arguments())
	})

	t.Run("has arguments", func(t *testing.T) {
		raw := Raw("SELECT * FROM users WHERE id = ?", 1)
		assert.Equal(t, "SELECT * FROM users WHERE id = ?", raw.Raw())
		assert.Equal(t, []interface{}{1}, raw.Arguments())
		assert.Equal(t, []LogicalExpr{raw}, raw.Expressions())
		assert.Equal(t, LogicalNone, raw.Operator())
		assert.False(t, raw.Empty(), "should not be empty")
	})
}

func TestBool(t *testing.T) {
	assert.Equal(t, "TRUE", Bool(true).String())
	assert.Equal(t, "FALSE", Bool(false).String())
}

func TestFunc(t *testing.T) {
	tests := []struct {
		name string
		args []interface{}
	}{
		{
			name: "HELLO",
			args: nil,
		},
		{
			name: "MOD",
			args: []interface{}{29, 9},
		},
		{
			name: "CONCAT",
			args: []interface{}{"a", "b", "c"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fn := Func(test.name, test.args...)
			assert.Equal(t, test.name, fn.Name())
			assert.Equal(t, test.args, fn.Arguments())
		})
	}
}
