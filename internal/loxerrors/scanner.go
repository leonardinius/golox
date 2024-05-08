package loxerrors

import (
	"errors"
	"fmt"
)

var (
	ErrScanUnexpectedCharacter = errors.New("Unexpected character.")
	ErrScanUnterminatedString  = errors.New("Unterminated string.")
	ErrScanUnterminatedComment = errors.New("Unterminated comment.")
)

type ScannerError struct {
	line    int
	cause   error
	details string
}

func NewScanError(line int, cause error, details string) *ScannerError {
	return &ScannerError{line, cause, details}
}

// Error implements error.
func (s *ScannerError) Error() string {
	details := s.details
	if details != "" {
		details = " " + details
	}
	return fmt.Sprintf("[line %d] syntax error: %v%s", s.line, s.cause, details)
}

func (s *ScannerError) Unwrap() error {
	return s.cause
}

var _ error = (*ScannerError)(nil)
var _ unwrapInterface = (*ScannerError)(nil)
