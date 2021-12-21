// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package exql

import (
	"strings"

	"github.com/pkg/errors"
)

// StatementType is the type of SQL query the statement represents.
type StatementType uint

const (
	_ = StatementType(iota)

	StatementCount
	StatementDelete
	StatementDropDatabase
	StatementDropTable
	StatementInsert
	StatementSelect
	StatementTruncate
	StatementUpdate

	StatementSQL
)

// Statement is an AST for constructing SQL statements.
//
// CAUTION: Modifications to the fields will not be reflected after the
// statement has been compiled once. Create a new statement instance and copy
// over values of unchanged fields instead.
type Statement struct {
	hash hash

	Type         StatementType
	Database     *DatabaseFragment
	Table        *TableFragment
	Columns      *ColumnsFragment
	Values       Fragment
	Distinct     bool
	ColumnValues Fragment
	OrderBy      Fragment
	GroupBy      Fragment
	Joins        Fragment
	Where        Fragment
	Returning    Fragment

	Limit  Limit
	Offset Offset

	SQL string

	amendFn func(string) string
}

// todo

type (
	// Limit represents the SQL limit in a query.
	Limit int
	// Offset represents the SQL offset in a query.
	Offset int
)

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
	case StatementTruncate:
		return layout.TruncateLayout, nil
	case StatementDropTable:
		return layout.DropTableLayout, nil
	case StatementDropDatabase:
		return layout.DropDatabaseLayout, nil
	case StatementCount:
		return layout.CountLayout, nil
	case StatementInsert:
		return layout.InsertLayout, nil
	case StatementSelect:
		return layout.SelectLayout, nil
	case StatementUpdate:
		return layout.UpdateLayout, nil
	case StatementDelete:
		return layout.DeleteLayout, nil
	}
	return "", errUnknownTemplateType
}

// Compile transforms the Statement into an equivalent SQL query.
func (s *Statement) Compile(layout *Template) (string, error) {
	if s.Type == StatementSQL {
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
		Type: StatementSQL,
		SQL:  s,
	}
}
