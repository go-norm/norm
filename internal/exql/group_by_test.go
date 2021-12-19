// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package exql

import (
	"testing"
)

func TestGroupBy(t *testing.T) {
	columns := GroupByColumns(
		&ColumnFragment{Name: "id"},
		&ColumnFragment{Name: "customer"},
		&ColumnFragment{Name: "service_id"},
		&ColumnFragment{Name: "role.name"},
		&ColumnFragment{Name: "role.id"},
	)

	s := mustTrim(columns.Compile(defaultTemplate))
	e := `GROUP BY "id", "customer", "service_id", "role"."name", "role"."id"`
	if s != e {
		t.Fatalf("Got: %s, Expecting: %s", s, e)
	}
}

func BenchmarkGroupByColumns(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = GroupByColumns(
			&ColumnFragment{Name: "a"},
			&ColumnFragment{Name: "b"},
			&ColumnFragment{Name: "c"},
		)
	}
}

func BenchmarkGroupByHash(b *testing.B) {
	c := GroupByColumns(
		&ColumnFragment{Name: "id"},
		&ColumnFragment{Name: "customer"},
		&ColumnFragment{Name: "service_id"},
		&ColumnFragment{Name: "role.name"},
		&ColumnFragment{Name: "role.id"},
	)
	for i := 0; i < b.N; i++ {
		c.Hash()
	}
}

func BenchmarkGroupByCompile(b *testing.B) {
	c := GroupByColumns(
		&ColumnFragment{Name: "id"},
		&ColumnFragment{Name: "customer"},
		&ColumnFragment{Name: "service_id"},
		&ColumnFragment{Name: "role.name"},
		&ColumnFragment{Name: "role.id"},
	)
	for i := 0; i < b.N; i++ {
		_, _ = c.Compile(defaultTemplate)
	}
}

func BenchmarkGroupByCompileNoCache(b *testing.B) {
	for i := 0; i < b.N; i++ {
		c := GroupByColumns(
			&ColumnFragment{Name: "id"},
			&ColumnFragment{Name: "customer"},
			&ColumnFragment{Name: "service_id"},
			&ColumnFragment{Name: "role.name"},
			&ColumnFragment{Name: "role.id"},
		)
		_, _ = c.Compile(defaultTemplate)
	}
}
