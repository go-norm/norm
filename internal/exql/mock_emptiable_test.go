// Code generated by go-mockgen 1.1.2; DO NOT EDIT.

package exql

import "sync"

// MockEmptiable is a mock implementation of the emptiable interface (from
// the package unknwon.dev/norm/internal/exql) used for unit testing.
type MockEmptiable struct {
	// EmptyFunc is an instance of a mock function object controlling the
	// behavior of the method Empty.
	EmptyFunc *EmptiableEmptyFunc
}

// NewMockEmptiable creates a new mock of the emptiable interface. All
// methods return zero values for all results, unless overwritten.
func NewMockEmptiable() *MockEmptiable {
	return &MockEmptiable{
		EmptyFunc: &EmptiableEmptyFunc{
			defaultHook: func() bool {
				return false
			},
		},
	}
}

// NewStrictMockEmptiable creates a new mock of the emptiable interface. All
// methods panic on invocation, unless overwritten.
func NewStrictMockEmptiable() *MockEmptiable {
	return &MockEmptiable{
		EmptyFunc: &EmptiableEmptyFunc{
			defaultHook: func() bool {
				panic("unexpected invocation of MockEmptiable.Empty")
			},
		},
	}
}

// surrogateMockEmptiable is a copy of the emptiable interface (from the
// package unknwon.dev/norm/internal/exql). It is redefined here as it is
// unexported in the source package.
type surrogateMockEmptiable interface {
	Empty() bool
}

// NewMockEmptiableFrom creates a new mock of the MockEmptiable interface.
// All methods delegate to the given implementation, unless overwritten.
func NewMockEmptiableFrom(i surrogateMockEmptiable) *MockEmptiable {
	return &MockEmptiable{
		EmptyFunc: &EmptiableEmptyFunc{
			defaultHook: i.Empty,
		},
	}
}

// EmptiableEmptyFunc describes the behavior when the Empty method of the
// parent MockEmptiable instance is invoked.
type EmptiableEmptyFunc struct {
	defaultHook func() bool
	hooks       []func() bool
	history     []EmptiableEmptyFuncCall
	mutex       sync.Mutex
}

// Empty delegates to the next hook function in the queue and stores the
// parameter and result values of this invocation.
func (m *MockEmptiable) Empty() bool {
	r0 := m.EmptyFunc.nextHook()()
	m.EmptyFunc.appendCall(EmptiableEmptyFuncCall{r0})
	return r0
}

// SetDefaultHook sets function that is called when the Empty method of the
// parent MockEmptiable instance is invoked and the hook queue is empty.
func (f *EmptiableEmptyFunc) SetDefaultHook(hook func() bool) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// Empty method of the parent MockEmptiable instance invokes the hook at the
// front of the queue and discards it. After the queue is empty, the default
// hook function is invoked for any future action.
func (f *EmptiableEmptyFunc) PushHook(hook func() bool) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultDefaultHook with a function that returns
// the given values.
func (f *EmptiableEmptyFunc) SetDefaultReturn(r0 bool) {
	f.SetDefaultHook(func() bool {
		return r0
	})
}

// PushReturn calls PushDefaultHook with a function that returns the given
// values.
func (f *EmptiableEmptyFunc) PushReturn(r0 bool) {
	f.PushHook(func() bool {
		return r0
	})
}

func (f *EmptiableEmptyFunc) nextHook() func() bool {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *EmptiableEmptyFunc) appendCall(r0 EmptiableEmptyFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of EmptiableEmptyFuncCall objects describing
// the invocations of this function.
func (f *EmptiableEmptyFunc) History() []EmptiableEmptyFuncCall {
	f.mutex.Lock()
	history := make([]EmptiableEmptyFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// EmptiableEmptyFuncCall is an object that describes an invocation of
// method Empty on an instance of MockEmptiable.
type EmptiableEmptyFuncCall struct {
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 bool
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c EmptiableEmptyFuncCall) Args() []interface{} {
	return []interface{}{}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c EmptiableEmptyFuncCall) Results() []interface{} {
	return []interface{}{c.Result0}
}
