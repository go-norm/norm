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
	iadapter "unknwon.dev/norm/internal/adapter"
	"unknwon.dev/norm/internal/exql"
	"unknwon.dev/norm/internal/sqlbuilder"
)

// OpenOptions contains options for opening a PostgreSQL database connection.
type OpenOptions struct {
	// NowFunc is a function to return the current time. Default is time.Now().
	norm.NowFunc
}

// Open opens a PostgreSQL database connection using given DSN and options.
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

	tmpl, err := exql.DefaultTemplate()
	if err != nil {
		return nil, errors.Wrap(err, "get template")
	}

	db := stdlib.OpenDB(*config)
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

type postgresDBDriver struct {
	*iadapter.BaseDBDriver
}

func newPostgresDBDriver(db *sql.DB) *postgresDBDriver {
	return &postgresDBDriver{
		BaseDBDriver: iadapter.NewBaseDBDriver(db),
	}
}

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

type postgresTxDriver struct {
	*iadapter.BaseTxDriver
}

func newPostgresTxDriver(tx *sql.Tx) *postgresTxDriver {
	return &postgresTxDriver{
		BaseTxDriver: iadapter.NewBaseTxDriver(tx),
	}
}
