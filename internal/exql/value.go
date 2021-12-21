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
//
// NOTE: Fields are public purely for the purpose of being hashable. Direct
// modifications to them after construction may not take effect depends on
// whether the hash has been computed.
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

var _ Fragment = (*ValuesFragment)(nil)

// ValuesFragment is a list of ValueFragment.
//
// NOTE: Fields are public purely for the purpose of being hashable. Direct
// modifications to them after construction may not take effect depends on
// whether the hash has been computed.
type ValuesFragment struct {
	hash   hash
	Values []*ValueFragment
}

// Values constructs a ValuesFragment with the given values.
func Values(values ...*ValueFragment) *ValuesFragment {
	return &ValuesFragment{
		Values: values,
	}
}

func (vs *ValuesFragment) Hash() string {
	return vs.hash.Hash(vs)
}

func (vs *ValuesFragment) Compile(t *Template) (compiled string, err error) {
	if len(vs.Values) == 0 {
		return "", nil
	}

	if v, ok := t.Get(vs); ok {
		return v, nil
	}

	out := make([]string, len(vs.Values))
	for i := range vs.Values {
		out[i], err = vs.Values[i].Compile(t)
		if err != nil {
			return "", errors.Wrap(err, "compile value")
		}
	}

	compiled, err = t.Compile(LayoutClauseGroup, strings.Join(out, t.layouts[LayoutValueSeparator]))
	if err != nil {
		return "", errors.Wrapf(err, "compile LayoutClauseGroup with values %v", out)
	}

	t.Set(vs, compiled)
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
