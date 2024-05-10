package loxerrors

import (
	"errors"
	"fmt"

	"github.com/leonardinius/golox/internal/token"
)

var (
	ErrParseExpectedRightParenToken          = errors.New("expected ')' after expression")
	ErrParseExpectedLeftParentWhileToken     = errors.New("expected '(' after while")
	ErrParseExpectedRightParentWhileToken    = errors.New("expected ')' after condition")
	ErrParseExpectedRightCurlyBlockToken     = errors.New("expect '}' after block.")
	ErrParseExpectedSemicolonTokenAfterValue = errors.New("expect ';' after value.")
	ErrParseExpectedSemicolonTokenAfterExpr  = errors.New("expect ';' after value.")
	ErrParseExpectedSemicolonTokenAfterVar   = errors.New("expect ';' after variable declaration.")
	ErrParseUnexpectedToken                  = errors.New("expected expression")
	ErrParseUnexpectedVariableName           = errors.New("expect variable name.")
	ErrParseInvalidAssignmentTarget          = errors.New("invalid assignment target.")
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
