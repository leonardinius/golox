package interpreter

import (
	"fmt"

	"github.com/leonardinius/golox/internal/loxerrors"
	"github.com/leonardinius/golox/internal/token"
)

type environment struct {
	enclosing *environment
	values    map[string]Value
}

func NewEnvironment() *environment {
	return &environment{}
}

func (e *environment) Define(name string, value Value) {
	if e.values == nil {
		e.values = make(map[string]Value)
	}
	e.values[name] = value
}

func (e *environment) Get(name *token.Token) (Value, error) {
	if value, ok := e.values[name.Lexeme]; ok {
		return value, nil
	}

	if e.enclosing != nil {
		return e.enclosing.Get(name)
	}

	return nil, e.undefinedVariable(name)
}

func (e *environment) Assign(name *token.Token, value Value) error {
	if _, ok := e.values[name.Lexeme]; ok {
		e.values[name.Lexeme] = value
		return nil
	}

	if e.enclosing != nil {
		return e.enclosing.Assign(name, value)
	}

	return e.undefinedVariable(name)
}

func (e *environment) GetAt(distance int, name string) (Value, error) {
	depth := e.ancestor(distance)
	if value, ok := depth.values[name]; ok {
		return value, nil
	}

	err := fmt.Errorf("%w '%s'.", loxerrors.ErrRuntimeUndefinedVariable, name)
	return nil, err
}

func (e *environment) AssignAt(distance int, name *token.Token, value Value) (Value, error) {
	depth := e.ancestor(distance)
	if depth.values == nil {
		depth.values = make(map[string]Value)
	}
	depth.values[name.Lexeme] = value

	return value, nil
}

func (e *environment) Nest() *environment {
	env := NewEnvironment()
	env.enclosing = e
	return env
}

func (e *environment) Enclosing() *environment {
	return e.enclosing
}

func (e *environment) ancestor(distance int) *environment {
	self := e
	for distance > 0 {
		self = self.enclosing
		distance--
	}

	return self
}

func (e *environment) undefinedVariable(name *token.Token) error {
	err := fmt.Errorf("%w '%s'.", loxerrors.ErrRuntimeUndefinedVariable, name.Lexeme)
	return loxerrors.NewRuntimeError(name, err)
}

func (e *environment) String() string {
	w := ""

	for self := e; self != nil; self = self.enclosing {
		w += "{"
		for k, v := range self.values {
			w += fmt.Sprintf("%s=%v,", k, v)
		}
		w += "}"
		if self.enclosing != nil {
			w += " -> "
		}
	}

	return w
}

var _ fmt.Stringer = (*environment)(nil)
