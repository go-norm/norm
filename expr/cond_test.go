// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package expr

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCond_Expressions(t *testing.T) {
	c := Cond{
		"age >":  1,
		"name =": "Joe",
	}
	exprs := c.Expressions()
	strs := make([]string, 0, len(exprs))
	for _, e := range exprs {
		strs = append(strs, e.String())
	}
	assert.Equal(t, []string{"(AND age > 1)", "(AND name = Joe)"}, strs)
}

func TestCond_Empty(t *testing.T) {
	c := Cond{}
	assert.True(t, c.Empty(), "should be empty")

	c = Cond{"id": 1}
	assert.False(t, c.Empty(), "should not be empty")
}
