// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package exql

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHash(t *testing.T) {
	name := "alice"
	h := hash{}
	assert.Nil(t, h.v.Load())
	h.Hash(name)
	assert.NotNil(t, h.v.Load())

	h.Reset()
	got := h.v.Load().(string)
	assert.Empty(t, got)
}
