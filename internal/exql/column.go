// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package exql

import (
	"strings"

	"github.com/pkg/errors"
)

var _ Fragment = (*ColumnFragment)(nil)

// ColumnFragment is a column in the SQL statement.
//
// NOTE: Fields are public purely for the purpose of being hashable. Direct
// modifications to them after construction may not take effect depends on
// whether the hash has been computed.
type ColumnFragment struct {
	hash  hash
	Name  interface{}
	Alias string
}

// Column constructs a ColumnFragment with the given name, where the name can be
// a string or RawFragment.
//
// When a string is passed as the name, the alias is recognized with a
// case-insensitive "AS" or whitespace(s):
//
//   => users.name AS foo
//   Column("users.name AS foo")
//   Column("users.name foo")
func Column(name interface{}) *ColumnFragment {
	return &ColumnFragment{
		Name: name,
	}
}

func (c *ColumnFragment) Hash() string {
	return c.hash.Hash(c)
}

func (c *ColumnFragment) Compile(t *Template) (compiled string, err error) {
	if v, ok := t.Get(c); ok {
		return v, nil
	}

	alias := c.Alias
	switch v := c.Name.(type) {
	case string:
		input := trimString(v)
		chunks := separateByAS(input)
		if len(chunks) == 1 {
			chunks = separateBySpace(input)
		}

		name := chunks[0]
		nameChunks := strings.SplitN(name, t.layouts[LayoutColumnSeparator], 2)
		for i := range nameChunks {
			nameChunks[i] = trimString(nameChunks[i])
			if nameChunks[i] == "*" {
				continue
			}
			nameChunks[i], err = t.Compile(LayoutIdentifierQuote, Raw(nameChunks[i]))
			if err != nil {
				return "", errors.Wrapf(err, "compile LayoutIdentifierQuote with name %q", nameChunks[i])
			}
		}

		compiled = strings.Join(nameChunks, t.layouts[LayoutColumnSeparator])

		if len(chunks) > 1 {
			alias = trimString(chunks[1])
			alias, err = t.Compile(LayoutIdentifierQuote, Raw(alias))
			if err != nil {
				return "", errors.Wrapf(err, "compile LayoutIdentifierQuote with alias %q", alias)
			}
		}

	case *RawFragment:
		compiled = v.String()
	default:
		return "", errors.Errorf("unsupported column name type %T", v)
	}

	if alias != "" {
		data := map[string]string{
			"Name":  compiled,
			"Alias": alias,
		}
		compiled, err = t.Compile(LayoutColumnAlias, data)
		if err != nil {
			return "", errors.Wrapf(err, "compile LayoutColumnAlias with data %v", data)
		}
	}

	t.Set(c, compiled)
	return compiled, nil
}
