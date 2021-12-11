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
	fields := m.TypeMap(reflect.TypeOf(z))
	for _, key := range []string{"strategy_id", "STRATEGYNAME"} {
		fi := fields.GetByPath(key)
		assert.NotNil(t, fi, "mapping should exist")
	}
}

func TestMapping(t *testing.T) {
	type Person struct {
		ID           int
		Name         string
		WearsGlasses bool `db:"wears_glasses"`
	}
	z1 := Person{
		ID:           1,
		Name:         "Jason",
		WearsGlasses: true,
	}

	m := NewMapperFunc("db", strings.ToLower)
	fields := m.TypeMap(reflect.TypeOf(z1))
	for _, key := range []string{"id", "name", "wears_glasses"} {
		fi := fields.GetByPath(key)
		assert.NotNil(t, fi, "mapping should exist")
	}

	type SportsPerson struct {
		Weight int
		Age    int
		Person
	}
	z2 := SportsPerson{
		Weight: 100,
		Age:    30,
		Person: z1,
	}
	fields = m.TypeMap(reflect.TypeOf(z2))
	for _, key := range []string{"id", "name", "wears_glasses", "weight", "age"} {
		fi := fields.GetByPath(key)
		assert.NotNil(t, fi, "mapping should exist")
	}

	type RugbyPlayer struct {
		Position   int
		IsIntense  bool `db:"is_intense"`
		IsAllBlack bool `db:"-"`
		SportsPerson
	}
	z3 := RugbyPlayer{
		Position:     12,
		IsIntense:    true,
		SportsPerson: z2,
	}
	fields = m.TypeMap(reflect.TypeOf(z3))
	for _, key := range []string{"id", "name", "wears_glasses", "weight", "age", "position", "is_intense"} {
		fi := fields.GetByPath(key)
		assert.NotNil(t, fi, "mapping should exist")
	}

	fi := fields.GetByPath("isallblack")
	assert.Nil(t, fi, "mapping should be ignored")
}

func TestGetByTraversal(t *testing.T) {
	type (
		C struct {
			C0 int
			C1 int
		}
		B struct {
			B0 string
			B1 *C
		}
		A struct {
			A0 int
			A1 B
		}
	)

	m := NewMapperFunc("db", func(n string) string { return n })
	fields := m.TypeMap(reflect.TypeOf(A{}))

	tests := []struct {
		index    []int
		wantName string
		wantNil  bool
	}{
		{
			index:    []int{0},
			wantName: "A0",
		},
		{
			index:    []int{1, 0},
			wantName: "B0",
		},
		{
			index:    []int{1, 1, 1},
			wantName: "C1",
		},
		{
			index:   []int{3, 4, 5},
			wantNil: true,
		},
		{
			index:   []int{},
			wantNil: true,
		},
		{
			index:   nil,
			wantNil: true,
		},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			fi := fields.GetByTraversal(test.index)
			if test.wantNil {
				assert.Nil(t, fi)
				return
			}
			require.NotNil(t, fi)
			assert.Equal(t, test.wantName, fi.Name)
		})
	}
}

func TestMapperMethodsByName(t *testing.T) {
	type (
		C struct {
			C0 string
			C1 int
		}
		B struct {
			B0 *C     `db:"B0"`
			B1 C      `db:"B1"`
			B2 string `db:"B2"`
		}
		A struct {
			A0 *B `db:"A0"`
			B  `db:"A1"`
			A2 int
			a3 int //nolint:unused,structcheck
		}
	)
	z := &A{
		A0: &B{
			B0: &C{
				C0: "0",
				C1: 1,
			},
			B1: C{
				C0: "2",
				C1: 3,
			},
			B2: "4",
		},
		B: B{
			B0: nil,
			B1: C{
				C0: "5",
				C1: 6,
			},
			B2: "7",
		},
		A2: 8,
	}
	zv := reflect.ValueOf(z)

	m := NewMapperFunc("db", func(n string) string { return n })

	tests := []struct {
		name        string
		wantInvalid bool
		wantValue   interface{}
		wantIndexes []int
	}{
		{
			name:        "A0.B0.C0",
			wantValue:   "0",
			wantIndexes: []int{0, 0, 0},
		},
		{
			name:        "A0.B0.C1",
			wantValue:   1,
			wantIndexes: []int{0, 0, 1},
		},
		{
			name:        "A0.B1.C0",
			wantValue:   "2",
			wantIndexes: []int{0, 1, 0},
		},
		{
			name:        "A0.B1.C1",
			wantValue:   3,
			wantIndexes: []int{0, 1, 1},
		},
		{
			name:        "A0.B2",
			wantValue:   "4",
			wantIndexes: []int{0, 2},
		},
		{
			name:        "A1.B0.C0",
			wantValue:   "",
			wantIndexes: []int{1, 0, 0},
		},
		{
			name:        "A1.B0.C1",
			wantValue:   0,
			wantIndexes: []int{1, 0, 1},
		},
		{
			name:        "A1.B1.C0",
			wantValue:   "5",
			wantIndexes: []int{1, 1, 0},
		},
		{
			name:        "A1.B1.C1",
			wantValue:   6,
			wantIndexes: []int{1, 1, 1},
		},
		{
			name:        "A1.B2",
			wantValue:   "7",
			wantIndexes: []int{1, 2},
		},
		{
			name:        "A2",
			wantValue:   8,
			wantIndexes: []int{2},
		},
		{
			name:        "XYZ",
			wantInvalid: true,
			wantIndexes: []int{},
		},
		{
			name:        "a3",
			wantInvalid: true,
			wantIndexes: []int{},
		},
	}

	// Build the names array from the test cases
	names := make([]string, len(tests))
	for i, tc := range tests {
		names[i] = tc.name
	}
	values := m.FieldsByName(zv, names)
	require.Equal(t, len(tests), len(values))

	indexes := m.TraversalsByName(zv.Type(), names)
	require.Equal(t, len(tests), len(indexes))

	for i, val := range values {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			gotIndexes := indexes[i]
			require.Equal(t, test.wantIndexes, gotIndexes)

			val = reflect.Indirect(val)
			if test.wantInvalid {
				assert.False(t, val.IsValid(), "should be invalid")
				return
			}
			require.True(t, val.IsValid(), "should be valid")

			gotValue := reflect.Indirect(val).Interface()
			assert.Equal(t, test.wantValue, gotValue)
		})
	}
}

func TestFieldByIndexes(t *testing.T) {
	type (
		C struct {
			C0 bool
			C1 string
			C2 int
			C3 map[string]int
		}
		B struct {
			B1 C
			B2 *C
		}
		A struct {
			A1 B
			A2 *B
		}
	)
	tests := []struct {
		value     interface{}
		indexes   []int
		wantValue interface{}
		readOnly  bool
	}{
		{
			value: A{
				A1: B{B1: C{C0: true}},
			},
			indexes:   []int{0, 0, 0},
			wantValue: true,
			readOnly:  true,
		},
		{
			value: A{
				A2: &B{B2: &C{C1: "answer"}},
			},
			indexes:   []int{1, 1, 1},
			wantValue: "answer",
			readOnly:  true,
		},
		{
			value:     &A{},
			indexes:   []int{1, 1, 3},
			wantValue: map[string]int{},
		},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			v := FieldByIndexes(reflect.ValueOf(test.value), test.indexes)
			if test.wantValue == nil {
				assert.Nil(t, v.IsNil())
			} else {
				assert.Equal(t, test.wantValue, v.Interface())
			}

			if test.readOnly {
				v := FieldByIndexesReadOnly(reflect.ValueOf(test.value), test.indexes)
				if test.wantValue == nil {
					assert.Nil(t, v.IsNil())
				} else {
					assert.Equal(t, test.wantValue, v.Interface())
				}
			}
		})
	}
}

func TestMustBe(t *testing.T) {
	typ := reflect.TypeOf(E1{})
	mustBe(typ, reflect.Struct)

	defer func() {
		r := recover()
		require.NotNil(t, r, "should panic")

		valueErr, ok := r.(*reflect.ValueError)
		require.True(t, ok, "should panic with *reflect.ValueError")
		assert.Equal(t, "unknwon.dev/norm/internal/reflectx.TestMustBe", valueErr.Method)
		assert.Equal(t, reflect.String, valueErr.Kind)
	}()

	typ = reflect.TypeOf("string")
	mustBe(typ, reflect.Struct)
}
}
