// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package sqlbuilder

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"unknwon.dev/norm/expr"
)

func TestPlaceholderSimple(t *testing.T) {
	{
		ret, _, _ := ExpandQuery("?", []interface{}{1})
		assert.Equal(t, "?", ret)
	}
	{
		ret, _, _ := ExpandQuery("?", nil)
		assert.Equal(t, "?", ret)
	}
}

func TestPlaceholderMany(t *testing.T) {
	{
		ret, _, _ := ExpandQuery("?, ?, ?", []interface{}{1, 2, 3})
		assert.Equal(t, "?, ?, ?", ret)
	}
}

func TestPlaceholderArray(t *testing.T) {
	{
		ret, _, _ := ExpandQuery("?, ?, ?", []interface{}{1, 2, []interface{}{3, 4, 5}})
		assert.Equal(t, "?, ?, (?, ?, ?)", ret)
	}

	{
		ret, _, _ := ExpandQuery("?, ?, ?", []interface{}{[]interface{}{1, 2, 3}, 4, 5})
		assert.Equal(t, "(?, ?, ?), ?, ?", ret)
	}

	{
		ret, _, _ := ExpandQuery("?, ?, ?", []interface{}{1, []interface{}{2, 3, 4}, 5})
		assert.Equal(t, "?, (?, ?, ?), ?", ret)
	}

	{
		ret, _, _ := ExpandQuery("???", []interface{}{1, []interface{}{2, 3, 4}, 5})
		assert.Equal(t, "?(?, ?, ?)?", ret)
	}

	{
		ret, _, _ := ExpandQuery("??", []interface{}{[]interface{}{1, 2, 3}, []interface{}{}, []interface{}{4, 5}, []interface{}{}})
		assert.Equal(t, "(?, ?, ?)(NULL)", ret)
	}
}

func TestPlaceholderArguments(t *testing.T) {
	{
		_, args, _ := ExpandQuery("?, ?, ?", []interface{}{1, 2, []interface{}{3, 4, 5}})
		assert.Equal(t, []interface{}{1, 2, 3, 4, 5}, args)
	}

	{
		_, args, _ := ExpandQuery("?, ?, ?", []interface{}{1, []interface{}{2, 3, 4}, 5})
		assert.Equal(t, []interface{}{1, 2, 3, 4, 5}, args)
	}

	{
		_, args, _ := ExpandQuery("?, ?, ?", []interface{}{[]interface{}{1, 2, 3}, 4, 5})
		assert.Equal(t, []interface{}{1, 2, 3, 4, 5}, args)
	}

	{
		_, args, _ := ExpandQuery("?, ?", []interface{}{[]interface{}{1, 2, 3}, []interface{}{4, 5}})
		assert.Equal(t, []interface{}{1, 2, 3, 4, 5}, args)
	}
}

func TestPlaceholderReplace(t *testing.T) {
	{
		ret, args, _ := ExpandQuery("?, ?, ?", []interface{}{1, expr.Raw("foo"), 3})
		assert.Equal(t, "?, foo, ?", ret)
		assert.Equal(t, []interface{}{1, 3}, args)
	}
}
