// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package sqlbuilder

import (
	"github.com/pkg/errors"
)

var (
	ErrExpectingPointer                    = errors.New(`argument must be an address`)
	ErrExpectingSlicePointer               = errors.New(`argument must be a slice address`)
	ErrExpectingSliceMapStruct             = errors.New(`argument must be a slice address of maps or structs`)
	ErrExpectingMapOrStruct                = errors.New(`argument must be either a map or a struct`)
	ErrExpectingPointerToEitherMapOrStruct = errors.New(`expecting a pointer to either a map or a struct`)
	ErrNoMoreRows                          = errors.New(`no more rows in the result set`)
)
