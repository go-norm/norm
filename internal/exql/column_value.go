// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package exql

import (
	"strings"

	"github.com/pkg/errors"
	"unknwon.dev/norm/expr"
)

var _ Fragment = (*ColumnValueFragment)(nil)

// ColumnValueFragment is a bundle of a column, a value, and their comparison
// operator.
//
// NOTE: Fields are public purely for the purpose of being hashable. Direct
// modifications to them after construction may not take effect depends on
// whether the hash has been computed.
type ColumnValueFragment struct {
	hash     hash
	Column   *ColumnFragment
	Operator interface{}
	Value    Fragment
}

// ColumnValue constructs a ColumnValueFragment with the given column, value,
// and their comparison operator.
//
// Arguments:
//   - The column name can be a string or RawFragment.
//   - The operator can be a string or expr.ComparisonOperator.
func ColumnValue(column interface{}, operator interface{}, value Fragment) *ColumnValueFragment {
	return &ColumnValueFragment{
		Column:   Column(column),
		Operator: operator,
		Value:    value,
	}
}

func (cv *ColumnValueFragment) Hash() string {
	return cv.hash.Hash(cv)
}

func (cv *ColumnValueFragment) Compile(t *Template) (string, error) {
	if v, ok := t.Get(cv); ok {
		return v, nil
	}

	column, err := cv.Column.Compile(t)
	if err != nil {
		return "", errors.Wrapf(err, "compile column")
	}

	var operator string
	switch v := cv.Operator.(type) {
	case string:
		operator = v
	case expr.ComparisonOperator:
		operator = t.Operator(v)
	default:
		return "", errors.Errorf("unsupported operator type %T", v)
	}

	data := map[string]string{
		"Column":   column,
		"Operator": operator,
	}
	if cv.Value != nil {
		value, err := cv.Value.Compile(t)
		if err != nil {
			return "", errors.Wrapf(err, "compile value")
		}
		data["Value"] = value
	}

	compiled, err := t.Compile(LayoutColumnValue, data)
	if err != nil {
		return "", errors.Wrapf(err, "compile LayoutColumnValue with data %v", data)
	}

	t.Set(cv, compiled)
	return compiled, nil
}

var _ Fragment = (*ColumnValuesFragment)(nil)

// ColumnValuesFragment is a list of ColumnValueFragment.
//
// NOTE: Fields are public purely for the purpose of being hashable. Direct
// modifications to them after construction may not take effect depends on
// whether the hash has been computed.
type ColumnValuesFragment struct {
	hash         hash
	ColumnValues []*ColumnValueFragment
}

// ColumnValues constructs a ColumnValuesFragment with the given column values.
func ColumnValues(values ...*ColumnValueFragment) *ColumnValuesFragment {
	return &ColumnValuesFragment{
		ColumnValues: values,
	}
}

func (cvs *ColumnValuesFragment) Hash() string {
	return cvs.hash.Hash(cvs)
}

func (cvs *ColumnValuesFragment) Compile(t *Template) (compiled string, err error) {
	if len(cvs.ColumnValues) == 0 {
		return "", nil
	}

	if v, ok := t.Get(cvs); ok {
		return v, nil
	}

	out := make([]string, len(cvs.ColumnValues))
	for i := range cvs.ColumnValues {
		out[i], err = cvs.ColumnValues[i].Compile(t)
		if err != nil {
			return "", errors.Wrap(err, "compile column value")
		}
	}

	compiled = strings.TrimSpace(strings.Join(out, t.layouts[LayoutIdentifierSeparator]))
	t.Set(cvs, compiled)
	return compiled, nil
}

// Append appends given column values to the ColumnValuesFragment.
func (cvs *ColumnValuesFragment) Append(values ...*ColumnValueFragment) *ColumnValuesFragment {
	cvs.ColumnValues = append(cvs.ColumnValues, values...)
	cvs.hash.Reset()
	return cvs
}
