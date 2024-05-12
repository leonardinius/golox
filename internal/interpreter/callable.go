package interpreter

import (
	"context"
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
	Call(ctx context.Context, interpreter *interpreter, arguments []any) (any, error)
}

// ========  ========  ========  ========  ========  ========  ========

type NativeFunctionVarArgs func(ctx context.Context, interpeter *interpreter, args ...any) (any, error)
type NativeFunction0 func(ctx context.Context, interpeter *interpreter) (any, error)
type NativeFunction1 func(ctx context.Context, interpeter *interpreter, arg1 any) (any, error)
type NativeFunction2 func(ctx context.Context, interpeter *interpreter, arg1, arg2 any) (any, error)
type NativeFunction3 func(ctx context.Context, interpeter *interpreter, arg1, arg2, arg3 any) (any, error)
type NativeFunction4 func(ctx context.Context, interpeter *interpreter, arg1, arg2, arg3, arg4 any) (any, error)
type NativeFunction5 func(ctx context.Context, interpeter *interpreter, arg1, arg2, arg3, arg4, arg5 any) (any, error)
type nativeFunctionN struct {
	arity Arity
	fn    func(ctx context.Context, interpeter *interpreter, args ...any) (any, error)
}

// Arity implements Callable.
func (n NativeFunctionVarArgs) Arity() Arity {
	return ArityVarArgs
}

// Call implements Callable.
func (n NativeFunctionVarArgs) Call(ctx context.Context, interpreter *interpreter, arguments []any) (any, error) {
	return n(ctx, interpreter, arguments...)
}

// String implements fmt.Stringer.
func (n NativeFunctionVarArgs) String() string {
	return nativeName(n.Arity())
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
func (n NativeFunction0) Call(ctx context.Context, interpreter *interpreter, arguments []any) (any, error) {
	return n(ctx, interpreter)
}

// String implements fmt.Stringer.
func (n NativeFunction0) String() string {
	return nativeName(n.Arity())
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
func (n NativeFunction1) Call(ctx context.Context, interpreter *interpreter, arguments []any) (any, error) {
	return n(ctx, interpreter, arguments[0])
}

// String implements fmt.Stringer.
func (n NativeFunction1) String() string {
	return nativeName(n.Arity())
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
func (n NativeFunction2) Call(ctx context.Context, interpreter *interpreter, arguments []any) (any, error) {
	return n(ctx, interpreter, arguments[0], arguments[1])
}

// String implements fmt.Stringer.
func (n NativeFunction2) String() string {
	return nativeName(n.Arity())
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
func (n NativeFunction3) Call(ctx context.Context, interpreter *interpreter, arguments []any) (any, error) {
	return n(ctx, interpreter, arguments[0], arguments[1], arguments[2])
}

// String implements fmt.Stringer.
func (n NativeFunction3) String() string {
	return nativeName(n.Arity())
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
func (n NativeFunction4) Call(ctx context.Context, interpreter *interpreter, arguments []any) (any, error) {
	return n(ctx, interpreter, arguments[0], arguments[1], arguments[2], arguments[3])
}

// String implements fmt.Stringer.
func (n NativeFunction4) String() string {
	return nativeName(n.Arity())
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
func (n NativeFunction5) Call(ctx context.Context, interpreter *interpreter, arguments []any) (any, error) {
	return n(ctx, interpreter, arguments[0], arguments[1], arguments[2], arguments[3], arguments[4])
}

// String implements fmt.Stringer.
func (n NativeFunction5) String() string {
	return nativeName(n.Arity())
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
func (n *nativeFunctionN) Call(ctx context.Context, interpreter *interpreter, arguments []any) (any, error) {
	return n.fn(ctx, interpreter, arguments...)
}

// String implements fmt.Stringer.
func (n *nativeFunctionN) String() string {
	return nativeName(n.Arity())
}

// GoString implements fmt.GoStringer.
func (n *nativeFunctionN) GoString() string {
	return n.String()
}

var _ Callable = (NativeFunctionVarArgs)(nil)
var _ fmt.GoStringer = (NativeFunctionVarArgs)(nil)
var _ fmt.Stringer = (NativeFunctionVarArgs)(nil)
var _ Callable = (NativeFunction0)(nil)
var _ fmt.Stringer = (NativeFunction0)(nil)
var _ fmt.GoStringer = (NativeFunction0)(nil)
var _ Callable = (NativeFunction1)(nil)
var _ fmt.Stringer = (NativeFunction1)(nil)
var _ fmt.GoStringer = (NativeFunction1)(nil)
var _ Callable = (NativeFunction2)(nil)
var _ fmt.Stringer = (NativeFunction2)(nil)
var _ fmt.GoStringer = (NativeFunction2)(nil)
var _ Callable = (NativeFunction3)(nil)
var _ fmt.Stringer = (NativeFunction3)(nil)
var _ fmt.GoStringer = (NativeFunction3)(nil)
var _ Callable = (NativeFunction4)(nil)
var _ fmt.Stringer = (NativeFunction4)(nil)
var _ fmt.GoStringer = (NativeFunction4)(nil)
var _ Callable = (NativeFunction5)(nil)
var _ fmt.Stringer = (NativeFunction5)(nil)
var _ fmt.GoStringer = (NativeFunction5)(nil)
var _ Callable = (*nativeFunctionN)(nil)
var _ fmt.Stringer = (*nativeFunctionN)(nil)
var _ fmt.GoStringer = (*nativeFunctionN)(nil)

func nativeName(arity Arity) string {
	return "<native fn/" + arity.String() + ">"
}
