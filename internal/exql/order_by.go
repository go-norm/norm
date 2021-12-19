// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package exql

import (
	"strings"

	"github.com/pkg/errors"
)

var _ Fragment = (*OrderByFragment)(nil)

// OrderByFragment is a ORDER BY clause in the SQL statement.
//
// NOTE: Fields are public purely for the purpose of being hashable. Direct
// modifications to them after construction may not take effect depends on
// whether the hash has been computed.
type OrderByFragment struct {
	hash    hash
	Columns []*SortColumnFragment
}

// OrderBy constructs a OrderByFragment with the given SortColumnFragment.
func OrderBy(columns ...*SortColumnFragment) *OrderByFragment {
	return &OrderByFragment{
		Columns: columns,
	}
}

func (ob *OrderByFragment) Hash() string {
	return ob.hash.Hash(ob)
}

func (ob *OrderByFragment) Compile(t *Template) (compiled string, err error) {
	if len(ob.Columns) == 0 {
		return "", nil
	}

	if v, ok := t.Get(ob); ok {
		return v, nil
	}

	out := make([]string, len(ob.Columns))
	for i := range ob.Columns {
		out[i], err = ob.Columns[i].Compile(t)
		if err != nil {
			return "", errors.Wrap(err, "compile column")
		}
	}

	sortColumns := strings.TrimSpace(strings.Join(out, t.layouts[LayoutIdentifierSeparator]))
	data := map[string]interface{}{
		"Columns": sortColumns,
	}
	compiled, err = t.Compile(LayoutOrderBy, data)
	if err != nil {
		return "", errors.Wrapf(err, "compile LayoutOrderBy with data %v", data)
	}

	t.Set(ob, compiled)
	return compiled, nil
}

// SortOrder represents the order in which SQL results are sorted.
type SortOrder uint8

const (
	SortAscendant = SortOrder(iota)
	SortDescendent
)

func (so SortOrder) compile(t *Template) string {
	if so == SortDescendent {
		return t.layouts[LayoutDescKeyword]
	}
	return t.layouts[LayoutAscKeyword]
}

var _ Fragment = (*SortColumnFragment)(nil)

// SortColumnFragment is a column-order relation within an ORDER BY clause in
// the SQL statement.
//
// NOTE: Fields are public purely for the purpose of being hashable. Direct
// modifications to them after construction may not take effect depends on
// whether the hash has been computed.
type SortColumnFragment struct {
	hash   hash
	Column *ColumnFragment
	Order  SortOrder
}

// SortColumn constructs a SortColumnFragment with the given column and an
// optional order. If the order is omitted, the default SortAscendant is used.
func SortColumn(column *ColumnFragment, order ...SortOrder) *SortColumnFragment {
	sc := &SortColumnFragment{
		Column: column,
	}
	if len(order) > 0 {
		sc.Order = order[0]
	}
	return sc
}

func (sc *SortColumnFragment) Hash() string {
	return sc.hash.Hash(sc)
}

func (sc *SortColumnFragment) Compile(t *Template) (string, error) {
	if v, ok := t.Get(sc); ok {
		return v, nil
	}

	column, err := sc.Column.Compile(t)
	if err != nil {
		return "", errors.Wrap(err, "compile column")
	}

	data := map[string]interface{}{
		"Column": column,
		"Order":  sc.Order.compile(t),
	}
	compiled, err := t.Compile(LayoutSortByColumn, data)
	if err != nil {
		return "", errors.Wrapf(err, "compile LayoutSortByColumn with data %v", data)
	}

	t.Set(sc, compiled)
	return compiled, nil
}
