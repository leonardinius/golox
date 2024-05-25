package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"runtime/pprof"
	"strings"

	"github.com/chzyer/readline"

	"github.com/leonardinius/golox/internal/interpreter"
	"github.com/leonardinius/golox/internal/loxerrors"
	"github.com/leonardinius/golox/internal/parser"
	"github.com/leonardinius/golox/internal/scanner"
)

type LoxApp struct {
	err        error
	interpeter interpreter.Interpreter
}

func NewLoxApp() *LoxApp {
	return &LoxApp{interpeter: interpreter.NewInterpreter()}
}

// ReportPanic implements loxerrors.ErrReporter.
func (app *LoxApp) ReportPanic(err error) {
	app.err = err
	loxerrors.DefaultReportPanic(os.Stderr, err)
}

// ReportError implements loxerrors.ErrReporter.
func (app *LoxApp) ReportError(err error) {
	app.err = err
	loxerrors.DefaultReportError(os.Stderr, err)
}

func (app *LoxApp) Main(args []string) int {
	profile := "default"
	if len(args) > 0 && strings.HasPrefix(args[0], "-profile=") {
		profile = strings.TrimPrefix(args[0], "-profile=")
		args = args[1:]
	}

	var err error
	switch len(args) {
	case 1:
		err = app.runFile(profile, args[0])
	case 0:
		err = app.runPrompt(profile)
	default:
		err = errors.New("Usage: golox [script]")
	}

	if app.err == nil && err != nil {
		app.ReportPanic(err)
	}

	return app.exitcode(app.err)
}

func (app *LoxApp) resetError() {
	app.err = nil
}

func (app *LoxApp) runPrompt(profile string) error {
	rl, err := readline.New("> ")
	if err != nil {
		return err
	}
	defer func() { _ = rl.Close() }()

	for {
		var value any
		line, err := rl.Readline()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}

		value, err = app.run(profile, line)
		if err == nil {
			fmt.Println(value)
		} else {
			app.ReportPanic(err)
			app.resetError()
		}
	}
}

func (app *LoxApp) runFile(profile, scriptPath string) error {
	bytes, err := os.ReadFile(scriptPath) //nolint:gosec // exppected here
	if err != nil {
		return err
	}

	_, err = app.run(profile, string(bytes))
	return err
}

func (app *LoxApp) run(profile, input string) (any, error) {
	s := scanner.NewScanner(input, app)

	tokens, err := s.Scan()
	if err != nil {
		return nil, err
	}

	p := parser.NewParser(tokens, app)
	stmts, err := p.Parse()
	if err != nil {
		return nil, err
	}

	if err := app.resolve(profile, stmts); err != nil {
		return nil, err
	}

	return app.interpret(stmts)
}

func (app *LoxApp) resolve(profile string, stmts []parser.Stmt) error {
	resolver := interpreter.NewResolver(app.interpeter, profile)
	return resolver.Resolve(stmts)
}

func (app *LoxApp) interpret(stmts []parser.Stmt) (any, error) {
	f, e := os.Create("cpuprofile.prof")
	if e != nil {
		panic(e)
	}
	e = pprof.StartCPUProfile(f)
	if e != nil {
		panic(e)
	}
	defer pprof.StopCPUProfile()
	v, e := app.interpeter.Interpret(stmts)
	return v, e
}

func (app *LoxApp) exitcode(err error) int {
	if match, code := app._exitcode(err); match {
		return code
	}

	if iface, ok := err.(interface{ Unwrap() []error }); ok {
		errs := iface.Unwrap()
		for _, err := range errs {
			if match, code := app._exitcode(err); match {
				return code
			}
		}
	}

	_, code := app._exitcode(err)
	return code
}

func (app *LoxApp) _exitcode(err error) (bool, int) {
	if err == nil {
		return false, 0
	}

	switch err.(type) { //nolint:errorlint // exppected here
	case *loxerrors.ParserError, *loxerrors.ScannerError:
		return true, 65
	case *loxerrors.RuntimeError:
		return true, 70
	default:
		return false, 71
	}
}

var _ loxerrors.ErrReporter = (*LoxApp)(nil)
