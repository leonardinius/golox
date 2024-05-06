package parser

import (
	"fmt"
	"strings"
)

type AstPrinter struct{}

func NewAstPrinter() *AstPrinter {
	return &AstPrinter{}
}

// VisitBinary implements Visitor.
func (p *AstPrinter) VisitBinary(expr *Binary) any {
	return p.parenthesize(expr.Operator.Lexeme, expr.Left, expr.Right)
}

// VisitGrouping implements Visitor.
func (p *AstPrinter) VisitGrouping(expr *Grouping) any {
	return p.parenthesize("group", expr.Expression)
}

// VisitLiteral implements Visitor.
func (p *AstPrinter) VisitLiteral(expr *Literal) any {
	return fmt.Sprintf("%v", expr.Value)
}

// VisitUnary implements Visitor.
func (p *AstPrinter) VisitUnary(expr *Unary) any {
	return p.parenthesize(expr.Operator.Lexeme, expr.Right)
}

func (p *AstPrinter) parenthesize(name string, exprs ...Expr) string {
	out := new(strings.Builder)
	_, _ = out.WriteString("(")
	_, _ = out.WriteString(name)
	for _, expr := range exprs {
		_, _ = out.WriteString(" ")
		_, _ = out.WriteString(p.asStr(expr.Accept(p)))
	}
	_, _ = out.WriteString(")")
	return out.String()
}

func (p *AstPrinter) Print(expr Expr) string {
	return p.asStr(expr.Accept(p))
}

func (p *AstPrinter) asStr(v any) string {
	if v == nil {
		return "<nil>"
	}

	return v.(string)
}

var _ Visitor = (*AstPrinter)(nil)
