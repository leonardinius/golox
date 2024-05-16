package loxerrors

import (
	"errors"
	"fmt"

	"github.com/leonardinius/golox/internal/token"
)

var (
	ErrRuntimeOperandMustBeNumber          = errors.New("operand must be a number.")
	ErrRuntimeOperandsMustBeNumbers        = errors.New("operands must be numbers.")
	ErrRuntimeOperandsMustNumbersOrStrings = errors.New("operands must be two numbers or two strings.")
	ErrRuntimeUndefinedVariable            = errors.New("undefined variable")
	ErrRuntimeCalleeMustBeCallable         = errors.New("can only call functions and classes.")
	ErrRuntimeOnlyInstancesHaveProperties  = errors.New("only instances have properties.")
)

func ErrRuntimeCalleeArityError(expectedArity int, actualArity int) error {
	return fmt.Errorf("expected %d arguments but got %d.", expectedArity, actualArity)
}

func ErrRuntimeUndefinedProperty(name string) error {
	return fmt.Errorf("undefined property %s.", name)
}

func NewRuntimeError(tok *token.Token, cause error) error {
	return &RuntimeError{tok, cause}
}

type RuntimeError struct {
	tok   *token.Token
	cause error
}

// Error implements error.
func (r *RuntimeError) Error() string {
	return fmt.Sprintf("[line %d] at %s: %v", r.tok.Line, r.tok.Lexeme, r.cause)
}

func (r *RuntimeError) Unwrap() error {
	return r.cause
}

var _ error = (*RuntimeError)(nil)
var _ unwrapInterface = (*RuntimeError)(nil)
