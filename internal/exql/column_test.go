// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package exql

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestColumn(t *testing.T) {
	tmpl := defaultTemplate(t)
	tests := []struct {
		name   string
		column *Column
		want   string
	}{
		{
			name:   "normal",
			column: ColumnWithName("users.name"),
			want:   `"users"."name"`,
		},
		{
			name:   "explicit as",
			column: ColumnWithName("users.name as foo"),
			want:   `"users"."name" AS "foo"`,
		},
		{
			name:   "implicit as",
			column: ColumnWithName("users.name foo"),
			want:   `"users"."name" AS "foo"`,
		},
		{
			name: "raw",
			column: &Column{
				Name: RawValue("users.name As foo"),
			},
			want: `users.name As foo`,
		},
		{
			name:   "with asterisk",
			column: ColumnWithName("*"),
			want:   `*`,
		},
		{
			name: "default fallback",
			column: &Column{
				Name: 857,
			},
			want: "857",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := test.column.Compile(tmpl)
			assert.NoError(t, err)
			assert.Equal(t, test.want, got)
		})
	}

	t.Run("cache hit", func(t *testing.T) {
		got, err := ColumnWithName("users.name").Compile(tmpl)
		assert.NoError(t, err)
		assert.Equal(t, `"users"."name"`, got)
	})
}

func TestColumn_Hash(t *testing.T) {
	got := ColumnWithName("users.name").Hash()
	want := "*exql.Column:3121935903895129804"
	assert.Equal(t, want, got)
}
