// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package postgres

import (
	"context"
	"time"

	"github.com/pkg/errors"

	"unknwon.dev/norm"
	"unknwon.dev/norm/adapter"
)

// postgresTX 实现了事务状态下的数据库方法集。
type postgresTX struct {
	now     norm.NowFunc
	driver  *postgresTxDriver
	adapter *postgresTxAdapter
	norm.SQL
}

func (tx *postgresTX) Now() time.Time {
	return tx.now()
}

func (tx *postgresTX) Driver() norm.Driver {
	return tx.driver
}

func (tx *postgresTX) Adapter() adapter.Adapter {
	return tx.adapter
}

func (tx *postgresTX) Close() error {
	return errors.New("cannot close connection within a transaction")
}

func (tx *postgresTX) Transaction(ctx context.Context, fn func(tx norm.DB) error, opts ...*norm.TxOptions) error {
	return fn(tx)
}
