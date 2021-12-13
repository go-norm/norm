// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package sqlbuilder

import (
	"context"
	"database/sql"
	"math"
	"strings"

	"github.com/pkg/errors"

	"unknwon.dev/norm"
	"unknwon.dev/norm/expr"
	"unknwon.dev/norm/internal/immutable"
)

var (
	errMissingCursorColumn = errors.New("Missing cursor column")
)

type paginatorQuery struct {
	sel norm.Selector

	cursorColumn       string
	cursorValue        interface{}
	cursorCond         expr.Cond
	cursorReverseOrder bool

	pageSize   uint
	pageNumber uint
}

func newPaginator(sel norm.Selector, pageSize uint) norm.Paginator {
	pag := &paginator{}
	return pag.frame(func(pq *paginatorQuery) error {
		pq.pageSize = pageSize
		pq.sel = sel
		return nil
	}).Page(1)
}

func (pq *paginatorQuery) count(ctx context.Context) (uint64, error) {
	var count uint64

	row, err := pq.sel.(*selector).setColumns(expr.Raw("count(1) AS _t")).
		Limit(0).
		Offset(0).
		OrderBy(nil).
		QueryRow(ctx)
	if err != nil {
		return 0, err
	}

	err = row.Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

type paginator struct {
	fn   func(*paginatorQuery) error
	prev *paginator
}

var _ = immutable.Immutable(&paginator{})

func (pag *paginator) frame(fn func(*paginatorQuery) error) *paginator {
	return &paginator{prev: pag, fn: fn}
}

func (pag *paginator) Page(pageNumber uint) norm.Paginator {
	return pag.frame(func(pq *paginatorQuery) error {
		if pageNumber < 1 {
			pageNumber = 1
		}
		pq.pageNumber = pageNumber
		return nil
	})
}

func (pag *paginator) Cursor(column string) norm.Paginator {
	return pag.frame(func(pq *paginatorQuery) error {
		pq.cursorColumn = column
		pq.cursorValue = nil
		return nil
	})
}

func (pag *paginator) NextPage(cursorValue interface{}) norm.Paginator {
	return pag.frame(func(pq *paginatorQuery) error {
		if pq.cursorValue != nil && pq.cursorColumn == "" {
			return errMissingCursorColumn
		}
		pq.cursorValue = cursorValue
		pq.cursorReverseOrder = false
		if strings.HasPrefix(pq.cursorColumn, "-") {
			pq.cursorCond = expr.Cond{
				pq.cursorColumn[1:]: expr.Lt(cursorValue),
			}
		} else {
			pq.cursorCond = expr.Cond{
				pq.cursorColumn: expr.Gt(cursorValue),
			}
		}
		return nil
	})
}

func (pag *paginator) PrevPage(cursorValue interface{}) norm.Paginator {
	return pag.frame(func(pq *paginatorQuery) error {
		if pq.cursorValue != nil && pq.cursorColumn == "" {
			return errMissingCursorColumn
		}
		pq.cursorValue = cursorValue
		pq.cursorReverseOrder = true
		if strings.HasPrefix(pq.cursorColumn, "-") {
			pq.cursorCond = expr.Cond{
				pq.cursorColumn[1:]: expr.Gt(cursorValue),
			}
		} else {
			pq.cursorCond = expr.Cond{
				pq.cursorColumn: expr.Lt(cursorValue),
			}
		}
		return nil
	})
}

func (pag *paginator) TotalPages(ctx context.Context) (uint, error) {
	pq, err := pag.build()
	if err != nil {
		return 0, err
	}

	count, err := pq.count(ctx)
	if err != nil {
		return 0, err
	}
	if count < 1 {
		return 0, nil
	}

	if pq.pageSize < 1 {
		return 1, nil
	}

	pages := uint(math.Ceil(float64(count) / float64(pq.pageSize)))
	return pages, nil
}

func (pag *paginator) All(ctx context.Context, dest interface{}) error {
	pq, err := pag.buildWithCursor()
	if err != nil {
		return err
	}
	err = pq.sel.All(ctx, dest)
	if err != nil {
		return err
	}
	return nil
}

func (pag *paginator) One(ctx context.Context, dest interface{}) error {
	pq, err := pag.buildWithCursor()
	if err != nil {
		return err
	}
	return pq.sel.One(ctx, dest)
}

func (pag *paginator) Iterator(ctx context.Context) norm.Iterator {
	pq, err := pag.buildWithCursor()
	if err != nil {
		return &iterator{pq.sel.(*selector).Builder().Adapter, nil, err}
	}
	return pq.sel.Iterate(ctx)
}

func (pag *paginator) String() string {
	pq, err := pag.buildWithCursor()
	if err != nil {
		panic(err.Error())
	}
	return pq.sel.String()
}

func (pag *paginator) Arguments() []interface{} {
	pq, err := pag.buildWithCursor()
	if err != nil {
		return nil
	}
	return pq.sel.Arguments()
}

func (pag *paginator) Compile() (string, error) {
	pq, err := pag.buildWithCursor()
	if err != nil {
		return "", err
	}
	return pq.sel.(*selector).Compile()
}

func (pag *paginator) Query(ctx context.Context) (*sql.Rows, error) {
	pq, err := pag.buildWithCursor()
	if err != nil {
		return nil, err
	}
	return pq.sel.Query(ctx)
}

func (pag *paginator) QueryRow(ctx context.Context) (*sql.Row, error) {
	pq, err := pag.buildWithCursor()
	if err != nil {
		return nil, err
	}
	return pq.sel.QueryRow(ctx)
}

func (pag *paginator) Prepare(ctx context.Context) (*sql.Stmt, error) {
	pq, err := pag.buildWithCursor()
	if err != nil {
		return nil, err
	}
	return pq.sel.Prepare(ctx)
}

func (pag *paginator) TotalEntries(ctx context.Context) (uint64, error) {
	pq, err := pag.build()
	if err != nil {
		return 0, err
	}
	return pq.count(ctx)
}

func (pag *paginator) build() (*paginatorQuery, error) {
	pq, err := immutable.FastForward(pag)
	if err != nil {
		return nil, err
	}
	return pq.(*paginatorQuery), nil
}

func (pag *paginator) buildWithCursor() (*paginatorQuery, error) {
	pq, err := immutable.FastForward(pag)
	if err != nil {
		return nil, err
	}

	pqq := pq.(*paginatorQuery)

	if pqq.cursorReverseOrder {
		orderBy := pqq.cursorColumn

		if orderBy == "" {
			return nil, errMissingCursorColumn
		}

		if strings.HasPrefix(orderBy, "-") {
			orderBy = orderBy[1:]
		} else {
			orderBy = "-" + orderBy
		}

		pqq.sel = pqq.sel.OrderBy(orderBy)
	}

	if pqq.pageSize > 0 {
		pqq.sel = pqq.sel.Limit(int(pqq.pageSize))
		if pqq.pageNumber > 1 {
			pqq.sel = pqq.sel.Offset(int(pqq.pageSize * (pqq.pageNumber - 1)))
		}
	}

	if pqq.cursorCond != nil {
		pqq.sel = pqq.sel.Where(pqq.cursorCond).Offset(0)
	}

	if pqq.cursorColumn != "" {
		if pqq.cursorReverseOrder {
			pqq.sel = pqq.sel.(*selector).Builder().
				SelectFrom(expr.Raw("? AS p0", pqq.sel)).
				OrderBy(pqq.cursorColumn)
		} else {
			pqq.sel = pqq.sel.OrderBy(pqq.cursorColumn)
		}
	}

	return pqq, nil
}

func (pag *paginator) Prev() immutable.Immutable {
	if pag == nil {
		return nil
	}
	return pag.prev
}

func (pag *paginator) Fn(in interface{}) error {
	if pag.fn == nil {
		return nil
	}
	return pag.fn(in.(*paginatorQuery))
}

func (pag *paginator) Base() interface{} {
	return &paginatorQuery{}
}
