// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package exql

import (
	"fmt"
)

var _ Fragment = (*RawFragment)(nil)

// RawFragment is a value that is meant to be used in a query without escaping.
//
// NOTE: Fields are public purely for the purpose of being hashable. Direct
// modifications to them after construction may not take effect depends on
// whether the hash has been computed.
type RawFragment struct {
	hash  hash
	Value string
}

// Raw constructs a RawFragment with the given value.
func Raw(value string) *RawFragment {
	return &RawFragment{
		Value: value,
	}
}

func (r *RawFragment) Hash() string {
	return r.hash.Hash(r)
}

func (r *RawFragment) Compile(*Template) (string, error) {
	return r.Value, nil
}

var _ fmt.Stringer = (*RawFragment)(nil)

func (r *RawFragment) String() string {
	return r.Value
}
