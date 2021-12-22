// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package expr

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConstraint(t *testing.T) {
	c := NewConstraint("name", "alice")
	assert.Equal(t, "name", c.Key())
	assert.Equal(t, "alice", c.Value())
}
