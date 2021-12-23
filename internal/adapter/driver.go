// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package adapter

import (
	"database/sql"
	"time"
)

type BaseDBDriver struct {
	*sql.DB
}

func NewBaseDBDriver(db *sql.DB) *BaseDBDriver {
	return &BaseDBDriver{
		DB: db,
	}
}

func (d *BaseDBDriver) SetConnMaxLifetime(duration time.Duration) {
	d.DB.SetConnMaxLifetime(duration)
}

func (d *BaseDBDriver) SetConnMaxIdleTime(duration time.Duration) {
	d.DB.SetConnMaxIdleTime(duration)
}

func (d *BaseDBDriver) SetMaxIdleConns(n int) {
	d.DB.SetMaxIdleConns(n)
}

func (d *BaseDBDriver) SetMaxOpenConns(n int) {
	d.DB.SetMaxOpenConns(n)
}

type BaseTxDriver struct {
	*sql.Tx
}

func NewBaseTxDriver(tx *sql.Tx) *BaseTxDriver {
	return &BaseTxDriver{
		Tx: tx,
	}
}

func (d *BaseTxDriver) SetConnMaxLifetime(time.Duration) {
	panic("SetConnMaxLifetime is not available within a transaction")
}

func (d *BaseTxDriver) SetConnMaxIdleTime(time.Duration) {
	panic("SetConnMaxIdleTime is not available within a transaction")
}

func (d *BaseTxDriver) SetMaxIdleConns(int) {
	panic("SetMaxIdleConns is not available within a transaction")
}

func (d *BaseTxDriver) SetMaxOpenConns(int) {
	panic("SetMaxOpenConns is not available within a transaction")
}
