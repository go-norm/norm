// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package expr

// Constraint represents a single condition like "a = 1", where "a" is the key
// and "1" is the value.
type Constraint interface {
	// Key is the leftmost part of the constraint and usually contains a column
	// name.
	Key() interface{}
	// Value if the rightmost part of the constraint and usually contains a column
	// value.
	Value() interface{}
}

var _ Constraint = (*constraint)(nil)

type constraint struct {
	k interface{}
	v interface{}
}

// NewConstraint constructs a new constraint with the given key and value.
func NewConstraint(key interface{}, value interface{}) Constraint {
	return &constraint{
		k: key,
		v: value,
	}
}

func (c constraint) Key() interface{} {
	return c.k
}

func (c constraint) Value() interface{} {
	return c.v
}

// Constraints represents a list of constraints, like "a = 1, b = 2, c = 3".
type Constraints interface {
	// Constraints returns the list of constraints.
	Constraints() []Constraint
}
