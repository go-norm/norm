// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package exql

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"unknwon.dev/norm/expr"
)

func TestTemplate(t *testing.T) {
	tmpl, err := NewTemplate(
		map[TemplateLayout]string{
			LayoutWhere: "{{.}}",
		},
		map[expr.ComparisonOperator]string{
			expr.ComparisonEqual: "=",
		},
	)
	require.NoError(t, err)

	t.Run("layout", func(t *testing.T) {
		got := tmpl.Layout(LayoutWhere)
		assert.Equal(t, "{{.}}", got)

		got = tmpl.Layout(LayoutOn)
		assert.Equal(t, "<undefined layout 19>", got)
	})

	t.Run("operator", func(t *testing.T) {
		got := tmpl.Operator(expr.ComparisonEqual)
		assert.Equal(t, "=", got)

		got = tmpl.Operator(expr.ComparisonGreaterThan)
		assert.Equal(t, "<undefined operator 5>", got)
	})
}

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
	tmpl, err := DefaultTemplate()
	if err != nil {
		t.Fatalf("Failed to get default template: %v", err)
	}
	return tmpl
}
