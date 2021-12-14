// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package sqlbuilder

import (

	"github.com/pkg/errors"

	"unknwon.dev/norm/internal/exql"
)

type template struct {
}

func (t *template) toWhereClause(conds interface{}) (where *exql.Where, args []interface{}, err error) {
	return nil, nil, errors.Errorf("unexpected condition type %T", conds)
}
