package parser_test

import (
	"testing"

	"github.com/leonardinius/golox/internal/parser"
	"github.com/leonardinius/golox/internal/token"
	"github.com/stretchr/testify/assert"
)

func TestAstPrinterVisitor(t *testing.T) {
	var tree parser.Expr = &parser.Binary{
		Left: &parser.Unary{
			Operator: token.NewTokenHeap(token.MINUS, "-", nil, 1),
			Right: &parser.Literal{
				Value: 123,
			},
		},
		Operator: token.NewTokenHeap(token.STAR, "*", nil, 1),
		Right: &parser.Grouping{
			Expression: &parser.Literal{
				Value: 45.67,
			},
		}}

	p := parser.NewAstPrinter()
	out := p.Print(tree)
	assert.Equal(t, "(* (- 123) (group 45.67))", out)
}
