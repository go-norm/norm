// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package cache

import (
	"strconv"

	"github.com/mitchellh/hashstructure/v2"
)

// Hashable represents an object that can compute a hash.
type Hashable interface {
	// Hash returns the computed hash of the object.
	Hash() string
}

// Hash computes the hash of the given object.
func Hash(v interface{}) string {
	q, err := hashstructure.Hash(v, hashstructure.FormatV2, nil)
	if err != nil {
		panic("unable to hash the object: " + err.Error())
	}
	return strconv.FormatUint(q, 10)
}
