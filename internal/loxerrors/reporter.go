package loxerrors

import (
	"fmt"
	"io"
)

type ErrReporter interface {
	ReportError(err error)
	ReportWarning(err error)
}

type errReporter struct {
	w io.Writer
}

func NewErrReporter(w io.Writer) *errReporter {
	return &errReporter{w: w}
}

// ReportError implements ErrReporter.
func (e *errReporter) ReportError(err error) {
	DefaultReportError(e.w, err)
}

// ReportWarning implements ErrReporter.
func (e *errReporter) ReportWarning(err error) {
	DefaultReportWarning(e.w, err)
}

// DefaultReportError is the default implementation of ErrReporter.ReportError.
func DefaultReportError(w io.Writer, err error) {
	fmt.Fprintf(w, "ERR  %v\n", err)
}

// DefaultReportWarning is the default implementation of ErrReporter.ReportWarning.
func DefaultReportWarning(w io.Writer, err error) {
	fmt.Fprintf(w, "WARN %v\n", err)
}

var _ ErrReporter = (*errReporter)(nil)
