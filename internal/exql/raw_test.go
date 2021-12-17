// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package exql

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"unknwon.dev/norm/internal/cache"
)

func TestRaw_Hash(t *testing.T) {
	t.Run("equal for same value", func(t *testing.T) {
		h1 := Raw("foo").Hash()
		h2 := Raw("foo").Hash()
		assert.Equal(t, h1, h2)
	})

	t.Run("not equal for different values", func(t *testing.T) {
		set := map[string]bool{}
		for _, v := range []cache.Hashable{
			Raw("foo"),
			Raw("bar"),
		} {
			require.False(t, set[v.Hash()], "should not have duplicates")
		}
	})
}

func TestRaw_Compile(t *testing.T) {
	got, err := Raw("foo").Compile(defaultTemplate(t))
	assert.NoError(t, err)
	assert.Equal(t, "foo", got)
}

func TestRaw_String(t *testing.T) {
	got, err := Raw("foo").Compile(defaultTemplate(t))
	assert.NoError(t, err)
	assert.Equal(t, "foo", got)
}
