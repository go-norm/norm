// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package sqlbuilder

import (
	"unknwon.dev/norm"
	"unknwon.dev/norm/adapter"
	"unknwon.dev/norm/internal/exql"
)

type sqlBuilder struct {
	adapter.Adapter
	*exql.Template
}

func New(adapter adapter.Adapter, t *exql.Template) norm.SQL {
	return &sqlBuilder{
		Adapter:  adapter,
		Template: t,
	}
}

func (b *sqlBuilder) Select(columns ...interface{}) norm.Selector {
	sel := &selector{
		builder: b,
	}
	return sel.Columns(columns...)
}

func (b *sqlBuilder) SelectFrom(tables ...interface{}) norm.Selector {
	sel := &selector{
		builder: b,
	}
	return sel.From(tables...)
}
