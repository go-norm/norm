// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package exql

// func TestTableSimple(t *testing.T) {
// 	var s, e string
//
// 	table := Table("artist")
//
// 	s = mustTrim(table.Compile(defaultTemplate))
// 	e = `"artist"`
//
// 	if s != e {
// 		t.Fatalf("Got: %s, Expecting: %s", s, e)
// 	}
// }
//
// func TestTableCompound(t *testing.T) {
// 	var s, e string
//
// 	table := Table("artist.foo")
//
// 	s = mustTrim(table.Compile(defaultTemplate))
// 	e = `"artist"."foo"`
//
// 	if s != e {
// 		t.Fatalf("Got: %s, Expecting: %s", s, e)
// 	}
// }
//
// func TestTableCompoundAlias(t *testing.T) {
// 	var s, e string
//
// 	table := Table("artist.foo AS baz")
//
// 	s = mustTrim(table.Compile(defaultTemplate))
// 	e = `"artist"."foo" AS "baz"`
//
// 	if s != e {
// 		t.Fatalf("Got: %s, Expecting: %s", s, e)
// 	}
// }
//
// func TestTableImplicitAlias(t *testing.T) {
// 	var s, e string
//
// 	table := Table("artist.foo baz")
//
// 	s = mustTrim(table.Compile(defaultTemplate))
// 	e = `"artist"."foo" AS "baz"`
//
// 	if s != e {
// 		t.Fatalf("Got: %s, Expecting: %s", s, e)
// 	}
// }
//
// func TestTableMultiple(t *testing.T) {
// 	var s, e string
//
// 	table := Table("artist.foo, artist.bar, artist.baz")
//
// 	s = mustTrim(table.Compile(defaultTemplate))
// 	e = `"artist"."foo", "artist"."bar", "artist"."baz"`
//
// 	if s != e {
// 		t.Fatalf("Got: %s, Expecting: %s", s, e)
// 	}
// }
//
// func TestTableMultipleAlias(t *testing.T) {
// 	var s, e string
//
// 	table := Table("artist.foo AS foo, artist.bar as bar, artist.baz As baz")
//
// 	s = mustTrim(table.Compile(defaultTemplate))
// 	e = `"artist"."foo" AS "foo", "artist"."bar" AS "bar", "artist"."baz" AS "baz"`
//
// 	if s != e {
// 		t.Fatalf("Got: %s, Expecting: %s", s, e)
// 	}
// }
//
// func TestTableMinimal(t *testing.T) {
// 	var s, e string
//
// 	table := Table("a")
//
// 	s = mustTrim(table.Compile(defaultTemplate))
// 	e = `"a"`
//
// 	if s != e {
// 		t.Fatalf("Got: %s, Expecting: %s", s, e)
// 	}
// }
//
// func TestTableEmpty(t *testing.T) {
// 	var s, e string
//
// 	table := Table("")
//
// 	s = mustTrim(table.Compile(defaultTemplate))
// 	e = ``
//
// 	if s != e {
// 		t.Fatalf("Got: %s, Expecting: %s", s, e)
// 	}
// }

