// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package exql

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
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
