package scanner

import "fmt"

type scanError struct {
	Line  int
	Where string
}

func (e *scanError) Error() string {
	return fmt.Sprintf("[line %d] syntax error%s", e.Line, e.Where)
}

func NewScanError(line int, where string, message error, details string) error {
	if details != "" {
		details = " " + details
	}

	se := &scanError{Line: line, Where: where}

	return fmt.Errorf("%w: %w%s", se, message, details)
}
