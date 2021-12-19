// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package exql

// Database represents a SQL database.
type Database struct {
	Name string
	hash hash
}

var _ = Fragment(&Database{})

// DatabaseWithName returns a Database with the given name.
func DatabaseWithName(name string) *Database {
	return &Database{Name: name}
}

// Hash returns a unique identifier for the struct.
func (d *Database) Hash() string {
	return d.hash.Hash(d)
}

// Compile transforms the Database into an equivalent SQL representation.
func (d *Database) Compile(layout *Template) (string, error) {
	if c, ok := layout.Get(d); ok {
		return c, nil
	}

	compiled := layout.Compile(layout.IdentifierQuote, RawFragment{Value: d.Name})
	layout.Set(d, compiled)
	return compiled, nil
}
