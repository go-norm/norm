// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"
	"github.com/pkg/errors"

	"unknwon.dev/norm"
	"unknwon.dev/norm/adapter"
	"unknwon.dev/norm/internal/exql"
	"unknwon.dev/norm/internal/sqlbuilder"
)

// OpenOptions 包含了开启数据库连接的选项。
type OpenOptions struct {
	// NowFunc 用于指定返回数据库连接当前时间的函数。
	norm.NowFunc
}

func Open(dsn string, opts ...OpenOptions) (norm.DB, error) {
	config, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, errors.Wrap(err, "parse config")
	}

	var opt OpenOptions
	if len(opts) > 0 {
		opt = opts[1]
	}
	if opt.NowFunc == nil {
		opt.NowFunc = norm.Now
	}

	db := stdlib.OpenDB(*config)
	tmpl := newTemplate()
	adp := newPostgresDBAdapter(db, tmpl)
	return &postgresDB{
		now:      opt.NowFunc,
		template: tmpl,
		driver:   newPostgresDBDriver(db),
		adapter:  adp,
		SQL:      sqlbuilder.New(adp, tmpl),
	}, nil
}

type postgresDB struct {
	now      norm.NowFunc
	template *exql.Template
	driver   *postgresDBDriver
	adapter  *postgresDBAdapter
	norm.SQL
}

func (db *postgresDB) Now() time.Time {
	return db.now()
}

func (db *postgresDB) Driver() norm.Driver {
	return db.driver
}

func (db *postgresDB) Adapter() adapter.Adapter {
	return db.adapter
}

func (db *postgresDB) Close() error {
	return db.driver.Close()
}

func (db *postgresDB) Transaction(ctx context.Context, fn func(tx norm.DB) error, opts ...*norm.TxOptions) error {
	var err error
	var tx *sql.Tx
	if len(opts) == 0 {
		tx, err = db.driver.BeginTx(ctx, nil)
	} else {
		tx, err = db.driver.BeginTx(ctx,
			&sql.TxOptions{
				Isolation: opts[0].Isolation,
				ReadOnly:  opts[0].ReadOnly,
			},
		)
	}
	if err != nil {
		return errors.Wrap(err, "begin")
	}

	adp := newPostgresTxAdapter(tx, db.template)
	err = fn(
		&postgresTX{
			now:     db.now,
			driver:  newPostgresTxDriver(tx),
			adapter: adp,
			SQL:     sqlbuilder.New(adp, db.template),
		},
	)
	if err != nil {
		errRollback := tx.Rollback()
		if errRollback != nil {
			return errors.Wrapf(err, "unable to rollback with %q", errRollback)
		}
		return err
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "commit")
	}
	return nil
}
