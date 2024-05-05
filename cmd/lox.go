package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/chzyer/readline"

	"github.com/leonardinius/golox/internal/scanner"
)

type LoxApp struct {
	err error
}

func NewLoxApp() *LoxApp {
	return &LoxApp{}
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

	for _, t := range tokens {
		fmt.Printf("%#v\n", t)
	}

	return nil
}
