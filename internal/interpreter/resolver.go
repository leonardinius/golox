package interpreter

import (
	"container/list"
	"context"
	"errors"
	"fmt"

	"github.com/leonardinius/golox/internal/loxerrors"
	"github.com/leonardinius/golox/internal/parser"
	"github.com/leonardinius/golox/internal/token"
)

type Resolver interface {
	Resolve(ctx context.Context, statements []parser.Stmt) error
}

type resolver struct {
	interpreter *interpreter
	scopes      *list.List
	err         []error
}

// Resolve implements Resolver.
func (r *resolver) Resolve(ctx context.Context, statements []parser.Stmt) error {
	r.err = nil
	r.resolveStmts(ctx, statements)
	return errors.Join(r.err...)
}

func NewResolver(interpreterInstance Interpreter) Resolver {
	interpreterStructPtr, ok := interpreterInstance.(*interpreter)
	if !ok {
		panic(fmt.Errorf("failed to cast interpreter to struct *interpreter"))
	}
	return &resolver{interpreter: interpreterStructPtr, scopes: list.New()}
}

// VisitStmtBlock implements parser.StmtVisitor.
func (r *resolver) VisitStmtBlock(ctx context.Context, stmtBlock *parser.StmtBlock) (any, error) {
	r.beginScope(ctx)
	defer r.endScope(ctx)
	r.resolveStmts(ctx, stmtBlock.Statements)
	return nil, nil
}

// VisitStmtBreak implements parser.StmtVisitor.
func (r *resolver) VisitStmtBreak(ctx context.Context, stmtBreak *parser.StmtBreak) (any, error) {
	return nil, nil
}

// VisitStmtContinue implements parser.StmtVisitor.
func (r *resolver) VisitStmtContinue(ctx context.Context, stmtContinue *parser.StmtContinue) (any, error) {
	return nil, nil
}

// VisitStmtExpression implements parser.StmtVisitor.
func (r *resolver) VisitStmtExpression(ctx context.Context, stmtExpression *parser.StmtExpression) (any, error) {
	r.resolveExpr(ctx, stmtExpression.Expression)
	return nil, nil
}

// VisitStmtFor implements parser.StmtVisitor.
func (r *resolver) VisitStmtFor(ctx context.Context, stmtFor *parser.StmtFor) (any, error) {
	if stmtFor.Initializer != nil {
		r.resolveStmt(ctx, stmtFor.Initializer)
	}
	if stmtFor.Condition != nil {
		r.resolveExpr(ctx, stmtFor.Condition)
	}
	if stmtFor.Increment != nil {
		r.resolveExpr(ctx, stmtFor.Increment)
	}
	r.resolveStmt(ctx, stmtFor.Body)
	return nil, nil
}

// VisitStmtFunction implements parser.StmtVisitor.
func (r *resolver) VisitStmtFunction(ctx context.Context, stmtFunction *parser.StmtFunction) (any, error) {
	r.declare(ctx, stmtFunction.Name)
	r.define(ctx, stmtFunction.Name)

	r.resolveFunction(ctx, stmtFunction.Fn)
	return nil, nil
}

// VisitStmtIf implements parser.StmtVisitor.
func (r *resolver) VisitStmtIf(ctx context.Context, stmtIf *parser.StmtIf) (any, error) {
	r.resolveExpr(ctx, stmtIf.Condition)
	r.resolveStmt(ctx, stmtIf.ThenBranch)
	if stmtIf.ElseBranch != nil {
		r.resolveStmt(ctx, stmtIf.ElseBranch)
	}
	return nil, nil
}

// VisitStmtPrint implements parser.StmtVisitor.
func (r *resolver) VisitStmtPrint(ctx context.Context, stmtPrint *parser.StmtPrint) (any, error) {
	r.resolveExpr(ctx, stmtPrint.Expression)
	return nil, nil
}

// VisitStmtReturn implements parser.StmtVisitor.
func (r *resolver) VisitStmtReturn(ctx context.Context, stmtReturn *parser.StmtReturn) (any, error) {
	if stmtReturn.Value != nil {
		r.resolveExpr(ctx, stmtReturn.Value)
	}
	return nil, nil
}

// VisitStmtVar implements parser.StmtVisitor.
func (r *resolver) VisitStmtVar(ctx context.Context, stmtVar *parser.StmtVar) (any, error) {
	r.declare(ctx, stmtVar.Name)
	if stmtVar.Initializer != nil {
		r.resolveExpr(ctx, stmtVar.Initializer)
	}
	r.define(ctx, stmtVar.Name)
	return nil, nil
}

// VisitStmtWhile implements parser.StmtVisitor.
func (r *resolver) VisitStmtWhile(ctx context.Context, stmtWhile *parser.StmtWhile) (any, error) {
	r.resolveExpr(ctx, stmtWhile.Condition)
	r.resolveStmt(ctx, stmtWhile.Body)
	return nil, nil
}

// VisitExprAssign implements parser.ExprVisitor.
func (r *resolver) VisitExprAssign(ctx context.Context, exprAssign *parser.ExprAssign) (any, error) {
	r.resolveExpr(ctx, exprAssign.Value)
	r.resolveLocal(ctx, exprAssign, exprAssign.Name)
	return nil, nil
}

// VisitExprBinary implements parser.ExprVisitor.
func (r *resolver) VisitExprBinary(ctx context.Context, exprBinary *parser.ExprBinary) (any, error) {
	r.resolveExpr(ctx, exprBinary.Left)
	r.resolveExpr(ctx, exprBinary.Right)
	return nil, nil
}

// VisitExprCall implements parser.ExprVisitor.
func (r *resolver) VisitExprCall(ctx context.Context, exprCall *parser.ExprCall) (any, error) {
	r.resolveExpr(ctx, exprCall.Callee)
	for _, arg := range exprCall.Arguments {
		r.resolveExpr(ctx, arg)
	}
	return nil, nil
}

// VisitExprFunction implements parser.ExprVisitor.
func (r *resolver) VisitExprFunction(ctx context.Context, exprFunction *parser.ExprFunction) (any, error) {
	r.resolveFunction(ctx, exprFunction)
	return nil, nil
}

// VisitExprGrouping implements parser.ExprVisitor.
func (r *resolver) VisitExprGrouping(ctx context.Context, exprGrouping *parser.ExprGrouping) (any, error) {
	r.resolveExpr(ctx, exprGrouping.Expression)
	return nil, nil
}

// VisitExprLiteral implements parser.ExprVisitor.
func (r *resolver) VisitExprLiteral(ctx context.Context, exprLiteral *parser.ExprLiteral) (any, error) {
	return nil, nil
}

// VisitExprLogical implements parser.ExprVisitor.
func (r *resolver) VisitExprLogical(ctx context.Context, exprLogical *parser.ExprLogical) (any, error) {
	r.resolveExpr(ctx, exprLogical.Left)
	r.resolveExpr(ctx, exprLogical.Right)
	return nil, nil
}

// VisitExprUnary implements parser.ExprVisitor.
func (r *resolver) VisitExprUnary(ctx context.Context, exprUnary *parser.ExprUnary) (any, error) {
	r.resolveExpr(ctx, exprUnary.Right)
	return nil, nil
}

// VisitExprVariable implements parser.ExprVisitor.
func (r *resolver) VisitExprVariable(ctx context.Context, exprVariable *parser.ExprVariable) (any, error) {
	var err error
	if defined, ok := r.peekScopeVar(ctx, exprVariable.Name.Lexeme); ok && !defined {
		r.reportError(exprVariable.Name, loxerrors.ErrParseCantInitVarSelfReference)
	}
	r.resolveLocal(ctx, exprVariable, exprVariable.Name)
	return nil, err
}

func (r *resolver) beginScope(_ context.Context) {
	r.debugScopes(">>> scopes")
	r.scopes.PushBack(map[string]bool{})
}

func (r *resolver) endScope(_ context.Context) {
	r.debugScopes("<<< scopes")
	r.scopes.Remove(r.scopes.Back())
}

func (r *resolver) resolveStmts(ctx context.Context, stmts []parser.Stmt) {
	for _, stmt := range stmts {
		r.resolveStmt(ctx, stmt)
	}
}

func (r *resolver) resolveStmt(ctx context.Context, stmt parser.Stmt) {
	_, _ = stmt.Accept(ctx, r)
}

func (r *resolver) resolveExpr(ctx context.Context, expr parser.Expr) {
	_, _ = expr.Accept(ctx, r)
}

func (r *resolver) resolveFunction(ctx context.Context, function *parser.ExprFunction) {
	r.beginScope(ctx)
	defer r.endScope(ctx)

	for _, param := range function.Parameters {
		r.declare(ctx, param)
		r.define(ctx, param)
	}

	r.resolveStmts(ctx, function.Body)
}

func (r *resolver) resolveLocal(ctx context.Context, expr parser.Expr, tok *token.Token) {
	depth := r.scopes.Len()
	back := r.scopes.Back()
	for i := 0; i < depth; i = i + 1 {
		scope := r.scopeFromListElem(back)
		if _, ok := scope[tok.Lexeme]; ok {
			r.interpreter.resolve(ctx, expr, i)
			return
		}
		back = back.Prev()
	}
}

func (r *resolver) declare(ctx context.Context, tok *token.Token) {
	if scope, ok := r.peekScope(ctx); ok {
		if _, ok := scope[tok.Lexeme]; ok {
			r.reportError(tok, loxerrors.ErrParseCantDuplicateVariableDefinition)
		}
		scope[tok.Lexeme] = false
	}
}

func (r *resolver) define(ctx context.Context, tok *token.Token) {
	if scope, ok := r.peekScope(ctx); ok {
		scope[tok.Lexeme] = true
	}
}

func (r *resolver) peekScope(_ context.Context) (map[string]bool, bool) {
	if r.scopes.Len() == 0 {
		return nil, false
	}
	return r.scopeFromListElem(r.scopes.Back()), true
}

func (r *resolver) peekScopeVar(ctx context.Context, name string) (bool, bool) {
	if scope, ok := r.peekScope(ctx); ok {
		if value, ok := scope[name]; ok {
			return value, true
		}
	}
	return false, false
}

func (r *resolver) scopeFromListElem(el *list.Element) map[string]bool {
	return el.Value.(map[string]bool)
}

func (r *resolver) reportError(tok *token.Token, err error) {
	r.err = append(r.err, loxerrors.NewParseError(tok, err))
}

func (r *resolver) debugScopes(_ string) {
	// w := new(strings.Builder)

	// i := 0
	// el := r.scopes.Front()
	// for el != nil {
	// 	_, _ = fmt.Fprintf(w, "%d{%v} ", i, el.Value.(map[string]bool))
	// 	i++
	// 	el = el.Next()
	// }
	// fmt.Printf("%s %s\n", message, w.String())
}

var _ parser.ExprVisitor = (*resolver)(nil)
var _ parser.StmtVisitor = (*resolver)(nil)
var _ Resolver = (*resolver)(nil)