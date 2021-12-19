// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package exql

import (
	"github.com/pkg/errors"
)

var _ Fragment = (*GroupByFragment)(nil)

// GroupByFragment is a GROUP BY clause in the SQL statement.
//
// NOTE: Fields are public purely for the purpose of being hashable. Direct
// modifications to them after construction may not take effect depends on
// whether the hash has been computed.
type GroupByFragment ColumnsFragment

// GroupBy constructs a GroupByFragment with the given columns.
func GroupBy(columns ...*ColumnFragment) *GroupByFragment {
	return &GroupByFragment{
		Columns: columns,
	}
}

func (gb *GroupByFragment) Hash() string {
	cs := ColumnsFragment(*gb)
	return `GroupByFragment(` + cs.Hash() + `)`
}

func (gb *GroupByFragment) Compile(t *Template) (string, error) {
	cs := ColumnsFragment(*gb)
	if cs.Empty() {
		return "", nil
	}

	if v, ok := t.Get(gb); ok {
		return v, nil
	}

	columns, err := cs.Compile(t)
	if err != nil {
		return "", errors.Wrap(err, "compile columns")
	}

	data := map[string]interface{}{
		"Columns": columns,
	}
	compiled, err := t.Compile(LayoutGroupBy, data)
	if err != nil {
		return "", errors.Wrapf(err, "compile LayoutGroupBy with data %v", data)
	}

	t.Set(gb, compiled)
	return compiled, nil
}
