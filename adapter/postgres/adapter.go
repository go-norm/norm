// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package postgres

import (
	"database/sql"

	"unknwon.dev/norm/adapter"
	"unknwon.dev/norm/internal/exql"
)

type postgresDBAdapter struct {
	executor *postgresDBExecutor
	typer    postgresTyper
}

func newPostgresDBAdapter(db *sql.DB, t *exql.Template) *postgresDBAdapter {
	return &postgresDBAdapter{
		executor: newPostgresDBExecutor(db, t),
	}
}

func (*postgresDBAdapter) Name() adapter.Name {
	return adapter.PostgreSQL
}

func (adp *postgresDBAdapter) Executor() adapter.Executor {
	return adp.executor
}

func (adp *postgresDBAdapter) Typer() adapter.Typer {
	return adp.typer
}

type postgresTxAdapter struct {
	executor *postgresTxExecutor
	typer    postgresTyper
}

func newPostgresTxAdapter(tx *sql.Tx, t *exql.Template) *postgresTxAdapter {
	return &postgresTxAdapter{
		executor: newPostgresTxExecutor(tx, t),
	}
}

func (*postgresTxAdapter) Name() adapter.Name {
	return adapter.PostgreSQL
}

func (adp *postgresTxAdapter) Executor() adapter.Executor {
	return adp.executor
}

func (adp *postgresTxAdapter) Typer() adapter.Typer {
	return adp.typer
}
