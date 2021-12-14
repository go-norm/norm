// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package sqlbuilder

import (
	"database/sql"

	"unknwon.dev/norm/adapter"
)

type iterator struct {
	adapter adapter.Adapter
	cursor  *sql.Rows // This is the main query cursor. It starts as a nil value.
	err     error
}

type sqlBuilder struct {
	adapter.Adapter
	t *template
}
