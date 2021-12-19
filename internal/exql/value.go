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

var _ Fragment = (*ValueFragment)(nil)

// ValueFragment is an escaped value in the SQL statement.
type ValueFragment struct {
	hash  hash
	Value interface{}
}

// Value constructs a ValueFragment with the given value.
func Value(v interface{}) *ValueFragment {
	return &ValueFragment{
		Value: v,
	}
}

func (v *ValueFragment) Hash() string {
	return v.hash.Hash(v)
}

func (v *ValueFragment) Compile(t *Template) (compiled string, err error) {
	if vv, ok := t.Get(v); ok {
		return vv, nil
	}

	switch vv := v.Value.(type) {
	case Fragment:
		compiled, err = vv.Compile(t)
		if err != nil {
			return "", errors.Wrap(err, "compile fragment")
		}

	default:
		compiled, err = t.Compile(LayoutValueQuote, Raw(fmt.Sprintf("%v", v.Value)))
		if err != nil {
			return "", errors.Wrapf(err, "compile LayoutValueQuote with value %v", v.Value)
		}
	}

	t.Set(v, compiled)
	return compiled, nil
}

// todo

// ValueGroups represents an array of value groups.
type ValueGroups struct {
	Values []*Values
	hash   hash
}

func (vg *ValueGroups) Empty() bool {
	if vg == nil || len(vg.Values) < 1 {
		return true
	}
	for i := range vg.Values {
		if !vg.Values[i].Empty() {
			return false
		}
	}
	return true
}

var _ = Fragment(&ValueGroups{})

// Values represents an array of Value.
type Values struct {
	Values []Fragment
	hash   hash
}

func (vs *Values) Empty() bool {
	if vs == nil || len(vs.Values) < 1 {
		return true
	}
	return false
}

var _ = Fragment(&Values{})

// NewValueGroup creates and returns an array of values.
func NewValueGroup(v ...Fragment) *Values {
	return &Values{Values: v}
}

// Hash returns a unique identifier for the struct.
func (vs *Values) Hash() string {
	return vs.hash.Hash(vs)
}

// Compile transforms the Values into an equivalent SQL representation.
func (vs *Values) Compile(layout *Template) (string, error) {
	if c, ok := layout.Get(vs); ok {
		return c, nil
	}

	var compiled string
	l := len(vs.Values)
	if l > 0 {
		chunks := make([]string, 0, l)
		for i := 0; i < l; i++ {
			chunk, err := vs.Values[i].Compile(layout)
			if err != nil {
				return "", err
			}
			chunks = append(chunks, chunk)
		}
		compiled = layout.Compile(layout.ClauseGroup, strings.Join(chunks, layout.ValueSeparator))
	}
	layout.Set(vs, compiled)
	return compiled, nil
}

// Hash returns a unique identifier for the struct.
func (vg *ValueGroups) Hash() string {
	return vg.hash.Hash(vg)
}

// Compile transforms the ValueGroups into an equivalent SQL representation.
func (vg *ValueGroups) Compile(layout *Template) (string, error) {
	if c, ok := layout.Get(vg); ok {
		return c, nil
	}

	var compiled string
	l := len(vg.Values)
	if l > 0 {
		chunks := make([]string, 0, l)
		for i := 0; i < l; i++ {
			chunk, err := vg.Values[i].Compile(layout)
			if err != nil {
				return "", err
			}
			chunks = append(chunks, chunk)
		}
		compiled = strings.Join(chunks, layout.ValueSeparator)
	}

	layout.Set(vg, compiled)
	return compiled, nil
}

// JoinValueGroups creates a new *ValueGroups object.
func JoinValueGroups(values ...*Values) *ValueGroups {
	return &ValueGroups{Values: values}
}
