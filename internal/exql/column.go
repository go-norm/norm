// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package exql

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

var _ Fragment = (*Column)(nil)

// Column represents a SQL column.
type Column struct {
	Name  interface{}
	Alias string
	hash  hash
}

// ColumnWithName creates and returns a Column with the given name.
func ColumnWithName(name string) *Column {
	return &Column{Name: name}
}

func (c *Column) Hash() string {
	return c.hash.Hash(c)
}

// Compile transforms the ColumnValue into an equivalent SQL representation.
func (c *Column) Compile(t *Template) (compiled string, err error) {
	if z, ok := t.Get(c); ok {
		return z, nil
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
			nameChunks[i], err = t.Compile(LayoutIdentifierQuote, Raw{Value: nameChunks[i]})
			if err != nil {
				return "", errors.Wrapf(err, "compile LayoutIdentifierQuote with name %q", nameChunks[i])
			}
		}

		compiled = strings.Join(nameChunks, t.layouts[LayoutColumnSeparator])

		if len(chunks) > 1 {
			alias = trimString(chunks[1])
			alias, err = t.Compile(LayoutIdentifierQuote, Raw{Value: alias})
			if err != nil {
				return "", errors.Wrapf(err, "compile LayoutIdentifierQuote with alias %q", alias)
			}
		}

	case Raw:
		compiled = v.String()
	default:
		compiled = fmt.Sprintf("%v", c.Name)
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
