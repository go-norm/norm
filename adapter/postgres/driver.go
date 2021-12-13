// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package postgres

import (
	"database/sql"

	"unknwon.dev/norm/internal/adapter"
)

type postgresDBDriver struct {
	*adapter.BaseDBDriver
}

func newPostgresDBDriver(db *sql.DB) *postgresDBDriver {
	return &postgresDBDriver{
		BaseDBDriver: adapter.NewBaseDBDriver(db),
	}
}

type postgresTxDriver struct {
	*adapter.BaseTxDriver
}

func newPostgresTxDriver(tx *sql.Tx) *postgresTxDriver {
	return &postgresTxDriver{
		BaseTxDriver: adapter.NewBaseTxDriver(tx),
	}
}
