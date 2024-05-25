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
	if len(args) != 3 {
		fmt.Printf("Usage: ast <expressions.go> <statements.go> <package>\n")
		return 1
	}

	expressionsOutFile := args[0]
	statementsOutFile := args[1]
	packageName := args[2]

	if err := defineAst(expressionsOutFile, packageName, "Expr",
		"ExprAssign   : Name *token.Token, Value Expr",
		"ExprBinary   : Left Expr, Operator *token.Token, Right Expr",
		"ExprCall     : Callee Expr, CloseParen *token.Token, Arguments []Expr",
		"ExprFunction : Parameters []*token.Token, Body []Stmt",
		"ExprGet      : Instance Expr, Name *token.Token",
		"ExprGrouping : Expression Expr",
		"ExprLiteral  : Value any",
		"ExprLogical  : Left Expr, Operator *token.Token, Right Expr",
		"ExprSet      : Instance Expr, Name *token.Token, Value Expr",
		"ExprSuper    : Keyword *token.Token, Method *token.Token",
		"ExprThis     : Keyword *token.Token",
		"ExprUnary    : Operator *token.Token, Right Expr",
		"ExprVariable : Name *token.Token",
	); err != nil {
		fmt.Printf("Error: %v", err)
		return 1
	}

	if err := defineAst(statementsOutFile, packageName, "Stmt",
		"StmtBlock      : Statements []Stmt",
		"StmtClass      : Name *token.Token, SuperClass *ExprVariable, Methods []*StmtFunction, ClassMethods []*StmtFunction",
		"StmtExpression : Expression Expr",
		"StmtFunction   : Name *token.Token, Fn *ExprFunction",
		"StmtIf         : Condition Expr, ThenBranch Stmt, ElseBranch Stmt",
		"StmtPrint      : Expression Expr",
		"StmtReturn     : Keyword  *token.Token, Value Expr",
		"StmtVar        : Name *token.Token, Initializer Expr",
		"StmtWhile      : Condition Expr, Body Stmt",
		"StmtFor        : Initializer Stmt, Condition Expr, Increment Expr, Body Stmt",
		"StmtBreak      :",
		"StmtContinue   :",
	); err != nil {
		fmt.Printf("Error: %v", err)
		return 1
	}

	return 0
}

func defineAst(outFile, packageName, baseClass string, types ...string) error {
	f, err := os.Create(outFile)
	defer func() { _ = f.Close() }()

	fprintfln := func(message string, args ...any) {
		if err == nil {
			_, err = fmt.Fprintf(f, message+"\n", args...)
		}
	}

	fprintfln("// Code generated by tools/gen/ast. DO NOT EDIT.\n")
	fprintfln("package %s\n", packageName)

	for _, typeDef := range types {
		if strings.Contains(typeDef, "token.Token") {
			fprintfln("import %s\n", strconv.Quote("github.com/leonardinius/golox/internal/token"))
			break
		}
	}

	fprintfln("// %sVisitor is the interface that wraps the Visit method.", baseClass)
	fprintfln("//")
	fprintfln("type %sVisitor interface {", baseClass)
	for _, typeDef := range types {
		exprClassName := strings.TrimSpace(strings.Split(typeDef, ":")[0])
		fprintfln("\tVisit%s(%s *%s) (Value, error)", exprClassName, varify(exprClassName), exprClassName)
	}
	fprintfln("}\n")

	fprintfln("type %s interface{", baseClass)
	fprintfln("\tAccept(v %sVisitor) (Value, error)", baseClass)
	fprintfln("}\n")

	for _, typeDef := range types {
		exprClassName := strings.TrimSpace(strings.Split(typeDef, ":")[0])
		fields := strings.Split(strings.TrimSpace(strings.Split(typeDef, ":")[1]), ",")
		for i, field := range fields {
			fields[i] = strings.TrimSpace(field)
		}

		defineType(fprintfln, baseClass, exprClassName, fields)
	}

	return err
}

func defineType(fprintf func(message string, args ...any), baseClass, exprClassName string, fields []string) {
	fprintf("type %s struct {", exprClassName)
	for _, field := range fields {
		fprintf("\t%s", field)
	}
	fprintf("}\n")

	fprintf("var _ %s = (*%s)(nil)\n", baseClass, exprClassName)

	fprintf("func (e *%s) Accept(v %sVisitor) (Value, error) {", exprClassName, baseClass)
	fprintf("\treturn v.Visit%s(e)", exprClassName)
	fprintf("}\n")
}

func varify(exprClassName string) string {
	return strings.ToLower(exprClassName[0:1]) + exprClassName[1:]
}
