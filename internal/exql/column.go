// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package exql

import (
	"fmt"
	"strings"
)

type columnT struct {
	Name  string
	Alias string
}

// Column represents a SQL column.
type Column struct {
	Name  interface{}
	Alias string
	hash  hash
}

var _ = Fragment(&Column{})

// ColumnWithName creates and returns a Column with the given name.
func ColumnWithName(name string) *Column {
	return &Column{Name: name}
}

// Hash returns a unique identifier for the struct.
func (c *Column) Hash() string {
	return c.hash.Hash(c)
}

// Compile transforms the ColumnValue into an equivalent SQL representation.
func (c *Column) Compile(layout *Template) (string, error) {
	if z, ok := layout.Get(c); ok {
		return z, nil
	}

	var compiled string
	alias := c.Alias
	switch value := c.Name.(type) {
	case string:
		input := trimString(value)

		chunks := separateByAS(input)

		if len(chunks) == 1 {
			chunks = separateBySpace(input)
		}

		name := chunks[0]

		nameChunks := strings.SplitN(name, layout.ColumnSeparator, 2)

		for i := range nameChunks {
			nameChunks[i] = trimString(nameChunks[i])
			if nameChunks[i] == "*" {
				continue
			}
			nameChunks[i] = layout.MustCompile(layout.IdentifierQuote, Raw{Value: nameChunks[i]})
		}

		compiled = strings.Join(nameChunks, layout.ColumnSeparator)

		if len(chunks) > 1 {
			alias = trimString(chunks[1])
			alias = layout.MustCompile(layout.IdentifierQuote, Raw{Value: alias})
		}
	case Raw:
		compiled = value.String()
	default:
		compiled = fmt.Sprintf("%v", c.Name)
	}

	if alias != "" {
		compiled = layout.MustCompile(layout.ColumnAliasLayout, columnT{compiled, alias})
	}

	layout.Set(c, compiled)
	return compiled, nil
}
