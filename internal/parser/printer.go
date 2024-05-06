package parser

import (
	"fmt"
	"strings"
)

type DebugPrinter struct{}

func NewDebugPrintVisitor() *DebugPrinter {
	return &DebugPrinter{}
}

// VisitBinary implements Visitor.
func (p *DebugPrinter) VisitBinary(expr *Binary) any {
	return p.parenthesize(expr.Operator.Lexeme, expr.Left, expr.Right)
}

// VisitGrouping implements Visitor.
func (p *DebugPrinter) VisitGrouping(expr *Grouping) any {
	return p.parenthesize("group", expr.Expression)
}

// VisitLiteral implements Visitor.
func (p *DebugPrinter) VisitLiteral(expr *Literal) any {
	return fmt.Sprintf("%v", expr.Value)
}

// VisitUnary implements Visitor.
func (p *DebugPrinter) VisitUnary(expr *Unary) any {
	return p.parenthesize(expr.Operator.Lexeme, expr.Right)
}

func (p *DebugPrinter) parenthesize(name string, exprs ...Expr) string {
	out := new(strings.Builder)
	_, _ = out.WriteString("(")
	_, _ = out.WriteString(name)
	for _, expr := range exprs {
		_, _ = out.WriteString(" ")
		_, _ = out.WriteString(fmt.Sprintf("%v", expr.Accept(p)))
	}
	_, _ = out.WriteString(")")
	return out.String()
}

func (p *DebugPrinter) Print(expr Expr) string {
	return fmt.Sprintf("%v", expr.Accept(p))
}

var _ Visitor = (*DebugPrinter)(nil)
