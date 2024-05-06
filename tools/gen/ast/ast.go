package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func main() {
	os.Exit(Main(os.Args[1:]))
}

func Main(args []string) int {
	if len(args) != 2 {
		fmt.Printf("Usage: ast <file.go> <package>\n")
		return 1
	}

	outFile := args[0]
	packageName := args[1]

	if err := defineAst(outFile, packageName,
		"Binary   : Left Expr, Operator scanner.Token, Right Expr",
		"Grouping : Expression Expr",
		"Literal  : Value any",
		"Unary    : Operator scanner.Token, Right Expr",
	); err != nil {
		fmt.Printf("Error: %v", err)
		return 1
	}

	return 0
}

func defineAst(outFile, packageName string, types ...string) error {
	f, err := os.Create(outFile)
	defer func() { _ = f.Close() }()

	fprintfln := func(message string, args ...any) {
		if err == nil {
			_, err = fmt.Fprintf(f, message+"\n", args...)
		}
	}

	fprintfln("package %s\n", packageName)
	fprintfln("import %s\n", strconv.Quote("github.com/leonardinius/golox/internal/scanner"))

	fprintfln("// Visitor is the interface that wraps the Visit method.")
	fprintfln("//")
	fprintfln("// Visit is called for every node in the tree.")
	fprintfln("type Visitor interface {")
	for _, typeDef := range types {
		exprClassName := strings.TrimSpace(strings.Split(typeDef, ":")[0])
		fprintfln("\tVisit%s(expr *%s) any", exprClassName, exprClassName)
	}
	fprintfln("}\n")

	fprintfln("type Expr interface{")
	fprintfln("\tAccept(v Visitor) any")
	fprintfln("}\n")

	for _, typeDef := range types {
		exprClassName := strings.TrimSpace(strings.Split(typeDef, ":")[0])
		fields := strings.Split(strings.TrimSpace(strings.Split(typeDef, ":")[1]), ",")
		for i, field := range fields {
			fields[i] = strings.TrimSpace(field)
		}

		defineType(fprintfln, exprClassName, fields)
	}

	return err
}

func defineType(fprintf func(message string, args ...any), exprClassName string, fields []string) {
	fprintf("type %s struct {", exprClassName)
	for _, field := range fields {
		fprintf("\t%s", field)
	}
	fprintf("}\n")

	fprintf("var _ Expr = (*%s)(nil)\n", exprClassName)

	fprintf("func (e *%s) Accept(v Visitor) any {", exprClassName)
	fprintf("\treturn v.Visit%s(e)", exprClassName)
	fprintf("}\n")
}