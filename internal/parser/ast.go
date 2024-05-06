package parser

import "github.com/leonardinius/golox/internal/scanner"

// Visitor is the interface that wraps the Visit method.
//
// Visit is called for every node in the tree.
type Visitor interface {
	VisitBinary(expr *Binary) any
	VisitGrouping(expr *Grouping) any
	VisitLiteral(expr *Literal) any
	VisitUnary(expr *Unary) any
}

type Expr interface {
	Accept(v Visitor) any
}

type Binary struct {
	Left     Expr
	Operator scanner.Token
	Right    Expr
}

var _ Expr = (*Binary)(nil)

func (e *Binary) Accept(v Visitor) any {
	return v.VisitBinary(e)
}

type Grouping struct {
	Expression Expr
}

var _ Expr = (*Grouping)(nil)

func (e *Grouping) Accept(v Visitor) any {
	return v.VisitGrouping(e)
}

type Literal struct {
	Value any
}

var _ Expr = (*Literal)(nil)

func (e *Literal) Accept(v Visitor) any {
	return v.VisitLiteral(e)
}

type Unary struct {
	Operator scanner.Token
	Right    Expr
}

var _ Expr = (*Unary)(nil)

func (e *Unary) Accept(v Visitor) any {
	return v.VisitUnary(e)
}
