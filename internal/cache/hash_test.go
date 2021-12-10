// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHash(t *testing.T) {
	a := "a"
	want := Hash(a)
	got := Hash(a)
	assert.Equal(t, want, got)

	b := "b"
	got = Hash(b)
	assert.NotEqual(t, want, got)
}
