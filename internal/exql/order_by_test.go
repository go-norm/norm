// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package exql

import (
	"testing"
)

func TestOrderBy(t *testing.T) {
	o := JoinWithOrderBy(
		JoinSortColumns(
			&SortColumn{Column: &Column{Name: "foo"}},
		),
	)

	s := mustTrim(o.Compile(defaultTemplate))
	e := `ORDER BY "foo"`
	if s != e {
		t.Fatalf("Got: %s, Expecting: %s", s, e)
	}
}

func TestOrderByRaw(t *testing.T) {
	o := JoinWithOrderBy(
		JoinSortColumns(
			&SortColumn{Column: RawValue("CASE WHEN id IN ? THEN 0 ELSE 1 END")},
		),
	)

	s := mustTrim(o.Compile(defaultTemplate))
	e := `ORDER BY CASE WHEN id IN ? THEN 0 ELSE 1 END`
	if s != e {
		t.Fatalf("Got: %s, Expecting: %s", s, e)
	}
}

func TestOrderByDesc(t *testing.T) {
	o := JoinWithOrderBy(
		JoinSortColumns(
			&SortColumn{Column: &Column{Name: "foo"}, Order: Descendent},
		),
	)

	s := mustTrim(o.Compile(defaultTemplate))
	e := `ORDER BY "foo" DESC`
	if s != e {
		t.Fatalf("Got: %s, Expecting: %s", s, e)
	}
}

func BenchmarkOrderBy(b *testing.B) {
	for i := 0; i < b.N; i++ {
		JoinWithOrderBy(
			JoinSortColumns(
				&SortColumn{Column: &Column{Name: "foo"}},
			),
		)
	}
}

func BenchmarkOrderByHash(b *testing.B) {
	o := OrderBy{
		SortColumns: JoinSortColumns(
			&SortColumn{Column: &Column{Name: "foo"}},
		),
	}
	for i := 0; i < b.N; i++ {
		o.Hash()
	}
}

func BenchmarkCompileOrderByCompile(b *testing.B) {
	o := OrderBy{
		SortColumns: JoinSortColumns(
			&SortColumn{Column: &Column{Name: "foo"}},
		),
	}
	for i := 0; i < b.N; i++ {
		_, _ = o.Compile(defaultTemplate)
	}
}

func BenchmarkCompileOrderByCompileNoCache(b *testing.B) {
	for i := 0; i < b.N; i++ {
		o := JoinWithOrderBy(
			JoinSortColumns(
				&SortColumn{Column: &Column{Name: "foo"}},
			),
		)
		_, _ = o.Compile(defaultTemplate)
	}
}

func BenchmarkCompileOrderCompile(b *testing.B) {
	o := Descendent
	for i := 0; i < b.N; i++ {
		_, _ = o.Compile(defaultTemplate)
	}
}

func BenchmarkCompileOrderCompileNoCache(b *testing.B) {
	for i := 0; i < b.N; i++ {
		o := Descendent
		_, _ = o.Compile(defaultTemplate)
	}
}

func BenchmarkSortColumnHash(b *testing.B) {
	s := &SortColumn{Column: &Column{Name: "foo"}}
	for i := 0; i < b.N; i++ {
		s.Hash()
	}
}

func BenchmarkSortColumnCompile(b *testing.B) {
	s := &SortColumn{Column: &Column{Name: "foo"}}
	for i := 0; i < b.N; i++ {
		_, _ = s.Compile(defaultTemplate)
	}
}

func BenchmarkSortColumnCompileNoCache(b *testing.B) {
	for i := 0; i < b.N; i++ {
		s := &SortColumn{Column: &Column{Name: "foo"}}
		_, _ = s.Compile(defaultTemplate)
	}
}

func BenchmarkSortColumnsHash(b *testing.B) {
	s := JoinSortColumns(
		&SortColumn{Column: &Column{Name: "foo"}},
		&SortColumn{Column: &Column{Name: "bar"}},
	)
	for i := 0; i < b.N; i++ {
		s.Hash()
	}
}

func BenchmarkSortColumnsCompile(b *testing.B) {
	s := JoinSortColumns(
		&SortColumn{Column: &Column{Name: "foo"}},
		&SortColumn{Column: &Column{Name: "bar"}},
	)
	for i := 0; i < b.N; i++ {
		_, _ = s.Compile(defaultTemplate)
	}
}

func BenchmarkSortColumnsCompileNoCache(b *testing.B) {
	for i := 0; i < b.N; i++ {
		s := JoinSortColumns(
			&SortColumn{Column: &Column{Name: "foo"}},
			&SortColumn{Column: &Column{Name: "bar"}},
		)
		_, _ = s.Compile(defaultTemplate)
	}
}
