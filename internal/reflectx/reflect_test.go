// Copyright 2013 Jason Moiron. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package reflectx

import (
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func intval(v reflect.Value) int {
	return v.Interface().(int)
}

func strval(v reflect.Value) string {
	return v.Interface().(string)
}

func TestBasic(t *testing.T) {
	type Foo struct {
		A int
		B int
		C int
	}

	f := Foo{A: 1, B: 2, C: 3}
	fv := reflect.ValueOf(f)
	m := NewMapperFunc("", func(s string) string { return s })

	v := m.FieldByName(fv, "A")
	assert.Equal(t, f.A, intval(v))

	v = m.FieldByName(fv, "B")
	assert.Equal(t, f.B, intval(v))

	v = m.FieldByName(fv, "C")
	assert.Equal(t, f.C, intval(v))
}

func TestBasicEmbedded(t *testing.T) {
	type (
		Foo struct {
			A int
		}
		Bar struct {
			Foo // `db:""` is implied for an embedded struct
			B   int
			C   int `db:"-"`
		}
		Baz struct {
			A   int
			Bar `db:"Bar"`
		}
	)
	z := Baz{
		A: 1,
		Bar: Bar{
			Foo: Foo{
				A: 3,
			},
			B: 2,
			C: 4,
		},
	}
	zv := reflect.ValueOf(z)

	m := NewMapperFunc("db", func(s string) string { return s })
	fields := m.TypeMap(reflect.TypeOf(z))
	assert.Len(t, fields.Index, 5, "number of fields")

	v := m.FieldByName(zv, "A")
	assert.Equal(t, z.A, intval(v))

	v = m.FieldByName(zv, "Bar.B")
	assert.Equal(t, z.B, intval(v))

	v = m.FieldByName(zv, "Bar.A")
	assert.Equal(t, z.Bar.Foo.A, intval(v))

	v = m.FieldByName(zv, "Bar.C")
	_, ok := v.Interface().(int)
	assert.False(t, ok, "Bar.C should not exist")

	fi := fields.GetByPath("Bar.C")
	assert.Nil(t, fi, "Bar.C should not exist")
}

func TestEmbeddedSimple(t *testing.T) {
	type UUID string
	type MyID struct {
		UUID
	}
	type Item struct {
		ID MyID
	}
	z := Item{
		ID: MyID{
			UUID: "6d1f719e-43f5-44e8-a4b1-6e048258829a",
		},
	}
	zv := reflect.ValueOf(z)

	m := NewMapper("db")
	fields := m.TypeMap(reflect.TypeOf(z))
	assert.Len(t, fields.Index, 2, "number of fields")

	v := m.FieldByName(zv, "ID")
	assert.Equal(t, z.ID, v.Interface().(MyID))
}

func TestBasicEmbeddedWithTags(t *testing.T) {
	type (
		Foo struct {
			A int `db:"a"`
		}
		Bar struct {
			Foo     // `db:""` is implied for an embedded struct
			B   int `db:"b"`
		}
		Baz struct {
			A   int `db:"a"`
			Bar     // `db:""` is implied for an embedded struct
		}
	)
	z := Baz{
		A: 1,
		Bar: Bar{
			Foo: Foo{
				A: 3,
			},
			B: 2,
		},
	}
	zv := reflect.ValueOf(z)

	m := NewMapper("db")
	fields := m.TypeMap(reflect.TypeOf(z))
	assert.Len(t, fields.Index, 5, "number of fields")

	v := m.FieldByName(zv, "a")
	assert.Equal(t, z.A, intval(v))

	v = m.FieldByName(zv, "b")
	assert.Equal(t, z.B, intval(v))
}

func TestBasicEmbeddedWithSameName(t *testing.T) {
	type (
		Foo struct {
			A   int `db:"a"`
			Foo int `db:"Foo"` // Same name as the embedded struct
		}
		FooExt struct {
			Foo
			B int `db:"b"`
		}
	)
	z := FooExt{
		Foo: Foo{
			A:   1,
			Foo: 3,
		},
		B: 2,
	}
	zv := reflect.ValueOf(z)

	m := NewMapper("db")
	fields := m.TypeMap(reflect.TypeOf(z))
	assert.Len(t, fields.Index, 4, "number of fields")

	v := m.FieldByName(zv, "a")
	assert.Equal(t, z.A, intval(v)) // the dominant field

	v = m.FieldByName(zv, "b")
	assert.Equal(t, z.B, intval(v))

	v = m.FieldByName(zv, "Foo")
	assert.Equal(t, z.Foo.Foo, intval(v))
}

func TestFlatTags(t *testing.T) {
	type (
		Asset struct {
			Title string `db:"title"`
		}
		Post struct {
			Author string `db:"author,required"`
			Asset  Asset  `db:""`
		}
	)
	z := Post{
		Author: "Joe",
		Asset: Asset{
			Title: "Hello",
		},
	}
	zv := reflect.ValueOf(z)

	m := NewMapper("db")

	v := m.FieldByName(zv, "author")
	assert.Equal(t, z.Author, strval(v))

	v = m.FieldByName(zv, "title")
	assert.Equal(t, z.Asset.Title, strval(v))
}
