package interpreter

import (
	"context"
	"fmt"

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
	SuperClass      *LoxClass

	Name    string
	Methods map[string]*LoxFunction
	Init    *LoxFunction
}

func NewLoxClass(name string, superClass *LoxClass, methods, classMethods map[string]*LoxFunction) *LoxClass {
	metaClass := &LoxClass{Name: name + " metaclass", Methods: classMethods}

	if init, ok := methods["init"]; ok {
		return &LoxClass{Name: name, Methods: methods, Init: init}
	}

	return &LoxClass{Name: name, SuperClass: superClass, Methods: methods, MetaClass: metaClass}
}

// Arity implements Callable.
func (l *LoxClass) Arity() Arity {
	if init := l.FindInit(); init != nil {
		return init.Arity()
	}
	return Arity(0)
}

// Call implements Callable.
func (l *LoxClass) Call(ctx context.Context, interpreter *interpreter, arguments []any) (any, error) {
	newInstance := &objectInstance{Class: l, Fields: make(map[string]any)}
	if init := l.FindInit(); init != nil {
		return init.Bind(newInstance).Call(ctx, interpreter, arguments)
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

	if l.SuperClass != nil {
		if method := l.SuperClass.MetaClass.FindMethod(ctx, name.Lexeme); method != nil {
			boundMethod := method.Bind(l)
			return boundMethod, nil
		}
	}

	return nil, loxerrors.NewRuntimeError(name, loxerrors.ErrRuntimeUndefinedProperty(name.Lexeme))
}

func (l *LoxClass) Set(_ context.Context, name *token.Token, value any) (any, error) {
	if l.MetaClassFields == nil {
		l.MetaClassFields = make(map[string]any)
	}
	l.MetaClassFields[name.Lexeme] = value
	return value, nil
}

func (l *LoxClass) FindMethod(_ context.Context, name string) *LoxFunction {
	cl := l
	for cl != nil {
		if method, ok := cl.Methods[name]; ok {
			return method
		}
		cl = cl.SuperClass
	}

	return nil
}

func (l *LoxClass) FindInit() *LoxFunction {
	cl := l
	for cl != nil {
		if cl.Init != nil {
			return cl.Init
		}
		cl = cl.SuperClass
	}
	return nil
}

// String implements fmt.Stringer.
func (l *LoxClass) String() string {
	return l.Name
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
	return l.Class.Name + " instance"
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
	return value, nil
}

var (
	_ Callable       = (*LoxClass)(nil)
	_ LoxInstance    = (*LoxClass)(nil)
	_ fmt.Stringer   = (*LoxClass)(nil)
	_ fmt.GoStringer = (*LoxClass)(nil)
)

var (
	_ LoxInstance    = (*objectInstance)(nil)
	_ fmt.Stringer   = (*objectInstance)(nil)
	_ fmt.GoStringer = (*objectInstance)(nil)
)
