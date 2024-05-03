package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/chzyer/readline"

	"github.com/leonardinius/golox/internal/scanner"
)

func main() {
	if err := Main(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(64)
	}
	os.Exit(0)
}

func Main(args []string) error {

	switch len(args) {
	case 1:
		return runFile(args[0])
	case 0:
		return runPrompt()
	default:
		return fmt.Errorf("Usage: golox [script]")
	}

}

func runPrompt() error {
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

		err = run(line)
		if err != nil {
			fmt.Printf("[ERROR] eval: %v\n", err)
		}
	}
}

func runFile(scriptPath string) error {
	bytes, err := os.ReadFile(scriptPath)
	if err != nil {
		return err
	}

	return run(string(bytes))
}

func run(input string) error {
	s := scanner.NewScanner(input)

	tokens, err := s.Scan()
	if err != nil {
		return err
	}

	for _, t := range tokens {
		fmt.Println(t)
	}

	return nil
}
