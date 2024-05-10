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
	Env    *environment
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

func NewInterpreter(options ...InterpreterOption) Interpreter {
	opts := newInterpreterOpts(options...)
	return &interpreter{
		Env:    opts.env,
		Stdin:  opts.stdin,
		Stdout: opts.stdout,
		Stderr: opts.stderr,
	}
}

// Interpret implements Interpreter.
func (i *interpreter) Interpret(ctx context.Context, stmts []parser.Stmt) (string, error) {
	var v any
	var err error

	ctx = i.Env.AsContext(ctx)

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

func (i *interpreter) print(v any) {
	if v == nil {
		v = "nil"
	}
	_, _ = fmt.Fprintln(i.Stdout, v)
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
	var err error
	for {
		if condition, err = i.evaluate(ctx, stmtWhile.Condition); err != nil {
			break
		}

		if !i.isTruthy(condition) {
			break
		}

		if _, err = i.execute(ctx, stmtWhile.Body); err != nil {
			switch {
			case err == errBreak:
				// return immediatelly
				return nil, nil
			case err == errContinue:
				// continue to next iteration
				err = nil
			default:
				break
			}
		}
	}

	return nil, err
}

// VisitStmtFor implements parser.StmtVisitor.
func (i *interpreter) VisitStmtFor(ctx context.Context, stmtFor *parser.StmtFor) (any, error) {
	var condition any
	var err error

	if stmtFor.Initializer != nil {
		if _, err = i.execute(ctx, stmtFor.Initializer); err != nil {
			return nil, err
		}
	}

	for {
		if condition, err = i.evaluate(ctx, stmtFor.Condition); err != nil {
			break
		}

		if !i.isTruthy(condition) {
			break
		}

		if _, err = i.execute(ctx, stmtFor.Body); err != nil {
			switch {
			case err == errBreak:
				// return immediatelly
				return nil, nil
			case err == errContinue:
				// continue to next iteration
				err = nil
			default:
				break
			}
		}

		if stmtFor.Increment != nil {
			if _, err = i.evaluate(ctx, stmtFor.Increment); err != nil {
				return nil, err
			}
		}
	}

	return nil, err
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
	return i.executeBlock(env.NewNestContext(ctx), block.Statements)
}

// VisitVariable implements parser.ExprVisitor.
func (i *interpreter) VisitExprVariable(ctx context.Context, expr *parser.ExprVariable) (any, error) {
	env := EnvFromContext(ctx)
	return env.Get(expr.Name)
}

// VisitExprAssign implements parser.ExprVisitor.
func (i *interpreter) VisitExprAssign(ctx context.Context, assign *parser.ExprAssign) (any, error) {
	value, err := i.evaluate(ctx, assign.Value)
	if err != nil {
		return nil, err
	}

	env := EnvFromContext(ctx)
	if err = env.Assign(assign.Name, value); err != nil {
		return nil, err
	}

	return value, nil
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
		return nil, i.runtimeError(expr.Operator, loxerrors.ErrRuntimeOperandsMustNumbersOrStrings)
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
	return stmt.Accept(ctx, i)
}

func (i *interpreter) executeBlock(ctx context.Context, stmt []parser.Stmt) (value any, err error) {

	for _, stmt := range stmt {
		if value, err = i.execute(ctx, stmt); err != nil {
			return nil, err
		}
	}

	return value, nil
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

func (i *interpreter) runtimeError(tok *token.Token, err error) error {
	return loxerrors.NewRuntimeError(tok, err)
}

func (i *interpreter) unreachable() (any, error) {
	panic("unreachable")
}

var _ parser.ExprVisitor = (*interpreter)(nil)
var _ parser.StmtVisitor = (*interpreter)(nil)
var _ Interpreter = (*interpreter)(nil)
