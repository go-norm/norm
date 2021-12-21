// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package exql

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsBlankSymbol(t *testing.T) {
	t.Run("yes", func(t *testing.T) {
		for _, c := range []byte(" \n\t\r") {
			assert.True(t, isBlankSymbol(c))
		}
	})

	t.Run("no", func(t *testing.T) {
		for _, c := range []byte("xyz") {
			assert.False(t, isBlankSymbol(c))
		}
	})
}

func TestTrimString(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{
			input: "  \t\nHello World!     \n",
			want:  "Hello World!",
		},
		{
			input: "Nope",
			want:  "Nope",
		},
		{
			input: "",
			want:  "",
		},
		{
			input: " ",
			want:  "",
		},
	}
	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			got := trimString(test.input)
			assert.Equal(t, test.want, got)
		})
	}
}

func TestSeparateByAS(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{
			input: "table.Name AS myTableAlias",
			want:  []string{"table.Name", "myTableAlias"},
		},
		{
			input: "table.Name     AS         myTableAlias",
			want:  []string{"table.Name", "myTableAlias"},
		},
		{
			input: "table.Name\tAS\nmyTableAlias",
			want:  []string{"table.Name", "myTableAlias"},
		},
		{
			input: "a",
			want:  []string{"a"},
		},
		{
			input: "",
			want:  []string{""},
		},
		{
			input: "  A Single Table ",
			want:  []string{"A Single Table"},
		},
		{
			input: "a AS b",
			want:  []string{"a", "b"},
		},
		{
			input: "   a    AS    b ",
			want:  []string{"a", "b"},
		},
		{
			input: "   a    AS    bb ",
			want:  []string{"a", "bb"},
		},
	}
	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			got := separateByAS(test.input)
			assert.Equal(t, test.want, got)
		})
	}
}

func TestSeparateBySpace(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{
			input: "       Hello        World!        Enjoy",
			want:  []string{"Hello", "World!", "Enjoy"},
		},
		{
			input: "",
			want:  []string{""},
		},
	}
	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			got := separateBySpace(test.input)
			assert.Equal(t, test.want, got)
		})
	}
}

func stripWhitespace(s string) string {
	s = InvisibleCharsRegexp.ReplaceAllString(s, ` `)
	s = strings.TrimSpace(s)
	s = strings.NewReplacer("( ", "(", " )", ")").Replace(s)
	return s
}
