package loxerrors

import (
	"errors"
	"fmt"

	"github.com/leonardinius/golox/internal/token"
)

var (
	ErrRuntimeOperandMustBeNumber          = errors.New("Operand must be a number.")
	ErrRuntimeOperandsMustBeNumbers        = errors.New("Operands must be numbers.")
	ErrRuntimeOperandsMustNumbersOrStrings = errors.New("Operands must be two numbers or two strings.")
)

type RuntimeError struct {
	tok   *token.Token
	cause error
}

func NewRuntimeError(tok *token.Token, cause error) *RuntimeError {
	return &RuntimeError{tok, cause}
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
