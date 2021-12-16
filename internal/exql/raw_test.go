// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package exql

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRaw_Compile(t *testing.T) {
	got, err := RawValue(" foo").Compile(defaultTemplate(t))
	assert.NoError(t, err)
	assert.Equal(t, "foo", got)
}

func TestRaw_String(t *testing.T) {
	got, err := RawValue("foo ").Compile(defaultTemplate(t))
	assert.NoError(t, err)
	assert.Equal(t, "foo", got)
}

func TestRaw_Hash(t *testing.T) {
	got := RawValue(" foo ").Hash()
	want := "*exql.Raw:14597207904236602666"
	assert.Equal(t, want, got)
}
