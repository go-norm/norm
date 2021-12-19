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
)

func TestOrderBy(t *testing.T) {
	tmpl := defaultTemplate(t)

	t.Run("empty", func(t *testing.T) {
		got, err := OrderBy().Compile(tmpl)
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	gb := GroupBy(
		Column("id"),
		Column("customer"),
		Column("service_id"),
		Column("users.name"),
		Column("users.id"),
	)

	got, err := gb.Compile(tmpl)
	require.NoError(t, err)

	want := `GROUP BY "id", "customer", "service_id", "users"."name", "users"."id"`
	assert.Equal(t, want, strings.TrimSpace(got))

	t.Run("cache hit", func(t *testing.T) {
		got, err := gb.Compile(tmpl)
		assert.NoError(t, err)
		assert.Equal(t, want, strings.TrimSpace(got))
	})
}

func TestOrderBy2(t *testing.T) {
	o := OrderBy(
		JoinSortColumns(
			&SortColumnFragment{Column: &ColumnFragment{Name: "foo"}},
		),
	)

	s := mustTrim(o.Compile(defaultTemplate))
	e := `ORDER BY "foo"`
	if s != e {
		t.Fatalf("Got: %s, Expecting: %s", s, e)
	}
}

func TestOrderByRaw(t *testing.T) {
	o := OrderBy(
		JoinSortColumns(
			&SortColumnFragment{Column: RawValue("CASE WHEN id IN ? THEN 0 ELSE 1 END")},
		),
	)

	s := mustTrim(o.Compile(defaultTemplate))
	e := `ORDER BY CASE WHEN id IN ? THEN 0 ELSE 1 END`
	if s != e {
		t.Fatalf("Got: %s, Expecting: %s", s, e)
	}
}

func TestOrderByDesc(t *testing.T) {
	o := OrderBy(
		JoinSortColumns(
			&SortColumnFragment{Column: &ColumnFragment{Name: "foo"}, SortOrder: SortDescendent},
		),
	)

	s := mustTrim(o.Compile(defaultTemplate))
	e := `ORDER BY "foo" DESC`
	if s != e {
		t.Fatalf("Got: %s, Expecting: %s", s, e)
	}
}
