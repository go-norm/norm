// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package exql

import (
	"bytes"
	"reflect"
	"regexp"
	"sync"
	"text/template"

	"unknwon.dev/norm/expr"
	"unknwon.dev/norm/internal/cache"
)

// Type is the type of SQL query the statement represents.
type Type uint

// Values for Type.
const (
	_ = Type(iota)

	Truncate
	DropTable
	DropDatabase
	Count
	Insert
	Select
	Update
	Delete

	SQL
)

type (
	// Limit represents the SQL limit in a query.
	Limit int
	// Offset represents the SQL offset in a query.
	Offset int
)

// Template is an SQL template.
type Template struct {
	AndKeyword          string
	AscKeyword          string
	AssignmentOperator  string
	ClauseGroup         string
	ClauseOperator      string
	ColumnAliasLayout   string
	ColumnSeparator     string
	ColumnValue         string
	CountLayout         string
	DeleteLayout        string
	DescKeyword         string
	DropDatabaseLayout  string
	DropTableLayout     string
	GroupByLayout       string
	IdentifierQuote     string
	IdentifierSeparator string
	InsertLayout        string
	JoinLayout          string
	OnLayout            string
	OrKeyword           string
	OrderByLayout       string
	SelectLayout        string
	SortByColumnLayout  string
	TableAliasLayout    string
	TruncateLayout      string
	UpdateLayout        string
	UsingLayout         string
	ValueQuote          string
	ValueSeparator      string
	WhereLayout         string

	ComparisonOperator map[expr.ComparisonOperator]string

	templateMutex sync.RWMutex
	templateMap   map[string]*template.Template

	*cache.LRU
	BufferPool *sync.Pool

	FormatSQL func(sql string) string
}

func (t *Template) MustCompile(templateText string, data interface{}) string {
	buf := t.BufferPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		t.BufferPool.Put(buf)
	}()

	v, ok := t.getTemplate(templateText)
	if !ok {
		v = template.
			Must(template.New("").
				Funcs(map[string]interface{}{
					"defined": func(in Fragment) bool {
						if in == nil || reflect.ValueOf(in).IsNil() {
							return false
						}
						if check, ok := in.(hasIsEmpty); ok {
							if check.IsEmpty() {
								return false
							}
						}
						return true
					},
					"compile": func(in Fragment) (string, error) {
						s, err := t.doCompile(in)
						if err != nil {
							return "", err
						}
						return s, nil
					},
				}).
				Parse(templateText))

		t.setTemplate(templateText, v)
	}

	if err := v.Execute(buf, data); err != nil {
		panic("There was an error compiling the following template:\n" + templateText + "\nError was: " + err.Error())
	}

	return buf.String()
}

func (t *Template) getTemplate(k string) (*template.Template, bool) {
	t.templateMutex.RLock()
	defer t.templateMutex.RUnlock()

	// todo: this could cause race conditions
	if t.templateMap == nil {
		t.templateMap = make(map[string]*template.Template)
	}

	v, ok := t.templateMap[k]
	return v, ok
}

func (t *Template) setTemplate(k string, v *template.Template) {
	t.templateMutex.Lock()
	defer t.templateMutex.Unlock()

	t.templateMap[k] = v
}

var InvisibleCharsRegexp = regexp.MustCompile(`[\s\r\n\t]+`)
