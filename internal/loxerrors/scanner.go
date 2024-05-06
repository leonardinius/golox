package loxerrors

import (
	"errors"
	"fmt"
)

var (
	ErrScanWrapError           = errors.New("syntax error")
	ErrScanUnexpectedCharacter = errors.New("Unexpected character.")
	ErrScanUnterminatedString  = errors.New("Unterminated string.")
	ErrScanUnterminatedComment = errors.New("Unterminated comment.")
)

func NewScanError(line int, where string, cause error, details string) error {
	if details != "" {
		details = " " + details
	}

	return fmt.Errorf("[line %d] %w%s: %w%s", line, ErrScanWrapError, where, cause, details)
}
