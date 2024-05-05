package scanner

import "fmt"

type scanError struct {
	Line    int
	Where   string
	Message string
	Details string
}

func (e *scanError) Error() string {
	details := e.Details
	if details != "" {
		details = " " + details
	}
	return fmt.Sprintf("[line %d] Error%s: %s%s", e.Line, e.Where, e.Message, details)
}

func NewScanError(line int, where string, message, details string) *scanError {
	return &scanError{Line: line, Where: where, Message: message, Details: details}
}
