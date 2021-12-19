// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package exql

import (
	"strings"

	"github.com/pkg/errors"
)

var _ Fragment = (*WhereFragment)(nil)

// WhereFragment is a WHERE clause in the SQL statement.
//
// NOTE: Fields are public purely for the purpose of being hashable. Direct
// modifications to them after construction may not take effect depends on
// whether the hash has been computed.
type WhereFragment struct {
	hash       hash
	Conditions []Fragment
}

// Where constructs a WhereFragment with the given conditions.
func Where(conds ...Fragment) *WhereFragment {
	return &WhereFragment{
		Conditions: conds,
	}
}

func (w *WhereFragment) Hash() string {
	return w.hash.Hash(w)
}

func groupConditions(t *Template, conds []Fragment, groupKeyword string) (compiled string, err error) {
	chunks := make([]string, len(conds))
	for i := range conds {
		chunks[i], err = conds[i].Compile(t)
		if err != nil {
			return "", errors.Wrap(err, "compile condition")
		}
	}
	return t.Compile(LayoutClauseGroup, strings.Join(chunks, groupKeyword))
}

func (w *WhereFragment) Compile(t *Template) (string, error) {
	if len(w.Conditions) == 0 {
		return "", nil
	}

	if v, ok := t.Get(w); ok {
		return v, nil
	}

	groupKeyword, err := t.Compile(LayoutClauseOperator, t.layouts[LayoutAndKeyword])
	if err != nil {
		return "", errors.Wrapf(err, "compile LayoutClauseOperator with keyword %q", t.layouts[LayoutAndKeyword])
	}

	grouped, err := groupConditions(t, w.Conditions, groupKeyword)
	if err != nil {
		return "", errors.Wrap(err, "group conditions")
	}

	data := map[string]interface{}{
		"Conds": grouped,
	}
	compiled, err := t.Compile(LayoutWhere, data)
	if err != nil {
		return "", errors.Wrapf(err, "compile LayoutWhere with data %v", data)
	}

	t.Set(w, compiled)
	return compiled, nil
}

// Append appends given conditions to the WhereFragment.
func (w *WhereFragment) Append(conds ...Fragment) *WhereFragment {
	w.Conditions = append(w.Conditions, conds...)
	w.hash.Reset()
	return w
}

var _ Fragment = (*AndFragment)(nil)

// AndFragment is a clause with AND operator in the SQL statement.
type AndFragment WhereFragment

// And constructs a AndFragment with the given conditions.
func And(conds ...Fragment) *AndFragment {
	return &AndFragment{
		Conditions: conds,
	}
}

func (and *AndFragment) Hash() string {
	w := WhereFragment(*and)
	return `AndFragment(` + w.Hash() + `)`
}

func (and *AndFragment) Compile(t *Template) (string, error) {
	if len(and.Conditions) == 0 {
		return "", nil
	}

	if v, ok := t.Get(and); ok {
		return v, nil
	}

	groupKeyword, err := t.Compile(LayoutClauseOperator, t.layouts[LayoutAndKeyword])
	if err != nil {
		return "", errors.Wrapf(err, "compile LayoutClauseOperator with keyword %q", t.layouts[LayoutAndKeyword])
	}

	compiled, err := groupConditions(t, and.Conditions, groupKeyword)
	if err != nil {
		return "", errors.Wrap(err, "group conditions")
	}

	t.Set(and, compiled)
	return compiled, nil
}

var _ Fragment = (*OrFragment)(nil)

// OrFragment is a clause with OR operator in the SQL statement.
type OrFragment WhereFragment

// Or constructs a OrFragment with the given conditions.
func Or(conds ...Fragment) *OrFragment {
	return &OrFragment{
		Conditions: conds,
	}
}

func (or *OrFragment) Hash() string {
	w := WhereFragment(*or)
	return `OrFragment(` + w.Hash() + `)`
}

func (or *OrFragment) Compile(t *Template) (string, error) {
	if len(or.Conditions) == 0 {
		return "", nil
	}

	if v, ok := t.Get(or); ok {
		return v, nil
	}

	groupKeyword, err := t.Compile(LayoutClauseOperator, t.layouts[LayoutOrKeyword])
	if err != nil {
		return "", errors.Wrapf(err, "compile LayoutClauseOperator with keyword %q", t.layouts[LayoutOrKeyword])
	}

	compiled, err := groupConditions(t, or.Conditions, groupKeyword)
	if err != nil {
		return "", errors.Wrap(err, "group conditions")
	}

	t.Set(or, compiled)
	return compiled, nil
}
