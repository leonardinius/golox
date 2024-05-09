package interpreter

import (
	"fmt"

	"github.com/leonardinius/golox/internal/loxerrors"
	"github.com/leonardinius/golox/internal/parser"
	"github.com/leonardinius/golox/internal/token"
)

type Interpreter interface {
	// Interpret interprets the given statements.
	// Returns the stringified result of the last statement and an error if any.
	// The error is nil if the statement is valid.
	//
	// Not thread safe.
	Interpret(stmt []parser.Stmt) (string, error)

	// Evaluate evaluates the given statement.
	// Returns an error if any.
	// The error is nil if the statement is valid.
	//
	// Not thread safe.
	Evaluate(parser.Stmt) (any, error)
}

type interpreter struct {
	env *environment
}

func NewInterpreter() Interpreter {
	return &interpreter{
		env: newEnvironment(),
	}
}

// Interpret implements Interpreter.
func (i *interpreter) Interpret(statements []parser.Stmt) (string, error) {
	var v any
	var err error

	for _, stmt := range statements {
		if v, err = i.Evaluate(stmt); err != nil {
			return "", err
		}
	}

	return i.stringify(v), nil
}

// Evaluate implements Interpreter.
func (i *interpreter) Evaluate(stmt parser.Stmt) (any, error) {
	return i.execute(stmt)
}

func (i *interpreter) print(v any) {
	if v == nil {
		v = "nil"
	}
	fmt.Println(v)
}

func (i *interpreter) stringify(v any) string {
	if v == nil {
		return "nil"
	}
	return fmt.Sprintf("%#v", v)
}

// VisitExpression implements parser.StmtVisitor.
func (i *interpreter) VisitStmtExpression(expr *parser.StmtExpression) (any, error) {
	return i.evaluate(expr.Expression)
}

// VisitPrint implements parser.StmtVisitor.
func (i *interpreter) VisitStmtPrint(expr *parser.StmtPrint) (any, error) {
	value, err := i.evaluate(expr.Expression)
	if err == nil {
		i.print(value)
	}
	return nil, err
}

// VisitVar implements parser.StmtVisitor.
func (i *interpreter) VisitStmtVar(stmt *parser.StmtVar) (any, error) {
	var value any
	var err error
	if stmt.Initializer != nil {
		if value, err = i.evaluate(stmt.Initializer); err != nil {
			return nil, err
		}
	}

	i.env.Define(stmt.Name.Lexeme, value)

	return nil, nil
}

// VisitVariable implements parser.ExprVisitor.
func (i *interpreter) VisitExprVariable(expr *parser.ExprVariable) (any, error) {
	return i.env.Get(expr.Name)
}

// VisitExprAssign implements parser.ExprVisitor.
func (i *interpreter) VisitExprAssign(assign *parser.ExprAssign) (any, error) {
	value, err := i.evaluate(assign.Value)
	if err != nil {
		return nil, err
	}

	if err = i.env.Assign(assign.Name, value); err != nil {
		return nil, err
	}

	return value, nil
}

// VisitBinary implements parser.Visitor.
func (i *interpreter) VisitExprBinary(expr *parser.ExprBinary) (any, error) {
	left, err := i.evaluate(expr.Left)
	if err != nil {
		return nil, err
	}
	right, err := i.evaluate(expr.Right)
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
func (i *interpreter) VisitExprGrouping(expr *parser.ExprGrouping) (any, error) {
	return i.evaluate(expr.Expression)
}

// VisitLiteral implements parser.Visitor.
func (i *interpreter) VisitExprLiteral(expr *parser.ExprLiteral) (any, error) {
	return expr.Value, nil
}

// VisitUnary implements parser.Visitor.
func (i *interpreter) VisitExprUnary(expr *parser.ExprUnary) (any, error) {
	right, err := i.evaluate(expr.Right)
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

func (i *interpreter) execute(stmt parser.Stmt) (any, error) {
	return stmt.Accept(i)
}

func (i *interpreter) evaluate(expr parser.Expr) (any, error) {
	return expr.Accept(i)
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
