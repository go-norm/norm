// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package exql

import (
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

var _ Fragment = (*Statement)(nil)

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
	Values       *ValuesGroupsFragment
	Distinct     bool
	ColumnValues *ColumnValuesFragment
	OrderBy      *OrderByFragment
	GroupBy      *GroupByFragment
	Joins        *JoinsFragment
	Where        *WhereFragment
	Returning    *ReturningFragment

	Limit  int
	Offset int

	SQL string

	amendFn func(string) string
}

func (s *Statement) Hash() string {
	return s.hash.Hash(s)
}

func (s *Statement) SetAmend(amendFn func(string) string) {
	s.amendFn = amendFn
}

func (s *Statement) Amend(in string) string {
	if s.amendFn == nil {
		return in
	}
	return s.amendFn(in)
}

func (s *Statement) layout() (TemplateLayout, error) {
	switch s.Type {
	case StatementCount:
		return LayoutCount, nil
	case StatementDelete:
		return LayoutDelete, nil
	case StatementDropDatabase:
		return LayoutDropDatabase, nil
	case StatementDropTable:
		return LayoutDropTable, nil
	case StatementInsert:
		return LayoutInsert, nil
	case StatementSelect:
		return LayoutSelect, nil
	case StatementTruncate:
		return LayoutTruncate, nil
	case StatementUpdate:
		return LayoutUpdate, nil
	}
	return LayoutNone, errors.Errorf("unexpected type %v", s.Type)
}

func (s *Statement) Compile(t *Template) (string, error) {
	if s.Type == StatementSQL {
		return s.SQL, nil
	}

	if z, ok := t.Get(s); ok {
		return s.Amend(z), nil
	}

	layout, err := s.layout()
	if err != nil {
		return "", errors.Wrap(err, "get layout")
	}

	compiled, err := t.Compile(layout, s)
	if err != nil {
		return "", errors.Wrap(err, "compile")
	}

	t.Set(s, compiled)
	return s.Amend(compiled), nil
}

// RawSQL constructs a Statement with the given SQL query.
func RawSQL(sql string) *Statement {
	return &Statement{
		Type: StatementSQL,
		SQL:  sql,
	}
}
