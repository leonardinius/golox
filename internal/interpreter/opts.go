package interpreter

import (
	"io"
	"os"

	"github.com/leonardinius/golox/internal/loxerrors"
)

type interpreterOpts struct {
	env      *environment
	stdin    io.Reader
	stdout   io.Writer
	stderr   io.Writer
	reporter loxerrors.ErrReporter
}

var defaultInterpreterOpts = interpreterOpts{
	stdin:    os.Stdin,
	stdout:   os.Stdout,
	stderr:   os.Stderr,
	reporter: loxerrors.NewErrReporter(os.Stderr),
}

type InterpreterOption func(*interpreterOpts)

func WithEnvironment(env *environment) InterpreterOption {
	return func(opts *interpreterOpts) {
		opts.env = env
	}
}

func WithStdin(stdin io.Reader) InterpreterOption {
	return func(opts *interpreterOpts) {
		opts.stdin = stdin
	}
}

func WithStdout(stdout io.Writer) InterpreterOption {
	return func(opts *interpreterOpts) {
		opts.stdout = stdout
	}
}

func WithStderr(stderr io.Writer) InterpreterOption {
	return func(opts *interpreterOpts) {
		opts.stderr = stderr
	}
}

func WithErrorReporter(r loxerrors.ErrReporter) InterpreterOption {
	return func(opts *interpreterOpts) {
		opts.reporter = r
	}
}

func newInterpreterOpts(options ...InterpreterOption) *interpreterOpts {
	opts := defaultInterpreterOpts
	for _, opt := range options {
		opt(&opts)
	}

	if opts.env == nil {
		opts.env = NewEnvironment()
	}

	return &opts
}
