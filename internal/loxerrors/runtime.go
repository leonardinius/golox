package loxerrors

import (
	"fmt"

	"github.com/leonardinius/golox/internal/token"
)

type RuntimeError struct {
	tok *token.Token
	msg string
}

func NewRuntimeError(tok *token.Token, msg string) *RuntimeError {
	return &RuntimeError{tok, msg}
}

// Error implements error.
func (r *RuntimeError) Error() string {
	return fmt.Sprintf("[line %d] at %s: %s", r.tok.Line, r.tok.Lexeme, r.msg)
}

var _ error = (*RuntimeError)(nil)
