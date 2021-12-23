// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package postgres

import (
	"database/sql"
	"database/sql/driver"
	"fmt"

	"github.com/jackc/pgtype"

	"unknwon.dev/norm/types"
)

type postgresTyper struct{}

func (t postgresTyper) Scanner(v interface{}) interface{} {
	switch v := v.(type) {
	case *types.Int64Array:
		return (*int64Array)(v)
	}
	return v
}

func (t postgresTyper) Valuer(v interface{}) interface{} {
	switch v := v.(type) {
	case types.Int64Array:
		return (*int64Array)(&v)
	}
	return v
}

type int64Array []int64

var _ sql.Scanner = (*int64Array)(nil)

func (v *int64Array) Scan(src interface{}) error {
	var dst []int64
	t := pgtype.Int8Array{}
	if err := t.Scan(src); err != nil {
		return err
	}
	if err := t.AssignTo(&dst); err != nil {
		return err
	}
	*v = dst
	return nil
}

var _ driver.Valuer = (*int64Array)(nil)

func (v int64Array) Value() (driver.Value, error) {
	t := pgtype.Int8Array{}
	if err := t.Set(v); err != nil {
		return nil, err
	}
	return t.Value()
}

var _ fmt.Stringer = (*int64Array)(nil)

func (v *int64Array) String() string {
	val, err := v.Value()
	if err != nil {
		return fmt.Sprintf("<norm.postgres.int64Array.Value: %v>", err)
	}
	return fmt.Sprintf("%v", val)
}
