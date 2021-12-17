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

	t.Run("unsupported type", func(t *testing.T) {
		_, err := Column(857).Compile(tmpl)
		assert.Error(t, err)
	})

	tests := []struct {
		name   string
		column *ColumnFragment
		want   string
	}{
		{
			name:   "normal",
			column: Column("users.name"),
			want:   `"users"."name"`,
		},
		{
			name:   "explicit as",
			column: Column("users.name as foo"),
			want:   `"users"."name" AS "foo"`,
		},
		{
			name:   "implicit as",
			column: Column("users.name foo"),
			want:   `"users"."name" AS "foo"`,
		},
		{
			name:   "raw",
			column: Column(Raw("users.name As foo")),
			want:   `users.name As foo`,
		},
		{
			name:   "with asterisk",
			column: Column("*"),
			want:   `*`,
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
		got, err := Column("users.name").Compile(tmpl)
		assert.NoError(t, err)
		assert.Equal(t, `"users"."name"`, got)
	})
}
