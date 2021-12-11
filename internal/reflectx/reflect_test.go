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

func boolval(v reflect.Value) bool {
	return v.Interface().(bool)
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

func TestNestedStruct(t *testing.T) {
	type (
		Details struct {
			Active bool `db:"active"`
		}
		Asset struct {
			Title   string  `db:"title"`
			Details Details `db:"details"`
		}
		Post struct {
			Author string `db:"author,required"`
			Asset  `db:"asset"`
		}
	)
	z := Post{
		Author: "Joe",
		Asset: Asset{
			Title: "Hello",
			Details: Details{
				Active: true,
			},
		},
	}
	zv := reflect.ValueOf(z)

	m := NewMapper("db")

	v := m.FieldByName(zv, "author")
	assert.Equal(t, z.Author, strval(v))

	v = m.FieldByName(zv, "title")
	_, ok := v.Interface().(string)
	assert.False(t, ok, "title should not exist")

	v = m.FieldByName(zv, "asset.title")
	assert.Equal(t, z.Asset.Title, strval(v))

	v = m.FieldByName(zv, "asset.details.active")
	assert.Equal(t, z.Asset.Details.Active, boolval(v))
}

func TestInlineStruct(t *testing.T) {
	type (
		Employee struct {
			Name string
			ID   int
		}
		Boss   Employee
		person struct {
			Employee `db:"employee"`
			Boss     `db:"boss"`
		}
	)
	z := person{
		Employee: Employee{
			Name: "Joe",
			ID:   2,
		},
		Boss: Boss{
			Name: "Dick",
			ID:   1,
		},
	}
	zv := reflect.ValueOf(z)

	m := NewMapperTagFunc("db", strings.ToLower, nil)

	fields := m.TypeMap(reflect.TypeOf(z))
	assert.Len(t, fields.Index, 6, "number of fields")

	v := m.FieldByName(zv, "employee.name")
	assert.Equal(t, z.Employee.Name, strval(v))

	v = m.FieldByName(zv, "boss.id")
	assert.Equal(t, z.Boss.ID, intval(v))
}

func TestRecursiveStruct(t *testing.T) {
	type Person struct {
		Parent *Person
		Name   string
	}
	z := &Person{
		Parent: &Person{
			Name: "parent",
		},
		Name: "child",
	}
	zv := reflect.ValueOf(z)

	m := NewMapperFunc("db", strings.ToLower)

	v := m.FieldByName(zv, "parent.name")
	assert.Equal(t, z.Parent.Name, strval(v))
}

func TestFieldsEmbedded(t *testing.T) {
	type (
		Person struct {
			Name string `db:"name,size=64"`
		}
		Place struct {
			Name string `db:"name"`
		}
		Article struct {
			Title string `db:"title"`
		}
		PP struct {
			Person  `db:"person,required"`
			Place   `db:",someflag"`
			Article `db:",required"`
		}
	)
	z := PP{
		Person: Person{
			Name: "Peter",
		},
		Place: Place{
			Name: "Toronto",
		},
		Article: Article{
			Title: "Best city ever",
		},
	}
	zv := reflect.ValueOf(z)

	m := NewMapper("db")
	fields := m.TypeMap(reflect.TypeOf(z))

	v := m.FieldByName(zv, "person.name")
	assert.Equal(t, z.Person.Name, strval(v))

	v = m.FieldByName(zv, "name")
	assert.Equal(t, z.Place.Name, strval(v))

	v = m.FieldByName(zv, "title")
	assert.Equal(t, z.Article.Title, strval(v))

	fi := fields.GetByPath("person")
	require.NotNil(t, fi)
	_, ok := fi.Options["required"]
	assert.True(t, ok, "required option")
	assert.True(t, fi.Embedded, "field should be embedded")
	require.Len(t, fi.Index, 1, "length of index")
	assert.Equal(t, 0, fi.Index[0])

	fi = fields.GetByPath("person.name")
	require.NotNil(t, fi)
	assert.Equal(t, "person.name", fi.Path)
	assert.Equal(t, "64", fi.Options["size"])

	fi = fields.GetByTraversal([]int{1, 0})
	require.NotNil(t, fi)
	assert.Equal(t, "name", fi.Path)

	fi = fields.GetByTraversal([]int{2})
	require.NotNil(t, fi)
	_, ok = fi.Options["required"]
	assert.True(t, ok, "required option")

	got := m.TraversalsByName(reflect.TypeOf(z), []string{"person.name", "name", "title"})
	want := [][]int{
		{0, 0},
		{1, 0},
		{2, 0},
	}
	assert.Equal(t, want, got)
}

func TestAnonymousPointerFields(t *testing.T) {
	type (
		Asset struct {
			Title string
		}
		Post struct {
			*Asset `db:"asset"`
			Author string
		}
	)
	z := &Post{
		Author: "Joe",
		Asset: &Asset{
			Title: "Hiyo",
		},
	}
	zv := reflect.ValueOf(z)

	m := NewMapperTagFunc("db", strings.ToLower, nil)
	fields := m.TypeMap(reflect.TypeOf(z))
	assert.Len(t, fields.Index, 3, "number of fields")

	v := m.FieldByName(zv, "asset.title")
	assert.Equal(t, z.Asset.Title, strval(v))

	v = m.FieldByName(zv, "author")
	assert.Equal(t, z.Author, strval(v))
}

func TestNamedPointerFields(t *testing.T) {
	type (
		User struct {
			Name string
		}
		Asset struct {
			Title string

			Owner *User `db:"owner"`
		}
		Post struct {
			Author string

			Asset1 *Asset `db:"asset1"`
			Asset2 *Asset `db:"asset2"`
		}
	)
	z := &Post{
		Author: "Joe",
		Asset1: &Asset{
			Title: "Hiyo",
			Owner: &User{
				Name: "Username",
			},
		},
		// Let Asset2 be nil
	}
	zv := reflect.ValueOf(z)

	m := NewMapperTagFunc("db", strings.ToLower, nil)
	fields := m.TypeMap(reflect.TypeOf(z))
	assert.Len(t, fields.Index, 9, "number of fields")

	v := m.FieldByName(zv, "asset1.title")
	assert.Equal(t, z.Asset1.Title, strval(v))

	v = m.FieldByName(zv, "asset1.owner.name")
	assert.Equal(t, z.Asset1.Owner.Name, strval(v))

	v = m.FieldByName(zv, "asset2.title")
	assert.Equal(t, z.Asset2.Title, strval(v))

	v = m.FieldByName(zv, "asset2.owner.name")
	assert.Equal(t, z.Asset2.Owner.Name, strval(v))

	v = m.FieldByName(zv, "author")
	assert.Equal(t, z.Author, strval(v))
}

func TestFieldMap(t *testing.T) {
	type Foo struct {
		A int
		B int
		C int
	}
	z := Foo{
		A: 1,
		B: 2,
		C: 3,
	}

	m := NewMapperFunc("db", strings.ToLower)
	fm := m.FieldMap(reflect.ValueOf(z))
	assert.Len(t, fm, 3, "number of fields")
	assert.Equal(t, 1, intval(fm["a"]))
	assert.Equal(t, 2, intval(fm["b"]))
	assert.Equal(t, 3, intval(fm["c"]))
}

func TestTagNameMapping(t *testing.T) {
	type Strategy struct {
		StrategyID   string `protobuf:"bytes,1,opt,name=strategy_id" json:"strategy_id,omitempty"`
		StrategyName string
	}
	z := Strategy{
		StrategyID:   "1",
		StrategyName: "Alpah",
	}

	m := NewMapperTagFunc("json", strings.ToUpper, func(value string) string {
		if strings.Contains(value, ",") {
			return strings.Split(value, ",")[0]
		}
		return value
	})
	fm := m.TypeMap(reflect.TypeOf(z))
	for _, key := range []string{"strategy_id", "STRATEGYNAME"} {
		fi := fm.GetByPath(key)
		assert.NotNil(t, fi, "mapping should exist")
	}
}
