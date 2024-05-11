package loxerrors

import (
	"errors"
	"fmt"

	"github.com/leonardinius/golox/internal/token"
)

var (
	ErrParseUnexpectedToken                       = errors.New("expected expression.")
	ErrParseUnexpectedVariableName                = errors.New("expect variable name.")
	ErrParseInvalidAssignmentTarget               = errors.New("invalid assignment target.")
	ErrParseExpectedRightParenToken               = errors.New("expected ')' after expression.")
	ErrParseExpectedLeftParentIfToken             = errors.New("expected '(' after if.")
	ErrParseExpectedRightParentIfToken            = errors.New("expected ')' after if condition.")
	ErrParseExpectedLeftParentWhileToken          = errors.New("expected '(' after while.")
	ErrParseExpectedRightParentWhileToken         = errors.New("expected ')' after condition.")
	ErrParseExpectedLeftParentForToken            = errors.New("expected '(' after for.")
	ErrParseExpectedRightParentForToken           = errors.New("expected ')' after for clauses.")
	ErrParseExpectedRightCurlyBlockToken          = errors.New("expect '}' after block.")
	ErrParseExpectedSemicolonTokenAfterPrintValue = errors.New("expect ';' after print value.")
	ErrParseExpectedSemicolonTokenAfterExpr       = errors.New("expect ';' after value.")
	ErrParseExpectedSemicolonTokenAfterVar        = errors.New("expect ';' after variable declaration.")
	ErrParseExpectedSemicolonAfterForLoopCond     = errors.New("expect ';' after loop condition.")
	ErrParseExpectedSemicolonTokenAfterBreak      = errors.New("expect ';' after 'break'.")
	ErrParseExpectedSemicolonTokenAfterContinue   = errors.New("expect ';' after 'continue'.")
	ErrParseBreakOutsideLoop                      = errors.New("must be inside a loop to use 'break'.")
	ErrParseContinueOutsideLoop                   = errors.New("must be inside a loop to use 'continue'.")
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
