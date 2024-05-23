// Code generated by tools/gen/ast. DO NOT EDIT.

package parser

import "github.com/leonardinius/golox/internal/token"

// StmtVisitor is the interface that wraps the Visit method.
type StmtVisitor interface {
	VisitStmtBlock(stmtBlock *StmtBlock) (any, error)
	VisitStmtClass(stmtClass *StmtClass) (any, error)
	VisitStmtExpression(stmtExpression *StmtExpression) (any, error)
	VisitStmtFunction(stmtFunction *StmtFunction) (any, error)
	VisitStmtIf(stmtIf *StmtIf) (any, error)
	VisitStmtPrint(stmtPrint *StmtPrint) (any, error)
	VisitStmtReturn(stmtReturn *StmtReturn) (any, error)
	VisitStmtVar(stmtVar *StmtVar) (any, error)
	VisitStmtWhile(stmtWhile *StmtWhile) (any, error)
	VisitStmtFor(stmtFor *StmtFor) (any, error)
	VisitStmtBreak(stmtBreak *StmtBreak) (any, error)
	VisitStmtContinue(stmtContinue *StmtContinue) (any, error)
}

type Stmt interface {
	Accept(v StmtVisitor) (any, error)
}

type StmtBlock struct {
	Statements []Stmt
}

var _ Stmt = (*StmtBlock)(nil)

func (e *StmtBlock) Accept(v StmtVisitor) (any, error) {
	return v.VisitStmtBlock(e)
}

type StmtClass struct {
	Name         *token.Token
	SuperClass   *ExprVariable
	Methods      []*StmtFunction
	ClassMethods []*StmtFunction
}

var _ Stmt = (*StmtClass)(nil)

func (e *StmtClass) Accept(v StmtVisitor) (any, error) {
	return v.VisitStmtClass(e)
}

type StmtExpression struct {
	Expression Expr
}

var _ Stmt = (*StmtExpression)(nil)

func (e *StmtExpression) Accept(v StmtVisitor) (any, error) {
	return v.VisitStmtExpression(e)
}

type StmtFunction struct {
	Name *token.Token
	Fn   *ExprFunction
}

var _ Stmt = (*StmtFunction)(nil)

func (e *StmtFunction) Accept(v StmtVisitor) (any, error) {
	return v.VisitStmtFunction(e)
}

type StmtIf struct {
	Condition  Expr
	ThenBranch Stmt
	ElseBranch Stmt
}

var _ Stmt = (*StmtIf)(nil)

func (e *StmtIf) Accept(v StmtVisitor) (any, error) {
	return v.VisitStmtIf(e)
}

type StmtPrint struct {
	Expression Expr
}

var _ Stmt = (*StmtPrint)(nil)

func (e *StmtPrint) Accept(v StmtVisitor) (any, error) {
	return v.VisitStmtPrint(e)
}

type StmtReturn struct {
	Keyword *token.Token
	Value   Expr
}

var _ Stmt = (*StmtReturn)(nil)

func (e *StmtReturn) Accept(v StmtVisitor) (any, error) {
	return v.VisitStmtReturn(e)
}

type StmtVar struct {
	Name        *token.Token
	Initializer Expr
}

var _ Stmt = (*StmtVar)(nil)

func (e *StmtVar) Accept(v StmtVisitor) (any, error) {
	return v.VisitStmtVar(e)
}

type StmtWhile struct {
	Condition Expr
	Body      Stmt
}

var _ Stmt = (*StmtWhile)(nil)

func (e *StmtWhile) Accept(v StmtVisitor) (any, error) {
	return v.VisitStmtWhile(e)
}

type StmtFor struct {
	Initializer Stmt
	Condition   Expr
	Increment   Expr
	Body        Stmt
}

var _ Stmt = (*StmtFor)(nil)

func (e *StmtFor) Accept(v StmtVisitor) (any, error) {
	return v.VisitStmtFor(e)
}

type StmtBreak struct {
}

var _ Stmt = (*StmtBreak)(nil)

func (e *StmtBreak) Accept(v StmtVisitor) (any, error) {
	return v.VisitStmtBreak(e)
}

type StmtContinue struct {
}

var _ Stmt = (*StmtContinue)(nil)

func (e *StmtContinue) Accept(v StmtVisitor) (any, error) {
	return v.VisitStmtContinue(e)
}
