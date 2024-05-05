package parser

import "github.com/leonardinius/golox/internal/scanner"

// Visitor is the interface that wraps the Visit method.
//
// Visit is called for every node in the tree.
type Visitor[R any] interface {
	VisitBinary(expr Expr) R
	VisitGrouping(expr Expr) R
	VisitLiteral(expr Expr) R
	VisitUnary(expr Expr) R
}

type Expr interface {
	Accept(v Visitor[any]) any
}

type Binary struct {
	Left     *Expr
	Operator scanner.Token
	Right    *Expr
}

var _ Expr = (*Binary)(nil)

func (e *Binary) Accept(v Visitor[any]) any {
	return v.VisitBinary(e)
}

type Grouping struct {
	Expression *Expr
}

var _ Expr = (*Grouping)(nil)

func (e *Grouping) Accept(v Visitor[any]) any {
	return v.VisitGrouping(e)
}

type Literal struct {
	Value any
}

var _ Expr = (*Literal)(nil)

func (e *Literal) Accept(v Visitor[any]) any {
	return v.VisitLiteral(e)
}

type Unary struct {
	Operator scanner.Token
	Right    *Expr
}

var _ Expr = (*Unary)(nil)

func (e *Unary) Accept(v Visitor[any]) any {
	return v.VisitUnary(e)
}
