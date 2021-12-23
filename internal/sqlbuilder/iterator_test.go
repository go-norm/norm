// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package sqlbuilder

import (
	"context"
	"database/sql"
	"testing"

	mockrequire "github.com/derision-test/go-mockgen/testutil/require"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:generate go-mockgen --force unknwon.dev/norm/adapter -i Adapter -i Executor -i Rows -i Typer -o mock_adapter_test.go
func TestIterator_All(t *testing.T) {
	ctx := context.Background()

	t.Run("nil pointer", func(t *testing.T) {
		var dest map[string]interface{}
		err := newIterator(NewMockAdapter(), NewMockCursor()).All(ctx, dest)
		assert.Error(t, err)
	})

	t.Run("not a pointer", func(t *testing.T) {
		dest := make(map[string]interface{})
		err := newIterator(NewMockAdapter(), NewMockCursor()).All(ctx, dest)
		assert.Error(t, err)
	})

	t.Run("not a slice", func(t *testing.T) {
		dest := make(map[string]interface{})
		err := newIterator(NewMockAdapter(), NewMockCursor()).All(ctx, &dest)
		assert.Error(t, err)
	})

	t.Run("context cancelled", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(ctx, 0)
		defer cancel()

		cursor := NewMockCursor()
		cursor.NextFunc.PushReturn(true)

		dest := make([]map[string]interface{}, 0)
		err := newIterator(NewMockAdapter(), cursor).All(ctx, &dest)
		assert.EqualError(t, err, "context deadline exceeded")
	})

	t.Run("map", func(t *testing.T) {
		// Mock two results
		cursor := NewMockCursor()
		cursor.ColumnsFunc.SetDefaultReturn([]string{"name", "email"}, nil)
		cursor.NextFunc.PushReturn(true)
		cursor.NextFunc.PushReturn(true)
		cursor.ScanFunc.PushHook(func(dest ...interface{}) error {
			assert.Len(t, dest, 2)
			return nil
		})

		iter := newIterator(NewMockAdapter(), cursor)

		dest := make([]map[string]interface{}, 0)
		err := iter.All(ctx, &dest)
		require.NoError(t, err)
		mockrequire.Called(t, cursor.ScanFunc)
	})

	t.Run("struct", func(t *testing.T) {
		// Mock two results
		cursor := NewMockCursor()
		cursor.ColumnsFunc.SetDefaultReturn([]string{"name", "email"}, nil)
		cursor.NextFunc.PushReturn(true)
		cursor.NextFunc.PushReturn(true)
		cursor.ScanFunc.PushHook(func(dest ...interface{}) error {
			assert.Len(t, dest, 2)
			return nil
		})

		iter := newIterator(NewMockAdapter(), cursor)

		dest := make([]struct{}, 0)
		err := iter.All(ctx, &dest)
		require.NoError(t, err)
		mockrequire.Called(t, cursor.ScanFunc)
	})
}

func TestIterator_One(t *testing.T) {
	ctx := context.Background()

	t.Run("nil pointer", func(t *testing.T) {
		var dest map[string]interface{}
		err := newIterator(NewMockAdapter(), NewMockCursor()).One(ctx, dest)
		assert.Error(t, err)
	})

	t.Run("not a pointer", func(t *testing.T) {
		dest := make(map[string]interface{})
		err := newIterator(NewMockAdapter(), NewMockCursor()).One(ctx, dest)
		assert.Error(t, err)
	})

	t.Run("no rows", func(t *testing.T) {
		dest := make([]map[string]interface{}, 0)
		err := newIterator(NewMockAdapter(), NewMockCursor()).One(ctx, &dest)
		assert.EqualError(t, err, sql.ErrNoRows.Error())
	})

	t.Run("map", func(t *testing.T) {
		cursor := NewMockCursor()
		cursor.ColumnsFunc.SetDefaultReturn([]string{"name", "email"}, nil)
		cursor.NextFunc.PushReturn(true)
		cursor.ScanFunc.PushHook(func(dest ...interface{}) error {
			assert.Len(t, dest, 2)
			return nil
		})

		iter := newIterator(NewMockAdapter(), cursor)

		dest := make(map[string]interface{})
		err := iter.One(ctx, &dest)
		require.NoError(t, err)
		mockrequire.Called(t, cursor.ScanFunc)
	})

	t.Run("struct", func(t *testing.T) {
		adapter := NewMockAdapter()
		adapter.TyperFunc.SetDefaultReturn(NewMockTyper())

		cursor := NewMockCursor()
		cursor.ColumnsFunc.SetDefaultReturn([]string{"name", "email"}, nil)
		cursor.NextFunc.PushReturn(true)
		cursor.ScanFunc.PushHook(func(dest ...interface{}) error {
			assert.Len(t, dest, 2)
			return nil
		})

		iter := newIterator(adapter, cursor)

		dest := new(struct {
			Name string `db:"name"`
		})
		err := iter.One(ctx, dest)
		require.NoError(t, err)
		mockrequire.Called(t, cursor.ScanFunc)
	})
}
