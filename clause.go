// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package norm

import (
	"context"
	"fmt"
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
	// From constructs a FROM clause for where the data to be retrieved from.
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
	//    q.Columns(...).From("users u").Where("u.name = ?", ...)
	From(tables ...interface{}) Selector
	// Distinct constructs a DISTINCT clause with given columns.
	//
	// It is used to ask the database to return only values that are different:
	//
	//    q.From(...).Distinct("name")
	Distinct(columns ...interface{}) Selector
	// As constructs an alias for the table.
	//
	// It can only be called after From() which defines the table:
	//
	//   q.From("users").As("u")
	As(alias string) Selector

	// Where constructs a WHERE clause to specify the conditions that columns must
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
	// GroupBy constructs a GROUP BY clause.
	//
	// It defines which columns should be used to aggregate and group results:
	//
	//   q.GroupBy("country_id")
	//   q.GroupBy("country_id", "city_id")
	//
	// Subsequent calls to GroupBy() replace the previously set clause.
	GroupBy(columns ...interface{}) Selector
	// OrderBy constructs a ORDER BY clause.
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
	Join(table ...interface{}) Selector
	// FullJoin is similar to Join() but is for FULL JOIN.
	FullJoin(table ...interface{}) Selector
	// CrossJoin is similar to Join() but is for CROSS JOIN.
	CrossJoin(table ...interface{}) Selector
	// RightJoin is similar to Join() but is for RIGHT JOIN.
	RightJoin(table ...interface{}) Selector
	// LeftJoin is similar to Join() but is for LEFT JOIN.
	LeftJoin(table ...interface{}) Selector
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

	// Limit constructs a LIMIT clause.
	//
	// It is used to define the maximum number of rows to be returned from the
	// query:
	//
	//   q.Limit(42)
	//
	// A negative limit cancels any previously set limit.
	Limit(n int) Selector
	// Offset constructs a OFFSET clause.
	//
	// It is used to define how many results are going to be skipped before starting to
	// return results.
	//
	//   q.Offset(56)
	//
	// A negative offset cancels any previously set offset.
	Offset(n int) Selector

	// Iterate creates an Iterator to iterate over query results.
	Iterate(ctx context.Context) Iterator
	ResultMapper

	// String returns a complied SQL query string.
	String() string
	// Arguments returns the arguments that are prepared for this query.
	Arguments() []interface{}

	// // Amend lets you alter the query's text just before sending it to the
	// // database server.
	// Amend(func(queryIn string) (queryOut string)) Selector
	//
	// // Paginate returns a paginator that can display a paginated lists of items.
	// // Paginators ignore previous Offset and Limit settings. Page numbering
	// // starts at 1.
	// Paginate(uint) Paginator
	//
	// // SQLPreparer provides methods for creating prepared statements.
	// SQLPreparer
	//
	// // SQLGetter provides methods to compile and execute a query that returns
	// // results.
	// SQLGetter
}

// Inserter represents a SQL query builder for the INSERT statement.
type Inserter interface {
	// Columns represents the COLUMNS clause.
	//
	// COLUMNS defines the columns that we are going to provide values for.
	//
	//   i.Columns("name", "last_name").Values(...)
	Columns(...string) Inserter

	// Values represents the VALUES clause.
	//
	// VALUES defines the values of the columns.
	//
	//   i.Columns(...).Values("María", "Méndez")
	//
	//   i.Values(map[string][string]{"name": "María"})
	Values(...interface{}) Inserter

	// Arguments returns the arguments that are prepared for this query.
	Arguments() []interface{}

	// Returning represents a RETURNING clause.
	//
	// RETURNING specifies which columns should be returned after INSERT.
	//
	// RETURNING may not be supported by all SQL databases.
	Returning(columns ...string) Inserter

	// Iterate provides methods to iterate over the results returned by
	// the Inserter. This is only possible when using Returning().
	// Iterate(ctx context.Context) Iterate

	// Amend lets you alter the query's text just before sending it to the
	// database server.
	Amend(func(queryIn string) (queryOut string)) Inserter

	// Batch provies a BatchInserter that can be used to insert many elements at
	// once by issuing several calls to Values(). It accepts a size parameter
	// which defines the batch size. If size is < 1, the batch size is set to 1.
	Batch(size int) BatchInserter

	// SQLExecer provides the Exec method.
	SQLExecer

	// SQLPreparer provides methods for creating prepared statements.
	SQLPreparer

	// SQLGetter provides methods to return query results from INSERT statements
	// that support such feature (e.g.: queries with Returning).
	SQLGetter

	// fmt.Stringer provides `String() string`, you can use `String()` to compile
	// the `Inserter` into a string.
	fmt.Stringer
}

// Updater represents a SQL query builder for the UPDATE statement.
type Updater interface {
	// Set represents the SET clause.
	Set(...interface{}) Updater

	// Where represents the WHERE clause.
	//
	// See Selector.Where for documentation and usage examples.
	Where(...interface{}) Updater

	// And appends more constraints to the WHERE clause without overwriting
	// conditions that have been already set.
	And(conds ...interface{}) Updater

	// Limit represents the LIMIT parameter.
	//
	// See Selector.Limit for documentation and usage examples.
	Limit(int) Updater

	// SQLPreparer provides methods for creating prepared statements.
	SQLPreparer

	// SQLExecer provides the Exec method.
	SQLExecer

	// fmt.Stringer provides `String() string`, you can use `String()` to compile
	// the `Inserter` into a string.
	fmt.Stringer

	// Arguments returns the arguments that are prepared for this query.
	Arguments() []interface{}

	// Amend lets you alter the query's text just before sending it to the
	// database server.
	Amend(func(queryIn string) (queryOut string)) Updater
}

// Deleter represents a SQL query builder for the DELETE statement.
type Deleter interface {
	// Where represents the WHERE clause.
	//
	// See Selector.Where for documentation and usage examples.
	Where(...interface{}) Deleter

	// And appends more constraints to the WHERE clause without overwriting
	// conditions that have been already set.
	And(conds ...interface{}) Deleter

	// Limit represents the LIMIT clause.
	//
	// See Selector.Limit for documentation and usage examples.
	Limit(int) Deleter

	// Amend lets you alter the query's text just before sending it to the
	// database server.
	Amend(func(queryIn string) (queryOut string)) Deleter

	// SQLPreparer provides methods for creating prepared statements.
	SQLPreparer

	// SQLExecer provides the Exec method.
	SQLExecer

	// fmt.Stringer provides `String() string`, you can use `String()` to compile
	// the `Inserter` into a string.
	fmt.Stringer

	// Arguments returns the arguments that are prepared for this query.
	Arguments() []interface{}
}

type Alter interface {
}

type Creator interface {
}

type Dropper interface {
}

// Paginator provides tools for splitting the results of a query into chunks
// containing a fixed number of items.
type Paginator interface {
	// Page sets the page number.
	Page(uint) Paginator

	// Cursor defines the column that is going to be taken as basis for
	// cursor-based pagination.
	//
	// Example:
	//
	//   a = q.Paginate(10).Cursor("id")
	//	 b = q.Paginate(12).Cursor("-id")
	//
	// You can set "" as cursorColumn to disable cursors.
	Cursor(cursorColumn string) Paginator

	// NextPage returns the next page according to the cursor. It expects a
	// cursorValue, which is the value the cursor column has on the last item of
	// the current result set (lower bound).
	//
	// Example:
	//
	//   p = q.NextPage(items[len(items)-1].ID)
	NextPage(cursorValue interface{}) Paginator

	// PrevPage returns the previous page according to the cursor. It expects a
	// cursorValue, which is the value the cursor column has on the fist item of
	// the current result set (upper bound).
	//
	// Example:
	//
	//   p = q.PrevPage(items[0].ID)
	PrevPage(cursorValue interface{}) Paginator

	// TotalPages returns the total number of pages in the query.
	TotalPages(ctx context.Context) (uint, error)

	// TotalEntries returns the total number of entries in the query.
	TotalEntries(ctx context.Context) (uint64, error)

	// SQLPreparer provides methods for creating prepared statements.
	SQLPreparer

	// SQLGetter provides methods to compile and execute a query that returns
	// results.
	SQLGetter

	// Iterator provides methods to iterate over the results returned by
	// the Selector.
	Iterator(ctx context.Context) Iterator

	// ResultMapper provides methods to retrieve and map results.
	ResultMapper

	// fmt.Stringer provides `String() string`, you can use `String()` to compile
	// the `Selector` into a string.
	fmt.Stringer

	// Arguments returns the arguments that are prepared for this query.
	Arguments() []interface{}
}

// BatchInserter provides an interface to do massive insertions in batches.
type BatchInserter interface {
	// Values pushes column values to be inserted as part of the batch.
	Values(...interface{}) BatchInserter

	// NextResult dumps the next slice of results to dst, which can mean having
	// the IDs of all inserted elements in the batch.
	NextResult(ctx context.Context, dst interface{}) bool

	// Done signals that no more elements are going to be added.
	Done()

	// Wait blocks until the whole batch is executed.
	Wait(ctx context.Context) error

	// Err returns the last error that happened while executing the batch (or nil
	// if no error happened).
	Err() error
}

// Iterator defines a collection of methods to iterate over query results.
type Iterator interface {
	// // Scan dumps the current result into the given pointer variable pointers.
	// Scan(dest ...interface{}) error
	// // NextScan advances the iterator and performs Scan.
	// NextScan(dest ...interface{}) error
	// // ScanOne advances the iterator, performs Scan and closes the iterator.
	// ScanOne(dest ...interface{}) error
	// // Next dumps the current element into the given destination, which could be
	// // a pointer to either a map or a struct.
	// Next(dest ...interface{}) bool
	// // Err returns the last error produced by the cursor.
	// Err() error
	// // Close closes the iterator and frees up the cursor.
	// Close() error

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
	// which can be a pointer to a map or a struct.
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
