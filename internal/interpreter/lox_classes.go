package interpreter

import (
	"fmt"

	"github.com/leonardinius/golox/internal/loxerrors"
	"github.com/leonardinius/golox/internal/token"
)

type LoxInstance interface {
	Get(name *token.Token) (any, error)
	Set(name *token.Token, value any) (any, error)
}

type LoxClass struct {
	// Static Class Inheritance. Class prototype.
	// Static methods are stored at MetaClass.Methods
	MetaClass *LoxClass
	// Static fields are stored at MetaClass.MetaClassFields
	MetaClassFields map[string]any

	// Instance Inheritance chain. Superclass.
	SuperClass *LoxClass

	// Class name
	Name string

	// Class methods
	Methods map[string]*LoxFunction
	// Constructor method
	Init *LoxFunction
}

func NewLoxClass(name string, superClass *LoxClass, methods, classMethods map[string]*LoxFunction) *LoxClass {
	metaClass := &LoxClass{Name: name + " metaclass", Methods: classMethods}

	if init, ok := methods["init"]; ok {
		return &LoxClass{Name: name, SuperClass: superClass, Methods: methods, Init: init}
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
func (l *LoxClass) Call(interpreter *interpreter, arguments []any) (any, error) {
	newInstance := &objectInstance{Class: l, Fields: make(map[string]any)}
	if init := l.FindInit(); init != nil {
		return init.Bind(newInstance).Call(interpreter, arguments)
	}
	return newInstance, nil
}

func (l *LoxClass) Get(name *token.Token) (any, error) {
	if value, ok := l.MetaClassFields[name.Lexeme]; ok {
		return value, nil
	}

	if method := l.MetaClass.FindMethod(name.Lexeme); method != nil {
		boundMethod := method.Bind(l)
		return boundMethod, nil
	}

	if l.SuperClass != nil {
		if method := l.SuperClass.MetaClass.FindMethod(name.Lexeme); method != nil {
			boundMethod := method.Bind(l)
			return boundMethod, nil
		}
	}

	return nil, loxerrors.NewRuntimeError(name, loxerrors.ErrRuntimeUndefinedProperty(name.Lexeme))
}

func (l *LoxClass) Set(name *token.Token, value any) (any, error) {
	if l.MetaClassFields == nil {
		l.MetaClassFields = make(map[string]any)
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

func (l *objectInstance) Get(name *token.Token) (any, error) {
	if value, ok := l.Fields[name.Lexeme]; ok {
		return value, nil
	}

	if method := l.Class.FindMethod(name.Lexeme); method != nil {
		boundMethod := method.Bind(l)
		return boundMethod, nil
	}

	return nil, loxerrors.NewRuntimeError(name, loxerrors.ErrRuntimeUndefinedProperty(name.Lexeme))
}

func (l *objectInstance) Set(name *token.Token, value any) (any, error) {
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
