// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package sqlbuilder

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"unknwon.dev/norm"
	"unknwon.dev/norm/expr"
	"unknwon.dev/norm/internal/exql"
	"unknwon.dev/norm/internal/immutable"
)

var _ norm.Selector = (*selector)(nil)

type selector struct {
	builder *sqlBuilder

	prev *selector
	fn   func(*selectorQuery) error
}

func (s *selector) frame(fn func(*selectorQuery) error) *selector {
	return &selector{
		prev: s,
		fn:   fn,
	}
}

func (s *selector) Builder() *sqlBuilder {
	if s.prev == nil {
		return s.builder
	}
	return s.prev.Builder()
}

func (s *selector) Columns(columns ...interface{}) norm.Selector {
	if len(columns) == 0 {
		return s
	}
	return s.frame(func(sq *selectorQuery) error {
		return errors.Wrap(sq.pushColumns(columns), "Columns")
	})
}

func (s *selector) From(tables ...interface{}) norm.Selector {
	if len(tables) == 0 {
		return s
	}
	return s.frame(func(sq *selectorQuery) error {
		fragments, args, err := parseColumnExpressions(tables)
		if err != nil {
			return errors.Wrap(err, "From: parse column expressions")
		}
		sq.table = exql.JoinColumns(fragments...)
		sq.tableArgs = args
		return nil
	})
}

func (s *selector) Distinct(columns ...interface{}) norm.Selector {
	if len(columns) == 0 {
		return s
	}
	return s.frame(func(sq *selectorQuery) error {
		sq.distinct = true
		return sq.pushColumns(columns)
	})
}

func (s *selector) As(alias string) norm.Selector {
	return s.frame(func(sq *selectorQuery) error {
		if sq.table.IsEmpty() {
			return errors.New("As: cannot use As() without a table")
		}

		last := len(sq.table.Columns) - 1
		raw, ok := sq.table.Columns[last].(*exql.Raw)
		if !ok {
			return nil
		}

		compiled, err := exql.ColumnWithName(alias).Compile(s.Builder().t.Template)
		if err != nil {
			return errors.Wrap(err, "As: compile column with alias")
		}

		sq.table.Columns[last] = exql.RawValue(raw.Value + " AS " + compiled)
		return nil
	})
}

func (s *selector) Where(conds ...interface{}) norm.Selector {
	if len(conds) == 0 {
		return s
	}
	return s.frame(func(sq *selectorQuery) error {
		sq.where, sq.whereArgs = nil, nil
		return errors.Wrap(sq.and(s.Builder().t, conds...), "Where")
	})
}

func (s *selector) And(conds ...interface{}) norm.Selector {
	if len(conds) == 0 {
		return s
	}
	return s.frame(func(sq *selectorQuery) error {
		return errors.Wrap(sq.and(s.Builder().t, conds...), "And")
	})
}

func (s *selector) GroupBy(columns ...interface{}) norm.Selector {
	if len(columns) == 0 {
		return s
	}
	return s.frame(func(sq *selectorQuery) error {
		fragments, args, err := parseColumnExpressions(columns)
		if err != nil {
			return errors.Wrap(err, "GroupBy: parse column expressions")
		}

		sq.groupBy = exql.GroupByColumns(fragments...)
		sq.groupByArgs = args
		return nil
	})
}

func (s *selector) OrderBy(columns ...interface{}) norm.Selector {
	if len(columns) == 0 {
		return s
	}
	return s.frame(func(sq *selectorQuery) error {
		sorts := &exql.SortColumns{
			Columns: make([]exql.Fragment, len(columns)),
		}
		orderByArgs := []interface{}{}
		for i, c := range columns {
			var sort *exql.SortColumn
			switch v := c.(type) {
			case *expr.RawExpr:
				q, args := Preprocess(v.Raw(), v.Arguments())
				sort = &exql.SortColumn{
					Column: exql.RawValue(q),
				}
				orderByArgs = append(orderByArgs, args...)

			case *expr.FuncExpr:
				fnName, fnArgs := v.Name(), v.Arguments()
				if len(fnArgs) == 0 {
					fnName = fnName + "()"
				} else {
					fnName = fnName + "(?" + strings.Repeat("?, ", len(fnArgs)-1) + ")"
				}
				fnName, fnArgs = Preprocess(fnName, fnArgs)
				sort = &exql.SortColumn{
					Column: exql.RawValue(fnName),
				}
				orderByArgs = append(orderByArgs, fnArgs...)

			case string:
				if strings.HasPrefix(v, "-") {
					sort = &exql.SortColumn{
						Column: exql.ColumnWithName(v[1:]),
						Order:  exql.Descendent,
					}
				} else {
					chunks := strings.SplitN(v, " ", 2)
					order := exql.Ascendant
					if len(chunks) > 1 && strings.ToUpper(chunks[1]) == "DESC" {
						order = exql.Descendent
					}

					sort = &exql.SortColumn{
						Column: exql.ColumnWithName(chunks[0]),
						Order:  order,
					}
				}

			default:
				return errors.Errorf("unexpected type %T", v)
			}

			sorts.Columns[i] = sort
		}

		sq.orderBy = &exql.OrderBy{
			SortColumns: sorts,
		}
		sq.orderByArgs = orderByArgs
		return nil
	})
}

func (s *selector) Join(table ...interface{}) norm.Selector {
	return s.frame(func(sq *selectorQuery) error {
		return sq.pushJoin("", table)
	})
}

func (s *selector) FullJoin(table ...interface{}) norm.Selector {
	return s.frame(func(sq *selectorQuery) error {
		return sq.pushJoin("FULL", table)
	})
}

func (s *selector) CrossJoin(table ...interface{}) norm.Selector {
	return s.frame(func(sq *selectorQuery) error {
		return sq.pushJoin("CROSS", table)
	})
}

func (s *selector) RightJoin(table ...interface{}) norm.Selector {
	return s.frame(func(sq *selectorQuery) error {
		return sq.pushJoin("RIGHT", table)
	})
}

func (s *selector) LeftJoin(table ...interface{}) norm.Selector {
	return s.frame(func(sq *selectorQuery) error {
		return sq.pushJoin("LEFT", table)
	})
}

func (s *selector) On(conds ...interface{}) norm.Selector {
	if len(conds) == 0 {
		return s
	}
	return s.frame(func(sq *selectorQuery) error {
		joins := len(sq.joins)
		if joins == 0 {
			return errors.New("On: cannot use On() without a preceding JOIN expression")
		}

		lastJoin := sq.joins[joins-1]
		if lastJoin.On != nil {
			return errors.New(`On: cannot use multiple Using() or On() with the same JOIN expression`)
		}

		where, args, err := s.Builder().t.toWhereClause(conds)
		if err != nil {
			return errors.Wrap(err, "convert to WHERE clause")
		}

		on := exql.On(*where)
		lastJoin.On = &on
		sq.joinsArgs = append(sq.joinsArgs, args...)
		return nil
	})
}

func (s *selector) Using(columns ...interface{}) norm.Selector {
	if len(columns) == 0 {
		return s
	}
	return s.frame(func(sq *selectorQuery) error {
		joins := len(sq.joins)
		if joins == 0 {
			return errors.New("Using: cannot use Using() without a preceding JOIN expression")
		}

		lastJoin := sq.joins[joins-1]
		if lastJoin.On != nil {
			return errors.New(`Using: cannot use multiple Using() or On() with the same JOIN expression`)
		}

		fragments, args, err := parseColumnExpressions(columns)
		if err != nil {
			return errors.Wrap(err, "Using: parse column expressions")
		}

		lastJoin.Using = exql.UsingColumns(fragments...)
		sq.joinsArgs = append(sq.joinsArgs, args...)
		return nil
	})
}

func (s *selector) Limit(n int) norm.Selector {
	return s.frame(func(sq *selectorQuery) error {
		if n < 0 {
			n = 0
		}
		sq.limit = exql.Limit(n)
		return nil
	})
}

func (s *selector) Offset(n int) norm.Selector {
	return s.frame(func(sq *selectorQuery) error {
		if n < 0 {
			n = 0
		}
		sq.offset = exql.Offset(n)
		return nil
	})
}

func (s *selector) Iterate(ctx context.Context) norm.Iterator {
	sq, err := s.build()
	if err != nil {
		return &iterator{err: errors.Wrap(err, "build query")}
	}

	adapter := s.Builder().Adapter
	rows, err := adapter.Executor().Query(ctx, sq.statement(), sq.arguments()...) //nolint:rowserrcheck
	return &iterator{
		adapter: adapter,
		cursor:  rows,
		err:     errors.Wrap(err, "execute query"),
	}
}

func (s *selector) All(ctx context.Context, destSlice interface{}) error {
	return s.Iterate(ctx).All(ctx, destSlice)
}

func (s *selector) One(ctx context.Context, dest interface{}) error {
	return s.Iterate(ctx).One(ctx, dest)
}

func (s *selector) String() string {
	q, err := s.Compile()
	if err != nil {
		panic("unable to compile SELECT query: " + err.Error())
	}
	return s.Builder().t.FormatSQL(q)
}

var _ compilable = (*selector)(nil)

func (s *selector) Compile() (string, error) {
	sq, err := s.build()
	if err != nil {
		return "", errors.Wrap(err, "build")
	}
	return sq.statement().Compile(s.Builder().t.Template)
}

func (s *selector) build() (*selectorQuery, error) {
	sq, err := immutable.FastForward(s)
	if err != nil {
		return nil, errors.Wrap(err, "construct *selectorQuery")
	}
	return sq.(*selectorQuery), nil
}

func (s *selector) Arguments() []interface{} {
	sq, err := s.build()
	if err != nil {
		panic("unable to build SELECT query: " + err.Error())
	}

	args := sq.arguments()
	for i := range args {
		args[i] = s.Builder().Typer().Valuer(args[i])
	}
	return args
}

var _ immutable.Immutable = (*selector)(nil)

func (s *selector) Prev() immutable.Immutable {
	if s == nil {
		return nil
	}
	return s.prev
}

func (s *selector) Fn(in interface{}) error {
	if s.fn == nil {
		return nil
	}
	return s.fn(in.(*selectorQuery))
}

func (s *selector) Base() interface{} {
	return &selectorQuery{}
}

// todo

func (s *selector) clone() norm.Selector {
	return s.frame(func(*selectorQuery) error {
		return nil
	})
}

func (s *selector) setColumns(columns ...interface{}) norm.Selector {
	return s.frame(func(sq *selectorQuery) error {
		sq.columns = nil
		return sq.pushColumns(columns)
	})
}

func (s *selector) Amend(fn func(string) string) norm.Selector {
	return s.frame(func(sq *selectorQuery) error {
		sq.amendFn = fn
		return nil
	})
}

func (s *selector) QueryRow(ctx context.Context) (*sql.Row, error) {
	sq, err := s.build()
	if err != nil {
		return nil, err
	}

	return s.Builder().Executor().QueryRow(ctx, sq.statement(), sq.arguments()...)
}

func (s *selector) Prepare(ctx context.Context) (*sql.Stmt, error) {
	sq, err := s.build()
	if err != nil {
		return nil, err
	}
	return s.Builder().Executor().Prepare(ctx, sq.statement())
}

func (s *selector) Query(ctx context.Context) (*sql.Rows, error) {
	sq, err := s.build()
	if err != nil {
		return nil, err
	}
	return s.Builder().Executor().Query(ctx, sq.statement(), sq.arguments()...)
}

func (s *selector) Paginate(pageSize uint) norm.Paginator {
	return newPaginator(s.clone(), pageSize)
}

type selectorQuery struct {
	table     *exql.Columns
	tableArgs []interface{}

	distinct bool

	where     *exql.Where
	whereArgs []interface{}

	groupBy     *exql.GroupBy
	groupByArgs []interface{}

	orderBy     *exql.OrderBy
	orderByArgs []interface{}

	limit  exql.Limit
	offset exql.Offset

	columns     *exql.Columns
	columnsArgs []interface{}

	joins     []*exql.Join
	joinsArgs []interface{}

	amendFn func(string) string
}

func (sq *selectorQuery) arguments() []interface{} {
	return joinArguments(
		sq.columnsArgs,
		sq.tableArgs,
		sq.joinsArgs,
		sq.whereArgs,
		sq.groupByArgs,
		sq.orderByArgs,
	)
}

func parseColumnExpressions(exprs []interface{}) (fragments []exql.Fragment, args []interface{}, err error) {
	fragments = make([]exql.Fragment, len(exprs))
	args = []interface{}{}
	for i, e := range exprs {
		switch v := e.(type) {
		case compilable:
			q, err := v.Compile()
			if err != nil {
				return nil, nil, errors.Wrap(err, "compile")
			}

			q, a := Preprocess(q, v.Arguments())
			if _, ok := v.(norm.Selector); ok {
				q = "(" + q + ")"
			}
			fragments[i] = exql.RawValue(q)
			args = append(args, a...)

		case *expr.FuncExpr:
			fnName, fnArgs := v.Name(), v.Arguments()
			if len(fnArgs) == 0 {
				fnName = fnName + "()"
			} else {
				fnName = fnName + "(?" + strings.Repeat("?, ", len(fnArgs)-1) + ")"
			}
			fnName, fnArgs = Preprocess(fnName, fnArgs)
			fragments[i] = exql.RawValue(fnName)
			args = append(args, fnArgs...)

		case *expr.RawExpr:
			q, a := Preprocess(v.Raw(), v.Arguments())
			fragments[i] = exql.RawValue(q)
			args = append(args, a...)

		case exql.Fragment:
			fragments[i] = v
		case string:
			fragments[i] = exql.ColumnWithName(v)
		case int:
			fragments[i] = exql.RawValue(strconv.Itoa(v))
		case interface{}:
			fragments[i] = exql.ColumnWithName(fmt.Sprintf("%v", v))
		default:
			return nil, nil, errors.Errorf("unexpected type %T", v)
		}
	}
	return fragments, args, nil
}

func (sq *selectorQuery) pushColumns(exprs []interface{}) error {
	fragments, args, err := parseColumnExpressions(exprs)
	if err != nil {
		return errors.Wrap(err, "parse column expressions")
	}

	columns := exql.JoinColumns(fragments...)
	if sq.columns != nil {
		sq.columns.Append(columns)
	} else {
		sq.columns = columns
	}

	sq.columnsArgs = append(sq.columnsArgs, args...)
	return nil
}

func (sq *selectorQuery) and(t *templateWithUtils, conds ...interface{}) error {
	where, whereArgs, err := t.toWhereClause(conds)
	if err != nil {
		return errors.Wrap(err, "convert to WHERE clause")
	}

	if sq.where == nil {
		sq.where, sq.whereArgs = &exql.Where{}, []interface{}{}
	}
	sq.where.Append(where)
	sq.whereArgs = append(sq.whereArgs, whereArgs...)
	return nil
}

func (sq *selectorQuery) pushJoin(t string, tables []interface{}) error {
	fragments, args, err := parseColumnExpressions(tables)
	if err != nil {
		return errors.Wrap(err, "parse column expressions")
	}

	sq.joins = append(sq.joins,
		&exql.Join{
			Type:  t,
			Table: exql.JoinColumns(fragments...),
		},
	)
	sq.joinsArgs = append(sq.joinsArgs, args...)
	return nil
}

func (sq *selectorQuery) statement() *exql.Statement {
	stmt := &exql.Statement{
		Type:     exql.Select,
		Table:    sq.table,
		Columns:  sq.columns,
		Distinct: sq.distinct,
		OrderBy:  sq.orderBy,
		GroupBy:  sq.groupBy,
		Where:    sq.where,
		Limit:    sq.limit,
		Offset:   sq.offset,
	}

	if len(sq.joins) > 0 {
		stmt.Joins = exql.JoinConditions(sq.joins...)
	}

	stmt.SetAmendment(sq.amendFn)
	return stmt
}
