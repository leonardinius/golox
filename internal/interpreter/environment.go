package interpreter

import (
	"context"
	"fmt"

	"github.com/leonardinius/golox/internal/loxerrors"
	"github.com/leonardinius/golox/internal/token"
)

type environment struct {
	enclosing *environment
	values    map[string]interface{}
}

type envCtxKey struct{}

func NewEnvironment() *environment {
	return &environment{}
}

func EnvFromContext(ctx context.Context) *environment {
	return ctx.Value(envCtxKey{}).(*environment)
}

func (e *environment) Define(name string, value any) {
	if e.values == nil {
		e.values = make(map[string]interface{})
	}
	e.values[name] = value
}

func (e *environment) Assign(name *token.Token, value any) error {
	if _, ok := e.values[name.Lexeme]; ok {
		e.values[name.Lexeme] = value
		return nil
	}

	if e.enclosing != nil {
		return e.enclosing.Assign(name, value)
	}

	return e.undefinedVariable(name)
}

func (e *environment) Get(name *token.Token) (any, error) {
	if value, ok := e.values[name.Lexeme]; ok {
		return value, nil
	}

	if e.enclosing != nil {
		return e.enclosing.Get(name)
	}

	return nil, e.undefinedVariable(name)
}

func (e *environment) Nest() *environment {
	env := NewEnvironment()
	env.enclosing = e
	return env
}

func (e *environment) AsContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, envCtxKey{}, e)
}

func (e *environment) undefinedVariable(name *token.Token) error {
	err := fmt.Errorf("%w '%s'.", loxerrors.ErrRuntimeUndefinedVariable, name.Lexeme)
	return loxerrors.NewRuntimeError(name, err)
}
