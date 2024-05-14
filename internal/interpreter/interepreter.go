package interpreter

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/leonardinius/golox/internal/loxerrors"
	"github.com/leonardinius/golox/internal/parser"
	"github.com/leonardinius/golox/internal/token"
)

var (
	errBreak    = errors.New("eval:break")
	errContinue = errors.New("eval:continue")
)

type Interpreter interface {
	// Interpret interprets the given statements.
	// Returns the stringified result of the last statement and an error if any.
	// The error is nil if the statement is valid.
	//
	// Not thread safe.
	Interpret(ctx context.Context, stmts []parser.Stmt) (string, error)

	// Evaluate evaluates the given statement.
	// Returns an error if any.
	// The error is nil if the statement is valid.
	//
	// Not thread safe.
	Evaluate(ctx context.Context, stmt parser.Stmt) (any, error)
}

type interpreter struct {
	Globals     *environment
	Stdin       io.Reader
	Stdout      io.Writer
	Stderr      io.Writer
	ErrReporter loxerrors.ErrReporter
	Locals      map[parser.Expr]int
}

func NewInterpreter(options ...InterpreterOption) *interpreter {
	opts := newInterpreterOpts(options...)
	globals := opts.globals
	globals.Define("clock", NativeFunction0(StdFnTime))
	globals.Define("pprint", NativeFunctionVarArgs(StdFnPPrint))

	return &interpreter{
		Globals:     opts.globals,
		Stdin:       opts.stdin,
		Stdout:      opts.stdout,
		Stderr:      opts.stderr,
		ErrReporter: opts.reporter,
		Locals:      make(map[parser.Expr]int),
	}
}

// Interpret implements Interpreter.
func (i *interpreter) Interpret(ctx context.Context, stmts []parser.Stmt) (string, error) {
	var v any
	var err error

	ctx = i.Globals.AsContext(ctx)

	for _, stmt := range stmts {
		if v, err = i.Evaluate(ctx, stmt); err != nil {
			return "", err
		}
	}

	return i.stringify(v), nil
}

// Evaluate implements Interpreter.
func (i *interpreter) Evaluate(ctx context.Context, stmt parser.Stmt) (any, error) {
	return i.execute(ctx, stmt)
}

func (i *interpreter) print(v ...any) {
	for i, vv := range v {
		if vv == nil {
			v[i] = "nil"
		}
	}

	_, _ = fmt.Fprintln(i.Stdout, v...)
}

func (i *interpreter) stringify(v any) string {
	if v == nil {
		return "nil"
	}
	return fmt.Sprintf("%#v", v)
}

// VisitExpression implements parser.StmtVisitor.
func (i *interpreter) VisitStmtExpression(ctx context.Context, expr *parser.StmtExpression) (any, error) {
	return i.evaluate(ctx, expr.Expression)
}

// VisitStmtFunction implements parser.StmtVisitor.
func (i *interpreter) VisitStmtFunction(ctx context.Context, stmtFunction *parser.StmtFunction) (any, error) {
	env := EnvFromContext(ctx)
	function := NewLoxFunction(stmtFunction.Name, stmtFunction.Fn, env)
	env.Define(stmtFunction.Name.Lexeme, function)
	return nil, nil
}

// VisitStmtIf implements parser.StmtVisitor.
func (i *interpreter) VisitStmtIf(ctx context.Context, stmtIf *parser.StmtIf) (any, error) {
	condition, err := i.evaluate(ctx, stmtIf.Condition)
	if err != nil {
		return nil, err
	}

	if i.isTruthy(condition) {
		return i.execute(ctx, stmtIf.ThenBranch)
	} else if stmtIf.ElseBranch != nil {
		return i.execute(ctx, stmtIf.ElseBranch)
	}

	return nil, nil
}

// VisitPrint implements parser.StmtVisitor.
func (i *interpreter) VisitStmtPrint(ctx context.Context, expr *parser.StmtPrint) (any, error) {
	value, err := i.evaluate(ctx, expr.Expression)
	if err == nil {
		i.print(value)
	}
	return nil, err
}

// VisitStmtReturn implements parser.StmtVisitor.
func (i *interpreter) VisitStmtReturn(ctx context.Context, stmtReturn *parser.StmtReturn) (value any, err error) {
	if stmtReturn.Value != nil {
		if value, err = i.evaluate(ctx, stmtReturn.Value); err != nil {
			return nil, err
		}
	}

	return nil, &ReturnValue{Value: value}
}

// VisitVar implements parser.StmtVisitor.
func (i *interpreter) VisitStmtVar(ctx context.Context, stmt *parser.StmtVar) (any, error) {
	var value any
	var err error
	if stmt.Initializer != nil {
		if value, err = i.evaluate(ctx, stmt.Initializer); err != nil {
			return nil, err
		}
	}

	env := EnvFromContext(ctx)
	env.Define(stmt.Name.Lexeme, value)

	return nil, nil
}

// VisitStmtWhile implements parser.StmtVisitor.
func (i *interpreter) VisitStmtWhile(ctx context.Context, stmtWhile *parser.StmtWhile) (any, error) {
	var condition any
	var value any
	var err error

	for err == nil {
		if condition, err = i.evaluate(ctx, stmtWhile.Condition); err != nil {
			break
		}

		if !i.isTruthy(condition) {
			break
		}

		if value, err = i.execute(ctx, stmtWhile.Body); err != nil {
			switch {
			case err == errBreak:
				// return immediatelly
				return nil, nil
			case err == errContinue:
				// continue to next iteration
				err = nil
			}
		}
	}

	return value, err
}

// VisitStmtFor implements parser.StmtVisitor.
func (i *interpreter) VisitStmtFor(ctx context.Context, stmtFor *parser.StmtFor) (any, error) {
	var condition any
	var value any
	var err error

	if stmtFor.Initializer != nil {
		_, err = i.execute(ctx, stmtFor.Initializer)
	}

	for err == nil {
		if condition, err = i.evaluate(ctx, stmtFor.Condition); err != nil {
			break
		}

		if !i.isTruthy(condition) {
			break
		}

		if value, err = i.execute(ctx, stmtFor.Body); err != nil {
			switch {
			case err == errBreak:
				// return immediatelly
				return nil, nil
			case err == errContinue:
				// continue to next iteration
				err = nil
			}
		}

		if err == nil && stmtFor.Increment != nil {
			_, err = i.evaluate(ctx, stmtFor.Increment)
		}
	}

	return value, err
}

// VisitStmtBreak implements parser.StmtVisitor.
func (*interpreter) VisitStmtBreak(ctx context.Context, stmtBreak *parser.StmtBreak) (any, error) {
	return nil, errBreak
}

// VisitStmtContinue implements parser.StmtVisitor.
func (*interpreter) VisitStmtContinue(ctx context.Context, stmtContinue *parser.StmtContinue) (any, error) {
	return nil, errContinue
}

// VisitStmtBlock implements parser.StmtVisitor.
func (i *interpreter) VisitStmtBlock(ctx context.Context, block *parser.StmtBlock) (any, error) {
	env := EnvFromContext(ctx)
	newCtx := env.Nest().AsContext(ctx)
	return i.executeBlock(newCtx, block.Statements)
}

// VisitVariable implements parser.ExprVisitor.
func (i *interpreter) VisitExprVariable(ctx context.Context, expr *parser.ExprVariable) (any, error) {
	return i.lookupVariable(ctx, expr.Name, expr)
}

// VisitExprAssign implements parser.ExprVisitor.
func (i *interpreter) VisitExprAssign(ctx context.Context, assign *parser.ExprAssign) (any, error) {
	value, err := i.evaluate(ctx, assign.Value)
	if err != nil {
		return nil, err
	}

	return i.assignVariable(ctx, assign, value)
}

// VisitBinary implements parser.Visitor.
func (i *interpreter) VisitExprBinary(ctx context.Context, expr *parser.ExprBinary) (any, error) {
	left, err := i.evaluate(ctx, expr.Left)
	if err != nil {
		return nil, err
	}
	right, err := i.evaluate(ctx, expr.Right)
	if err != nil {
		return nil, err
	}

	switch expr.Operator.Type {
	case token.GREATER:
		if err = i.checkNumberOperands(expr.Operator, left, right); err != nil {
			return nil, err
		}
		return left.(float64) > right.(float64), nil
	case token.GREATER_EQUAL:
		if err = i.checkNumberOperands(expr.Operator, left, right); err != nil {
			return nil, err
		}
		return left.(float64) >= right.(float64), nil
	case token.LESS:
		if err = i.checkNumberOperands(expr.Operator, left, right); err != nil {
			return nil, err
		}
		return left.(float64) < right.(float64), nil
	case token.LESS_EQUAL:
		if err = i.checkNumberOperands(expr.Operator, left, right); err != nil {
			return nil, err
		}
		return left.(float64) <= right.(float64), nil
	case token.BANG_EQUAL:
		return !i.isEqual(left, right), nil
	case token.EQUAL_EQUAL:
		return i.isEqual(left, right), nil
	case token.MINUS:
		if err = i.checkNumberOperands(expr.Operator, left, right); err != nil {
			return nil, err
		}
		return left.(float64) - right.(float64), nil
	case token.PLUS:
		if left, ok := left.(string); ok {
			if right, ok := right.(string); ok {
				return left + right, nil
			}
		}
		if left, ok := left.(float64); ok {
			if right, ok := right.(float64); ok {
				return left + right, nil
			}
		}
		return i.returnRuntimeError(expr.Operator, loxerrors.ErrRuntimeOperandsMustNumbersOrStrings)
	case token.SLASH:
		if err = i.checkNumberOperands(expr.Operator, left, right); err != nil {
			return nil, err
		}
		return left.(float64) / right.(float64), nil
	case token.STAR:
		if err = i.checkNumberOperands(expr.Operator, left, right); err != nil {
			return nil, err
		}
		return left.(float64) * right.(float64), nil
	}

	return i.unreachable()
}

// VisitExprFunction implements parser.ExprVisitor.
func (i *interpreter) VisitExprFunction(ctx context.Context, exprFunction *parser.ExprFunction) (any, error) {
	env := EnvFromContext(ctx)
	fn := NewLoxFunction(nil, exprFunction, env)
	return fn, nil
}

// VisitExprCall implements parser.ExprVisitor.
func (i *interpreter) VisitExprCall(ctx context.Context, exprCall *parser.ExprCall) (any, error) {
	callee, err := i.evaluate(ctx, exprCall.Callee)
	if err != nil {
		return nil, err
	}
	callable, ok := callee.(Callable)
	if !ok {
		return i.returnRuntimeError(exprCall.CloseParen, loxerrors.ErrRuntimeCalleeMustBeCallable)
	}

	var args []any
	for _, arg := range exprCall.Arguments {
		argValue, err := i.evaluate(ctx, arg)
		if err != nil {
			return nil, err
		}
		args = append(args, argValue)
	}

	if !callable.Arity().IsVarArgs() && len(args) != int(callable.Arity()) {
		return i.returnRuntimeError(exprCall.CloseParen,
			loxerrors.ErrRuntimeCalleeArityError(
				int(callable.Arity()),
				len(args),
			))
	}
	return callable.Call(ctx, i, args)
}

// VisitGrouping implements parser.Visitor.
func (i *interpreter) VisitExprGrouping(ctx context.Context, expr *parser.ExprGrouping) (any, error) {
	return i.evaluate(ctx, expr.Expression)
}

// VisitLiteral implements parser.Visitor.
func (i *interpreter) VisitExprLiteral(ctx context.Context, expr *parser.ExprLiteral) (any, error) {
	return expr.Value, nil
}

// VisitExprLogical implements parser.ExprVisitor.
func (i *interpreter) VisitExprLogical(ctx context.Context, exprLogical *parser.ExprLogical) (any, error) {
	switch exprLogical.Operator.Type {
	case token.AND:
		return i.evalLogicalAnd(ctx, exprLogical.Left, exprLogical.Right)
	case token.OR:
		return i.evalLogicalOr(ctx, exprLogical.Left, exprLogical.Right)
	default:
		return i.unreachable()
	}
}

func (i *interpreter) evalLogicalAnd(ctx context.Context, left parser.Expr, right parser.Expr) (any, error) {
	if leftValue, err := i.evaluate(ctx, left); err != nil || !i.isTruthy(leftValue) {
		return leftValue, err
	}

	return i.evaluate(ctx, right)
}

func (i *interpreter) evalLogicalOr(ctx context.Context, left parser.Expr, right parser.Expr) (any, error) {
	if leftValue, err := i.evaluate(ctx, left); err != nil || i.isTruthy(leftValue) {
		return leftValue, err
	}

	return i.evaluate(ctx, right)
}

// VisitUnary implements parser.Visitor.
func (i *interpreter) VisitExprUnary(ctx context.Context, expr *parser.ExprUnary) (any, error) {
	right, err := i.evaluate(ctx, expr.Right)
	if err != nil {
		return nil, err
	}

	switch expr.Operator.Type {
	case token.MINUS:
		if err = i.checkNumberOperand(expr.Operator, right); err != nil {
			return nil, err
		}
		return -right.(float64), nil
	case token.BANG:
		return !i.isTruthy(right), nil
	}

	return i.unreachable()
}

func (i *interpreter) execute(ctx context.Context, stmt parser.Stmt) (any, error) {
	value, err := stmt.Accept(ctx, i)
	return value, err
}

func (i *interpreter) executeBlock(ctx context.Context, stmt []parser.Stmt) (value any, err error) {

	for _, stmt := range stmt {
		if _, err = i.execute(ctx, stmt); err != nil {
			return nil, err
		}
	}

	return nil, nil
}

func (i *interpreter) evaluate(ctx context.Context, expr parser.Expr) (any, error) {
	return expr.Accept(ctx, i)
}

func (i *interpreter) isTruthy(value any) bool {
	if value == nil {
		return false
	}
	if expr, ok := value.(bool); ok {
		return expr
	}

	return true
}

func (i *interpreter) isEqual(left, right any) bool {
	if left == nil && right == nil {
		return true
	}
	return left == right
}

func (i *interpreter) checkNumberOperands(tok *token.Token, left, right any) error {
	if _, ok := left.(float64); !ok {
		return i.runtimeError(tok, loxerrors.ErrRuntimeOperandsMustBeNumbers)
	} else if _, ok := right.(float64); !ok {
		return i.runtimeError(tok, loxerrors.ErrRuntimeOperandsMustBeNumbers)
	}
	return nil
}

func (i *interpreter) checkNumberOperand(tok *token.Token, val any) error {
	if _, ok := val.(float64); !ok {
		return i.runtimeError(tok, loxerrors.ErrRuntimeOperandMustBeNumber)
	}

	return nil
}

func (i *interpreter) returnRuntimeError(tok *token.Token, err error) (any, error) {
	return nil, i.runtimeError(tok, err)
}

func (i *interpreter) runtimeError(tok *token.Token, err error) error {
	return loxerrors.NewRuntimeError(tok, err)
}

func (i *interpreter) resolve(_ context.Context, expr parser.Expr, depth int) {
	i.Locals[expr] = depth
}

func (i *interpreter) lookupVariable(ctx context.Context, name *token.Token, expr parser.Expr) (any, error) {
	env := EnvFromContext(ctx)
	if distance, ok := i.Locals[expr]; ok {
		return env.GetAt(distance, name.Lexeme)
	}

	return i.Globals.Get(name)
}

func (i *interpreter) assignVariable(ctx context.Context, expr *parser.ExprAssign, value any) (any, error) {
	env := EnvFromContext(ctx)
	if distance, ok := i.Locals[expr]; ok {
		return env.AssignAt(distance, expr.Name, value)
	}

	return value, i.Globals.Assign(expr.Name, value)
}

func (i *interpreter) unreachable() (any, error) {
	panic("unreachable")
}

var _ parser.ExprVisitor = (*interpreter)(nil)
var _ parser.StmtVisitor = (*interpreter)(nil)
var _ Interpreter = (*interpreter)(nil)
