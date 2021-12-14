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

// SortColumn represents the column-order relation in an ORDER BY clause.
type SortColumn struct {
	Column Fragment
	Order
	hash hash
}

// SortColumns represents the columns in an ORDER BY clause.
type SortColumns struct {
	Columns []Fragment
	hash    hash
}

// OrderBy represents an ORDER BY clause.
type OrderBy struct {
	SortColumns Fragment
	hash        hash
}
