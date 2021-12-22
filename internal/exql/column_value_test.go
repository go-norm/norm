// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package exql

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"unknwon.dev/norm/expr"
)

func TestColumnValue(t *testing.T) {
	tmpl := defaultTemplate(t)
	tests := []struct {
		name        string
		columnValue *ColumnValueFragment
		want        string
	}{
		{
			name:        "compare value",
			columnValue: ColumnValue("id", expr.ComparisonEqual, Raw("1")),
			want:        `"id" = 1`,
		},
		{
			name:        "compare func",
			columnValue: ColumnValue("date", expr.ComparisonGreaterThan, Raw("NOW()")),
			want:        `"date" > NOW()`,
		},
		{
			name:        "compare func",
			columnValue: ColumnValue(Raw(`'{"a":1,"b":2}'::json`), "->", Value("b")),
			want:        `'{"a":1,"b":2}'::json -> 'b'`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := test.columnValue.Compile(tmpl)
			assert.NoError(t, err)
			assert.Equal(t, test.want, got)
		})
	}

	t.Run("cache hit", func(t *testing.T) {
		got, err := ColumnValue("id", expr.ComparisonEqual, Raw("1")).Compile(tmpl)
		assert.NoError(t, err)
		assert.Equal(t, `"id" = 1`, got)
	})
}

func TestColumnValues(t *testing.T) {
	cvs := ColumnValues(
		ColumnValue("id", expr.ComparisonGreaterThan, Raw("8")),
		ColumnValue("other.id", expr.ComparisonLessThan, Raw("100")),
		ColumnValue("name", expr.ComparisonEqual, Raw(`'Haruki Murakami'`)),
		ColumnValue("created", expr.ComparisonGreaterThanOrEqualTo, Raw("NOW()")),
		ColumnValue("modified", expr.ComparisonLessThanOrEqualTo, Raw("NOW()")),
	)
	tmpl := defaultTemplate(t)

	got, err := cvs.Compile(tmpl)
	require.NoError(t, err)

	want := `"id" > 8, "other"."id" < 100, "name" = 'Haruki Murakami', "created" >= NOW(), "modified" <= NOW()`
	assert.Equal(t, want, got)

	t.Run("cache hit", func(t *testing.T) {
		got, err := cvs.Compile(tmpl)
		assert.NoError(t, err)
		assert.Equal(t, want, got)
	})
}

func TestColumnValues_Append(t *testing.T) {
	cvs := ColumnValues()
	tmpl := defaultTemplate(t)

	got, err := cvs.Compile(tmpl)
	require.NoError(t, err)
	assert.Empty(t, got)

	cvs.Append(
		ColumnValue("id", expr.ComparisonGreaterThan, Raw("8")),
	)
	got, err = cvs.Compile(tmpl)
	require.NoError(t, err)

	want := `"id" > 8`
	assert.Equal(t, want, got)
}
