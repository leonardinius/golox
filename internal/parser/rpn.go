package parser

import (
	"fmt"
	"strings"

	"github.com/leonardinius/golox/internal/grammar"
)

type RPNPrinter struct{}

func NewRPNPrinter() *RPNPrinter {
	return &RPNPrinter{}
}

// VisitBinary implements Visitor.
func (p *RPNPrinter) VisitBinary(expr *Binary) any {
	return p.reverse(expr.Operator.Lexeme, expr.Left, expr.Right)
}

// VisitGrouping implements Visitor.
func (p *RPNPrinter) VisitGrouping(expr *Grouping) any {
	return p.reverse("", expr.Expression)
}

// VisitLiteral implements Visitor.
func (p *RPNPrinter) VisitLiteral(expr *Literal) any {
	return fmt.Sprintf("%v", expr.Value)
}

// VisitUnary implements Visitor.
func (p *RPNPrinter) VisitUnary(expr *Unary) any {
	operator := expr.Operator.Lexeme
	if expr.Operator.Type == grammar.MINUS {
		operator = "~"
	}
	return p.reverse(operator, expr.Right)
}

func (p *RPNPrinter) reverse(name string, exprs ...Expr) string {
	out := new(strings.Builder)
	for _, expr := range exprs {
		_, _ = out.WriteString(fmt.Sprintf("%v", expr.Accept(p)))
		_, _ = out.WriteString(" ")
	}
	_, _ = out.WriteString(name)
	v := out.String()
	return strings.TrimSuffix(v, " ")
}

func (p *RPNPrinter) Print(expr Expr) string {
	return fmt.Sprintf("%v", expr.Accept(p))
}

var _ Visitor = (*RPNPrinter)(nil)
