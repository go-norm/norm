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

func TestOrderBy(t *testing.T) {
	tmpl := defaultTemplate(t)

	t.Run("empty", func(t *testing.T) {
		got, err := OrderBy().Compile(tmpl)
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	ob := OrderBy(
		SortColumn("id"),
		SortColumn("customer"),
		SortColumn("service_id", SortAscendant),
		SortColumn("users.name", SortDescendent),
		SortColumn("users.id"),
		SortColumn(Raw("CASE WHEN id IN ? THEN 0 ELSE 1 END")),
	)

	got, err := ob.Compile(tmpl)
	require.NoError(t, err)

	want := `ORDER BY "id", "customer", "service_id" ASC, "users"."name" DESC, "users"."id", CASE WHEN id IN ? THEN 0 ELSE 1 END`
	assert.Equal(t, want, stripWhitespace(got))

	t.Run("cache hit", func(t *testing.T) {
		got, err := ob.Compile(tmpl)
		assert.NoError(t, err)
		assert.Equal(t, want, stripWhitespace(got))
	})
}

func TestSortColumn(t *testing.T) {
	sc := SortColumn("id")
	tmpl := defaultTemplate(t)
	got, err := sc.Compile(tmpl)
	require.NoError(t, err)

	want := `"id"`
	assert.Equal(t, want, stripWhitespace(got))

	t.Run("cache hit", func(t *testing.T) {
		got, err := sc.Compile(tmpl)
		assert.NoError(t, err)
		assert.Equal(t, want, stripWhitespace(got))
	})
}
