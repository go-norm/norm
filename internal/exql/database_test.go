// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package exql

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDatabase(t *testing.T) {
	tmpl := defaultTemplate(t)
	d := Database("norm")

	got, err := d.Compile(tmpl)
	require.NoError(t, err)

	want := `"norm"`
	assert.Equal(t, want, strings.TrimSpace(got))

	t.Run("cache hit", func(t *testing.T) {
		got, err := Database("norm").Compile(tmpl)
		assert.NoError(t, err)
		assert.Equal(t, want, got)
	})
}
