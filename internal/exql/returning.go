// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package exql

import (
	"github.com/pkg/errors"
)

var _ Fragment = (*ReturningFragment)(nil)

// ReturningFragment is a RETURNING clause in the SQL statement.
//
// NOTE: Fields are public purely for the purpose of being hashable. Direct
// modifications to them after construction may not take effect depends on
// whether the hash has been computed.
type ReturningFragment ColumnsFragment

// Returning constructs a ReturningFragment with the given columns.
func Returning(columns ...*ColumnFragment) *ReturningFragment {
	return &ReturningFragment{
		Columns: columns,
	}
}

func (r *ReturningFragment) Hash() string {
	cs := ColumnsFragment(*r)
	return `ReturningFragment(` + cs.Hash() + `)`
}

func (r *ReturningFragment) Compile(t *Template) (string, error) {
	cs := ColumnsFragment(*r)
	if cs.Empty() {
		return "", nil
	}

	if v, ok := t.Get(r); ok {
		return v, nil
	}

	columns, err := cs.Compile(t)
	if err != nil {
		return "", errors.Wrap(err, "compile columns")
	}

	data := map[string]interface{}{
		"Columns": columns,
	}
	compiled, err := t.Compile(LayoutReturning, data)
	if err != nil {
		return "", errors.Wrapf(err, "compile LayoutReturning with data %v", data)
	}

	t.Set(r, compiled)
	return compiled, nil
}
