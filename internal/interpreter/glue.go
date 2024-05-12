package interpreter

import (
	"context"
	"errors"
	"fmt"

	"github.com/leonardinius/golox/internal/parser"
)

type LoxFunction struct {
	Fn *parser.StmtFunction
}

// Arity implements Callable.
func (l *LoxFunction) Arity() Arity {
	return Arity(len(l.Fn.Parameters))
}

// Call implements Callable.
func (l *LoxFunction) Call(ctx context.Context, interpreter *interpreter, arguments []any) (any, error) {
	env := EnvFromContext(ctx).Nest()

	for idx, e := range l.Fn.Parameters {
		env.Define(e.Lexeme, arguments[idx])
	}

	value, err := interpreter.executeBlock(env.AsContext(ctx), l.Fn.Body)
	if err != nil {
		return l.returnValue(err)
	}
	return value, nil
}

func (l *LoxFunction) returnValue(err error) (any, error) {
	var ret *ReturnValue
	if errors.As(err, &ret) {
		return ret.Value, nil
	}
	return nil, err
}

// String implements fmt.Stringer.
func (l *LoxFunction) String() string {
	return fmt.Sprintf("<fn:%s/%s>", l.Fn.Name.Lexeme, l.Arity())
}

// GoString implements fmt.GoStringer.
func (l *LoxFunction) GoString() string {
	return l.String()
}

var _ Callable = (*LoxFunction)(nil)
var _ fmt.Stringer = (*LoxFunction)(nil)
var _ fmt.GoStringer = (*LoxFunction)(nil)

type ReturnValue struct {
	Value any
}

func (r *ReturnValue) Error() string {
	return fmt.Sprintf("fatal value: %v", r.Value)
}
