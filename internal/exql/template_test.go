// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package exql

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"unknwon.dev/norm/expr"
)

func TestNewTemplate(t *testing.T) {
	t.Run("bad layout", func(t *testing.T) {
		_, err := NewTemplate(
			map[TemplateLayout]string{
				LayoutWhere: "{{",
			},
			nil,
		)
		assert.Error(t, err)
	})

	t.Run("good layout", func(t *testing.T) {
		_, err := NewTemplate(
			map[TemplateLayout]string{
				LayoutWhere: `
{{if .Conds}}
  WHERE {{.Conds}}
{{end}}
`,
			},
			nil,
		)
		assert.NoError(t, err)
	})
}

func TestTemplate_Compile(t *testing.T) {
	t.Run("no such template", func(t *testing.T) {
		tmpl, err := NewTemplate(
			map[TemplateLayout]string{
				LayoutWhere: "",
			},
			nil,
		)
		assert.NoError(t, err)

		_, err = tmpl.Compile(LayoutAndKeyword, nil)
		assert.Error(t, err)
	})

	tests := []struct {
		name string
		data func() interface{}
		want string
	}{
		{
			name: "normal",
			data: func() interface{} {
				f := NewMockFragment()
				f.CompileFunc.SetDefaultReturn("test string", nil)
				return map[string]Fragment{
					"Columns": f,
				}
			},
			want: "test string",
		},

		{
			name: "nil fragment",
			data: func() interface{} {
				return map[string]Fragment{
					"Columns": (Fragment)(nil),
				}
			},
			want: "*",
		},
		{
			name: "empty fragment",
			data: func() interface{} {
				e := NewMockEmptiable()
				e.EmptyFunc.SetDefaultReturn(true)
				ef := struct {
					Fragment
					emptiable
				}{
					Fragment:  NewMockFragment(),
					emptiable: e,
				}
				return map[string]Fragment{
					"Columns": &ef,
				}
			},
			want: "*",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tmpl, err := NewTemplate(
				map[TemplateLayout]string{
					LayoutWhere: `
{{if defined .Columns}}
  {{.Columns | compile}}
{{else}}
  *
{{end}}
`,
				},
				nil,
			)
			assert.NoError(t, err)

			got, err := tmpl.Compile(LayoutWhere, test.data())
			assert.NoError(t, err)
			assert.Equal(t, test.want, strings.TrimSpace(got))
		})
	}
}

func defaultTemplate(t testing.TB) *Template {
	const (
		defaultAndKeyword         = `AND`
		defaultAscKeyword         = `ASC`
		defaultAssignmentOperator = `=`
		defaultClauseGroup        = `({{.}})`
		defaultClauseOperator     = ` {{.}} `
		defaultColumnAlias        = `{{.Name}}{{if .Alias}} AS {{.Alias}}{{end}}`
		defaultColumnSeparator    = `.`
		defaultColumnValue        = `{{.Column}} {{.Operator}} {{.Value}}`
		defaultCount              = `
SELECT
  COUNT(1) AS _t
FROM {{.Table | compile}}
  {{.Where | compile}}

  {{if .Limit}}
	LIMIT {{.Limit | compile}}
  {{end}}

  {{if .Offset}}
	OFFSET {{.Offset}}
  {{end}}
  `
		defaultDelete = `
DELETE
  FROM {{.Table | compile}}
  {{.Where | compile}}
{{if .Limit}}
  LIMIT {{.Limit}}
{{end}}
{{if .Offset}}
  OFFSET {{.Offset}}
{{end}}
  `
		defaultDescKeyword  = `DESC`
		defaultDropDatabase = `DROP DATABASE {{.Database | compile}}`
		defaultDropTable    = `DROP TABLE {{.Table | compile}}`
		defaultGroupBy      = `
    {{if .GroupColumns}}
      GROUP BY {{.GroupColumns}}
    {{end}}
  `
		defaultIdentifierQuote     = `"{{.}}"`
		defaultIdentifierSeparator = `, `
		defaultInsert              = `
INSERT INTO {{.Table | compile}}
  {{if .Columns }}({{.Columns | compile}}){{end}}
VALUES
  {{.Values | compile}}
{{if .Returning}}
  RETURNING {{.Returning | compile}}
{{end}}
  `
		defaultJoin = `
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
		defaultOn = `
{{if .Conds}}
  ON {{.Conds}}
{{end}}
  `
		defaultOrKeyword = `OR`
		defaultOrderBy   = `
{{if .SortColumns}}
  ORDER BY {{.SortColumns}}
{{end}}
  `
		defaultSelect = `
SELECT
  {{if .Distinct}}
	DISTINCT
  {{end}}

  {{if .Columns}}
	{{.Columns | compile}}
  {{else}}
	*
  {{end}}

  {{if defined .Table}}
	FROM {{.Table | compile}}
  {{end}}

  {{.Joins | compile}}

  {{.Where | compile}}

  {{.GroupBy | compile}}

  {{.OrderBy | compile}}

  {{if .Limit}}
	LIMIT {{.Limit}}
  {{end}}

  {{if .Offset}}
	OFFSET {{.Offset}}
  {{end}}
  `
		defaultSortByColumn = `{{.Column}} {{.Order}}`
		defaultTableAlias   = `{{.Name}}{{if .Alias}} AS {{.Alias}}{{end}}`
		defaultTruncate     = `TRUNCATE TABLE {{.Table | compile}}`
		defaultUpdate       = `
UPDATE
  {{.Table | compile}}
SET {{.ColumnValues | compile}}
  {{.Where | compile}}
  `
		defaultUsing = `
{{if .Columns}}
  USING ({{.Columns}})
{{end}}
  `
		defaultValueQuote     = `'{{.}}'`
		defaultValueSeparator = `, `
		defaultWhere          = `
{{if .Conds}}
  WHERE {{.Conds}}
{{end}}
  `
	)

	tmpl, err := NewTemplate(
		map[TemplateLayout]string{
			LayoutAndKeyword:          defaultAndKeyword,
			LayoutAscKeyword:          defaultAscKeyword,
			LayoutAssignmentOperator:  defaultAssignmentOperator,
			LayoutClauseGroup:         defaultClauseGroup,
			LayoutClauseOperator:      defaultClauseOperator,
			LayoutColumnAlias:         defaultColumnAlias,
			LayoutColumnSeparator:     defaultColumnSeparator,
			LayoutColumnValue:         defaultColumnValue,
			LayoutCount:               defaultCount,
			LayoutDelete:              defaultDelete,
			LayoutDescKeyword:         defaultDescKeyword,
			LayoutDropDatabase:        defaultDropDatabase,
			LayoutDropTable:           defaultDropTable,
			LayoutGroupBy:             defaultGroupBy,
			LayoutIdentifierQuote:     defaultIdentifierQuote,
			LayoutIdentifierSeparator: defaultIdentifierSeparator,
			LayoutInsert:              defaultInsert,
			LayoutJoin:                defaultJoin,
			LayoutOn:                  defaultOn,
			LayoutOrKeyword:           defaultOrKeyword,
			LayoutOrderBy:             defaultOrderBy,
			LayoutSelect:              defaultSelect,
			LayoutSortByColumn:        defaultSortByColumn,
			LayoutTableAlias:          defaultTableAlias,
			LayoutTruncate:            defaultTruncate,
			LayoutUpdate:              defaultUpdate,
			LayoutUsing:               defaultUsing,
			LayoutValueQuote:          defaultValueQuote,
			LayoutValueSeparator:      defaultValueSeparator,
			LayoutWhere:               defaultWhere,
		},
		map[expr.ComparisonOperator]string{
			expr.ComparisonEqual:                "=",
			expr.ComparisonNotEqual:             "!=",
			expr.ComparisonLessThan:             "<",
			expr.ComparisonGreaterThan:          ">",
			expr.ComparisonLessThanOrEqualTo:    "<=",
			expr.ComparisonGreaterThanOrEqualTo: ">=",
		},
	)
	if err != nil {
		t.Fatalf("Failed to create new template: %v", err)
	}
	return tmpl
}
