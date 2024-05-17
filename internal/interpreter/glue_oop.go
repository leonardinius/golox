package interpreter

import (
	"context"
	"fmt"
	"reflect"

	"github.com/leonardinius/golox/internal/loxerrors"
	"github.com/leonardinius/golox/internal/token"
)

type LoxInstance interface {
	Get(ctx context.Context, name *token.Token) (any, error)
	Set(ctx context.Context, name *token.Token, value any) (any, error)
}

type LoxClass struct {
	MetaClass       *LoxClass
	MetaClassFields map[string]any

	Name    string
	Methods map[string]*LoxFunction
	Init    *LoxFunction
}

func NewLoxClass(name string, methods map[string]*LoxFunction, classMethods map[string]*LoxFunction) *LoxClass {
	metaClass := &LoxClass{Name: name + " metaclass", Methods: classMethods}

	if init, ok := methods["init"]; ok {
		return &LoxClass{Name: name, Methods: methods, Init: init}
	}

	return &LoxClass{Name: name, Methods: methods, MetaClass: metaClass}
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
	newInstance := &objectInstance{Class: l, Fields: make(map[string]any)}
	if l.Init != nil {
		return l.Init.Bind(newInstance).Call(ctx, interpreter, arguments)
	}
	return newInstance, nil
}

func (l *LoxClass) Get(ctx context.Context, name *token.Token) (any, error) {
	if value, ok := l.MetaClassFields[name.Lexeme]; ok {
		return value, nil
	}

	if method := l.MetaClass.FindMethod(ctx, name.Lexeme); method != nil {
		boundMethod := method.Bind(l)
		return boundMethod, nil
	}

	return nil, loxerrors.NewRuntimeError(name, loxerrors.ErrRuntimeUndefinedProperty(name.Lexeme))
}

func (l *LoxClass) Set(_ context.Context, name *token.Token, value any) (any, error) {
	if l.MetaClassFields == nil {
		l.MetaClassFields = make(map[string]any)
	}
	l.MetaClassFields[name.Lexeme] = value
	return nil, nil
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

type objectInstance struct {
	Class  *LoxClass
	Fields map[string]any
}

func NewObjectInstance(class *LoxClass) *objectInstance {
	return &objectInstance{Class: class, Fields: make(map[string]any)}
}

// String implements fmt.Stringer.
func (l *objectInstance) String() string {
	ptr := reflect.ValueOf(l).Pointer()
	return fmt.Sprintf("<instance:%s#%d>", l.Class.Name, ptr)
}

// GoString implements fmt.GoStringer.
func (l *objectInstance) GoString() string {
	return l.String()
}

func (l *objectInstance) Get(ctx context.Context, name *token.Token) (any, error) {
	if value, ok := l.Fields[name.Lexeme]; ok {
		return value, nil
	}

	if method := l.Class.FindMethod(ctx, name.Lexeme); method != nil {
		boundMethod := method.Bind(l)
		return boundMethod, nil
	}

	return nil, loxerrors.NewRuntimeError(name, loxerrors.ErrRuntimeUndefinedProperty(name.Lexeme))
}

func (l *objectInstance) Set(_ context.Context, name *token.Token, value any) (any, error) {
	l.Fields[name.Lexeme] = value
	return nil, nil
}

var _ Callable = (*LoxClass)(nil)
var _ LoxInstance = (*LoxClass)(nil)
var _ fmt.Stringer = (*LoxClass)(nil)
var _ fmt.GoStringer = (*LoxClass)(nil)

var _ LoxInstance = (*objectInstance)(nil)
var _ fmt.Stringer = (*objectInstance)(nil)
var _ fmt.GoStringer = (*objectInstance)(nil)
