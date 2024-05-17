package interpreter

import (
	"context"
	"fmt"
	"reflect"

	"github.com/leonardinius/golox/internal/loxerrors"
	"github.com/leonardinius/golox/internal/token"
)

type LoxClass struct {
	Name    string
	Methods map[string]*LoxFunction
	Init    *LoxFunction
}

func NewLoxClass(name string, methods map[string]*LoxFunction) *LoxClass {
	if init, ok := methods["init"]; ok {
		return &LoxClass{Name: name, Methods: methods, Init: init}
	}
	return &LoxClass{Name: name, Methods: methods}
}

// Arity implements Callable.
func (l *LoxClass) Arity() Arity {
	if l.Init != nil {
		return l.Init.Arity()
	}
	return Arity(0)
}

// Call implements Callable.
func (l *LoxClass) Call(ctx context.Context, interpreter *interpreter, arguments []any) (any, error) {
	instance := &LoxInstance{Class: l, Fields: make(map[string]any)}
	if l.Init != nil {
		return l.Init.Bind(instance).Call(ctx, interpreter, arguments)
	}
	return instance, nil
}

func (l *LoxClass) FindMethod(_ context.Context, name string) *LoxFunction {
	if method, ok := l.Methods[name]; ok {
		return method
	}

	return nil
}

// String implements fmt.Stringer.
func (l *LoxClass) String() string {
	return fmt.Sprintf("<class:%s/%s>", l.Name, l.Arity())
}

// GoString implements fmt.GoStringer.
func (l *LoxClass) GoString() string {
	return l.String()
}

type LoxInstance struct {
	Class  *LoxClass
	Fields map[string]any
}

// String implements fmt.Stringer.
func (l *LoxInstance) String() string {
	ptr := reflect.ValueOf(l).Pointer()
	return fmt.Sprintf("<instance:%s#%d>", l.Class.Name, ptr)
}

// GoString implements fmt.GoStringer.
func (l *LoxInstance) GoString() string {
	return l.String()
}

// Arity implements Callable.
func (l *LoxInstance) Arity() Arity {
	panic("unimplemented")
}

func (l *LoxInstance) Get(ctx context.Context, name *token.Token) (any, error) {
	if value, ok := l.Fields[name.Lexeme]; ok {
		return value, nil
	}

	if method := l.Class.FindMethod(ctx, name.Lexeme); method != nil {
		boundMethod := method.Bind(l)
		return boundMethod, nil
	}

	return nil, loxerrors.NewRuntimeError(name, loxerrors.ErrRuntimeUndefinedProperty(name.Lexeme))
}

func (l *LoxInstance) Set(_ context.Context, name *token.Token, value any) (any, error) {
	l.Fields[name.Lexeme] = value
	return nil, nil
}

var _ Callable = (*LoxClass)(nil)
var _ fmt.Stringer = (*LoxClass)(nil)
var _ fmt.GoStringer = (*LoxClass)(nil)

var _ fmt.Stringer = (*LoxInstance)(nil)
var _ fmt.GoStringer = (*LoxInstance)(nil)
