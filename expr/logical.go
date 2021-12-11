// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package expr

import (
	"unknwon.dev/norm/internal/immutable"
)

// LogicalExpr represents an expression to be used in logical statements. It
// consists of a group of expressions joined by operators like
// LogicalOperatorAnd or LogicalOperatorOr.
type LogicalExpr interface {
	// Expressions returns all child expressions in the logical expression.
	Expressions() []LogicalExpr

	// Operator returns the Operator that joins all the child expressions in the
	// logical expression.
	Operator() LogicalOperator

	// Empty returns true if the logical expression has no child expressions.
	Empty() bool
}

// LogicalOperator is a logical operation on a compound statement.
type LogicalOperator uint

const (
	LogicalOperatorNone LogicalOperator = iota
	LogicalOperatorAnd
	LogicalOperatorOr
)

type logicalExpr struct {
	op LogicalOperator

	prev *logicalExpr
	fn   func(*[]LogicalExpr) error
}

// Logical constructs a LogicalExpr with given operator and expressions.
func Logical(op LogicalOperator, exprs ...LogicalExpr) LogicalExpr {
	l := &logicalExpr{op: op}
	if len(exprs) == 0 {
		return l
	}
	return l.frame(func(in *[]LogicalExpr) error {
		*in = append(*in, exprs...)
		return nil
	})
}

func (g *logicalExpr) Expressions() []LogicalExpr {
	exprs, err := immutable.FastForward(g)
	if err != nil {
		return nil
	}
	return *(exprs.(*[]LogicalExpr))
}

func (g *logicalExpr) Operator() LogicalOperator {
	return g.op
}

func (g *logicalExpr) Empty() bool {
	if g.fn != nil {
		return false
	}
	if g.prev != nil {
		return g.prev.Empty()
	}
	return true
}

func (g *logicalExpr) frame(fn func(*[]LogicalExpr) error) *logicalExpr {
	return &logicalExpr{prev: g, op: g.op, fn: fn}
}

var _ immutable.Immutable = (*logicalExpr)(nil)

func (g *logicalExpr) Prev() immutable.Immutable {
	if g == nil {
		return nil
	}
	return g.prev
}

func (g *logicalExpr) Fn(in interface{}) error {
	if g.fn == nil {
		return nil
	}
	return g.fn(in.(*[]LogicalExpr))
}

func (g *logicalExpr) Base() interface{} {
	return &[]LogicalExpr{}
}
