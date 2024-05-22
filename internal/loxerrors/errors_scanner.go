package loxerrors

import (
	"errors"
	"fmt"
)

var (
	ErrScanError               = errors.New("scan error.")
	ErrScanUnexpectedCharacter = errors.New("Unexpected character.")
	ErrScanUnterminatedString  = errors.New("Unterminated string.")
	ErrScanUnterminatedComment = errors.New("Unterminated comment.")
)

type ScannerError struct {
	line  int
	cause error
}

func NewScanError(line int, cause error) error {
	return &ScannerError{line, cause}
}

// Error implements error.
func (s *ScannerError) Error() string {
	return fmt.Sprintf("[line %d] Error: %v", s.line, s.cause)
}

func (s *ScannerError) Unwrap() error {
	return s.cause
}

var _ error = (*ScannerError)(nil)
var _ unwrapInterface = (*ScannerError)(nil)
