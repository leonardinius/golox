package loxerrors

import (
	"errors"
	"fmt"
)

var (
	ErrParseWrapError               = errors.New("parse error")
	ErrParseExpectedRightParenToken = errors.New("expected ')' after expression")
	ErrParseUnexpectedToken         = errors.New("expected expression")
)

func NewParseError(line int, where string, cause error) error {
	return fmt.Errorf("[line %d] %w%s: %w", line, ErrParseWrapError, where, cause)
}
