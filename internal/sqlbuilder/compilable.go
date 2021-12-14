// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package sqlbuilder

//go:generate go-mockgen --force unknwon.dev/norm/internal/sqlbuilder -i compilable -o mock_compilable_test.go
type compilable interface {
	Compile() (string, error)
	Arguments() []interface{}
}
