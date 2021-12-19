// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package exql

import (
	"testing"
)

func TestValue(t *testing.T) {
	val := NewValue(1)

	s, err := val.Compile(defaultTemplate)
	if err != nil {
		t.Fatal()
	}

	e := `'1'`
	if s != e {
		t.Fatalf("Got: %s, Expecting: %s", s, e)
	}

	val = NewValue(&RawFragment{Value: "NOW()"})

	s, err = val.Compile(defaultTemplate)
	if err != nil {
		t.Fatal()
	}

	e = `NOW()`
	if s != e {
		t.Fatalf("Got: %s, Expecting: %s", s, e)
	}
}

func TestValues(t *testing.T) {
	val := NewValueGroup(
		&ValueFragment{V: &RawFragment{Value: "1"}},
		&ValueFragment{V: &RawFragment{Value: "2"}},
		&ValueFragment{V: "3"},
	)

	s, err := val.Compile(defaultTemplate)
	if err != nil {
		t.Fatal()
	}

	e := `(1, 2, '3')`
	if s != e {
		t.Fatalf("Got: %s, Expecting: %s", s, e)
	}
}
