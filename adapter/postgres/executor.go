// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package postgres

import (
	"database/sql"

	"unknwon.dev/norm/internal/adapter"
	"unknwon.dev/norm/internal/exql"
)

type postgresDBExecutor struct {
	*adapter.BaseDBExecutor
}

func newPostgresDBExecutor(db *sql.DB, t *exql.Template) *postgresDBExecutor {
	return &postgresDBExecutor{
		BaseDBExecutor: adapter.NewBaseDBExecutor(db, t, postgresTyper{}),
	}
}

type postgresTxExecutor struct {
	*adapter.BaseTxExecutor
}

func newPostgresTxExecutor(tx *sql.Tx, t *exql.Template) *postgresTxExecutor {
	return &postgresTxExecutor{
		BaseTxExecutor: adapter.NewBaseTxExecutor(tx, t, postgresTyper{}),
	}
}
