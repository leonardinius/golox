package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"

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
	r := bufio.NewReader(os.Stdin)
	w := bufio.NewWriter(os.Stdout)
	rw := bufio.NewReadWriter(r, w)

	for {
		fmt.Fprint(rw, "> ")
		rw.Flush()

		var (
			line   []byte
			l      []byte
			prefix bool
			err    error
		)
		for l, prefix, err = rw.ReadLine(); err == nil; l, prefix, err = rw.ReadLine() {
			line = append(line, l...)
			if !prefix {
				break
			}
		}

		// check for EOF
		if err == io.EOF {
			return nil
		}

		if err != nil {
			return fmt.Errorf("error reading line: %w", err)
		}

		if err := run(string(line)); err != nil {
			log.Printf("[ERROR] eval: %v\n", err)
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
