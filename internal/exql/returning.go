// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package exql

// Returning represents a RETURNING clause.
type Returning struct {
	*Columns
	hash hash
}

// Hash returns a unique identifier for the struct.
func (r *Returning) Hash() string {
	return r.hash.Hash(r)
}

var _ = Fragment(&Returning{})

// ReturningColumns creates and returns an array of Column.
func ReturningColumns(columns ...Fragment) *Returning {
	return &Returning{Columns: &Columns{Columns: columns}}
}

// Compile transforms the clause into its equivalent SQL representation.
func (r *Returning) Compile(layout *Template) (string, error) {
	if z, ok := layout.Get(r); ok {
		return z, nil
	}

	compiled, err := r.Columns.Compile(layout)
	if err != nil {
		return "", err
	}

	layout.Set(r, compiled)
	return compiled, nil
}
