package interpreter

import (
	"fmt"
	"strconv"
)

type Arity int

const ArityVarArgs = Arity(-1)

func (a Arity) IsVarArgs() bool {
	return a == ArityVarArgs
}

func (a Arity) String() string {
	if a.IsVarArgs() {
		return "[...]"
	}
	return strconv.Itoa(int(a))
}

type Callable interface {
	Arity() Arity
	Call(interpreter *interpreter, arguments []Value) (Value, error)
}

// ========  ========  ========  ========  ========  ========  ========

type (
	NativeFunctionVarArgs func(interpeter *interpreter, args ...Value) (Value, error)
	NativeFunction0       func(interpeter *interpreter) (Value, error)
	NativeFunction1       func(interpeter *interpreter, arg1 Value) (Value, error)
	NativeFunction2       func(interpeter *interpreter, arg1, arg2 Value) (Value, error)
	NativeFunction3       func(interpeter *interpreter, arg1, arg2, arg3 Value) (Value, error)
	NativeFunction4       func(interpeter *interpreter, arg1, arg2, arg3, arg4 Value) (Value, error)
	NativeFunction5       func(interpeter *interpreter, arg1, arg2, arg3, arg4, arg5 Value) (Value, error)
	nativeFunctionN       struct {
		arity Arity
		fn    func(interpeter *interpreter, args ...Value) (Value, error)
	}
)

// Arity implements Callable.
func (n NativeFunctionVarArgs) Arity() Arity {
	return ArityVarArgs
}

// Call implements Callable.
func (n NativeFunctionVarArgs) Call(interpreter *interpreter, arguments []Value) (Value, error) {
	return n(interpreter, arguments...)
}

// String implements fmt.Stringer.
func (n NativeFunctionVarArgs) String() string {
	return nativeName()
}

// GoString implements fmt.GoStringer.
func (n NativeFunctionVarArgs) GoString() string {
	return n.String()
}

// Arity implements Callable.
func (n NativeFunction0) Arity() Arity {
	return 0
}

// Call implements Callable.
func (n NativeFunction0) Call(interpreter *interpreter, arguments []Value) (Value, error) {
	return n(interpreter)
}

// String implements fmt.Stringer.
func (n NativeFunction0) String() string {
	return nativeName()
}

// GoString implements fmt.GoStringer.
func (n NativeFunction0) GoString() string {
	return n.String()
}

// Arity implements Callable.
func (n NativeFunction1) Arity() Arity {
	return 1
}

// Call implements Callable.
func (n NativeFunction1) Call(interpreter *interpreter, arguments []Value) (Value, error) {
	return n(interpreter, arguments[0])
}

// String implements fmt.Stringer.
func (n NativeFunction1) String() string {
	return nativeName()
}

// GoString implements fmt.GoStringer.
func (n NativeFunction1) GoString() string {
	return n.String()
}

// Arity implements Callable.
func (n NativeFunction2) Arity() Arity {
	return 2
}

// Call implements Callable.
func (n NativeFunction2) Call(interpreter *interpreter, arguments []Value) (Value, error) {
	return n(interpreter, arguments[0], arguments[1])
}

// String implements fmt.Stringer.
func (n NativeFunction2) String() string {
	return nativeName()
}

// GoString implements fmt.GoStringer.
func (n NativeFunction2) GoString() string {
	return n.String()
}

// Arity implements Callable.
func (n NativeFunction3) Arity() Arity {
	return 3
}

// Call implements Callable.
func (n NativeFunction3) Call(interpreter *interpreter, arguments []Value) (Value, error) {
	return n(interpreter, arguments[0], arguments[1], arguments[2])
}

// String implements fmt.Stringer.
func (n NativeFunction3) String() string {
	return nativeName()
}

// GoString implements fmt.GoStringer.
func (n NativeFunction3) GoString() string {
	return n.String()
}

// Arity implements Callable.
func (n NativeFunction4) Arity() Arity {
	return 4
}

// Call implements Callable.
func (n NativeFunction4) Call(interpreter *interpreter, arguments []Value) (Value, error) {
	return n(interpreter, arguments[0], arguments[1], arguments[2], arguments[3])
}

// String implements fmt.Stringer.
func (n NativeFunction4) String() string {
	return nativeName()
}

// GoString implements fmt.GoStringer.
func (n NativeFunction4) GoString() string {
	return n.String()
}

// Arity implements Callable.
func (n NativeFunction5) Arity() Arity {
	return 5
}

// Call implements Callable.
func (n NativeFunction5) Call(interpreter *interpreter, arguments []Value) (Value, error) {
	return n(interpreter, arguments[0], arguments[1], arguments[2], arguments[3], arguments[4])
}

// String implements fmt.Stringer.
func (n NativeFunction5) String() string {
	return nativeName()
}

// GoString implements fmt.GoStringer.
func (n NativeFunction5) GoString() string {
	return n.String()
}

// Arity implements Callable.
func (n *nativeFunctionN) Arity() Arity {
	return n.arity
}

// Call implements Callable.
func (n *nativeFunctionN) Call(interpreter *interpreter, arguments []Value) (Value, error) {
	return n.fn(interpreter, arguments...)
}

// String implements fmt.Stringer.
func (n *nativeFunctionN) String() string {
	return nativeName()
}

// GoString implements fmt.GoStringer.
func (n *nativeFunctionN) GoString() string {
	return n.String()
}

var (
	_ Callable       = NativeFunctionVarArgs(nil)
	_ fmt.GoStringer = NativeFunctionVarArgs(nil)
	_ fmt.Stringer   = NativeFunctionVarArgs(nil)
	_ Callable       = NativeFunction0(nil)
	_ fmt.Stringer   = NativeFunction0(nil)
	_ fmt.GoStringer = NativeFunction0(nil)
	_ Callable       = NativeFunction1(nil)
	_ fmt.Stringer   = NativeFunction1(nil)
	_ fmt.GoStringer = NativeFunction1(nil)
	_ Callable       = NativeFunction2(nil)
	_ fmt.Stringer   = NativeFunction2(nil)
	_ fmt.GoStringer = NativeFunction2(nil)
	_ Callable       = NativeFunction3(nil)
	_ fmt.Stringer   = NativeFunction3(nil)
	_ fmt.GoStringer = NativeFunction3(nil)
	_ Callable       = NativeFunction4(nil)
	_ fmt.Stringer   = NativeFunction4(nil)
	_ fmt.GoStringer = NativeFunction4(nil)
	_ Callable       = NativeFunction5(nil)
	_ fmt.Stringer   = NativeFunction5(nil)
	_ fmt.GoStringer = NativeFunction5(nil)
	_ Callable       = (*nativeFunctionN)(nil)
	_ fmt.Stringer   = (*nativeFunctionN)(nil)
	_ fmt.GoStringer = (*nativeFunctionN)(nil)
)

func nativeName() string {
	return "<native fn>"
}
