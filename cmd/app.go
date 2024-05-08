package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/chzyer/readline"

	"github.com/leonardinius/golox/internal/interpreter"
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

func (app *LoxApp) reportError(err error) {
	fmt.Fprintln(os.Stderr, err)
	app.err = err
}

func (app *LoxApp) Main(args []string) int {

	var err error
	switch len(args) {
	case 1:
		err = app.runFile(args[0])
	case 0:
		err = app.runPrompt()
	default:
		err = fmt.Errorf("Usage: golox [script]")
	}

	if err != nil {
		app.reportError(err)
	}

	if app.err != nil {
		return 64
	}

	return 0
}

func (app *LoxApp) resetError() {
	app.err = nil
}

func (app *LoxApp) runPrompt() error {
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

		err = app.run(line)
		if err != nil {
			app.reportError(err)
			app.resetError()
		}
	}
}

func (app *LoxApp) runFile(scriptPath string) error {
	bytes, err := os.ReadFile(scriptPath)
	if err != nil {
		return err
	}

	return app.run(string(bytes))
}

func (app *LoxApp) run(input string) error {
	s := scanner.NewScanner(input)

	tokens, err := s.Scan()
	if err != nil {
		return err
	}

	p := parser.NewParser(tokens)
	statements, err := p.Parse()
	if err != nil {
		return err
	}

	return app.interpret(statements)
}

func (app *LoxApp) interpret(statements []parser.Stmt) error {

	if out, err := app.interpeter.Interpret(statements); err != nil {
		return err
	} else {
		fmt.Println(out)
	}

	return nil
}
