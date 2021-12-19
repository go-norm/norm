// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package exql

import (
	"github.com/pkg/errors"
)

// JoinType is the type of the JOIN clause.
type JoinType string

const ()

var _ Fragment = (*JoinFragment)(nil)

// JoinFragment is a JOIN clause in the SQL statement.
//
// NOTE: Fields are public purely for the purpose of being hashable. Direct
// modifications to them after construction may not take effect depends on
// whether the hash has been computed.
type JoinFragment struct {
	hash  hash
	Type  JoinType
	Table Fragment
	On    Fragment
	Using Fragment
}

func (j *JoinFragment) Hash() string {
	return j.hash.Hash(j)
}

func (j *JoinFragment) Compile(t *Template) (string, error) {
	if v, ok := t.Get(j); ok {
		return v, nil
	}

	if j.Table == nil {
		return "", nil
	}

	table, err := j.Table.Compile(t)
	if err != nil {
		return "", errors.Wrap(err, "compile table")
	}

	on, err := t.compile(j.On)
	if err != nil {
		return "", errors.Wrap(err, "compile ON clause")
	}

	using, err := t.compile(j.Using)
	if err != nil {
		return "", errors.Wrap(err, "compile USING clause")
	}

	data := map[string]interface{}{
		"Type":  j.Type,
		"Table": table,
		"On":    on,
		"Using": using,
	}
	compiled, err := t.Compile(LayoutJoin, data)
	if err != nil {
		return "", errors.Wrapf(err, "compile LayoutJoin with data %v", data)
	}

	t.Set(j, compiled)
	return compiled, nil
}

// // Joins represents the union of different join conditions.
// type Joins struct {
// 	Conditions []Fragment
// 	hash       hash
// }
//
// var _ = Fragment(&Joins{})
//
// // Hash returns a unique identifier for the struct.
// func (j *Joins) Hash() string {
// 	return j.hash.Hash(j)
// }
//
// // Compile transforms the WhereFragment into an equivalent SQL representation.
// func (j *Joins) Compile(layout *Template) (string, error) {
// 	if c, ok := layout.Get(j); ok {
// 		return c, nil
// 	}
//
// 	l := len(j.Conditions)
//
// 	chunks := make([]string, 0, l)
//
// 	if l > 0 {
// 		for i := 0; i < l; i++ {
// 			chunk, err := j.Conditions[i].Compile(layout)
// 			if err != nil {
// 				return "", err
// 			}
// 			chunks = append(chunks, chunk)
// 		}
// 	}
//
// 	compiled := strings.Join(chunks, " ")
// 	layout.Set(j, compiled)
// 	return compiled, nil
// }
//
// // JoinConditions creates a Joins object.
// func JoinConditions(joins ...*JoinFragment) *Joins {
// 	fragments := make([]Fragment, len(joins))
// 	for i := range fragments {
// 		fragments[i] = joins[i]
// 	}
// 	return &Joins{Conditions: fragments}
// }
//
// // On represents JOIN conditions.
// type On WhereFragment
//
// // OnConditions creates and retuens a new On.
// func OnConditions(conditions ...Fragment) *On {
// 	return &On{Conditions: conditions}
// }
//
// var _ = Fragment(&On{})
//
// // Hash returns a unique identifier.
// func (o *On) Hash() string {
// 	return o.hash.Hash(o)
// }
//
// // Compile transforms the On into an equivalent SQL representation.
// func (o *On) Compile(layout *Template) (string, error) {
// 	if c, ok := layout.Get(o); ok {
// 		return c, nil
// 	}
//
// 	grouped, err := groupConditions(layout, o.Conditions, layout.Compile(layout.ClauseOperator, layout.AndKeyword))
// 	if err != nil {
// 		return "", err
// 	}
//
// 	var compiled string
// 	if grouped != "" {
// 		compiled = layout.Compile(layout.OnLayout, conds{grouped})
// 	}
//
// 	layout.Set(o, compiled)
// 	return compiled, nil
// }
//
// // Using represents a USING function.
// type Using ColumnsFragment
//
// // UsingColumns builds a Using from the given columns.
// func UsingColumns(columns ...Fragment) *Using {
// 	return &Using{Columns: columns}
// }
//
// var _ = Fragment(&Using{})
//
// type usingT struct {
// 	Columns string
// }
//
// // Hash returns a unique identifier.
// func (u *Using) Hash() string {
// 	return u.hash.Hash(u)
// }
//
// // Compile transforms the Using into an equivalent SQL representation.
// func (u *Using) Compile(layout *Template) (string, error) {
// 	if u == nil {
// 		return "", nil
// 	}
//
// 	if c, ok := layout.Get(u); ok {
// 		return c, nil
// 	}
//
// 	var compiled string
// 	if len(u.Columns) > 0 {
// 		c := ColumnsFragment(*u)
// 		columns, err := c.Compile(layout)
// 		if err != nil {
// 			return "", err
// 		}
// 		data := usingT{Columns: columns}
// 		compiled = layout.Compile(layout.UsingLayout, data)
// 	}
//
// 	layout.Set(u, compiled)
// 	return compiled, nil
// }
