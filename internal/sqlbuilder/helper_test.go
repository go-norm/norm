// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package sqlbuilder

import (
	"testing"

	"unknwon.dev/norm/internal/exql"
)

func defaultTemplate(t testing.TB) *exql.Template {
	tmpl, err := exql.DefaultTemplate()
	if err != nil {
		t.Fatalf("Failed to get default template: %v", err)
	}
	return tmpl
}
