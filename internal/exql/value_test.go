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

func TestValue(t *testing.T) {
	tmpl := defaultTemplate(t)
	tests := []struct {
		name  string
		value *ValueFragment
		want  string
	}{
		{
			name:  "string",
			value: Value("John"),
			want:  `'John'`,
		},
		{
			name:  "int",
			value: Value(1),
			want:  `'1'`,
		},
		{
			name:  "raw",
			value: Value(Raw("NOW()")),
			want:  `NOW()`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := test.value.Compile(tmpl)
			assert.NoError(t, err)
			assert.Equal(t, test.want, got)
		})
	}

	t.Run("cache hit", func(t *testing.T) {
		got, err := Value(1).Compile(tmpl)
		assert.NoError(t, err)
		assert.Equal(t, `'1'`, got)
	})
}

func TestValuesGroup(t *testing.T) {
	tmpl := defaultTemplate(t)

	t.Run("empty", func(t *testing.T) {
		got, err := ValuesGroup().Compile(tmpl)
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	vg := ValuesGroup(
		Value("John"),
		Value(1),
		Value(Raw("NOW()")),
	)

	got, err := vg.Compile(tmpl)
	require.NoError(t, err)

	want := `('John', '1', NOW())`
	assert.Equal(t, want, got)

	t.Run("cache hit", func(t *testing.T) {
		got, err := vg.Compile(tmpl)
		assert.NoError(t, err)
		assert.Equal(t, want, got)
	})
}

func TestValuesGroups(t *testing.T) {
	tmpl := defaultTemplate(t)

	t.Run("empty", func(t *testing.T) {
		got, err := ValuesGroups().Compile(tmpl)
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	vgs := ValuesGroups(
		ValuesGroup(
			Value("John"),
			Value(1),
			Value(Raw("NOW()")),
		),
		ValuesGroup(
			Value("Joe"),
			Value(2),
			Value(Raw("NOW()")),
		),
	)

	got, err := vgs.Compile(tmpl)
	require.NoError(t, err)

	want := `('John', '1', NOW()), ('Joe', '2', NOW())`
	assert.Equal(t, want, got)

	t.Run("cache hit", func(t *testing.T) {
		got, err := vgs.Compile(tmpl)
		assert.NoError(t, err)
		assert.Equal(t, want, got)
	})
}
