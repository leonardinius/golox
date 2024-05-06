package parser_test

import (
	"testing"

	"github.com/leonardinius/golox/internal/grammar"
	"github.com/leonardinius/golox/internal/parser"
	"github.com/leonardinius/golox/internal/scanner"
	"github.com/stretchr/testify/assert"
)

func TestDebugPrintVisitor(t *testing.T) {
	var tree parser.Expr = &parser.Binary{
		Left: &parser.Unary{
			Operator: scanner.NewToken(grammar.MINUS, "-", nil, 1),
			Right: &parser.Literal{
				Value: 123,
			},
		},
		Operator: scanner.NewToken(grammar.STAR, "*", nil, 1),
		Right: &parser.Grouping{
			Expression: &parser.Literal{
				Value: 45.67,
			},
		}}

	p := parser.NewDebugPrintVisitor()
	out := p.Print(tree)
	assert.Equal(t, "(* (- 123) (group 45.67))", out)
}
