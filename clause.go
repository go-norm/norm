// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package norm

// Selector represents a SQL query builder for the SELECT statement.
type Selector interface {
}

// Inserter represents a SQL query builder for the INSERT statement.
type Inserter interface {
}

// Updater represents a SQL query builder for the UPDATE statement.
type Updater interface {
}

// Deleter represents a SQL query builder for the DELETE statement.
type Deleter interface {
}
