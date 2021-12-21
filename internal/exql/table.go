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

// TableFragment is a table in the SQL statement.
//
// NOTE: Fields are public purely for the purpose of being hashable. Direct
// modifications to them after construction may not take effect depends on
// whether the hash has been computed.
type TableFragment struct {
	hash  hash
	Name  interface{}
	Alias string
}

// Table constructs a TableFragment with the given name, where the name can be a
// string or RawFragment.
//
// When a string is passed as the name, the alias is recognized with a
// case-insensitive "AS" or whitespace(s):
//
//   => users.name AS foo
//   Table("users AS u")
//   Table("users u")
func Table(name interface{}) *TableFragment {
	return &TableFragment{
		Name: name,
	}
}

func (t *TableFragment) Hash() string {
	return t.hash.Hash(t)
}

func (t *TableFragment) Compile(tmpl *Template) (compiled string, err error) {
	if v, ok := tmpl.Get(t); ok {
		return v, nil
	}

	alias := t.Alias
	switch v := t.Name.(type) {
	case string:
		input := trimString(v)
		chunks := separateByAS(input)
		if len(chunks) == 1 {
			chunks = separateBySpace(input)
		}

		name := chunks[0]
		nameChunks := strings.SplitN(name, tmpl.layouts[LayoutColumnSeparator], 2)
		for i := range nameChunks {
			nameChunks[i] = trimString(nameChunks[i])
			nameChunks[i], err = tmpl.Compile(LayoutIdentifierQuote, Raw(nameChunks[i]))
			if err != nil {
				return "", errors.Wrapf(err, "compile LayoutIdentifierQuote with name %q", nameChunks[i])
			}
		}

		compiled = strings.Join(nameChunks, tmpl.layouts[LayoutColumnSeparator])

		if len(chunks) > 1 {
			alias = trimString(chunks[1])
			alias, err = tmpl.Compile(LayoutIdentifierQuote, Raw(alias))
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
		compiled, err = tmpl.Compile(LayoutTableAlias, data)
		if err != nil {
			return "", errors.Wrapf(err, "compile LayoutTableAlias with data %v", data)
		}
	}

	tmpl.Set(t, compiled)
	return compiled, nil
}

var _ Fragment = (*TablesFragment)(nil)

// TablesFragment is a list of TableFragment.
//
// NOTE: Fields are public purely for the purpose of being hashable. Direct
// modifications to them after construction may not take effect depends on
// whether the hash has been computed.
type TablesFragment struct {
	hash   hash
	Tables []*TableFragment
}

// Tables constructs a TablesFragment with the given tables.
func Tables(tables ...*TableFragment) *TablesFragment {
	return &TablesFragment{
		Tables: tables,
	}
}

func (ts *TablesFragment) Hash() string {
	return ts.hash.Hash(ts)
}

func (ts *TablesFragment) Compile(t *Template) (compiled string, err error) {
	if len(ts.Tables) == 0 {
		return "", nil
	}

	if v, ok := t.Get(ts); ok {
		return v, nil
	}

	out := make([]string, len(ts.Tables))
	for i := range ts.Tables {
		out[i], err = ts.Tables[i].Compile(t)
		if err != nil {
			return "", errors.Wrap(err, "compile table")
		}
	}

	compiled = strings.TrimSpace(strings.Join(out, t.layouts[LayoutIdentifierSeparator]))
	t.Set(ts, compiled)
	return compiled, nil
}
