// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package postgres

import (
	"bytes"
	"strconv"
	"sync"

	"unknwon.dev/norm/expr"
	"unknwon.dev/norm/internal/cache"
	"unknwon.dev/norm/internal/exql"
)

const (
	tmplColumnSeparator     = `.`
	tmplIdentifierSeparator = `, `
	tmplIdentifierQuote     = `"{{.Value}}"`
	tmplValueSeparator      = `, `
	tmplValueQuote          = `'{{.}}'`
	tmplAndKeyword          = `AND`
	tmplOrKeyword           = `OR`
	tmplDescKeyword         = `DESC`
	tmplAscKeyword          = `ASC`
	tmplAssignmentOperator  = `=`
	tmplClauseGroup         = `({{.}})`
	tmplClauseOperator      = ` {{.}} `
	tmplColumnValue         = `{{.Column}} {{.Operator}} {{.Value}}`
	tmplTableAliasLayout    = `{{.Name}}{{if .Alias}} AS {{.Alias}}{{end}}`
	tmplColumnAliasLayout   = `{{.Name}}{{if .Alias}} AS {{.Alias}}{{end}}`
	tmplSortByColumnLayout  = `{{.Column}} {{.Order}}`

	tmplOrderByLayout = `
    {{if .SortColumns}}
      ORDER BY {{.SortColumns}}
    {{end}}
  `

	tmplWhereLayout = `
    {{if .Conds}}
      WHERE {{.Conds}}
    {{end}}
  `

	tmplUsingLayout = `
    {{if .Columns}}
      USING ({{.Columns}})
    {{end}}
  `

	tmplJoinLayout = `
    {{if .Table}}
      {{ if .On }}
        {{.Type}} JOIN {{.Table}}
        {{.On}}
      {{ else if .Using }}
        {{.Type}} JOIN {{.Table}}
        {{.Using}}
      {{ else if .Type | eq "CROSS" }}
        {{.Type}} JOIN {{.Table}}
      {{else}}
        NATURAL {{.Type}} JOIN {{.Table}}
      {{end}}
    {{end}}
  `

	tmplOnLayout = `
    {{if .Conds}}
      ON {{.Conds}}
    {{end}}
  `

	tmplSelectLayout = `
    SELECT
      {{if .Distinct}}
        DISTINCT
      {{end}}

      {{if defined .Columns}}
        {{.Columns | compile}}
      {{else}}
        *
      {{end}}

      {{if defined .Table}}
        FROM {{.Table | compile}}
      {{end}}

      {{.Joins | compile}}

      {{.Where | compile}}

      {{if defined .GroupBy}}
        {{.GroupBy | compile}}
      {{end}}

      {{.OrderBy | compile}}

      {{if .Limit}}
        LIMIT {{.Limit}}
      {{end}}

      {{if .Offset}}
        OFFSET {{.Offset}}
      {{end}}
  `
	tmplDeleteLayout = `
    DELETE
      FROM {{.Table | compile}}
      {{.Where | compile}}
  `
	tmplUpdateLayout = `
    UPDATE
      {{.Table | compile}}
    SET {{.ColumnValues | compile}}
      {{.Where | compile}}
  `

	tmplSelectCountLayout = `
    SELECT
      COUNT(1) AS _t
    FROM {{.Table | compile}}
      {{.Where | compile}}
  `

	tmplInsertLayout = `
    INSERT INTO {{.Table | compile}}
      {{if defined .Columns}}({{.Columns | compile}}){{end}}
    VALUES
    {{if defined .Values}}
      {{.Values | compile}}
    {{else}}
      (default)
    {{end}}
    {{if defined .Returning}}
      RETURNING {{.Returning | compile}}
    {{end}}
  `

	tmplTruncateLayout = `
    TRUNCATE TABLE {{.Table | compile}} RESTART IDENTITY
  `

	tmplDropDatabaseLayout = `
    DROP DATABASE {{.Database | compile}}
  `

	tmplDropTableLayout = `
    DROP TABLE {{.Table | compile}}
  `

	tmplGroupByLayout = `
    {{if .GroupColumns}}
      GROUP BY {{.GroupColumns}}
    {{end}}
  `
)

func newTemplate() *exql.Template {
	bufferPool := &sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}
	return &exql.Template{
		AndKeyword:          tmplAndKeyword,
		AscKeyword:          tmplAscKeyword,
		AssignmentOperator:  tmplAssignmentOperator,
		ClauseGroup:         tmplClauseGroup,
		ClauseOperator:      tmplClauseOperator,
		ColumnAliasLayout:   tmplColumnAliasLayout,
		ColumnSeparator:     tmplColumnSeparator,
		ColumnValue:         tmplColumnValue,
		CountLayout:         tmplSelectCountLayout,
		DeleteLayout:        tmplDeleteLayout,
		DescKeyword:         tmplDescKeyword,
		DropDatabaseLayout:  tmplDropDatabaseLayout,
		DropTableLayout:     tmplDropTableLayout,
		GroupByLayout:       tmplGroupByLayout,
		IdentifierQuote:     tmplIdentifierQuote,
		IdentifierSeparator: tmplIdentifierSeparator,
		InsertLayout:        tmplInsertLayout,
		JoinLayout:          tmplJoinLayout,
		OnLayout:            tmplOnLayout,
		OrKeyword:           tmplOrKeyword,
		OrderByLayout:       tmplOrderByLayout,
		SelectLayout:        tmplSelectLayout,
		SortByColumnLayout:  tmplSortByColumnLayout,
		TableAliasLayout:    tmplTableAliasLayout,
		TruncateLayout:      tmplTruncateLayout,
		UpdateLayout:        tmplUpdateLayout,
		UsingLayout:         tmplUsingLayout,
		ValueQuote:          tmplValueQuote,
		ValueSeparator:      tmplValueSeparator,
		WhereLayout:         tmplWhereLayout,
		ComparisonOperator: map[expr.ComparisonOperator]string{
			expr.ComparisonRegexp:    "~",
			expr.ComparisonNotRegexp: "!~",
		},
		LRU:        cache.NewLRU(),
		BufferPool: bufferPool,
		FormatSQL: func(sql string) string {
			buf := bufferPool.Get().(*bytes.Buffer)
			defer func() {
				buf.Reset()
				bufferPool.Put(buf)
			}()

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

			out := exql.InvisibleCharsRegexp.ReplaceAll(buf.Bytes(), []byte(` `))
			return string(bytes.TrimSpace(out))
		},
	}
}
