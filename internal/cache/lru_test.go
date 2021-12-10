// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package cache

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

type hashable struct {
	Name string
}

func (h *hashable) Hash() string {
	return Hash(h)
}

func TestLRU(t *testing.T) {
	c := NewLRU()

	h := &hashable{Name: "foo"}
	_, ok := c.Get(h)
	assert.False(t, ok, "cache miss")

	c.Set(h, "bar")
	got, ok := c.Get(h)
	assert.True(t, ok, "cache hit")
	assert.Equal(t, "bar", got)
}

func TestLRU_Capacity(t *testing.T) {
	c := NewLRU(
		Options{
			Capacity: 2,
		},
	)

	h1 := &hashable{Name: "foo1"}
	h2 := &hashable{Name: "foo2"}
	h3 := &hashable{Name: "foo3"}

	c.Set(h1, "bar1")
	c.Set(h2, "bar2")

	_, ok := c.Get(h1)
	assert.True(t, ok, "cache hit")

	c.Set(h3, "bar3") // h2 evicted

	_, ok = c.Get(h1)
	assert.True(t, ok, "cache hit")
	_, ok = c.Get(h2)
	assert.False(t, ok, "cache miss")
	_, ok = c.Get(h3)
	assert.True(t, ok, "cache hit")
}

func BenchmarkSet(b *testing.B) {
	c := NewLRU()
	var s string
	for i := 0; i < b.N; i++ {
		s = strconv.Itoa(i)
		c.Set(&hashable{Name: s}, s)
	}
}

func BenchmarkSet_SameValue(b *testing.B) {
	c := NewLRU()
	h := &hashable{Name: "foo"}
	for i := 0; i < b.N; i++ {
		c.Set(h, "bar")
	}
}

func BenchmarkGet(b *testing.B) {
	c := NewLRU()
	for i := 0; i < 128; i++ {
		s := strconv.Itoa(i)
		c.Set(&hashable{Name: s}, s)
	}

	for i := 0; i < b.N; i++ {
		c.Get(&hashable{Name: "127"})
	}
}

func BenchmarkGet_NotFound(b *testing.B) {
	c := NewLRU()
	for i := 0; i < 128; i++ {
		s := strconv.Itoa(i)
		c.Set(&hashable{Name: s}, s)
	}

	for i := 0; i < b.N; i++ {
		c.Get(&hashable{Name: "128"})
	}
}
