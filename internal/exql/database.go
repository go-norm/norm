// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package exql

import (
	"github.com/pkg/errors"
)

var _ Fragment = (*GroupByFragment)(nil)

// DatabaseFragment is a database in the SQL statement.
//
// NOTE: Fields are public purely for the purpose of being hashable. Direct
// modifications to them after construction may not take effect depends on
// whether the hash has been computed.
type DatabaseFragment struct {
	hash hash
	Name string
}

// Database constructs a DatabaseFragment with the given name.
func Database(name string) *DatabaseFragment {
	return &DatabaseFragment{
		Name: name,
	}
}

func (d *DatabaseFragment) Hash() string {
	return d.hash.Hash(d)
}

func (d *DatabaseFragment) Compile(t *Template) (string, error) {
	if v, ok := t.Get(d); ok {
		return v, nil
	}

	compiled, err := t.Compile(LayoutIdentifierQuote, Raw(d.Name))
	if err != nil {
		return "", errors.Wrapf(err, "compile LayoutIdentifierQuote with name %q", d.Name)
	}

	t.Set(d, compiled)
	return compiled, nil
}
