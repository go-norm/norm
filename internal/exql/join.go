// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package exql

// Join represents a generic JOIN statement.
type Join struct {
	Type  string
	Table Fragment
	On    Fragment
	Using Fragment
	hash  hash
}

// On represents JOIN conditions.
type On Where
