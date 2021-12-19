// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package exql

import (
	"strings"

	"github.com/pkg/errors"
)

// Type is the type of SQL query the statement represents.
type Type uint

// Values for Type.
const (
	_ = Type(iota)

	Truncate
	DropTable
	DropDatabase
	Count
	Insert
	Select
	Update
	Delete

	SQL
)

type (
	// Limit represents the SQL limit in a query.
	Limit int
	// Offset represents the SQL offset in a query.
	Offset int
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
	amendFn func(string) string
}

// Hash returns a unique identifier for the struct.
func (s *Statement) Hash() string {
	return s.hash.Hash(s)
}

func (s *Statement) SetAmendment(amendFn func(string) string) {
	s.amendFn = amendFn
}

func (s *Statement) Amend(in string) string {
	if s.amendFn == nil {
		return in
	}
	return s.amendFn(in)
}

var errUnknownTemplateType = errors.New("unknown template type")

func (s *Statement) template(layout *Template) (string, error) {
	switch s.Type {
	case Truncate:
		return layout.TruncateLayout, nil
	case DropTable:
		return layout.DropTableLayout, nil
	case DropDatabase:
		return layout.DropDatabaseLayout, nil
	case Count:
		return layout.CountLayout, nil
	case Insert:
		return layout.InsertLayout, nil
	case Select:
		return layout.SelectLayout, nil
	case Update:
		return layout.UpdateLayout, nil
	case Delete:
		return layout.DeleteLayout, nil
	}
	return "", errUnknownTemplateType
}

// Compile transforms the Statement into an equivalent SQL query.
func (s *Statement) Compile(layout *Template) (string, error) {
	if s.Type == SQL {
		// No need to hit the cache.
		return s.SQL, nil
	}

	if z, ok := layout.Get(s); ok {
		return s.Amend(z), nil
	}

	tpl, err := s.template(layout)
	if err != nil {
		return "", errors.Wrap(err, "get template")
	}

	// todo: return error, not panic
	compiled := layout.Compile(tpl, s)
	compiled = strings.TrimSpace(compiled)
	layout.Set(s, compiled)
	return s.Amend(compiled), nil
}

// RawSQL represents a raw SQL statement.
func RawSQL(s string) *Statement {
	return &Statement{
		Type: SQL,
		SQL:  s,
	}
}
