// Copyright 2012 The upper.io/db authors. All rights reserved.
// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package immutable

// Immutable represents an immutable chain that, if passed to FastForward, Fn()
// is applied to every element in the chain. The first element of this chain is
// represented by Base().
type Immutable interface {
	// Prev returns the previous element in the chain.
	Prev() Immutable
	// Fn is a function that is able to modify the passed element.
	Fn(interface{}) error
	// Base returns the first element in the chain.
	Base() interface{}
}

// FastForward applies all Fn methods to the given chain.
func FastForward(curr Immutable) (interface{}, error) {
	prev := curr.Prev()
	if prev == nil {
		return curr.Base(), nil
	}
	in, err := FastForward(prev)
	if err != nil {
		return nil, err
	}
	err = curr.Fn(in)
	return in, err
}
