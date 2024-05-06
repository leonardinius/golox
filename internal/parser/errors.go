package parser

import "fmt"

type parseError struct {
	Line    int
	Where   string
	Message string
	Details string
}

func (e *parseError) Error() string {
	details := e.Details
	if details != "" {
		details = " " + details
	}
	return fmt.Sprintf("[line %d] Error%s: %s%s", e.Line, e.Where, e.Message, details)
}

func NewParseError(line int, where string, message, details string) *parseError {
	return &parseError{Line: line, Where: where, Message: message, Details: details}
}
