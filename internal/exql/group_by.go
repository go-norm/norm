// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package exql

// GroupBy represents a "GROUP BY" statement in a SQL query.
type GroupBy struct {
	Columns Fragment
	hash    hash
}

var _ = Fragment(&GroupBy{})

type groupByT struct {
	GroupColumns string
}

// Hash returns a unique identifier.
func (g *GroupBy) Hash() string {
	return g.hash.Hash(g)
}

// GroupByColumns creates and returns a GroupBy with the given columns.
func GroupByColumns(columns ...Fragment) *GroupBy {
	return &GroupBy{Columns: Columns(columns...)}
}

func (g *GroupBy) Empty() bool {
	if g == nil || g.Columns == nil {
		return true
	}
	return g.Columns.(emptiable).Empty()
}

// Compile transforms the GroupBy into an equivalent SQL representation.
func (g *GroupBy) Compile(layout *Template) (string, error) {
	if c, ok := layout.Get(g); ok {
		return c, nil
	}

	var compiled string
	if g.Columns != nil {
		columns, err := g.Columns.Compile(layout)
		if err != nil {
			return "", err
		}

		data := groupByT{
			GroupColumns: columns,
		}
		compiled = layout.Compile(layout.GroupByLayout, data)
	}

	layout.Set(g, compiled)
	return compiled, nil
}
