// Code generated by go-mockgen 1.1.2; DO NOT EDIT.

package exql

import "sync"

// MockFragment is a mock implementation of the Fragment interface (from the
// package unknwon.dev/norm/internal/exql) used for unit testing.
type MockFragment struct {
	// CompileFunc is an instance of a mock function object controlling the
	// behavior of the method Compile.
	CompileFunc *FragmentCompileFunc
	// HashFunc is an instance of a mock function object controlling the
	// behavior of the method Hash.
	HashFunc *FragmentHashFunc
}

// NewMockFragment creates a new mock of the Fragment interface. All methods
// return zero values for all results, unless overwritten.
func NewMockFragment() *MockFragment {
	return &MockFragment{
		CompileFunc: &FragmentCompileFunc{
			defaultHook: func(*Template) (string, error) {
				return "", nil
			},
		},
		HashFunc: &FragmentHashFunc{
			defaultHook: func() string {
				return ""
			},
		},
	}
}

// NewStrictMockFragment creates a new mock of the Fragment interface. All
// methods panic on invocation, unless overwritten.
func NewStrictMockFragment() *MockFragment {
	return &MockFragment{
		CompileFunc: &FragmentCompileFunc{
			defaultHook: func(*Template) (string, error) {
				panic("unexpected invocation of MockFragment.Compile")
			},
		},
		HashFunc: &FragmentHashFunc{
			defaultHook: func() string {
				panic("unexpected invocation of MockFragment.Hash")
			},
		},
	}
}

// NewMockFragmentFrom creates a new mock of the MockFragment interface. All
// methods delegate to the given implementation, unless overwritten.
func NewMockFragmentFrom(i Fragment) *MockFragment {
	return &MockFragment{
		CompileFunc: &FragmentCompileFunc{
			defaultHook: i.Compile,
		},
		HashFunc: &FragmentHashFunc{
			defaultHook: i.Hash,
		},
	}
}

// FragmentCompileFunc describes the behavior when the Compile method of the
// parent MockFragment instance is invoked.
type FragmentCompileFunc struct {
	defaultHook func(*Template) (string, error)
	hooks       []func(*Template) (string, error)
	history     []FragmentCompileFuncCall
	mutex       sync.Mutex
}

// Compile delegates to the next hook function in the queue and stores the
// parameter and result values of this invocation.
func (m *MockFragment) Compile(v0 *Template) (string, error) {
	r0, r1 := m.CompileFunc.nextHook()(v0)
	m.CompileFunc.appendCall(FragmentCompileFuncCall{v0, r0, r1})
	return r0, r1
}

// SetDefaultHook sets function that is called when the Compile method of
// the parent MockFragment instance is invoked and the hook queue is empty.
func (f *FragmentCompileFunc) SetDefaultHook(hook func(*Template) (string, error)) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// Compile method of the parent MockFragment instance invokes the hook at
// the front of the queue and discards it. After the queue is empty, the
// default hook function is invoked for any future action.
func (f *FragmentCompileFunc) PushHook(hook func(*Template) (string, error)) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultDefaultHook with a function that returns
// the given values.
func (f *FragmentCompileFunc) SetDefaultReturn(r0 string, r1 error) {
	f.SetDefaultHook(func(*Template) (string, error) {
		return r0, r1
	})
}

// PushReturn calls PushDefaultHook with a function that returns the given
// values.
func (f *FragmentCompileFunc) PushReturn(r0 string, r1 error) {
	f.PushHook(func(*Template) (string, error) {
		return r0, r1
	})
}

func (f *FragmentCompileFunc) nextHook() func(*Template) (string, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *FragmentCompileFunc) appendCall(r0 FragmentCompileFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of FragmentCompileFuncCall objects describing
// the invocations of this function.
func (f *FragmentCompileFunc) History() []FragmentCompileFuncCall {
	f.mutex.Lock()
	history := make([]FragmentCompileFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// FragmentCompileFuncCall is an object that describes an invocation of
// method Compile on an instance of MockFragment.
type FragmentCompileFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 *Template
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 string
	// Result1 is the value of the 2nd result returned from this method
	// invocation.
	Result1 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c FragmentCompileFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c FragmentCompileFuncCall) Results() []interface{} {
	return []interface{}{c.Result0, c.Result1}
}

// FragmentHashFunc describes the behavior when the Hash method of the
// parent MockFragment instance is invoked.
type FragmentHashFunc struct {
	defaultHook func() string
	hooks       []func() string
	history     []FragmentHashFuncCall
	mutex       sync.Mutex
}

// Hash delegates to the next hook function in the queue and stores the
// parameter and result values of this invocation.
func (m *MockFragment) Hash() string {
	r0 := m.HashFunc.nextHook()()
	m.HashFunc.appendCall(FragmentHashFuncCall{r0})
	return r0
}

// SetDefaultHook sets function that is called when the Hash method of the
// parent MockFragment instance is invoked and the hook queue is empty.
func (f *FragmentHashFunc) SetDefaultHook(hook func() string) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// Hash method of the parent MockFragment instance invokes the hook at the
// front of the queue and discards it. After the queue is empty, the default
// hook function is invoked for any future action.
func (f *FragmentHashFunc) PushHook(hook func() string) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultDefaultHook with a function that returns
// the given values.
func (f *FragmentHashFunc) SetDefaultReturn(r0 string) {
	f.SetDefaultHook(func() string {
		return r0
	})
}

// PushReturn calls PushDefaultHook with a function that returns the given
// values.
func (f *FragmentHashFunc) PushReturn(r0 string) {
	f.PushHook(func() string {
		return r0
	})
}

func (f *FragmentHashFunc) nextHook() func() string {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *FragmentHashFunc) appendCall(r0 FragmentHashFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of FragmentHashFuncCall objects describing the
// invocations of this function.
func (f *FragmentHashFunc) History() []FragmentHashFuncCall {
	f.mutex.Lock()
	history := make([]FragmentHashFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// FragmentHashFuncCall is an object that describes an invocation of method
// Hash on an instance of MockFragment.
type FragmentHashFuncCall struct {
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 string
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c FragmentHashFuncCall) Args() []interface{} {
	return []interface{}{}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c FragmentHashFuncCall) Results() []interface{} {
	return []interface{}{c.Result0}
}
