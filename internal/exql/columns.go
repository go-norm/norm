// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package exql

import (
	"strings"
)

// Columns represents an array of Column.
type Columns struct {
	Columns []Fragment
	hash    hash
}

var _ = Fragment(&Columns{})

// Hash returns a unique identifier.
func (c *Columns) Hash() string {
	return c.hash.Hash(c)
}

// Append appends the other Columns to the current one.
func (c *Columns) Append(other *Columns) *Columns {
	c.Columns = append(c.Columns, other.Columns...)
	return c
}

// IsEmpty returns true if there is no column in the Columns.
func (c *Columns) IsEmpty() bool {
	if c == nil || len(c.Columns) < 1 {
		return true
	}
	return false
}

// Compile transforms the Columns into an equivalent SQL representation.
func (c *Columns) Compile(layout *Template) (string, error) {
	if z, ok := layout.Get(c); ok {
		return z, nil
	}

	var compiled string
	l := len(c.Columns)
	if l > 0 {
		var err error
		out := make([]string, l)
		for i := 0; i < l; i++ {
			out[i], err = c.Columns[i].Compile(layout)
			if err != nil {
				return "", err
			}
		}

		compiled = strings.Join(out, layout.IdentifierSeparator)
	} else {
		compiled = "*"
	}

	layout.Set(c, compiled)
	return compiled, nil
}

// JoinColumns creates and returns an array of Column.
func JoinColumns(columns ...Fragment) *Columns {
	return &Columns{Columns: columns}
}

// OnConditions creates and retuens a new On.
func OnConditions(conditions ...Fragment) *On {
	return &On{Conditions: conditions}
}

// UsingColumns builds a Using from the given columns.
func UsingColumns(columns ...Fragment) *Using {
	return &Using{Columns: columns}
}
