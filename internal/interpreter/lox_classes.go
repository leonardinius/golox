package interpreter

import (
	"fmt"

	"github.com/leonardinius/golox/internal/loxerrors"
	"github.com/leonardinius/golox/internal/token"
)

type LoxObject interface {
	Get(name *token.Token) (Value, error)
	Set(name *token.Token, value Value) (Value, error)
}

type LoxClass struct {
	MetaClass       *LoxClass
	MetaClassFields map[string]Value
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
func (l *LoxClass) Call(interpreter *interpreter, arguments []Value) (Value, error) {
	newInstance := &objectInstance{Class: l, Fields: make(map[string]Value)}
	if init := l.FindInit(); init != nil {
		return init.Bind(newInstance).Call(interpreter, arguments)
	}
	return ValueObject{newInstance}, nil
}

func (l *LoxClass) Get(name *token.Token) (Value, error) {
	if value, ok := l.MetaClassFields[name.Lexeme]; ok {
		return value, nil
	}

	if method := l.MetaClass.FindMethod(name.Lexeme); method != nil {
		boundMethod := method.Bind(l)
		return ValueCallable{boundMethod}, nil
	}

	if l.SuperClass != nil {
		if method := l.SuperClass.MetaClass.FindMethod(name.Lexeme); method != nil {
			boundMethod := method.Bind(l)
			return ValueCallable{boundMethod}, nil
		}
	}

	return NilValue, loxerrors.NewRuntimeError(name, loxerrors.ErrRuntimeUndefinedProperty(name.Lexeme))
}

func (l *LoxClass) Set(name *token.Token, value Value) (Value, error) {
	if l.MetaClassFields == nil {
		l.MetaClassFields = make(map[string]Value)
	}
	l.MetaClassFields[name.Lexeme] = value
	return value, nil
}

func (l *LoxClass) FindMethod(name string) *LoxFunction {
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
	Fields map[string]Value
}

func NewObjectInstance(class *LoxClass) *objectInstance {
	return &objectInstance{Class: class, Fields: make(map[string]Value)}
}

// String implements fmt.Stringer.
func (l *objectInstance) String() string {
	return l.Class.Name + " instance"
}

// GoString implements fmt.GoStringer.
func (l *objectInstance) GoString() string {
	return l.String()
}

func (l *objectInstance) Get(name *token.Token) (Value, error) {
	if value, ok := l.Fields[name.Lexeme]; ok {
		return value, nil
	}

	if method := l.Class.FindMethod(name.Lexeme); method != nil {
		boundMethod := method.Bind(l)
		return ValueCallable{boundMethod}, nil
	}

	return NilValue, loxerrors.NewRuntimeError(name, loxerrors.ErrRuntimeUndefinedProperty(name.Lexeme))
}

func (l *objectInstance) Set(name *token.Token, value Value) (Value, error) {
	l.Fields[name.Lexeme] = value
	return value, nil
}

var (
	_ Callable       = (*LoxClass)(nil)
	_ LoxObject      = (*LoxClass)(nil)
	_ fmt.Stringer   = (*LoxClass)(nil)
	_ fmt.GoStringer = (*LoxClass)(nil)
)

var (
	_ LoxObject      = (*objectInstance)(nil)
	_ fmt.Stringer   = (*objectInstance)(nil)
	_ fmt.GoStringer = (*objectInstance)(nil)
)
