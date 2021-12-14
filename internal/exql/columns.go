// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package exql

// Columns represents an array of Column.
type Columns struct {
	Columns []Fragment
	hash    hash
}

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

// JoinColumns creates and returns an array of Column.
func JoinColumns(columns ...Fragment) *Columns {
	return &Columns{Columns: columns}
}

// UsingColumns builds a Using from the given columns.
func UsingColumns(columns ...Fragment) *Using {
	return &Using{Columns: columns}
}
