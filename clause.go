// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package norm

import (
	"context"
)

// Selector represents a SQL query builder for the SELECT statement.
type Selector interface {
	// Columns defines which columns to retrieve.
	//
	// The From() should be called along with Columns() to query data from a
	// specific table:
	//
	//   q.Columns("name", "last_name").From("users")
	//
	// Aliases may be used for the columns using the "AS" keyword:
	//
	//   q.Columns("name AS n")
	//
	// or using the shortcut:
	//
	//   q.Columns("name n")
	//
	// Use the `expr.Raw` to prevent column expressions to be escaped:
	//
	//   q.Columns(expr.Raw("MAX(id)"))
	//
	// The above statement is equivalent to:
	//
	//   q.Columns(expr.Func("MAX", "id"))
	//
	// Subsequent calls to Columns() append more columns to the retrieval list (i.e.
	// do not replace previously set columns).
	Columns(columns ...interface{}) Selector
	// From constructs the FROM clause for where the data to be retrieved from.
	//
	// It is typically used along with Columns():
	//
	//   q.Columns(...).From("users")
	//
	// It is also possible to use an alias for the table:
	//
	//   q.Columns(...).From("users AS u").Where("u.name = ?", ...)
	//
	// or using the shortcut:
	//
	//   q.Columns(...).From("users u").Where("u.name = ?", ...)
	From(tables ...interface{}) Selector
	// Distinct constructs the DISTINCT clause with given columns.
	//
	// If no column is given, the DISTINCT applies to all columns:
	//
	//   => SELECT "name", DISTINCT("email", "gender")
	//   q.Select("name").Distinct("email", "gender")
	//
	//   => SELECT DISTINCT "name", "email"
	//   q.Select("name").Distinct().Columns("email")
	Distinct(columns ...string) Selector
	// As constructs an alias for the table.
	//
	// It can only be called after From() which defines the table:
	//
	//   q.From("users").As("u")
	As(alias string) Selector

	// Where constructs the WHERE clause to specify the conditions that columns must
	// match in order to be retrieved.
	//
	// It accepts raw strings and `fmt.Stringer` to define conditions and uses
	// `interface{}` to specify bind parameters.
	//
	// CAUTION: It is possible to cause SQL injection when embedding values of bind
	// parameters to the SQL query string. The placeholder question mark ("?")
	// should be used instead:
	//
	//   q.Where("name = ?", "max")
	//   q.Where("name = ? AND last_name = ?", "Mary", "Doe")
	//   q.Where("last_name IS NULL")
	//   q.Where("online = ? AND last_logged <= ?", true, time.Now())
	//
	// Subsequent calls to Where() replace previously set conditions, use And()
	// instead for condition conjunctions.
	Where(conds ...interface{}) Selector
	// And appends more conditions to the WHERE clause.
	//
	// It can be called regardless of Where():
	//
	//   q.Where("name = ?", "max").And("last_name = ?", "Doe")
	//   q.And("last_name = ?", "Doe")
	And(conds ...interface{}) Selector
	// GroupBy constructs the GROUP BY clause.
	//
	// It defines which columns should be used to aggregate and group results:
	//
	//   q.GroupBy("country_id")
	//   q.GroupBy("country_id", "city_id")
	//
	// Subsequent calls to GroupBy() replace the previously set clause.
	GroupBy(columns ...interface{}) Selector
	// OrderBy constructs the ORDER BY clause.
	//
	// It is used to define which columns are going to be used to sort results, and
	// results are in ascending order by default:
	//
	//   => ORDER BY "last_name" ASC
	//   q.OrderBy("last_name")
	//   q.OrderBy("last_name ASC")
	//
	//   => ORDER BY "last_name" DESC, "name" ASC
	//   q.OrderBy("last_name DESC", "name ASC")
	//
	// Prefix the column name with the minus sign ("-") to sort results in
	// descending order:
	//
	//   => ORDER BY "last_name" DESC
	//   q.OrderBy("-last_name")
	//
	// Subsequent calls to OrderBy() replace the previously set clause.
	OrderBy(columns ...interface{}) Selector
	// Join constructs a JOIN clause.
	//
	// JOIN statements are used to define external tables to be included as part of
	// the result:
	//
	//   q.Join("authors")
	//
	// Use the On() to specify the JOIN conditions:
	//
	//   q.Join("authors").On("authors.id = books.author_id")
	//
	//
	// Use the Using() to specify the JOIN columns:
	//
	//   q.Join("employees").Using("department_id")
	//
	// The NATURAL JOIN is used when no conditions specified for the join.
	Join(table interface{}) Selector
	// FullJoin is similar to Join() but is for FULL JOIN.
	FullJoin(table interface{}) Selector
	// CrossJoin is similar to Join() but is for CROSS JOIN.
	CrossJoin(table interface{}) Selector
	// RightJoin is similar to Join() but is for RIGHT JOIN.
	RightJoin(table interface{}) Selector
	// LeftJoin is similar to Join() but is for LEFT JOIN.
	LeftJoin(table interface{}) Selector
	// On constructs a ON clause.
	//
	// It can only be called after any type of join method and accepts the same
	// types of arguments as Where():
	//
	//   q.Join(...).On("b.author_id = a.id")
	On(conds ...interface{}) Selector
	// Using constructs a USING clause.
	//
	// It is used to specify columns to join results and can only be called after
	// any type of join method:
	//
	//   q.LeftJoin(...).Using("country_id")
	Using(columns ...interface{}) Selector

	// Limit constructs the LIMIT clause.
	//
	// It is used to define the maximum number of rows to be returned from the
	// query:
	//
	//   q.Limit(42)
	//
	// A negative limit cancels any previously set limit.
	Limit(n int) Selector
	// Offset constructs the OFFSET clause.
	//
	// It is used to define how many results are going to be skipped before starting to
	// return results.
	//
	//   q.Offset(56)
	//
	// A negative offset cancels any previously set offset.
	Offset(n int) Selector

	// Amend alters the query string just before executing it.
	Amend(func(query string) string) Selector

	// Iterate creates an Iterator to iterate over query results.
	Iterate(ctx context.Context) Iterator
	ResultMapper

	// String returns a complied SQL query string.
	String() string
	// Arguments returns the arguments that are prepared for this query.
	Arguments() []interface{}
}

// Inserter represents a SQL query builder for the INSERT statement.
type Inserter interface {
	// Columns defines which columns that we are going to provide values for.
	//
	// The Values() should be called along with Columns() to provide actual values
	// to be inserted:
	//
	//   q.Columns("first_name", "last_name", "age").Values("María", "Méndez", 18)
	Columns(columns ...interface{}) Inserter
	// Values constructs the VALUES clause for values to be inserted as designated
	// columns.
	//
	// Example:
	//
	//   q.Columns("first_name", "last_name", "age").Values("María", "Méndez", 18)
	Values(values ...interface{}) Inserter

	// Returning constructs the RETURNING clause to specify which columns should be
	// returned upon successful insertion.
	Returning(columns ...interface{}) Inserter

	// Amend alters the query string just before executing it.
	Amend(func(query string) string) Inserter

	// Iterate creates an Iterator to iterate over query results. This is only
	// possible when using Returning().
	Iterate(ctx context.Context) Iterator
	ResultMapper

	// String returns a complied SQL query string.
	String() string
	// Arguments returns the arguments that are prepared for this query.
	Arguments() []interface{}
}

// Updater represents a SQL query builder for the UPDATE statement.
type Updater interface {
	// Set constructs the SET clause with pairs of key names and values.
	//
	// Examples:
	//
	//   q.Set("name", "John", "last_name", "Smith").Set("age", 18)
	Set(kvs ...interface{}) Updater

	// Where constructs the WHERE clause.
	//
	// See Selector.Where for documentation and usage examples.
	Where(conds ...interface{}) Updater
	// And appends more conditions to the WHERE clause.
	//
	// See Selector.And for documentation and usage examples.
	And(conds ...interface{}) Updater

	// Returning constructs the RETURNING clause to specify which columns should be
	// returned upon successful update.
	Returning(columns ...interface{}) Updater

	// Amend alters the query string just before executing it.
	Amend(func(query string) string) Updater

	// Iterate creates an Iterator to iterate over query results. This is only
	// possible when using Returning().
	Iterate(ctx context.Context) Iterator
	ResultMapper

	// String returns a complied SQL query string.
	String() string
	// Arguments returns the arguments that are prepared for this query.
	Arguments() []interface{}
}

// Deleter represents a SQL query builder for the DELETE statement.
type Deleter interface {
}

// Iterator defines a collection of methods to iterate over query results.
type Iterator interface {
	ResultMapper
}

// ResultMapper defines a collection of methods to map query results to
// destination objects.
type ResultMapper interface {
	// All dumps all the results into the destination, and expects a pointer to the
	// slice of maps or structs.
	//
	// The behaviour of One() is extended to each one of the results.
	All(ctx context.Context, destSlice interface{}) error

	// One maps the row that is in the current query cursor into the destination,
	// which can be a pointer to a map or a struct. The sql.ErrNoRows is returned
	// when there is no result to scan.
	//
	// If dest is a pointer to a map, each column creates a new map key and the
	// results are set as values of the keys. Depending on the type of map key and
	// value, the results columns and values may need to be transformed.
	//
	// If dest is a pointer to a struct, each one field will be tested for a `db`
	// tag which defines the column name mapping. The results are set as values of
	// the fields.
	One(ctx context.Context, dest interface{}) error
}
