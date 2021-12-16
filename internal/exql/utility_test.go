// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package exql

import (
	"testing"
)

func TestUtilIsBlankSymbol(t *testing.T) {
	if isBlankSymbol(' ') == false {
		t.Fail()
	}
	if isBlankSymbol('\n') == false {
		t.Fail()
	}
	if isBlankSymbol('\t') == false {
		t.Fail()
	}
	if isBlankSymbol('\r') == false {
		t.Fail()
	}
	if isBlankSymbol('x') == true {
		t.Fail()
	}
}

func TestUtilTrimString(t *testing.T) {

	trimmed := trimString("  \t\nHello World!     \n")
	if string(trimmed) != "Hello World!" {
		t.Fatalf("Got: %s\n", trimmed)
	}

	trimmed = trimString("Nope")
	if string(trimmed) != "Nope" {
		t.Fatalf("Got: %s\n", trimmed)
	}

	trimmed = trimString("")
	if string(trimmed) != "" {
		t.Fatalf("Got: %s\n", trimmed)
	}

	trimmed = trimString(" ")
	if string(trimmed) != "" {
		t.Fatalf("Got: %s\n", trimmed)
	}
}

func TestUtilSeparateBySpace(t *testing.T) {
	chunks := separateBySpace("       Hello        World!        Enjoy")

	if len(chunks) != 3 {
		t.Fatal()
	}

	if chunks[0] != "Hello" {
		t.Fatal()
	}
	if chunks[1] != "World!" {
		t.Fatal()
	}
	if chunks[2] != "Enjoy" {
		t.Fatal()
	}
}

func TestUtilSeparateByAS(t *testing.T) {
	var chunks []string

	var tests = []string{
		`table.Name AS myTableAlias`,
		`table.Name     AS         myTableAlias`,
		"table.Name\tAS\r\nmyTableAlias",
	}

	for _, test := range tests {
		chunks = separateByAS(test)

		if len(chunks) != 2 {
			t.Fatalf(`Expecting 2 results.`)
		}

		if chunks[0] != "table.Name" {
			t.Fatal(`Expecting first result to be "table.Name".`)
		}
		if chunks[1] != "myTableAlias" {
			t.Fatal(`Expecting second result to be myTableAlias.`)
		}
	}

	// Single character.
	chunks = separateByAS("a")

	if len(chunks) != 1 {
		t.Fatalf(`Expecting 1 results.`)
	}

	if chunks[0] != "a" {
		t.Fatal(`Expecting first result to be "a".`)
	}

	// Empty name
	chunks = separateByAS("")

	if len(chunks) != 1 {
		t.Fatalf(`Expecting 1 results.`)
	}

	if chunks[0] != "" {
		t.Fatal(`Expecting first result to be "".`)
	}

	// Single name
	chunks = separateByAS("  A Single Table ")

	if len(chunks) != 1 {
		t.Fatalf(`Expecting 1 results.`)
	}

	if chunks[0] != "A Single Table" {
		t.Fatal(`Expecting first result to be "ASingleTable".`)
	}

	// Minimal expression.
	chunks = separateByAS("a AS b")

	if len(chunks) != 2 {
		t.Fatalf(`Expecting 2 results.`)
	}

	if chunks[0] != "a" {
		t.Fatal(`Expecting first result to be "a".`)
	}

	if chunks[1] != "b" {
		t.Fatal(`Expecting first result to be "b".`)
	}

	// Minimal expression with spaces.
	chunks = separateByAS("   a    AS    b ")

	if len(chunks) != 2 {
		t.Fatalf(`Expecting 2 results.`)
	}

	if chunks[0] != "a" {
		t.Fatal(`Expecting first result to be "a".`)
	}

	if chunks[1] != "b" {
		t.Fatal(`Expecting first result to be "b".`)
	}

	// Minimal expression + 1 with spaces.
	chunks = separateByAS("   a    AS    bb ")

	if len(chunks) != 2 {
		t.Fatalf(`Expecting 2 results.`)
	}

	if chunks[0] != "a" {
		t.Fatal(`Expecting first result to be "a".`)
	}

	if chunks[1] != "bb" {
		t.Fatal(`Expecting first result to be "bb".`)
	}
}
