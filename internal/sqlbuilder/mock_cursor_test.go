// Code generated by go-mockgen 1.1.2; DO NOT EDIT.

package sqlbuilder

import "sync"

// MockCursor is a mock implementation of the cursor interface (from the
// package unknwon.dev/norm/internal/sqlbuilder) used for unit testing.
type MockCursor struct {
	// CloseFunc is an instance of a mock function object controlling the
	// behavior of the method Close.
	CloseFunc *CursorCloseFunc
	// ColumnsFunc is an instance of a mock function object controlling the
	// behavior of the method Columns.
	ColumnsFunc *CursorColumnsFunc
	// ErrFunc is an instance of a mock function object controlling the
	// behavior of the method Err.
	ErrFunc *CursorErrFunc
	// NextFunc is an instance of a mock function object controlling the
	// behavior of the method Next.
	NextFunc *CursorNextFunc
	// ScanFunc is an instance of a mock function object controlling the
	// behavior of the method Scan.
	ScanFunc *CursorScanFunc
}

// NewMockCursor creates a new mock of the cursor interface. All methods
// return zero values for all results, unless overwritten.
func NewMockCursor() *MockCursor {
	return &MockCursor{
		CloseFunc: &CursorCloseFunc{
			defaultHook: func() error {
				return nil
			},
		},
		ColumnsFunc: &CursorColumnsFunc{
			defaultHook: func() ([]string, error) {
				return nil, nil
			},
		},
		ErrFunc: &CursorErrFunc{
			defaultHook: func() error {
				return nil
			},
		},
		NextFunc: &CursorNextFunc{
			defaultHook: func() bool {
				return false
			},
		},
		ScanFunc: &CursorScanFunc{
			defaultHook: func(...interface{}) error {
				return nil
			},
		},
	}
}

// NewStrictMockCursor creates a new mock of the cursor interface. All
// methods panic on invocation, unless overwritten.
func NewStrictMockCursor() *MockCursor {
	return &MockCursor{
		CloseFunc: &CursorCloseFunc{
			defaultHook: func() error {
				panic("unexpected invocation of MockCursor.Close")
			},
		},
		ColumnsFunc: &CursorColumnsFunc{
			defaultHook: func() ([]string, error) {
				panic("unexpected invocation of MockCursor.Columns")
			},
		},
		ErrFunc: &CursorErrFunc{
			defaultHook: func() error {
				panic("unexpected invocation of MockCursor.Err")
			},
		},
		NextFunc: &CursorNextFunc{
			defaultHook: func() bool {
				panic("unexpected invocation of MockCursor.Next")
			},
		},
		ScanFunc: &CursorScanFunc{
			defaultHook: func(...interface{}) error {
				panic("unexpected invocation of MockCursor.Scan")
			},
		},
	}
}

// surrogateMockCursor is a copy of the cursor interface (from the package
// unknwon.dev/norm/internal/sqlbuilder). It is redefined here as it is
// unexported in the source package.
type surrogateMockCursor interface {
	Close() error
	Columns() ([]string, error)
	Err() error
	Next() bool
	Scan(...interface{}) error
}

// NewMockCursorFrom creates a new mock of the MockCursor interface. All
// methods delegate to the given implementation, unless overwritten.
func NewMockCursorFrom(i surrogateMockCursor) *MockCursor {
	return &MockCursor{
		CloseFunc: &CursorCloseFunc{
			defaultHook: i.Close,
		},
		ColumnsFunc: &CursorColumnsFunc{
			defaultHook: i.Columns,
		},
		ErrFunc: &CursorErrFunc{
			defaultHook: i.Err,
		},
		NextFunc: &CursorNextFunc{
			defaultHook: i.Next,
		},
		ScanFunc: &CursorScanFunc{
			defaultHook: i.Scan,
		},
	}
}

// CursorCloseFunc describes the behavior when the Close method of the
// parent MockCursor instance is invoked.
type CursorCloseFunc struct {
	defaultHook func() error
	hooks       []func() error
	history     []CursorCloseFuncCall
	mutex       sync.Mutex
}

// Close delegates to the next hook function in the queue and stores the
// parameter and result values of this invocation.
func (m *MockCursor) Close() error {
	r0 := m.CloseFunc.nextHook()()
	m.CloseFunc.appendCall(CursorCloseFuncCall{r0})
	return r0
}

// SetDefaultHook sets function that is called when the Close method of the
// parent MockCursor instance is invoked and the hook queue is empty.
func (f *CursorCloseFunc) SetDefaultHook(hook func() error) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// Close method of the parent MockCursor instance invokes the hook at the
// front of the queue and discards it. After the queue is empty, the default
// hook function is invoked for any future action.
func (f *CursorCloseFunc) PushHook(hook func() error) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultDefaultHook with a function that returns
// the given values.
func (f *CursorCloseFunc) SetDefaultReturn(r0 error) {
	f.SetDefaultHook(func() error {
		return r0
	})
}

// PushReturn calls PushDefaultHook with a function that returns the given
// values.
func (f *CursorCloseFunc) PushReturn(r0 error) {
	f.PushHook(func() error {
		return r0
	})
}

func (f *CursorCloseFunc) nextHook() func() error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *CursorCloseFunc) appendCall(r0 CursorCloseFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of CursorCloseFuncCall objects describing the
// invocations of this function.
func (f *CursorCloseFunc) History() []CursorCloseFuncCall {
	f.mutex.Lock()
	history := make([]CursorCloseFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// CursorCloseFuncCall is an object that describes an invocation of method
// Close on an instance of MockCursor.
type CursorCloseFuncCall struct {
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c CursorCloseFuncCall) Args() []interface{} {
	return []interface{}{}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c CursorCloseFuncCall) Results() []interface{} {
	return []interface{}{c.Result0}
}

// CursorColumnsFunc describes the behavior when the Columns method of the
// parent MockCursor instance is invoked.
type CursorColumnsFunc struct {
	defaultHook func() ([]string, error)
	hooks       []func() ([]string, error)
	history     []CursorColumnsFuncCall
	mutex       sync.Mutex
}

// Columns delegates to the next hook function in the queue and stores the
// parameter and result values of this invocation.
func (m *MockCursor) Columns() ([]string, error) {
	r0, r1 := m.ColumnsFunc.nextHook()()
	m.ColumnsFunc.appendCall(CursorColumnsFuncCall{r0, r1})
	return r0, r1
}

// SetDefaultHook sets function that is called when the Columns method of
// the parent MockCursor instance is invoked and the hook queue is empty.
func (f *CursorColumnsFunc) SetDefaultHook(hook func() ([]string, error)) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// Columns method of the parent MockCursor instance invokes the hook at the
// front of the queue and discards it. After the queue is empty, the default
// hook function is invoked for any future action.
func (f *CursorColumnsFunc) PushHook(hook func() ([]string, error)) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultDefaultHook with a function that returns
// the given values.
func (f *CursorColumnsFunc) SetDefaultReturn(r0 []string, r1 error) {
	f.SetDefaultHook(func() ([]string, error) {
		return r0, r1
	})
}

// PushReturn calls PushDefaultHook with a function that returns the given
// values.
func (f *CursorColumnsFunc) PushReturn(r0 []string, r1 error) {
	f.PushHook(func() ([]string, error) {
		return r0, r1
	})
}

func (f *CursorColumnsFunc) nextHook() func() ([]string, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *CursorColumnsFunc) appendCall(r0 CursorColumnsFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of CursorColumnsFuncCall objects describing
// the invocations of this function.
func (f *CursorColumnsFunc) History() []CursorColumnsFuncCall {
	f.mutex.Lock()
	history := make([]CursorColumnsFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// CursorColumnsFuncCall is an object that describes an invocation of method
// Columns on an instance of MockCursor.
type CursorColumnsFuncCall struct {
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 []string
	// Result1 is the value of the 2nd result returned from this method
	// invocation.
	Result1 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c CursorColumnsFuncCall) Args() []interface{} {
	return []interface{}{}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c CursorColumnsFuncCall) Results() []interface{} {
	return []interface{}{c.Result0, c.Result1}
}

// CursorErrFunc describes the behavior when the Err method of the parent
// MockCursor instance is invoked.
type CursorErrFunc struct {
	defaultHook func() error
	hooks       []func() error
	history     []CursorErrFuncCall
	mutex       sync.Mutex
}

// Err delegates to the next hook function in the queue and stores the
// parameter and result values of this invocation.
func (m *MockCursor) Err() error {
	r0 := m.ErrFunc.nextHook()()
	m.ErrFunc.appendCall(CursorErrFuncCall{r0})
	return r0
}

// SetDefaultHook sets function that is called when the Err method of the
// parent MockCursor instance is invoked and the hook queue is empty.
func (f *CursorErrFunc) SetDefaultHook(hook func() error) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// Err method of the parent MockCursor instance invokes the hook at the
// front of the queue and discards it. After the queue is empty, the default
// hook function is invoked for any future action.
func (f *CursorErrFunc) PushHook(hook func() error) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultDefaultHook with a function that returns
// the given values.
func (f *CursorErrFunc) SetDefaultReturn(r0 error) {
	f.SetDefaultHook(func() error {
		return r0
	})
}

// PushReturn calls PushDefaultHook with a function that returns the given
// values.
func (f *CursorErrFunc) PushReturn(r0 error) {
	f.PushHook(func() error {
		return r0
	})
}

func (f *CursorErrFunc) nextHook() func() error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *CursorErrFunc) appendCall(r0 CursorErrFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of CursorErrFuncCall objects describing the
// invocations of this function.
func (f *CursorErrFunc) History() []CursorErrFuncCall {
	f.mutex.Lock()
	history := make([]CursorErrFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// CursorErrFuncCall is an object that describes an invocation of method Err
// on an instance of MockCursor.
type CursorErrFuncCall struct {
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c CursorErrFuncCall) Args() []interface{} {
	return []interface{}{}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c CursorErrFuncCall) Results() []interface{} {
	return []interface{}{c.Result0}
}

// CursorNextFunc describes the behavior when the Next method of the parent
// MockCursor instance is invoked.
type CursorNextFunc struct {
	defaultHook func() bool
	hooks       []func() bool
	history     []CursorNextFuncCall
	mutex       sync.Mutex
}

// Next delegates to the next hook function in the queue and stores the
// parameter and result values of this invocation.
func (m *MockCursor) Next() bool {
	r0 := m.NextFunc.nextHook()()
	m.NextFunc.appendCall(CursorNextFuncCall{r0})
	return r0
}

// SetDefaultHook sets function that is called when the Next method of the
// parent MockCursor instance is invoked and the hook queue is empty.
func (f *CursorNextFunc) SetDefaultHook(hook func() bool) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// Next method of the parent MockCursor instance invokes the hook at the
// front of the queue and discards it. After the queue is empty, the default
// hook function is invoked for any future action.
func (f *CursorNextFunc) PushHook(hook func() bool) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultDefaultHook with a function that returns
// the given values.
func (f *CursorNextFunc) SetDefaultReturn(r0 bool) {
	f.SetDefaultHook(func() bool {
		return r0
	})
}

// PushReturn calls PushDefaultHook with a function that returns the given
// values.
func (f *CursorNextFunc) PushReturn(r0 bool) {
	f.PushHook(func() bool {
		return r0
	})
}

func (f *CursorNextFunc) nextHook() func() bool {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *CursorNextFunc) appendCall(r0 CursorNextFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of CursorNextFuncCall objects describing the
// invocations of this function.
func (f *CursorNextFunc) History() []CursorNextFuncCall {
	f.mutex.Lock()
	history := make([]CursorNextFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// CursorNextFuncCall is an object that describes an invocation of method
// Next on an instance of MockCursor.
type CursorNextFuncCall struct {
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 bool
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c CursorNextFuncCall) Args() []interface{} {
	return []interface{}{}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c CursorNextFuncCall) Results() []interface{} {
	return []interface{}{c.Result0}
}

// CursorScanFunc describes the behavior when the Scan method of the parent
// MockCursor instance is invoked.
type CursorScanFunc struct {
	defaultHook func(...interface{}) error
	hooks       []func(...interface{}) error
	history     []CursorScanFuncCall
	mutex       sync.Mutex
}

// Scan delegates to the next hook function in the queue and stores the
// parameter and result values of this invocation.
func (m *MockCursor) Scan(v0 ...interface{}) error {
	r0 := m.ScanFunc.nextHook()(v0...)
	m.ScanFunc.appendCall(CursorScanFuncCall{v0, r0})
	return r0
}

// SetDefaultHook sets function that is called when the Scan method of the
// parent MockCursor instance is invoked and the hook queue is empty.
func (f *CursorScanFunc) SetDefaultHook(hook func(...interface{}) error) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// Scan method of the parent MockCursor instance invokes the hook at the
// front of the queue and discards it. After the queue is empty, the default
// hook function is invoked for any future action.
func (f *CursorScanFunc) PushHook(hook func(...interface{}) error) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultDefaultHook with a function that returns
// the given values.
func (f *CursorScanFunc) SetDefaultReturn(r0 error) {
	f.SetDefaultHook(func(...interface{}) error {
		return r0
	})
}

// PushReturn calls PushDefaultHook with a function that returns the given
// values.
func (f *CursorScanFunc) PushReturn(r0 error) {
	f.PushHook(func(...interface{}) error {
		return r0
	})
}

func (f *CursorScanFunc) nextHook() func(...interface{}) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *CursorScanFunc) appendCall(r0 CursorScanFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of CursorScanFuncCall objects describing the
// invocations of this function.
func (f *CursorScanFunc) History() []CursorScanFuncCall {
	f.mutex.Lock()
	history := make([]CursorScanFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// CursorScanFuncCall is an object that describes an invocation of method
// Scan on an instance of MockCursor.
type CursorScanFuncCall struct {
	// Arg0 is a slice containing the values of the variadic arguments
	// passed to this method invocation.
	Arg0 []interface{}
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 error
}

// Args returns an interface slice containing the arguments of this
// invocation. The variadic slice argument is flattened in this array such
// that one positional argument and three variadic arguments would result in
// a slice of four, not two.
func (c CursorScanFuncCall) Args() []interface{} {
	trailing := []interface{}{}
	for _, val := range c.Arg0 {
		trailing = append(trailing, val)
	}

	return append([]interface{}{}, trailing...)
}

// Results returns an interface slice containing the results of this
// invocation.
func (c CursorScanFuncCall) Results() []interface{} {
	return []interface{}{c.Result0}
}
