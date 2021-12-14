// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package exql

// Or represents an SQL OR operator.
type Or Where

// And represents an SQL AND operator.
type And Where

// Where represents an SQL WHERE clause.
type Where struct {
	Conditions []Fragment
	hash       hash
}

func (w *Where) Append(a *Where) *Where {
	return w
}
