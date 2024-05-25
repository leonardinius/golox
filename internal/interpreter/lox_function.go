package interpreter

import (
	"fmt"

	"github.com/leonardinius/golox/internal/parser"
	"github.com/leonardinius/golox/internal/token"
)

type ReturnValueError struct {
	Value Value
}

func (r *ReturnValueError) Error() string {
	return fmt.Sprintf("fatal value: %v", r.Value)
}

type LoxFunction struct {
	Name        *token.Token
	Fn          *parser.ExprFunction
	Env         *environment
	IsIntialize bool
}

func NewLoxFunction(name *token.Token, fn *parser.ExprFunction, env *environment, isInitialize bool) *LoxFunction {
	return &LoxFunction{Name: name, Fn: fn, Env: env, IsIntialize: isInitialize}
}

// Arity implements Callable.
func (l *LoxFunction) Arity() Arity {
	return Arity(len(l.Fn.Parameters))
}

// Call implements Callable.
func (l *LoxFunction) Call(interpreter *interpreter, arguments []Value) (Value, error) {
	env := l.Env.Nest()

	for idx, e := range l.Fn.Parameters {
		env.Define(e.Lexeme, arguments[idx])
	}

	var value Value
	err := interpreter.executeBlock(env, l.Fn.Body)
	if err != nil {
		value, err = l.returnValue(err)
	}
	if err != nil {
		return nil, err
	}
	if l.IsIntialize {
		return l.Env.GetAt(0, "this")
	}
	return value, nil
}

func (l *LoxFunction) Bind(instance LoxObject) *LoxFunction {
	env := l.Env.Nest()
	env.Define("this", ValueObject{instance})
	return NewLoxFunction(l.Name, l.Fn, env, l.IsIntialize)
}

func (l *LoxFunction) returnValue(err error) (Value, error) {
	if ret, ok := err.(*ReturnValueError); ok {
		return ret.Value, nil
	}
	return nil, err
}

// String implements fmt.Stringer.
func (l *LoxFunction) String() string {
	if l.Name == nil {
		return "<fn #anon>"
	}
	return fmt.Sprintf("<fn %s>", l.Name.Lexeme)
}

// GoString implements fmt.GoStringer.
func (l *LoxFunction) GoString() string {
	return l.String()
}

var (
	_ Callable       = (*LoxFunction)(nil)
	_ fmt.Stringer   = (*LoxFunction)(nil)
	_ fmt.GoStringer = (*LoxFunction)(nil)
)
