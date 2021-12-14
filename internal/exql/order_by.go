// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package exql

type Order uint8

const (
	_ = Order(iota)
	Ascendant
	Descendent
)

// OrderBy represents an ORDER BY clause.
type OrderBy struct {
	SortColumns Fragment
	hash        hash
}
