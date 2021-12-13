// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package expr

// ConstraintValuer allows constraints to use specific values of their own.
type ConstraintValuer interface {
	ConstraintValue() interface{}
}

// Constraint interface represents a single condition, like "a = 1".  where `a`
// is the key and `1` is the value. This is an exported interface but it's
// rarely used directly, you may want to use the `db.Cond{}` map instead.
type Constraint interface {
	// Key is the leftmost part of the constraint and usually contains a column
	// name.
	Key() interface{}
	// Value if the rightmost part of the constraint and usually contains a
	// column value.
	Value() interface{}
}

// Constraints interface represents an array of constraints, like "a = 1, b =
// 2, c = 3".
type Constraints interface {
	// Constraints returns an array of constraints.
	Constraints() []Constraint
}

type constraint struct {
	k interface{}
	v interface{}
}

func (c constraint) Key() interface{} {
	return c.k
}

func (c constraint) Value() interface{} {
	if constraintValuer, ok := c.v.(ConstraintValuer); ok {
		return constraintValuer.ConstraintValue()
	}
	return c.v
}

// NewConstraint creates a constraint.
func NewConstraint(key interface{}, value interface{}) Constraint {
	return &constraint{k: key, v: value}
}

var (
	_ = Constraint(&constraint{})
)
