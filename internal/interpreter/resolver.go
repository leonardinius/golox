package interpreter

import (
	"container/list"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/leonardinius/golox/internal/loxerrors"
	"github.com/leonardinius/golox/internal/parser"
	"github.com/leonardinius/golox/internal/token"
)

type Resolver interface {
	Resolve(ctx context.Context, statements []parser.Stmt) error
}

type VarState int

const (
	VarStateDeclared VarState = iota
	VarStateDefined
	VarStateRead
)

type FunctionType int

const (
	FnTypeNone FunctionType = iota
	FnTypeExpr
	FnTypeFunction
	FnTypeMethod
	FnTypeClassMethod
	FnTypeInitializer
)

type ClassType int

const (
	CTypeNone ClassType = iota
	CTypeClass
	CTypeSubclass
)

type ResolverVariable struct {
	Name  *token.Token
	State VarState
}

type resolver struct {
	interpreter     *interpreter
	scopes          *list.List
	err             []error
	currentFunction FunctionType
	currentClass    ClassType
	profile         string
}

var profiles map[string][]error = map[string][]error{
	"default": {},
	"strict":  {},
	"non-strict": {
		loxerrors.ErrParseLocalVariableNotUsed,
	},
}

func NewResolver(interpreterInstance Interpreter, profile string) Resolver {
	interpreterPtr, ok := interpreterInstance.(*interpreter)
	if !ok {
		panic("failed to cast interpreter to struct *interpreter")
	}

	newResolver := &resolver{
		interpreter:     interpreterPtr,
		scopes:          list.New(),
		err:             nil,
		currentFunction: FnTypeNone,
		currentClass:    CTypeNone,
		profile:         profile,
	}

	return newResolver
}

// Resolve implements Resolver.
func (r *resolver) Resolve(ctx context.Context, statements []parser.Stmt) error {
	r.err = nil
	r.resolveStmts(ctx, statements)
	return errors.Join(r.err...)
}

// VisitStmtBlock implements parser.StmtVisitor.
func (r *resolver) VisitStmtBlock(ctx context.Context, stmtBlock *parser.StmtBlock) (any, error) {
	r.beginScope(ctx)
	defer r.endScope(ctx)
	r.resolveStmts(ctx, stmtBlock.Statements)
	return nil, errNilnil
}

// VisitStmtClass implements parser.StmtVisitor.
func (r *resolver) VisitStmtClass(ctx context.Context, stmtClass *parser.StmtClass) (any, error) {
	enclosingClass := r.currentClass
	defer func() { r.currentClass = enclosingClass }()
	r.currentClass = CTypeClass

	r.declare(ctx, stmtClass.Name)
	r.define(ctx, stmtClass.Name)

	if stmtClass.SuperClass != nil && stmtClass.Name.Lexeme == stmtClass.SuperClass.Name.Lexeme {
		r.reportError(stmtClass.SuperClass.Name, loxerrors.ErrParseClassCantInheritFromItself)
	}
	if stmtClass.SuperClass != nil {
		r.currentClass = CTypeSubclass
		r.resolveExpr(ctx, stmtClass.SuperClass)

		r.beginScope(ctx)
		defer r.endScope(ctx)
		r.defineInternal(ctx, "super")
	}

	r.beginScope(ctx)
	defer r.endScope(ctx)

	r.defineInternal(ctx, "this")

	for _, method := range stmtClass.ClassMethods {
		r.resolveFunction(ctx, method.Fn, FnTypeClassMethod)
	}

	for _, method := range stmtClass.Methods {
		functionType := FnTypeMethod
		if method.Name.Lexeme == "init" {
			functionType = FnTypeInitializer
		}
		r.resolveFunction(ctx, method.Fn, functionType)
	}

	return nil, errNilnil
}

// VisitStmtBreak implements parser.StmtVisitor.
func (r *resolver) VisitStmtBreak(ctx context.Context, stmtBreak *parser.StmtBreak) (any, error) {
	return nil, errNilnil
}

// VisitStmtContinue implements parser.StmtVisitor.
func (r *resolver) VisitStmtContinue(ctx context.Context, stmtContinue *parser.StmtContinue) (any, error) {
	return nil, errNilnil
}

// VisitStmtExpression implements parser.StmtVisitor.
func (r *resolver) VisitStmtExpression(ctx context.Context, stmtExpression *parser.StmtExpression) (any, error) {
	r.resolveExpr(ctx, stmtExpression.Expression)
	return nil, errNilnil
}

// VisitStmtFor implements parser.StmtVisitor.
func (r *resolver) VisitStmtFor(ctx context.Context, stmtFor *parser.StmtFor) (any, error) {
	if stmtFor.Initializer != nil {
		r.beginScope(ctx)
		defer r.endScope(ctx)
		r.resolveStmt(ctx, stmtFor.Initializer)
	}
	if stmtFor.Condition != nil {
		r.resolveExpr(ctx, stmtFor.Condition)
	}
	if stmtFor.Increment != nil {
		r.resolveExpr(ctx, stmtFor.Increment)
	}

	r.resolveStmt(ctx, stmtFor.Body)
	return nil, errNilnil
}

// VisitStmtFunction implements parser.StmtVisitor.
func (r *resolver) VisitStmtFunction(ctx context.Context, stmtFunction *parser.StmtFunction) (any, error) {
	r.declare(ctx, stmtFunction.Name)
	r.define(ctx, stmtFunction.Name)

	r.resolveFunction(ctx, stmtFunction.Fn, FnTypeFunction)
	return nil, errNilnil
}

// VisitStmtIf implements parser.StmtVisitor.
func (r *resolver) VisitStmtIf(ctx context.Context, stmtIf *parser.StmtIf) (any, error) {
	r.resolveExpr(ctx, stmtIf.Condition)
	r.resolveStmt(ctx, stmtIf.ThenBranch)
	if stmtIf.ElseBranch != nil {
		r.resolveStmt(ctx, stmtIf.ElseBranch)
	}
	return nil, errNilnil
}

// VisitStmtPrint implements parser.StmtVisitor.
func (r *resolver) VisitStmtPrint(ctx context.Context, stmtPrint *parser.StmtPrint) (any, error) {
	r.resolveExpr(ctx, stmtPrint.Expression)
	return nil, errNilnil
}

// VisitStmtReturn implements parser.StmtVisitor.
func (r *resolver) VisitStmtReturn(ctx context.Context, stmtReturn *parser.StmtReturn) (any, error) {
	if stmtReturn.Value != nil {
		if r.currentFunction == FnTypeInitializer {
			r.reportError(stmtReturn.Keyword, loxerrors.ErrParseCantReturnValueFromInitializer)
			return nil, errNilnil
		}
		r.resolveExpr(ctx, stmtReturn.Value)
	}
	return nil, errNilnil
}

// VisitStmtVar implements parser.StmtVisitor.
func (r *resolver) VisitStmtVar(ctx context.Context, stmtVar *parser.StmtVar) (any, error) {
	r.declare(ctx, stmtVar.Name)
	if stmtVar.Initializer != nil {
		r.resolveExpr(ctx, stmtVar.Initializer)
	}
	r.define(ctx, stmtVar.Name)
	return nil, errNilnil
}

// VisitStmtWhile implements parser.StmtVisitor.
func (r *resolver) VisitStmtWhile(ctx context.Context, stmtWhile *parser.StmtWhile) (any, error) {
	r.resolveExpr(ctx, stmtWhile.Condition)
	r.resolveStmt(ctx, stmtWhile.Body)
	return nil, errNilnil
}

// VisitExprAssign implements parser.ExprVisitor.
func (r *resolver) VisitExprAssign(ctx context.Context, exprAssign *parser.ExprAssign) (any, error) {
	r.resolveExpr(ctx, exprAssign.Value)
	r.resolveLocal(ctx, exprAssign, exprAssign.Name, false)
	return nil, errNilnil
}

// VisitExprBinary implements parser.ExprVisitor.
func (r *resolver) VisitExprBinary(ctx context.Context, exprBinary *parser.ExprBinary) (any, error) {
	r.resolveExpr(ctx, exprBinary.Left)
	r.resolveExpr(ctx, exprBinary.Right)
	return nil, errNilnil
}

// VisitExprCall implements parser.ExprVisitor.
func (r *resolver) VisitExprCall(ctx context.Context, exprCall *parser.ExprCall) (any, error) {
	r.resolveExpr(ctx, exprCall.Callee)
	for _, arg := range exprCall.Arguments {
		r.resolveExpr(ctx, arg)
	}
	return nil, errNilnil
}

// VisitExprGet implements parser.ExprVisitor.
func (r *resolver) VisitExprGet(ctx context.Context, exprGet *parser.ExprGet) (any, error) {
	r.resolveExpr(ctx, exprGet.Instance)
	return nil, errNilnil
}

// VisitExprFunction implements parser.ExprVisitor.
func (r *resolver) VisitExprFunction(ctx context.Context, exprFunction *parser.ExprFunction) (any, error) {
	r.resolveFunction(ctx, exprFunction, FnTypeExpr)
	return nil, errNilnil
}

// VisitExprGrouping implements parser.ExprVisitor.
func (r *resolver) VisitExprGrouping(ctx context.Context, exprGrouping *parser.ExprGrouping) (any, error) {
	r.resolveExpr(ctx, exprGrouping.Expression)
	return nil, errNilnil
}

// VisitExprLiteral implements parser.ExprVisitor.
func (r *resolver) VisitExprLiteral(ctx context.Context, exprLiteral *parser.ExprLiteral) (any, error) {
	return nil, errNilnil
}

// VisitExprLogical implements parser.ExprVisitor.
func (r *resolver) VisitExprLogical(ctx context.Context, exprLogical *parser.ExprLogical) (any, error) {
	r.resolveExpr(ctx, exprLogical.Left)
	r.resolveExpr(ctx, exprLogical.Right)
	return nil, errNilnil
}

// VisitExprSet implements parser.ExprVisitor.
func (r *resolver) VisitExprSet(ctx context.Context, exprSet *parser.ExprSet) (any, error) {
	r.resolveExpr(ctx, exprSet.Value)
	r.resolveExpr(ctx, exprSet.Instance)
	return nil, errNilnil
}

// VisitExprSuper implements parser.ExprVisitor.
func (r *resolver) VisitExprSuper(ctx context.Context, exprSuper *parser.ExprSuper) (any, error) {
	switch r.currentClass {
	case CTypeSubclass:
		break
	case CTypeNone:
		r.reportError(exprSuper.Keyword, loxerrors.ErrParseCantUseSuperOutsideClass)
	default:
		r.reportError(exprSuper.Keyword, loxerrors.ErrParseCantUseSuperInClassWithNoSuperclass)
	}

	if r.currentFunction == FnTypeClassMethod {
		r.reportError(exprSuper.Keyword, loxerrors.ErrParseCantUseSuperInClassMethod)
	}

	r.resolveLocal(ctx, exprSuper, exprSuper.Keyword, true)
	return nil, errNilnil
}

// VisitExprThis implements parser.ExprVisitor.
func (r *resolver) VisitExprThis(ctx context.Context, exprThis *parser.ExprThis) (any, error) {
	if r.currentClass == CTypeNone {
		r.reportError(exprThis.Keyword, loxerrors.ErrParseThisOutsideClass)
	}
	r.resolveLocal(ctx, exprThis, exprThis.Keyword, true)
	return nil, errNilnil
}

// VisitExprUnary implements parser.ExprVisitor.
func (r *resolver) VisitExprUnary(ctx context.Context, exprUnary *parser.ExprUnary) (any, error) {
	r.resolveExpr(ctx, exprUnary.Right)
	return nil, errNilnil
}

// VisitExprVariable implements parser.ExprVisitor.
func (r *resolver) VisitExprVariable(ctx context.Context, exprVariable *parser.ExprVariable) (any, error) {
	var err error
	if state, ok := r.peekScopeVar(ctx, exprVariable.Name.Lexeme); ok && state.State == VarStateDeclared {
		r.reportError(exprVariable.Name, loxerrors.ErrParseCantInitVarSelfReference)
	}
	r.resolveLocal(ctx, exprVariable, exprVariable.Name, true)
	return nil, err
}

func (r *resolver) beginScope(_ context.Context) {
	r.scopes.PushBack(map[string]*ResolverVariable{})
}

func (r *resolver) endScope(ctx context.Context) {
	if scope, ok := r.peekScope(ctx); ok {
		for _, name := range scope {
			if name.State == VarStateDefined {
				r.reportError(name.Name, loxerrors.ErrParseLocalVariableNotUsed)
			}
		}
	}

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

func (r *resolver) resolveFunction(ctx context.Context, function *parser.ExprFunction, declaration FunctionType) {
	enclosingFunction := r.currentFunction
	r.beginScope(ctx)
	r.currentFunction = declaration

	defer func() { r.currentFunction = enclosingFunction }()
	defer r.endScope(ctx)

	for _, param := range function.Parameters {
		r.declare(ctx, param)
		r.define(ctx, param)
	}

	r.resolveStmts(ctx, function.Body)
}

func (r *resolver) resolveLocal(ctx context.Context, expr parser.Expr, tok *token.Token, isRead bool) {
	depth := r.scopes.Len()
	back := r.scopes.Back()
	for i := range depth {
		scope := r.scopeFromListElem(back)
		if _, ok := scope[tok.Lexeme]; ok {
			r.interpreter.resolve(ctx, expr, i)

			if isRead {
				scope[tok.Lexeme].State = VarStateRead
			}
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
		scope[tok.Lexeme] = &ResolverVariable{Name: tok, State: VarStateDeclared}
	}
}

func (r *resolver) define(ctx context.Context, tok *token.Token) {
	if scope, ok := r.peekScope(ctx); ok {
		scope[tok.Lexeme].State = VarStateDefined
	}
}

func (r *resolver) defineInternal(ctx context.Context, name string) {
	if scope, ok := r.peekScope(ctx); ok {
		scope[name] = &ResolverVariable{Name: nil, State: VarStateRead}
	}
}

func (r *resolver) peekScope(_ context.Context) (map[string]*ResolverVariable, bool) {
	if r.scopes.Len() == 0 {
		return nil, false
	}
	return r.scopeFromListElem(r.scopes.Back()), true
}

func (r *resolver) peekScopeVar(ctx context.Context, name string) (*ResolverVariable, bool) {
	if scope, ok := r.peekScope(ctx); ok {
		if value, ok := scope[name]; ok {
			return value, true
		}
	}
	return nil, false
}

func (r *resolver) scopeFromListElem(el *list.Element) map[string]*ResolverVariable {
	return el.Value.(map[string]*ResolverVariable)
}

func (r *resolver) reportError(tok *token.Token, err error) {
	if ignoredErrors, ok := profiles[r.profile]; ok {
		for _, ignoredError := range ignoredErrors {
			if errors.Is(err, ignoredError) {
				return
			}
		}
	}

	r.err = append(r.err, loxerrors.NewParseError(tok, err))
}

func (r *resolver) String() string {
	w := new(strings.Builder)

	index := 0
	delimiter := ""
	element := r.scopes.Front()
	for element != nil {
		_, _ = fmt.Fprintf(w, "%s%d{%v}", delimiter, index, element.Value.(map[string]*ResolverVariable))
		index++
		element = element.Next()
		delimiter = " ->"
	}

	return fmt.Sprintf("resolver{err: %v, scopes: %s}", r.err, w)
}

var (
	_ parser.ExprVisitor = (*resolver)(nil)
	_ parser.StmtVisitor = (*resolver)(nil)
	_ Resolver           = (*resolver)(nil)
	_ fmt.Stringer       = (*resolver)(nil)
)
