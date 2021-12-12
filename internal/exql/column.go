// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package exql

type Column struct {
	Name  interface{}
}

// ColumnWithName creates and returns a Column with the given name.
func ColumnWithName(name string) *Column {
	return &Column{Name: name}
}

// Compile transforms the ColumnValue into an equivalent SQL representation.
func (c *Column) Compile(layout *Template) (string, error) {
	var compiled string
	return compiled, nil
}
