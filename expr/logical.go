// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package expr

import (
	"fmt"
	"strings"

	"unknwon.dev/norm/internal/immutable"
)

// LogicalExpr represents an expression to be used in logical statements. It
// consists of a group of expressions joined by operators like LogicalAnd or
// LogicalOr.
type LogicalExpr interface {
	// Expressions returns all child expressions in the logical expression.
	Expressions() []LogicalExpr
	// Operator returns the Operator that joins all the child expressions in the
	// logical expression.
	Operator() LogicalOperator
	// Empty returns true if the logical expression has no child expressions.
	Empty() bool

	fmt.Stringer
}

// LogicalOperator is a logical operation on a compound statement.
type LogicalOperator uint

const (
	LogicalNone LogicalOperator = iota
	LogicalAnd
	LogicalOr
)

func (op LogicalOperator) String() string {
	switch op {
	case LogicalAnd:
		return "AND"
	case LogicalOr:
		return "OR"
	}
	return "NONE"
}

type logicalExpr struct {
	op LogicalOperator

	prev *logicalExpr
	fn   frameFunc
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

func (e *logicalExpr) Expressions() []LogicalExpr {
	exprs, err := immutable.FastForward(e)
	if err != nil {
		return nil
	}
	return *(exprs.(*[]LogicalExpr))
}

func (e *logicalExpr) Operator() LogicalOperator {
	return e.op
}

func (e *logicalExpr) Empty() bool {
	if e.fn != nil {
		return false
	}
	if e.prev != nil {
		return e.prev.Empty()
	}
	return true
}

type frameFunc func(*[]LogicalExpr) error

func (e *logicalExpr) frame(fn frameFunc) *logicalExpr {
	return &logicalExpr{
		prev: e,
		op:   e.op,
		fn:   fn,
	}
}

func (e *logicalExpr) String() string {
	exprs := e.Expressions()
	if len(exprs) == 0 {
		return e.op.String()
	} else if len(exprs) == 1 {
		return exprs[0].String()
	}

	strs := make([]string, 0, len(exprs))
	for _, e := range exprs {
		strs = append(strs, e.String())
	}
	return fmt.Sprintf("(%s %s)", e.op, strings.Join(strs, " "))
}

var _ immutable.Immutable = (*logicalExpr)(nil)

func (e *logicalExpr) Prev() immutable.Immutable {
	if e == nil {
		return nil
	}
	return e.prev
}

func (e *logicalExpr) Fn(in interface{}) error {
	if e.fn == nil {
		return nil
	}
	return e.fn(in.(*[]LogicalExpr))
}

func (e *logicalExpr) Base() interface{} {
	return &[]LogicalExpr{}
}

var _ LogicalExpr = (*RawExpr)(nil)

// RawExpr is a raw expression that can bypass SQL filters.
type RawExpr struct {
	value string
	args  *[]interface{}
}

// Raw returns the value of the raw expression.
func (e RawExpr) Raw() string {
	return e.value
}

// Arguments returns arguments of the raw expression. It returns nil if there is
// no argument.
func (e *RawExpr) Arguments() []interface{} {
	if e.args != nil {
		return *e.args
	}
	return nil
}

func (e RawExpr) String() string {
	return e.Raw()
}

func (e *RawExpr) Expressions() []LogicalExpr {
	return []LogicalExpr{e}
}

func (e RawExpr) Operator() LogicalOperator {
	return LogicalNone
}

func (e *RawExpr) Empty() bool {
	return e.value == ""
}

// Raw keeps the given value and arguments unfiltered and passes them directly
// to the query .
//
// CAUTION: It is possible to cause SQL injection if user inputs are not
// properly sanitized before giving to this function.
//
// Example:
//
//   => SOUNDEX('Hello')
//   expr.Raw("SOUNDEX('Hello')")
func Raw(value string, args ...interface{}) *RawExpr {
	r := &RawExpr{value: value, args: nil}
	if len(args) > 0 {
		r.args = &args
	}
	return r
}

var _ LogicalExpr = (*FuncExpr)(nil)

// FuncExpr is similar to RawExpr but is designed for database functions.
type FuncExpr struct {
	*RawExpr
}

// Name returns the name of the function.
func (e *FuncExpr) Name() string {
	return e.Raw()
}

// Func returns a database function expression.
//
// Examples:
//
//	 => MOD(29, 9)
//   expr.Func("MOD", 29, 9)
//
//   => CONCAT("foo", "bar")
//   expr.Func("CONCAT", "foo", "bar")
//
//   => NOW()
//   expr.Func("NOW")
//
//   => RTRIM("Hello  ")
//   expr.Func("RTRIM", "Hello  ")
func Func(name string, args ...interface{}) *FuncExpr {
	return &FuncExpr{
		RawExpr: Raw(name, args...),
	}
}

var _ LogicalExpr = (*AndExpr)(nil)

// AndExpr is an expression that joins child expressions by logical conjunctions.
type AndExpr struct {
	*logicalExpr
}

// And adds more AND conditions to the expression.
func (e *AndExpr) And(ands ...LogicalExpr) *AndExpr {
	var fn frameFunc
	if len(ands) > 0 {
		fn = func(in *[]LogicalExpr) error {
			*in = append(*in, ands...)
			return nil
		}
	}
	return &AndExpr{logicalExpr: e.frame(fn)}
}

// And joins given expressions by logical conjunctions (LogicalAnd). Expressions
// can be represented by mixes of `expr.Cond{}`, `expr.Or()` and `expr.And()`.
//
// Examples:
//
//   => name = 'Peter' AND last_name = 'Parker'
//	 expr.And(
//	     expr.Cond{"name": "Peter"},
//	     expr.Cond{"last_name": "Parker "},
//	 )
//
//   => (name = 'Peter' OR name = 'Mickey') AND last_name = 'Mouse'
//	 expr.And(
//       expr.Or(
//           expr.Cond{"name": "Peter"},
//	         expr.Cond{"name": "Mickey"},
//       ),
//       expr.Cond{"last_name": "Mouse"},
//   )
func And(exprs ...LogicalExpr) *AndExpr {
	return &AndExpr{Logical(LogicalAnd, exprs...).(*logicalExpr)}
}

var _ LogicalExpr = (*OrExpr)(nil)

// OrExpr is an expression that joins child expressions by logical disjunction.
type OrExpr struct {
	*logicalExpr
}

// Or adds more OR conditions to the expression.
func (e *OrExpr) Or(ors ...LogicalExpr) *OrExpr {
	var fn frameFunc
	if len(ors) > 0 {
		fn = func(in *[]LogicalExpr) error {
			*in = append(*in, ors...)
			return nil
		}
	}
	return &OrExpr{logicalExpr: e.frame(fn)}
}

// Or joins given expressions by logical disjunction (LogicalOr). Expressions
// can be represented by mixes of `expr.Cond{}`, `expr.Or()` and `expr.And()`.
//
// Example:
//
//   => year = 2012 OR year = 1987
//   expr.Or(
//       expr.Cond{"year": 2012},
//       expr.Cond{"year": 1987},
//   )
func Or(exprs ...LogicalExpr) *OrExpr {
	joined := func(in []LogicalExpr) []LogicalExpr {
		for i := range in {
			cond, ok := in[i].(Cond)
			if ok && !cond.Empty() {
				in[i] = And(cond)
			}
		}
		return in
	}(exprs)
	return &OrExpr{
		logicalExpr: Logical(LogicalOr, joined...).(*logicalExpr),
	}
}
