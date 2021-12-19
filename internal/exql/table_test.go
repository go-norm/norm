// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package exql

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTable(t *testing.T) {
	tmpl := defaultTemplate(t)

	t.Run("unsupported type", func(t *testing.T) {
		_, err := Table(857).Compile(tmpl)
		assert.Error(t, err)
	})

	tests := []struct {
		name  string
		table *TableFragment
		want  string
	}{
		{
			name:  "normal",
			table: Table("users"),
			want:  `"users"`,
		},
		{
			name:  "explicit as",
			table: Table("users as foo"),
			want:  `"users" AS "foo"`,
		},
		{
			name:  "implicit as",
			table: Table("users foo"),
			want:  `"users" AS "foo"`,
		},
		{
			name:  "raw",
			table: Table(Raw("users As foo")),
			want:  `users As foo`,
		},

		{
			name:  "normal with column",
			table: Table("users.name"),
			want:  `"users"."name"`,
		},
		{
			name:  "explicit as with column",
			table: Table("users.name as foo"),
			want:  `"users"."name" AS "foo"`,
		},
		{
			name:  "implicit as with column",
			table: Table("users.name foo"),
			want:  `"users"."name" AS "foo"`,
		},
		{
			name:  "raw with column",
			table: Table(Raw("users.name As foo")),
			want:  `users.name As foo`,
		},
		{
			name:  "with asterisk",
			table: Table("*"),
			want:  `"*"`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := test.table.Compile(tmpl)
			assert.NoError(t, err)
			assert.Equal(t, test.want, got)
		})
	}

	t.Run("cache hit", func(t *testing.T) {
		got, err := Table("users").Compile(tmpl)
		assert.NoError(t, err)
		assert.Equal(t, `"users"`, got)
	})
}
