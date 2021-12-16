// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package exql

import (
	"fmt"
	"strings"
)

var _ Fragment = (*Raw)(nil)

// Raw represents a value that is meant to be used in a query without escaping.
type Raw struct {
	Value string // Value should not be modified after assigned.
	hash  hash
}

// RawValue creates and returns a new raw value, surrounding spaces are trimmed.
func RawValue(v string) *Raw {
	return &Raw{
		Value: strings.TrimSpace(v),
	}
}

func (r *Raw) Hash() string {
	return r.hash.Hash(r)
}

func (r *Raw) Compile(*Template) (string, error) {
	return r.Value, nil
}

var _ fmt.Stringer = (*Raw)(nil)

func (r *Raw) String() string {
	return r.Value
}
