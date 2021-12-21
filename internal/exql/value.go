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

	var value string
	switch vv := v.Value.(type) {
	case Fragment:
		value, err = vv.Compile(t)
		if err != nil {
			return "", errors.Wrap(err, "compile fragment")
		}

	case fmt.Stringer:
		value = vv.String()

	default:
		value = fmt.Sprintf("%v", v.Value)
	}

	compiled, err = t.Compile(LayoutValueQuote, value)
	if err != nil {
		return "", errors.Wrapf(err, "compile LayoutValueQuote with value %v", value)
	}

	t.Set(v, compiled)
	return compiled, nil
}

var _ Fragment = (*ValuesGroupFragment)(nil)

// ValuesGroupFragment is a group of ValueFragment.
//
// NOTE: Fields are public purely for the purpose of being hashable. Direct
// modifications to them after construction may not take effect depends on
// whether the hash has been computed.
type ValuesGroupFragment struct {
	hash   hash
	Values []Fragment
}

// ValuesGroup constructs a ValuesGroupFragment with the given values.
func ValuesGroup(values ...Fragment) *ValuesGroupFragment {
	return &ValuesGroupFragment{
		Values: values,
	}
}

func (vg *ValuesGroupFragment) Hash() string {
	return vg.hash.Hash(vg)
}

func (vg *ValuesGroupFragment) Compile(t *Template) (compiled string, err error) {
	if len(vg.Values) == 0 {
		return "", nil
	}

	if v, ok := t.Get(vg); ok {
		return v, nil
	}

	out := make([]string, len(vg.Values))
	for i := range vg.Values {
		out[i], err = vg.Values[i].Compile(t)
		if err != nil {
			return "", errors.Wrap(err, "compile value")
		}
	}

	compiled, err = t.Compile(LayoutClauseGroup, strings.Join(out, t.layouts[LayoutValueSeparator]))
	if err != nil {
		return "", errors.Wrapf(err, "compile LayoutClauseGroup with values %v", out)
	}

	t.Set(vg, compiled)
	return compiled, nil
}

var _ Fragment = (*ValuesGroupsFragment)(nil)

// ValuesGroupsFragment is a list of ValuesGroupFragment.
//
// NOTE: Fields are public purely for the purpose of being hashable. Direct
// modifications to them after construction may not take effect depends on
// whether the hash has been computed.
type ValuesGroupsFragment struct {
	hash   hash
	Groups []*ValuesGroupFragment
}

// ValuesGroups constructs a ValuesGroupsFragment with the given list of values
// groups.
func ValuesGroups(groups ...*ValuesGroupFragment) *ValuesGroupsFragment {
	return &ValuesGroupsFragment{
		Groups: groups,
	}
}

func (vgs *ValuesGroupsFragment) Hash() string {
	return vgs.hash.Hash(vgs)
}

func (vgs *ValuesGroupsFragment) Compile(t *Template) (compiled string, err error) {
	if len(vgs.Groups) == 0 {
		return "", nil
	}

	if v, ok := t.Get(vgs); ok {
		return v, nil
	}

	out := make([]string, len(vgs.Groups))
	for i := range vgs.Groups {
		out[i], err = vgs.Groups[i].Compile(t)
		if err != nil {
			return "", errors.Wrap(err, "compile values group")
		}
	}

	compiled = strings.TrimSpace(strings.Join(out, t.layouts[LayoutValueSeparator]))
	t.Set(vgs, compiled)
	return compiled, nil
}
