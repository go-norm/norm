// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package exql

import (
	"bytes"
	"fmt"
	"reflect"
	"text/template"

	"github.com/pkg/errors"

	"unknwon.dev/norm/expr"
	"unknwon.dev/norm/internal/cache"
)

// TemplateLayout indicates the type of the template layout.
type TemplateLayout uint8

const (
	LayoutNone = TemplateLayout(iota)

	LayoutAndKeyword
	LayoutAscKeyword
	LayoutAssignmentOperator
	LayoutClauseGroup
	LayoutClauseOperator
	LayoutColumnAlias
	LayoutColumnSeparator
	LayoutColumnValue
	LayoutCount
	LayoutDelete
	LayoutDescKeyword
	LayoutDropDatabase
	LayoutDropTable
	LayoutGroupBy
	LayoutIdentifierQuote
	LayoutIdentifierSeparator
	LayoutInsert
	LayoutJoin
	LayoutOn
	LayoutOrKeyword
	LayoutOrderBy
	LayoutReturning
	LayoutSelect
	LayoutSortByColumn
	LayoutTableAlias
	LayoutTruncate
	LayoutUpdate
	LayoutUsing
	LayoutValueQuote
	LayoutValueSeparator
	LayoutWhere
)

// Template is an SQL template.
type Template struct {
	layouts   map[TemplateLayout]string
	templates map[TemplateLayout]*template.Template
	operators map[expr.ComparisonOperator]string

	*cache.LRU // The cache of compiled SQLs.
}

// NewTemplate initializes a new Template with given layouts and operators, all
// layouts are complied to templates at the time of initialization.
func NewTemplate(layouts map[TemplateLayout]string, operators map[expr.ComparisonOperator]string) (*Template, error) {
	t := &Template{
		layouts:   layouts,
		templates: make(map[TemplateLayout]*template.Template, len(layouts)),
		operators: operators,
		LRU:       cache.NewLRU(),
	}

	funcMap := template.FuncMap{
		"defined": func(f Fragment) bool {
			if f == nil || reflect.ValueOf(f).IsNil() {
				return false
			}
			if e, ok := f.(emptiable); ok && e.Empty() {
				return false
			}
			return true
		},
		"compile": func(f Fragment) (string, error) {
			s, err := t.compile(f)
			if err != nil {
				return "", err
			}
			return s, nil
		},
	}
	for typ, layout := range layouts {
		tmpl, err := template.New("").Funcs(funcMap).Parse(layout)
		if err != nil {
			return nil, errors.Wrapf(err, "parse %v - %q", typ, layout)
		}

		t.templates[typ] = tmpl
	}
	return t, nil
}

// Layout returns the string value of the layout.
func (t *Template) Layout(typ TemplateLayout) string {
	s, ok := t.layouts[typ]
	if !ok {
		return fmt.Sprintf("<undefined layout %v>", typ)
	}
	return s
}

// Operator returns the string value of the comparison operator.
func (t *Template) Operator(op expr.ComparisonOperator) string {
	s, ok := t.operators[op]
	if !ok {
		return fmt.Sprintf("<undefined operator %v>", op)
	}
	return s
}

func (t *Template) compile(f Fragment) (string, error) {
	if f == nil || reflect.ValueOf(f).IsNil() {
		return "", nil
	}
	return f.Compile(t)
}

//go:generate go-mockgen --force unknwon.dev/norm/internal/exql -i emptiable -o mock_emptiable_test.go
type emptiable interface {
	Empty() bool
}

func (t *Template) Compile(layout TemplateLayout, data interface{}) (string, error) {
	tmpl, ok := t.templates[layout]
	if !ok {
		return "", errors.Errorf("no template for layout %v", layout)
	}

	var buf bytes.Buffer
	err := tmpl.Execute(&buf, data)
	if err != nil {
		return "", errors.Wrap(err, "execute")
	}
	return buf.String(), nil
}

// DefaultTemplate returns a template that uses PostgreSQL's syntax.
func DefaultTemplate() (*Template, error) {
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
  COUNT(*)
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
{{if .Columns}}
  GROUP BY {{.Columns}}
{{end}}
`
		defaultIdentifierQuote     = `"{{.}}"`
		defaultIdentifierSeparator = `, `
		defaultInsert              = `
INSERT INTO {{.Table | compile}}
  {{if .Columns }}({{.Columns | compile}}){{end}}
VALUES
  {{.Values | compile}}
{{.Returning | compile}}
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
{{if .Columns}}
  ORDER BY {{.Columns}}
{{end}}
`
		defaultReturning = `
{{if .Columns}}
  RETURNING {{.Columns}}
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
			LayoutReturning:           defaultReturning,
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
			expr.ComparisonEqual:    "=",
			expr.ComparisonNotEqual: "!=",

			expr.ComparisonLessThan:    "<",
			expr.ComparisonGreaterThan: ">",

			expr.ComparisonLessThanOrEqualTo:    "<=",
			expr.ComparisonGreaterThanOrEqualTo: ">=",

			expr.ComparisonBetween:    "BETWEEN",
			expr.ComparisonNotBetween: "NOT BETWEEN",

			expr.ComparisonIn:    "IN",
			expr.ComparisonNotIn: "NOT IN",

			expr.ComparisonIs:    "IS",
			expr.ComparisonIsNot: "IS NOT",

			expr.ComparisonLike:    "LIKE",
			expr.ComparisonNotLike: "NOT LIKE",

			expr.ComparisonRegexp:    "REGEXP",
			expr.ComparisonNotRegexp: "NOT REGEXP",
		},
	)
	if err != nil {
		return nil, errors.Wrap(err, "new template")
	}
	return tmpl, nil
}
