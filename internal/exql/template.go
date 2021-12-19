// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package exql

import (
	"bytes"
	"reflect"
	"text/template"

	"github.com/pkg/errors"

	"unknwon.dev/norm/expr"
	"unknwon.dev/norm/internal/cache"
)

// TemplateLayout indicates the type of the template layout.
type TemplateLayout uint8

const (
	_ = TemplateLayout(iota)

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
