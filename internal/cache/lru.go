// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package cache

import (
	"container/list"
	"sync"
)

// LRU is an LRU cache implementation.
type LRU struct {
	capacity int

	mu   sync.RWMutex
	data map[string]*list.Element
	list *list.List
}

// Options contains options for creating an LRU cache.
type Options struct {
	// Capacity is the number of items to hold before evicting. Default is 128.
	Capacity int
}

// NewLRU creates and returns a new LRU cache.
func NewLRU(opts ...Options) *LRU {
	var opt Options
	if len(opts) > 0 {
		opt = opts[0]
	}

	if opt.Capacity <= 0 {
		opt.Capacity = 128
	}

	return &LRU{
		capacity: opt.Capacity,
		data:     make(map[string]*list.Element),
		list:     list.New(),
	}
}

type item struct {
	key   string
	value interface{}
}

// Get attempts to retrieve a cached value as a string. It returns false if the
// value does not exist or cannot be type-casted to a string.
func (c *LRU) Get(h Hashable) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	e, ok := c.data[h.Hash()]
	if !ok {
		return "", false
	}

	c.list.MoveToFront(e)
	s, ok := e.Value.(*item).value.(string)
	return s, ok
}

// Set stores the given value to the cache.
func (c *LRU) Set(h Hashable, v interface{}) {
	key := h.Hash()

	c.mu.Lock()
	defer c.mu.Unlock()

	if e, ok := c.data[key]; ok {
		e.Value.(*item).value = v
		c.list.MoveToFront(e)
		return
	}

	c.data[key] = c.list.PushFront(
		&item{
			key:   key,
			value: v,
		},
	)

	if c.list.Len() > c.capacity {
		e := c.list.Remove(c.list.Back())
		delete(c.data, e.(*item).key)
	}
}
