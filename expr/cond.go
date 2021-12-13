// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package expr

import (
	"fmt"
	"sort"
	"strings"
)

var _ LogicalExpr = (*Cond)(nil)

// Cond is a map that defines conditions for a query.
//
// Each entry of the map represents a constraint (a column-value relation bound
// by a ComparisonOperator). The comparison can be specified after the column
// name, if no ComparisonOperator is provided the equality operator (LogicalAnd)
// is used as the default.
//
// Examples:
//
//   => age = 18
//   expr.Cond{"age": 18}
//
//   => age >= 18
//   expr.Cond{"age >=": 18}
//
//   => id IN (1, 2, 3)
//   expr.Cond{"id IN": []{1, 2, 3}}
//
//   => age > 32 AND age < 35
//    expr.Cond{"age >": 32, "age <": 35}
type Cond map[interface{}]interface{}

func (c Cond) keys() []interface{} {
	keys := make([]interface{}, 0, len(c))
	for k := range c {
		keys = append(keys, k)
	}
	if len(c) > 1 {
		sort.Slice(keys, func(i, j int) bool {
			return fmt.Sprintf("%v", keys[i]) < fmt.Sprintf("%v", keys[j])
		})
	}
	return keys
}

func (c Cond) Expressions() []LogicalExpr {
	exprs := make([]LogicalExpr, 0, len(c))
	for _, k := range c.keys() {
		exprs = append(exprs, Cond{k: c[k]})
	}
	return exprs
}

func (c Cond) Operator() LogicalOperator {
	return LogicalAnd
}

func (c Cond) Empty() bool {
	return len(c) == 0
}

func (c Cond) String() string {
	strs := make([]string, 0, len(c))
	for k, v := range c {
		strs = append(strs, fmt.Sprintf("%s %v", k, v))
	}
	return fmt.Sprintf("(%s %s)", c.Operator(), strings.Join(strs, " "))
}

var _ Constraints = (Cond)(nil)

// Constraints returns each one of the Cond map entires as a constraint.
func (c Cond) Constraints() []Constraint {
	z := make([]Constraint, 0, len(c))
	for _, k := range c.keys() {
		z = append(z, NewConstraint(k, c[k]))
	}
	return z
}
