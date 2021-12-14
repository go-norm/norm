// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package exql

import (
	"strings"
)

type Raw struct {
	Value string // Value should not be modified after assigned.
	hash  hash
}

func RawValue(v string) *Raw {
	return &Raw{
		Value: strings.TrimSpace(v),
	}
}
