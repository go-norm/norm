// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package exql

import (
	"reflect"
	"sync/atomic"

	"unknwon.dev/norm/internal/cache"
)

type hash struct {
	v atomic.Value
}

func (h *hash) Hash(i interface{}) string {
	v := h.v.Load()
	if r, ok := v.(string); ok && r != "" {
		return r
	}

	s := reflect.TypeOf(i).String() + ":" + cache.Hash(i)
	h.v.Store(s)
	return s
}

func (h *hash) Reset() {
	h.v.Store("")
}

type hashMap map[string]interface{}
