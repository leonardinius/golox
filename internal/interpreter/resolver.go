package interpreter

import (
	"container/list"
	"errors"
	"fmt"
	"strings"

	"github.com/leonardinius/golox/internal/loxerrors"
	"github.com/leonardinius/golox/internal/parser"
	"github.com/leonardinius/golox/internal/token"
)

type Resolver interface {
	Resolve(statements []parser.Stmt) error
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
func (r *resolver) Resolve(statements []parser.Stmt) error {
	r.err = nil
	r.resolveStmts(statements)
	return errors.Join(r.err...)
}

// VisitStmtBlock implements parser.StmtVisitor.
func (r *resolver) VisitStmtBlock(stmtBlock *parser.StmtBlock) (Value, error) {
	r.beginScope()
	defer r.endScope()
	r.resolveStmts(stmtBlock.Statements)
	return NilValue, ErrNilNil
}

// VisitStmtClass implements parser.StmtVisitor.
func (r *resolver) VisitStmtClass(stmtClass *parser.StmtClass) (Value, error) {
	enclosingClass := r.currentClass
	defer func() { r.currentClass = enclosingClass }()
	r.currentClass = CTypeClass

	r.declare(stmtClass.Name)
	r.define(stmtClass.Name)

	if stmtClass.SuperClass != nil && stmtClass.Name.Lexeme == stmtClass.SuperClass.Name.Lexeme {
		r.reportError(stmtClass.SuperClass.Name, loxerrors.ErrParseClassCantInheritFromItself)
	}
	if stmtClass.SuperClass != nil {
		r.currentClass = CTypeSubclass
		r.resolveExpr(stmtClass.SuperClass)

		r.beginScope()
		defer r.endScope()
		r.defineInternal("super")
	}

	r.beginScope()
	defer r.endScope()

	r.defineInternal("this")

	for _, method := range stmtClass.ClassMethods {
		r.resolveFunction(method.Fn, FnTypeClassMethod)
	}

	for _, method := range stmtClass.Methods {
		functionType := FnTypeMethod
		if method.Name.Lexeme == "init" {
			functionType = FnTypeInitializer
		}
		r.resolveFunction(method.Fn, functionType)
	}

	return NilValue, ErrNilNil
}

// VisitStmtBreak implements parser.StmtVisitor.
func (r *resolver) VisitStmtBreak(stmtBreak *parser.StmtBreak) (Value, error) {
	return NilValue, ErrNilNil
}

// VisitStmtContinue implements parser.StmtVisitor.
func (r *resolver) VisitStmtContinue(stmtContinue *parser.StmtContinue) (Value, error) {
	return NilValue, ErrNilNil
}

// VisitStmtExpression implements parser.StmtVisitor.
func (r *resolver) VisitStmtExpression(stmtExpression *parser.StmtExpression) (Value, error) {
	r.resolveExpr(stmtExpression.Expression)
	return NilValue, ErrNilNil
}

// VisitStmtFor implements parser.StmtVisitor.
func (r *resolver) VisitStmtFor(stmtFor *parser.StmtFor) (Value, error) {
	if stmtFor.Initializer != nil {
		r.beginScope()
		defer r.endScope()
		r.resolveStmt(stmtFor.Initializer)
	}
	if stmtFor.Condition != nil {
		r.resolveExpr(stmtFor.Condition)
	}
	if stmtFor.Increment != nil {
		r.resolveExpr(stmtFor.Increment)
	}

	r.resolveStmt(stmtFor.Body)
	return NilValue, ErrNilNil
}

// VisitStmtFunction implements parser.StmtVisitor.
func (r *resolver) VisitStmtFunction(stmtFunction *parser.StmtFunction) (Value, error) {
	r.declare(stmtFunction.Name)
	r.define(stmtFunction.Name)

	r.resolveFunction(stmtFunction.Fn, FnTypeFunction)
	return NilValue, ErrNilNil
}

// VisitStmtIf implements parser.StmtVisitor.
func (r *resolver) VisitStmtIf(stmtIf *parser.StmtIf) (Value, error) {
	r.resolveExpr(stmtIf.Condition)
	r.resolveStmt(stmtIf.ThenBranch)
	if stmtIf.ElseBranch != nil {
		r.resolveStmt(stmtIf.ElseBranch)
	}
	return NilValue, ErrNilNil
}

// VisitStmtPrint implements parser.StmtVisitor.
func (r *resolver) VisitStmtPrint(stmtPrint *parser.StmtPrint) (Value, error) {
	r.resolveExpr(stmtPrint.Expression)
	return NilValue, ErrNilNil
}

// VisitStmtReturn implements parser.StmtVisitor.
func (r *resolver) VisitStmtReturn(stmtReturn *parser.StmtReturn) (Value, error) {
	if stmtReturn.Value != nil {
		if r.currentFunction == FnTypeInitializer {
			r.reportError(stmtReturn.Keyword, loxerrors.ErrParseCantReturnValueFromInitializer)
			return NilValue, ErrNilNil
		}
		r.resolveExpr(stmtReturn.Value)
	}
	return NilValue, ErrNilNil
}

// VisitStmtVar implements parser.StmtVisitor.
func (r *resolver) VisitStmtVar(stmtVar *parser.StmtVar) (Value, error) {
	r.declare(stmtVar.Name)
	if stmtVar.Initializer != nil {
		r.resolveExpr(stmtVar.Initializer)
	}
	r.define(stmtVar.Name)
	return NilValue, ErrNilNil
}

// VisitStmtWhile implements parser.StmtVisitor.
func (r *resolver) VisitStmtWhile(stmtWhile *parser.StmtWhile) (Value, error) {
	r.resolveExpr(stmtWhile.Condition)
	r.resolveStmt(stmtWhile.Body)
	return NilValue, ErrNilNil
}

// VisitExprAssign implements parser.ExprVisitor.
func (r *resolver) VisitExprAssign(exprAssign *parser.ExprAssign) (Value, error) {
	r.resolveExpr(exprAssign.Value)
	r.resolveLocal(exprAssign, exprAssign.Name, false)
	return NilValue, ErrNilNil
}

// VisitExprBinary implements parser.ExprVisitor.
func (r *resolver) VisitExprBinary(exprBinary *parser.ExprBinary) (Value, error) {
	r.resolveExpr(exprBinary.Left)
	r.resolveExpr(exprBinary.Right)
	return NilValue, ErrNilNil
}

// VisitExprCall implements parser.ExprVisitor.
func (r *resolver) VisitExprCall(exprCall *parser.ExprCall) (Value, error) {
	r.resolveExpr(exprCall.Callee)
	for _, arg := range exprCall.Arguments {
		r.resolveExpr(arg)
	}
	return NilValue, ErrNilNil
}

// VisitExprGet implements parser.ExprVisitor.
func (r *resolver) VisitExprGet(exprGet *parser.ExprGet) (Value, error) {
	r.resolveExpr(exprGet.Instance)
	return NilValue, ErrNilNil
}

// VisitExprFunction implements parser.ExprVisitor.
func (r *resolver) VisitExprFunction(exprFunction *parser.ExprFunction) (Value, error) {
	r.resolveFunction(exprFunction, FnTypeExpr)
	return NilValue, ErrNilNil
}

// VisitExprGrouping implements parser.ExprVisitor.
func (r *resolver) VisitExprGrouping(exprGrouping *parser.ExprGrouping) (Value, error) {
	r.resolveExpr(exprGrouping.Expression)
	return NilValue, ErrNilNil
}

// VisitExprLiteral implements parser.ExprVisitor.
func (r *resolver) VisitExprLiteral(exprLiteral *parser.ExprLiteral) (Value, error) {
	return NilValue, ErrNilNil
}

// VisitExprLogical implements parser.ExprVisitor.
func (r *resolver) VisitExprLogical(exprLogical *parser.ExprLogical) (Value, error) {
	r.resolveExpr(exprLogical.Left)
	r.resolveExpr(exprLogical.Right)
	return NilValue, ErrNilNil
}

// VisitExprSet implements parser.ExprVisitor.
func (r *resolver) VisitExprSet(exprSet *parser.ExprSet) (Value, error) {
	r.resolveExpr(exprSet.Value)
	r.resolveExpr(exprSet.Instance)
	return NilValue, ErrNilNil
}

// VisitExprSuper implements parser.ExprVisitor.
func (r *resolver) VisitExprSuper(exprSuper *parser.ExprSuper) (Value, error) {
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

	r.resolveLocal(exprSuper, exprSuper.Keyword, true)
	return NilValue, ErrNilNil
}

// VisitExprThis implements parser.ExprVisitor.
func (r *resolver) VisitExprThis(exprThis *parser.ExprThis) (Value, error) {
	if r.currentClass == CTypeNone {
		r.reportError(exprThis.Keyword, loxerrors.ErrParseThisOutsideClass)
	}
	r.resolveLocal(exprThis, exprThis.Keyword, true)
	return NilValue, ErrNilNil
}

// VisitExprUnary implements parser.ExprVisitor.
func (r *resolver) VisitExprUnary(exprUnary *parser.ExprUnary) (Value, error) {
	r.resolveExpr(exprUnary.Right)
	return NilValue, ErrNilNil
}

// VisitExprVariable implements parser.ExprVisitor.
func (r *resolver) VisitExprVariable(exprVariable *parser.ExprVariable) (Value, error) {
	var err error
	if state, ok := r.peekScopeVar(exprVariable.Name.Lexeme); ok && state.State == VarStateDeclared {
		r.reportError(exprVariable.Name, loxerrors.ErrParseCantInitVarSelfReference)
	}
	r.resolveLocal(exprVariable, exprVariable.Name, true)
	return NilValue, err
}

func (r *resolver) beginScope() {
	r.scopes.PushBack(map[string]*ResolverVariable{})
}

func (r *resolver) endScope() {
	if scope, ok := r.peekScope(); ok {
		for _, name := range scope {
			if name.State == VarStateDefined {
				r.reportError(name.Name, loxerrors.ErrParseLocalVariableNotUsed)
			}
		}
	}

	r.scopes.Remove(r.scopes.Back())
}

func (r *resolver) resolveStmts(stmts []parser.Stmt) {
	for _, stmt := range stmts {
		r.resolveStmt(stmt)
	}
}

func (r *resolver) resolveStmt(stmt parser.Stmt) {
	_, _ = stmt.Accept(r)
}

func (r *resolver) resolveExpr(expr parser.Expr) {
	_, _ = expr.Accept(r)
}

func (r *resolver) resolveFunction(function *parser.ExprFunction, declaration FunctionType) {
	enclosingFunction := r.currentFunction
	r.beginScope()
	r.currentFunction = declaration

	defer func() { r.currentFunction = enclosingFunction }()
	defer r.endScope()

	for _, param := range function.Parameters {
		r.declare(param)
		r.define(param)
	}

	r.resolveStmts(function.Body)
}

func (r *resolver) resolveLocal(expr parser.Expr, tok *token.Token, isRead bool) {
	depth := r.scopes.Len()
	back := r.scopes.Back()
	for i := range depth {
		scope := r.scopeFromListElem(back)
		if _, ok := scope[tok.Lexeme]; ok {
			r.interpreter.resolve(expr, i)

			if isRead {
				scope[tok.Lexeme].State = VarStateRead
			}
			return
		}
		back = back.Prev()
	}
}

func (r *resolver) declare(tok *token.Token) {
	if scope, ok := r.peekScope(); ok {
		if _, ok := scope[tok.Lexeme]; ok {
			r.reportError(tok, loxerrors.ErrParseCantDuplicateVariableDefinition)
		}
		scope[tok.Lexeme] = &ResolverVariable{Name: tok, State: VarStateDeclared}
	}
}

func (r *resolver) define(tok *token.Token) {
	if scope, ok := r.peekScope(); ok {
		scope[tok.Lexeme].State = VarStateDefined
	}
}

func (r *resolver) defineInternal(name string) {
	if scope, ok := r.peekScope(); ok {
		scope[name] = &ResolverVariable{Name: nil, State: VarStateRead}
	}
}

func (r *resolver) peekScope() (map[string]*ResolverVariable, bool) {
	if r.scopes.Len() == 0 {
		return nil, false
	}
	return r.scopeFromListElem(r.scopes.Back()), true
}

func (r *resolver) peekScopeVar(name string) (*ResolverVariable, bool) {
	if scope, ok := r.peekScope(); ok {
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
