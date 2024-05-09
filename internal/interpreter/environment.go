package interpreter

import (
	"fmt"

	"github.com/leonardinius/golox/internal/loxerrors"
	"github.com/leonardinius/golox/internal/token"
)

type environment struct {
	parent *environment
	values map[string]interface{}
}

func newEnvironment() *environment {
	return &environment{
		values: make(map[string]interface{}),
	}
}

func (e *environment) Define(name string, value any) {
	if _, ok := e.values[name]; ok || e.parent == nil {
		e.values[name] = value
		return
	}

	if e.parent != nil {
		e.parent.Define(name, value)
	}
}

func (e *environment) Assign(name *token.Token, value any) error {
	if _, ok := e.values[name.Lexeme]; ok {
		e.values[name.Lexeme] = value
		return nil
	}

	if e.parent != nil {
		return e.parent.Assign(name, value)
	}

	return e.undefinedVariable(name)
}

func (e *environment) Get(name *token.Token) (any, error) {
	if value, ok := e.values[name.Lexeme]; ok {
		return value, nil
	}

	if e.parent != nil {
		return e.parent.Get(name)
	}

	return nil, e.undefinedVariable(name)
}

func (e *environment) undefinedVariable(name *token.Token) error {
	err := fmt.Errorf("%w '%s'.", loxerrors.ErrRuntimeUndefinedVariable, name.Lexeme)
	return loxerrors.NewRuntimeError(name, err)
}
