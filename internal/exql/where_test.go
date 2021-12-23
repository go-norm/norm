// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package exql

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"unknwon.dev/norm/expr"
)

func TestWhere(t *testing.T) {
	w := Where(
		ColumnValue("id", expr.ComparisonGreaterThan, Raw("8")),
		ColumnValue("other.id", expr.ComparisonLessThan, Raw("100")),
		ColumnValue("name", expr.ComparisonEqual, Raw(`'Haruki Murakami'`)),
		ColumnValue("created", expr.ComparisonGreaterThanOrEqualTo, Raw("NOW()")),
		ColumnValue("modified", expr.ComparisonLessThanOrEqualTo, Raw("NOW()")),
	)
	tmpl := defaultTemplate(t)

	got, err := w.Compile(tmpl)
	require.NoError(t, err)

	want := `WHERE "id" > 8 AND "other"."id" < 100 AND "name" = 'Haruki Murakami' AND "created" >= NOW() AND "modified" <= NOW()`
	assert.Equal(t, want, strings.TrimSpace(got))

	t.Run("cache hit", func(t *testing.T) {
		got, err := w.Compile(tmpl)
		assert.NoError(t, err)
		assert.Equal(t, want, strings.TrimSpace(got))
	})
}

func TestWhere_Append(t *testing.T) {
	w := Where()
	tmpl := defaultTemplate(t)

	got, err := w.Compile(tmpl)
	require.NoError(t, err)
	assert.Empty(t, got)

	w.Append(
		ColumnValue("id", expr.ComparisonGreaterThan, Raw("8")),
	)
	got, err = w.Compile(tmpl)
	require.NoError(t, err)

	want := `WHERE "id" > 8`
	assert.Equal(t, want, strings.TrimSpace(got))
}

func TestAnd(t *testing.T) {
	tmpl := defaultTemplate(t)

	t.Run("empty", func(t *testing.T) {
		got, err := And().Compile(tmpl)
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	and := And(
		ColumnValue("id", expr.ComparisonGreaterThan, Raw("8")),
		ColumnValue("other.id", expr.ComparisonLessThan, Raw("100")),
		ColumnValue("name", expr.ComparisonEqual, Raw(`'Haruki Murakami'`)),
		ColumnValue("created", expr.ComparisonGreaterThanOrEqualTo, Raw("NOW()")),
		ColumnValue("modified", expr.ComparisonLessThanOrEqualTo, Raw("NOW()")),
	)

	got, err := and.Compile(tmpl)
	require.NoError(t, err)

	want := `("id" > 8 AND "other"."id" < 100 AND "name" = 'Haruki Murakami' AND "created" >= NOW() AND "modified" <= NOW())`
	assert.Equal(t, want, got)

	t.Run("cache hit", func(t *testing.T) {
		got, err := and.Compile(tmpl)
		assert.NoError(t, err)
		assert.Equal(t, want, got)
	})
}

func TestOr(t *testing.T) {
	tmpl := defaultTemplate(t)

	t.Run("empty", func(t *testing.T) {
		got, err := Or().Compile(tmpl)
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	and := Or(
		ColumnValue("id", expr.ComparisonGreaterThan, Raw("8")),
		ColumnValue("other.id", expr.ComparisonLessThan, Raw("100")),
		ColumnValue("name", expr.ComparisonEqual, Raw(`'Haruki Murakami'`)),
		ColumnValue("created", expr.ComparisonGreaterThanOrEqualTo, Raw("NOW()")),
		ColumnValue("modified", expr.ComparisonLessThanOrEqualTo, Raw("NOW()")),
	)

	got, err := and.Compile(tmpl)
	require.NoError(t, err)

	want := `("id" > 8 OR "other"."id" < 100 OR "name" = 'Haruki Murakami' OR "created" >= NOW() OR "modified" <= NOW())`
	assert.Equal(t, want, got)

	t.Run("cache hit", func(t *testing.T) {
		got, err := and.Compile(tmpl)
		assert.NoError(t, err)
		assert.Equal(t, want, got)
	})
}

func TestWhere_And_Or(t *testing.T) {
	w := Where(
		And(
			ColumnValue("id", expr.ComparisonGreaterThan, Raw("8")),
			ColumnValue("id", expr.ComparisonLessThan, Raw("99")),
			Or(
				ColumnValue("age", expr.ComparisonLessThan, Raw("18")),
				ColumnValue("age", expr.ComparisonGreaterThan, Raw("41")),
			),
		),
		ColumnValue("name", expr.ComparisonEqual, Raw(`'John'`)),
		Or(
			ColumnValue("last_name", expr.ComparisonEqual, Raw(`'Smith'`)),
			ColumnValue("last_name", expr.ComparisonEqual, Raw(`'Reyes'`)),
		),
	).Append(
		Raw("city_id = 728"),
	)

	got, err := w.Compile(defaultTemplate(t))
	require.NoError(t, err)

	want := StripWhitespace(`
WHERE
	(
			"id" > 8
		AND "id" < 99
		AND ("age" < 18 OR "age" > 41)
	)
AND "name" = 'John'
AND ("last_name" = 'Smith' OR "last_name" = 'Reyes')
AND city_id = 728
`)
	assert.Equal(t, want, strings.TrimSpace(got))
}
