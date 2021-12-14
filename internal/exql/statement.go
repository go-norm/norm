// Copyright 2012 The upper.io/db authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package exql


type Type uint

const (
	Select

	SQL
)

// Statement is an AST for constructing SQL statements.
type Statement struct {
	Type
	Table        Fragment
	Database     Fragment
	Columns      Fragment
	Values       Fragment
	Distinct     bool
	ColumnValues Fragment
	OrderBy      Fragment
	GroupBy      Fragment
	Joins        Fragment
	Where        Fragment
	Returning    Fragment

	Limit
	Offset

	SQL string

	hash    hash
}

func (s *Statement) Compile(layout *Template) (string, error) {
	return "", nil
}
