// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package exql

import (
	"strings"

	"github.com/pkg/errors"
)

// JoinType is the type of the JOIN clause.
type JoinType string

const (
	DefaultJoin JoinType = ""
	FullJoin    JoinType = "FULL"
	CrossJoin   JoinType = "CROSS"
	RightJoin   JoinType = "RIGHT"
	LeftJoin    JoinType = "LEFT"
)

var _ Fragment = (*JoinFragment)(nil)

// JoinFragment is a JOIN clause in the SQL statement.
//
// NOTE: Fields are public purely for the purpose of being hashable. Direct
// modifications to them after construction may not take effect depends on
// whether the hash has been computed.
type JoinFragment struct {
	hash  hash
	Type  JoinType
	Table *TableFragment
	On    *OnFragment
	Using *UsingFragment
}

// Join constructs a JoinFragment with the given table as a NATURAL JOIN, where
// the table name can be a string or RawFragment.
func Join(table interface{}) *JoinFragment {
	var t *TableFragment
	if table != nil {
		t = Table(table)
	}
	return &JoinFragment{
		Type:  DefaultJoin,
		Table: t,
	}
}

// JoinOn constructs a JoinFragment with the given type, table and on clause,
// where the table name can be a string or RawFragment.
func JoinOn(typ JoinType, table interface{}, on *OnFragment) *JoinFragment {
	join := Join(table)
	join.Type = typ
	join.On = on
	return join
}

// JoinUsing constructs a JoinFragment with the given type, table and using
// clause, where the table name can be a string or RawFragment.
func JoinUsing(typ JoinType, table interface{}, using *UsingFragment) *JoinFragment {
	join := Join(table)
	join.Type = typ
	join.Using = using
	return join
}

func (j *JoinFragment) Hash() string {
	return j.hash.Hash(j)
}

func (j *JoinFragment) Compile(t *Template) (string, error) {
	if j.Table == nil {
		return "", nil
	}

	if v, ok := t.Get(j); ok {
		return v, nil
	}

	table, err := j.Table.Compile(t)
	if err != nil {
		return "", errors.Wrap(err, "compile table")
	}

	on, err := t.compile(j.On)
	if err != nil {
		return "", errors.Wrap(err, "compile ON clause")
	}

	using, err := t.compile(j.Using)
	if err != nil {
		return "", errors.Wrap(err, "compile USING clause")
	}

	data := map[string]interface{}{
		"Type":  j.Type,
		"Table": table,
		"On":    on,
		"Using": using,
	}
	compiled, err := t.Compile(LayoutJoin, data)
	if err != nil {
		return "", errors.Wrapf(err, "compile LayoutJoin with data %v", data)
	}

	t.Set(j, compiled)
	return compiled, nil
}

var _ Fragment = (*OnFragment)(nil)

// OnFragment is a ON clause within a JOIN clause in the SQL statement.
type OnFragment WhereFragment

// On constructs a OnFragment with the given conditions.
func On(conds ...Fragment) *OnFragment {
	return &OnFragment{
		Conditions: conds,
	}
}

func (on *OnFragment) Hash() string {
	w := WhereFragment(*on)
	return `OnFragment(` + w.Hash() + `)`
}

func (on *OnFragment) Compile(t *Template) (string, error) {
	if len(on.Conditions) == 0 {
		return "", nil
	}

	if v, ok := t.Get(on); ok {
		return v, nil
	}

	groupKeyword, err := t.Compile(LayoutClauseOperator, t.layouts[LayoutAndKeyword])
	if err != nil {
		return "", errors.Wrapf(err, "compile LayoutClauseOperator with keyword %q", t.layouts[LayoutAndKeyword])
	}

	grouped, err := groupConditions(t, on.Conditions, groupKeyword)
	if err != nil {
		return "", errors.Wrap(err, "group conditions")
	}

	data := map[string]interface{}{
		"Conds": grouped,
	}
	compiled, err := t.Compile(LayoutOn, data)
	if err != nil {
		return "", errors.Wrapf(err, "compile LayoutOn with data %v", data)
	}

	t.Set(on, compiled)
	return compiled, nil
}

var _ Fragment = (*UsingFragment)(nil)

// UsingFragment is a USING clause within a JOIN clause in the SQL statement.
type UsingFragment ColumnsFragment

// Using constructs a UsingFragment with the given columns.
func Using(columns ...*ColumnFragment) *UsingFragment {
	return &UsingFragment{
		Columns: columns,
	}
}

func (u *UsingFragment) Hash() string {
	cs := ColumnsFragment(*u)
	return `UsingFragment(` + cs.Hash() + `)`
}

func (u *UsingFragment) Compile(t *Template) (string, error) {
	cs := ColumnsFragment(*u)
	if cs.Empty() {
		return "", nil
	}

	if v, ok := t.Get(u); ok {
		return v, nil
	}

	columns, err := cs.Compile(t)
	if err != nil {
		return "", errors.Wrap(err, "compile columns")
	}

	data := map[string]interface{}{
		"Columns": columns,
	}
	compiled, err := t.Compile(LayoutUsing, data)
	if err != nil {
		return "", errors.Wrapf(err, "compile LayoutUsing with data %v", data)
	}

	t.Set(u, compiled)
	return compiled, nil
}

var _ Fragment = (*JoinsFragment)(nil)

// JoinsFragment is a list of JoinFragment.
//
// NOTE: Fields are public purely for the purpose of being hashable. Direct
// modifications to them after construction may not take effect depends on
// whether the hash has been computed.
type JoinsFragment struct {
	hash  hash
	Joins []*JoinFragment
}

// Joins constructs a JoinsFragment with the given joins.
func Joins(joins ...*JoinFragment) *JoinsFragment {
	return &JoinsFragment{
		Joins: joins,
	}
}

func (js *JoinsFragment) Hash() string {
	return js.hash.Hash(js)
}

func (js *JoinsFragment) Compile(t *Template) (compiled string, err error) {
	if len(js.Joins) == 0 {
		return "", nil
	}

	if v, ok := t.Get(js); ok {
		return v, nil
	}

	out := make([]string, len(js.Joins))
	for i := range js.Joins {
		out[i], err = js.Joins[i].Compile(t)
		if err != nil {
			return "", errors.Wrap(err, "compile join")
		}
	}

	compiled = strings.TrimSpace(strings.Join(out, " "))
	t.Set(js, compiled)
	return compiled, nil
}
