package loxerrors

import (
	"errors"
	"fmt"

	"github.com/leonardinius/golox/internal/token"
)

var (
	ErrParseExpectedRightParenToken = errors.New("expected ')' after expression")
	ErrParseUnexpectedToken         = errors.New("expected expression")
)

type ParserError struct {
	tok   *token.Token
	cause error
}

func NewParseError(tok *token.Token, cause error) *ParserError {
	return &ParserError{tok: tok, cause: cause}
}

// Error implements error.
func (p *ParserError) Error() string {
	where := "at end"
	if p.tok.Type != token.EOF {
		where = fmt.Sprintf("at '%s'", p.tok.Lexeme)
	}
	return fmt.Sprintf("[line %d] parse error %s: %v", p.tok.Line, where, p.cause)
}

func (p *ParserError) Unwrap() error {
	return p.cause
}

var _ error = (*ParserError)(nil)
var _ unwrapInterface = (*ParserError)(nil)
