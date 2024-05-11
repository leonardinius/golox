package loxerrors

import (
	"fmt"
	"io"
)

type ErrReporter interface {
	ReportPanic(err error)
	ReportError(err error)
}

type errReporter struct {
	w io.Writer
}

func NewErrReporter(w io.Writer) *errReporter {
	return &errReporter{w: w}
}

// ReportPanic implements ErrReporter.
func (e *errReporter) ReportPanic(err error) {
	DefaultReportPanic(e.w, err)
}

// ReportError implements ErrReporter.
func (e *errReporter) ReportError(err error) {
	DefaultReportError(e.w, err)
}

// DefaultReportPanic is the default implementation of ErrReporter.ReportPanic.
func DefaultReportPanic(w io.Writer, err error) {
	fmt.Fprintf(w, "FATAL %v\n", err)
}

// DefaultReportError is the default implementation of ErrReporter.ReportError.
func DefaultReportError(w io.Writer, err error) {
	fmt.Fprintf(w, "ERROR %v\n", err)
}

var _ ErrReporter = (*errReporter)(nil)
