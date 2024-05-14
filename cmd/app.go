package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

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
	ctx := context.Background()

	var err error
	switch len(args) {
	case 1:
		err = app.runFile(ctx, args[0])
	case 0:
		err = app.runPrompt(ctx)
	default:
		err = fmt.Errorf("Usage: golox [script]")
	}

	if err != nil {
		app.ReportPanic(err)
	}

	if app.err != nil {
		return 64
	}

	return 0
}

func (app *LoxApp) resetError() {
	app.err = nil
}

func (app *LoxApp) runPrompt(ctx context.Context) error {
	rl, err := readline.New("> ")
	if err != nil {
		return err
	}
	defer rl.Close()

	for {
		line, err := rl.Readline()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}

		err = app.run(ctx, line)
		if err != nil {
			app.ReportPanic(err)
			app.resetError()
		}
	}
}

func (app *LoxApp) runFile(ctx context.Context, scriptPath string) error {
	bytes, err := os.ReadFile(scriptPath)
	if err != nil {
		return err
	}

	return app.run(ctx, string(bytes))
}

func (app *LoxApp) run(ctx context.Context, input string) error {
	s := scanner.NewScanner(input, app)

	tokens, err := s.Scan()
	app.err = err

	p := parser.NewParser(tokens, app)
	stmts, err := p.Parse()
	if err != nil {
		return err
	}

	if err = app.resolve(ctx, stmts); err != nil {
		return err
	}

	return app.interpret(ctx, stmts)
}

func (app *LoxApp) resolve(ctx context.Context, stmts []parser.Stmt) error {
	resolver := interpreter.NewResolver(app.interpeter)
	return resolver.Resolve(ctx, stmts)
}

func (app *LoxApp) interpret(ctx context.Context, stmts []parser.Stmt) error {

	if eval, err := app.interpeter.Interpret(ctx, stmts); err != nil {
		return err
	} else {
		fmt.Println(eval)
	}

	return nil
}

var _ loxerrors.ErrReporter = (*LoxApp)(nil)
