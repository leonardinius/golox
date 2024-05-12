package interpreter

import (
	"context"
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
