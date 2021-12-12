// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package exql

// GroupBy represents a "GROUP BY" statement in a SQL query.
type GroupBy struct {
	Columns Fragment
}

// GroupByColumns creates and returns a GroupBy with the given columns.
func GroupByColumns(columns ...Fragment) *GroupBy {
	return &GroupBy{Columns: JoinColumns(columns...)}
}
