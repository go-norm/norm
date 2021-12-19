// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package exql

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestColumns(t *testing.T) {
	cs := Columns(
		Column("id"),
		Column("customer"),
		Column("service_id"),
		Column("users.name"),
		Column("users.id"),
	)
	tmpl := defaultTemplate(t)

	got, err := cs.Compile(tmpl)
	require.NoError(t, err)

	want := `"id", "customer", "service_id", "users"."name", "users"."id"`
	assert.Equal(t, want, got)

	t.Run("cache hit", func(t *testing.T) {
		got, err := cs.Compile(tmpl)
		assert.NoError(t, err)
		assert.Equal(t, want, got)
	})
}

func TestColumns_Append(t *testing.T) {
	cs := Columns()
	tmpl := defaultTemplate(t)

	got, err := cs.Compile(tmpl)
	require.NoError(t, err)
	assert.Empty(t, got)

	cs.Append(
		Column("id"),
	)
	got, err = cs.Compile(tmpl)
	require.NoError(t, err)

	want := `"id"`
	assert.Equal(t, want, got)
}
