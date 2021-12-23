// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package postgres

import (
	"bytes"
	"database/sql"
	"strconv"

	"unknwon.dev/norm/adapter"
	iadapter "unknwon.dev/norm/internal/adapter"
	"unknwon.dev/norm/internal/exql"
)

type postgresDBAdapter struct {
	executor *postgresDBExecutor
	typer    postgresTyper
}

func newPostgresDBAdapter(db *sql.DB, t *exql.Template) *postgresDBAdapter {
	adp := &postgresDBAdapter{}
	adp.executor = newPostgresDBExecutor(db, t, adp)
	return adp
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

func formatSQL(sql string) string {
	var buf bytes.Buffer
	j := 1
	for i := range sql {
		if sql[i] == '?' {
			buf.WriteByte('$')
			buf.WriteString(strconv.Itoa(j))
			j++
		} else {
			buf.WriteByte(sql[i])
		}
	}
	return exql.StripWhitespace(buf.String())
}

func (adp *postgresDBAdapter) FormatSQL(sql string) string {
	return formatSQL(sql)
}

type postgresDBExecutor struct {
	*iadapter.BaseDBExecutor
}

func newPostgresDBExecutor(db *sql.DB, t *exql.Template, adapter adapter.Adapter) *postgresDBExecutor {
	return &postgresDBExecutor{
		BaseDBExecutor: iadapter.NewBaseDBExecutor(db, t, adapter),
	}
}

type postgresTxAdapter struct {
	executor *postgresTxExecutor
	typer    postgresTyper
}

func newPostgresTxAdapter(tx *sql.Tx, t *exql.Template) *postgresTxAdapter {
	adp := &postgresTxAdapter{}
	adp.executor = newPostgresTxExecutor(tx, t, adp)
	return adp
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

func (adp *postgresTxAdapter) FormatSQL(sql string) string {
	return formatSQL(sql)
}

type postgresTxExecutor struct {
	*iadapter.BaseTxExecutor
}

func newPostgresTxExecutor(tx *sql.Tx, t *exql.Template, adapter adapter.Adapter) *postgresTxExecutor {
	return &postgresTxExecutor{
		BaseTxExecutor: iadapter.NewBaseTxExecutor(tx, t, adapter),
	}
}
