package parser

import "fmt"

type parseError struct {
	Line    int
	Where   string
	Message string
	Cause   error
}

func (e *parseError) Error() string {
	return fmt.Sprintf("[line %d] Error%s: %s", e.Line, e.Where, e.Message)
}

func NewParseError(line int, where string, message string, err error) error {
	pe := &parseError{Line: line, Where: where, Message: message, Cause: err}
	return fmt.Errorf("%w - %w", pe, err)
}
