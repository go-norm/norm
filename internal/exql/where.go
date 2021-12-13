// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package exql

import (
	"strings"
)

// Or represents an SQL OR operator.
type Or Where

// And represents an SQL AND operator.
type And Where

// Where represents an SQL WHERE clause.
type Where struct {
	Conditions []Fragment
	hash       hash
}

var _ = Fragment(&Where{})

type conds struct {
	Conds string
}

// WhereConditions creates and returns a new Where.
func WhereConditions(conds ...Fragment) *Where {
	return &Where{Conditions: conds}
}

// JoinWithOr creates and returns a new Or.
func JoinWithOr(conds ...Fragment) *Or {
	return &Or{Conditions: conds}
}

// JoinWithAnd creates and returns a new And.
func JoinWithAnd(conds ...Fragment) *And {
	return &And{Conditions: conds}
}

// Hash returns a unique identifier for the struct.
func (w *Where) Hash() string {
	return w.hash.Hash(w)
}

// Append adds the conditions to the ones that already exist.
func (w *Where) Append(a *Where) *Where {
	if a != nil {
		w.Conditions = append(w.Conditions, a.Conditions...)
	}
	return w
}

// Hash returns a unique identifier.
func (o *Or) Hash() string {
	w := Where(*o)
	return `Or(` + w.Hash() + `)`
}

// Hash returns a unique identifier.
func (a *And) Hash() string {
	w := Where(*a)
	return `And(` + w.Hash() + `)`
}

// Compile transforms the Or into an equivalent SQL representation.
func (o *Or) Compile(layout *Template) (string, error) {
	if z, ok := layout.Get(o); ok {
		return z, nil
	}

	compiled, err := groupCondition(layout, o.Conditions, layout.MustCompile(layout.ClauseOperator, layout.OrKeyword))
	if err != nil {
		return "", err
	}

	layout.Set(o, compiled)
	return compiled, nil
}

// Compile transforms the And into an equivalent SQL representation.
func (a *And) Compile(layout *Template) (string, error) {
	if c, ok := layout.Get(a); ok {
		return c, nil
	}

	compiled, err := groupCondition(layout, a.Conditions, layout.MustCompile(layout.ClauseOperator, layout.AndKeyword))
	if err != nil {
		return "", err
	}

	layout.Set(a, compiled)
	return compiled, nil
}

// Compile transforms the Where into an equivalent SQL representation.
func (w *Where) Compile(layout *Template) (string, error) {
	if c, ok := layout.Get(w); ok {
		return c, nil
	}

	grouped, err := groupCondition(layout, w.Conditions, layout.MustCompile(layout.ClauseOperator, layout.AndKeyword))
	if err != nil {
		return "", err
	}

	var compiled string
	if grouped != "" {
		compiled = layout.MustCompile(layout.WhereLayout, conds{grouped})
	}

	layout.Set(w, compiled)
	return compiled, nil
}

func groupCondition(layout *Template, terms []Fragment, joinKeyword string) (string, error) {
	l := len(terms)

	chunks := make([]string, 0, l)

	if l > 0 {
		for i := 0; i < l; i++ {
			chunk, err := terms[i].Compile(layout)
			if err != nil {
				return "", err
			}
			chunks = append(chunks, chunk)
		}
	}

	if len(chunks) > 0 {
		return layout.MustCompile(layout.ClauseGroup, strings.Join(chunks, joinKeyword)), nil
	}

	return "", nil
}
