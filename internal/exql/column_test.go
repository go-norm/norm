// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package exql

import (
	"testing"
)

func TestColumnHash(t *testing.T) {
	var s, e string

	column := Column{Name: "role.name"}

	s = column.Hash()
	e = "*exql.Column:6700949208506105459"

	if s != e {
		t.Fatalf("Got: %s, Expecting: %s", s, e)
	}
}

func TestColumnString(t *testing.T) {

	column := Column{Name: "role.name"}

	s, err := column.Compile(defaultTemplate())
	if err != nil {
		t.Fatal()
	}

	e := `"role"."name"`
	if s != e {
		t.Fatalf("Got: %s, Expecting: %s", s, e)
	}
}

func TestColumnAs(t *testing.T) {
	column := Column{Name: "role.name as foo"}

	s, err := column.Compile(defaultTemplate())
	if err != nil {
		t.Fatal()
	}

	e := `"role"."name" AS "foo"`
	if s != e {
		t.Fatalf("Got: %s, Expecting: %s", s, e)
	}
}

func TestColumnImplicitAs(t *testing.T) {
	column := Column{Name: "role.name foo"}

	s, err := column.Compile(defaultTemplate())
	if err != nil {
		t.Fatal()
	}

	e := `"role"."name" AS "foo"`
	if s != e {
		t.Fatalf("Got: %s, Expecting: %s", s, e)
	}
}

func TestColumnRaw(t *testing.T) {
	column := Column{Name: Raw{Value: "role.name As foo"}}

	s, err := column.Compile(defaultTemplate())
	if err != nil {
		t.Fatal()
	}

	e := `role.name As foo`
	if s != e {
		t.Fatalf("Got: %s, Expecting: %s", s, e)
	}
}
