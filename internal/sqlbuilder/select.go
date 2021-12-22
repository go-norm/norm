// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package sqlbuilder

import (
	"context"
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

func (sel *selector) frame(fn func(*selectorQuery) error) *selector {
	return &selector{
		prev: sel,
		fn:   fn,
	}
}

func (sel *selector) Builder() *sqlBuilder {
	if sel.prev == nil {
		return sel.builder
	}
	return sel.prev.Builder()
}

func (sel *selector) Columns(columns ...interface{}) norm.Selector {
	if len(columns) == 0 {
		return sel
	}
	return sel.frame(func(sq *selectorQuery) error {
		return errors.Wrap(sq.pushColumns(columns), "Columns")
	})
}

func (sel *selector) From(tables ...interface{}) norm.Selector {
	if len(tables) == 0 {
		return sel
	}
	return sel.frame(func(sq *selectorQuery) error {
		cs, args, err := toColumns(tables)
		if err != nil {
			return errors.Wrap(err, "From: convert to columns")
		}

		ts := make([]*exql.TableFragment, len(cs))
		for i := range cs {
			ts[i] = exql.Table(cs[i])
		}
		sq.table = exql.Tables(ts...)
		sq.tableArgs = args
		return nil
	})
}

func (sel *selector) Distinct(columns ...string) norm.Selector {
	return sel.frame(func(sq *selectorQuery) error {
		if len(columns) == 0 {
			sq.distinct = true
			return nil
		}

		cs := make([]*exql.ColumnFragment, len(columns))
		for i := range columns {
			cs[i] = exql.Column(columns[i])
		}
		cvs := exql.Columns(cs...)

		compiled, err := cvs.Compile(sel.Builder().Template)
		if err != nil {
			return errors.Wrap(err, "compile column")
		}

		distinct := expr.Raw("DISTINCT(" + compiled + ")")
		return sq.pushColumns([]interface{}{distinct})
	})
}

func (sel *selector) As(alias string) norm.Selector {
	return sel.frame(func(sq *selectorQuery) error {
		if sq.table.Empty() {
			return errors.New("As: cannot use As() without a table")
		}

		last := len(sq.table.Tables) - 1
		sq.table.Tables[last].Alias = alias
		return nil
	})
}

func (sel *selector) Where(conds ...interface{}) norm.Selector {
	if len(conds) == 0 {
		return sel
	}
	return sel.frame(func(sq *selectorQuery) error {
		sq.where, sq.whereArgs = nil, nil
		return errors.Wrap(sq.and(sel.Builder().Template, conds...), "WhereFragment")
	})
}

func (sel *selector) And(conds ...interface{}) norm.Selector {
	if len(conds) == 0 {
		return sel
	}
	return sel.frame(func(sq *selectorQuery) error {
		return errors.Wrap(sq.and(sel.Builder().Template, conds...), "AndFragment")
	})
}

func (sel *selector) GroupBy(columns ...interface{}) norm.Selector {
	if len(columns) == 0 {
		return sel
	}
	return sel.frame(func(sq *selectorQuery) error {
		cs, args, err := toColumns(columns)
		if err != nil {
			return errors.Wrap(err, "GroupBy: convert to columns")
		}

		groupBy := make([]*exql.ColumnFragment, len(cs))
		for i := range cs {
			groupBy[i] = exql.Column(cs[i])
		}
		sq.groupBy = exql.GroupBy(groupBy...)
		sq.groupByArgs = args
		return nil
	})
}

func (sel *selector) OrderBy(columns ...interface{}) norm.Selector {
	if len(columns) == 0 {
		return sel
	}
	return sel.frame(func(sq *selectorQuery) error {
		orderByColumns := make([]*exql.SortColumnFragment, len(columns))
		orderByArgs := []interface{}{}
		for i := range columns {
			var sc *exql.SortColumnFragment
			switch v := columns[i].(type) {
			case *expr.RawExpr:
				r, rArgs, err := ExpandQuery(v.Raw(), v.Arguments())
				if err != nil {
					return errors.Wrap(err, "OrderBy: expand query for *expr.RawExpr")
				}

				sc = exql.SortColumn(exql.Raw(r))
				orderByArgs = append(orderByArgs, rArgs...)

			case *expr.FuncExpr:
				fnName, fnArgs := v.Name(), v.Arguments()
				if len(fnArgs) == 0 {
					fnName = fnName + "()"
				} else {
					fnName = fnName + "(?" + strings.Repeat("?, ", len(fnArgs)-1) + ")"
				}

				placeholder, args, err := ExpandQuery(fnName, fnArgs)
				if err != nil {
					return errors.Wrap(err, "OrderBy: expand query for *expr.FuncExpr")
				}

				sc = exql.SortColumn(exql.Raw(placeholder))
				orderByArgs = append(orderByArgs, args...)

			case string:
				if strings.HasPrefix(v, "-") {
					sc = exql.SortColumn(v[1:], exql.SortDescendent)
				} else {
					chunks := strings.SplitN(v, " ", 2)
					order := exql.SortAscendant
					if len(chunks) > 1 && strings.ToUpper(chunks[1]) == "DESC" {
						order = exql.SortDescendent
					}
					sc = exql.SortColumn(chunks[0], order)
				}

			default:
				return errors.Errorf("OrderBy: unsupported type %T", v)
			}

			orderByColumns[i] = sc
		}

		sq.orderBy = exql.OrderBy(orderByColumns...)
		sq.orderByArgs = orderByArgs
		return nil
	})
}

func (sel *selector) Join(table interface{}) norm.Selector {
	return sel.frame(func(sq *selectorQuery) error {
		return sq.pushJoin(exql.DefaultJoin, table)
	})
}

func (sel *selector) FullJoin(table interface{}) norm.Selector {
	return sel.frame(func(sq *selectorQuery) error {
		return sq.pushJoin(exql.FullJoin, table)
	})
}

func (sel *selector) CrossJoin(table interface{}) norm.Selector {
	return sel.frame(func(sq *selectorQuery) error {
		return sq.pushJoin(exql.CrossJoin, table)
	})
}

func (sel *selector) RightJoin(table interface{}) norm.Selector {
	return sel.frame(func(sq *selectorQuery) error {
		return sq.pushJoin(exql.RightJoin, table)
	})
}

func (sel *selector) LeftJoin(table interface{}) norm.Selector {
	return sel.frame(func(sq *selectorQuery) error {
		return sq.pushJoin(exql.LeftJoin, table)
	})
}

func (sel *selector) On(conds ...interface{}) norm.Selector {
	if len(conds) == 0 {
		return sel
	}
	return sel.frame(func(sq *selectorQuery) error {
		joins := len(sq.joins)
		if joins == 0 {
			return errors.New("On: cannot use On() without a preceding JOIN expression")
		}

		lastJoin := sq.joins[joins-1]
		if lastJoin.On != nil || lastJoin.Using != nil {
			return errors.New(`On: cannot use multiple Using() or On() with the same JOIN expression`)
		}

		// Convert columns to *exql.ColumnValueFragment, e.g. "users.id = user_emails.id"
		for i := range conds {
			str, ok := conds[i].(string)
			if !ok {
				continue
			}

			chunks := strings.SplitN(str, " ", 3)
			if len(chunks) != 3 {
				continue
			}

			conds[i] = exql.ColumnValue(chunks[0], chunks[1], exql.Column(chunks[2]))
		}

		where, args, err := toWhere(sel.Builder().Template, conds)
		if err != nil {
			return errors.Wrap(err, "convert to WHERE clause")
		}

		lastJoin.On = exql.On(where.Conditions...)
		sq.joinsArgs = append(sq.joinsArgs, args...)
		return nil
	})
}

func (sel *selector) Using(columns ...interface{}) norm.Selector {
	if len(columns) == 0 {
		return sel
	}
	return sel.frame(func(sq *selectorQuery) error {
		joins := len(sq.joins)
		if joins == 0 {
			return errors.New("Using: cannot use Using() without a preceding JOIN expression")
		}

		lastJoin := sq.joins[joins-1]
		if lastJoin.On != nil || lastJoin.Using != nil {
			return errors.New(`Using: cannot use multiple Using() or On() with the same JOIN expression`)
		}

		cs, args, err := toColumns(columns)
		if err != nil {
			return errors.Wrap(err, "Using: convert to columns")
		}

		using := make([]*exql.ColumnFragment, len(cs))
		for i := range cs {
			using[i] = exql.Column(cs[i])
		}
		lastJoin.Using = exql.Using(using...)
		sq.joinsArgs = append(sq.joinsArgs, args...)
		return nil
	})
}

func (sel *selector) Limit(n int) norm.Selector {
	if n <= 0 {
		return sel
	}
	return sel.frame(func(sq *selectorQuery) error {
		sq.limit = n
		return nil
	})
}

func (sel *selector) Offset(n int) norm.Selector {
	if n <= 0 {
		return sel
	}
	return sel.frame(func(sq *selectorQuery) error {
		sq.offset = n
		return nil
	})
}

func (sel *selector) Iterate(ctx context.Context) norm.Iterator {
	sq, err := sel.build()
	if err != nil {
		return &iterator{err: errors.Wrap(err, "build query")}
	}

	adapter := sel.Builder().Adapter
	rows, err := adapter.Executor().Query(ctx, sq.statement(), sq.arguments()...) //nolint:rowserrcheck
	return &iterator{
		adapter: adapter,
		cursor:  rows,
		err:     errors.Wrap(err, "execute query"),
	}
}

func (sel *selector) All(ctx context.Context, destSlice interface{}) error {
	return sel.Iterate(ctx).All(ctx, destSlice)
}

func (sel *selector) One(ctx context.Context, dest interface{}) error {
	return sel.Iterate(ctx).One(ctx, dest)
}

func (sel *selector) String() string {
	q, err := sel.Compile()
	if err != nil {
		panic("unable to compile SELECT query: " + err.Error())
	}
	return sel.Builder().FormatSQL(q)
}

var _ compilable = (*selector)(nil)

func (sel *selector) Compile() (string, error) {
	sq, err := sel.build()
	if err != nil {
		return "", errors.Wrap(err, "build")
	}
	return sq.statement().Compile(sel.Builder().Template)
}

func (sel *selector) build() (*selectorQuery, error) {
	sq, err := immutable.FastForward(sel)
	if err != nil {
		return nil, errors.Wrap(err, "construct *selectorQuery")
	}
	return sq.(*selectorQuery), nil
}

func (sel *selector) Arguments() []interface{} {
	sq, err := sel.build()
	if err != nil {
		panic("unable to build SELECT query: " + err.Error())
	}

	args := sq.arguments()
	for i := range args {
		args[i] = sel.Builder().Typer().Valuer(args[i])
	}
	return args
}

var _ immutable.Immutable = (*selector)(nil)

func (sel *selector) Prev() immutable.Immutable {
	if sel == nil {
		return nil
	}
	return sel.prev
}

func (sel *selector) Fn(in interface{}) error {
	if sel.fn == nil {
		return nil
	}
	return sel.fn(in.(*selectorQuery))
}

func (sel *selector) Base() interface{} {
	return &selectorQuery{}
}

type selectorQuery struct {
	table     *exql.TablesFragment
	tableArgs []interface{}

	distinct bool

	where     *exql.WhereFragment
	whereArgs []interface{}

	groupBy     *exql.GroupByFragment
	groupByArgs []interface{}

	orderBy     *exql.OrderByFragment
	orderByArgs []interface{}

	limit  int
	offset int

	columns     *exql.ColumnsFragment
	columnsArgs []interface{}

	joins     []*exql.JoinFragment
	joinsArgs []interface{}

	amendFn func(string) string
}

func flattenArguments(args ...[]interface{}) []interface{} {
	total := 0
	for i := range args {
		total += len(args[i])
	}
	if total == 0 {
		return nil
	}

	flatten := make([]interface{}, 0, total)
	for i := range args {
		flatten = append(flatten, args[i]...)
	}
	return flatten
}

func (sq *selectorQuery) arguments() []interface{} {
	return flattenArguments(
		sq.columnsArgs,
		sq.tableArgs,
		sq.joinsArgs,
		sq.whereArgs,
		sq.groupByArgs,
		sq.orderByArgs,
	)
}

func (sq *selectorQuery) pushColumns(exprs []interface{}) error {
	cs, args, err := toColumns(exprs)
	if err != nil {
		return errors.Wrap(err, "convert to columns")
	}

	columns := make([]*exql.ColumnFragment, len(cs))
	for i := range cs {
		columns[i] = exql.Column(cs[i])
	}
	if sq.columns != nil {
		sq.columns.Append(columns...)
	} else {
		sq.columns = exql.Columns(columns...)
	}

	sq.columnsArgs = append(sq.columnsArgs, args...)
	return nil
}

func (sq *selectorQuery) and(t *exql.Template, conds ...interface{}) error {
	where, whereArgs, err := toWhere(t, conds)
	if err != nil {
		return errors.Wrap(err, "convert to WHERE clause")
	}

	if sq.where == nil {
		sq.where, sq.whereArgs = exql.Where(), []interface{}{}
	}
	sq.where.Append(where.Conditions...)
	sq.whereArgs = append(sq.whereArgs, whereArgs...)
	return nil
}

func (sq *selectorQuery) pushJoin(typ exql.JoinType, table interface{}) error {
	sq.joins = append(sq.joins, exql.JoinOn(typ, table, nil))
	return nil
}

func (sq *selectorQuery) statement() *exql.Statement {
	stmt := &exql.Statement{
		Type:     exql.StatementSelect,
		Table:    sq.table,
		Columns:  sq.columns,
		Distinct: sq.distinct,
		OrderBy:  sq.orderBy,
		GroupBy:  sq.groupBy,
		Joins:    exql.Joins(sq.joins...),
		Where:    sq.where,
		Limit:    sq.limit,
		Offset:   sq.offset,
	}
	stmt.SetAmend(sq.amendFn)
	return stmt
}