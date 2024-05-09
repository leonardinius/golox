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
	// Resets internal state on Interpret.
	Interpret(stmt []parser.Stmt) (string, error)

	// Evaluate evaluates the given statement.
	// Returns an error if any.
	// The error is nil if the statement is valid.
	//
	// Not thread safe.
	Evaluate(parser.Stmt) (any, error)
}

type interpreter struct {
	err error
	env *environment
}

func NewInterpreter() Interpreter {
	return &interpreter{
		env: newEnvironment(),
	}
}

// Interpret implements Interpreter.
func (i *interpreter) Interpret(statements []parser.Stmt) (string, error) {
	i.resetError()

	for _, stmt := range statements {
		if v, err := i.Evaluate(stmt); err != nil {
			return "", err
		} else {
			return i.stringify(v), nil
		}
	}

	return i.stringify(nil), nil
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
func (i *interpreter) VisitExpression(expr *parser.Expression) any {
	if v, err := i.evaluate(expr.Expression); err == nil {
		return v
	}
	return nil
}

// VisitPrint implements parser.StmtVisitor.
func (i *interpreter) VisitPrint(expr *parser.Print) any {
	if value, err := i.evaluate(expr.Expression); err == nil {
		i.print(value)
	}
	return nil
}

// VisitVar implements parser.StmtVisitor.
func (i *interpreter) VisitVar(stmt *parser.Var) any {
	var value any
	var err error
	if stmt.Initializer != nil {
		if value, err = i.evaluate(stmt.Initializer); err != nil {
			return nil
		}
	}

	i.env.Assign(stmt.Name.Lexeme, value)

	return nil
}

// VisitVariable implements parser.ExprVisitor.
func (i *interpreter) VisitVariable(expr *parser.Variable) any {
	value, err := i.env.Get(expr.Name)
	if err != nil {
		return i.reportError(expr.Name, err)
	}
	return value
}

// VisitBinary implements parser.Visitor.
func (i *interpreter) VisitBinary(expr *parser.Binary) any {
	left, err := i.evaluate(expr.Left)
	if err != nil {
		return err
	}
	right, err := i.evaluate(expr.Right)
	if err != nil {
		return err
	}

	switch expr.Operator.Type {
	case token.GREATER:
		if ok := i.checkNumberOperands(expr.Operator, left, right); !ok {
			return nil
		}
		return left.(float64) > right.(float64)
	case token.GREATER_EQUAL:
		if ok := i.checkNumberOperands(expr.Operator, left, right); !ok {
			return nil
		}
		return left.(float64) >= right.(float64)
	case token.LESS:
		if ok := i.checkNumberOperands(expr.Operator, left, right); !ok {
			return nil
		}
		return left.(float64) < right.(float64)
	case token.LESS_EQUAL:
		if ok := i.checkNumberOperands(expr.Operator, left, right); !ok {
			return nil
		}
		return left.(float64) <= right.(float64)
	case token.BANG_EQUAL:
		return !i.isEqual(left, right)
	case token.EQUAL_EQUAL:
		return i.isEqual(left, right)
	case token.MINUS:
		if ok := i.checkNumberOperands(expr.Operator, left, right); !ok {
			return nil
		}
		return left.(float64) - right.(float64)
	case token.PLUS:
		if left, ok := left.(string); ok {
			if right, ok := right.(string); ok {
				return left + right
			}
		}
		if left, ok := left.(float64); ok {
			if right, ok := right.(float64); ok {
				return left + right
			}
		}
		return i.reportError(expr.Operator, loxerrors.ErrRuntimeOperandsMustNumbersOrStrings)
	case token.SLASH:
		if ok := i.checkNumberOperands(expr.Operator, left, right); !ok {
			return nil
		}
		return left.(float64) / right.(float64)
	case token.STAR:
		if ok := i.checkNumberOperands(expr.Operator, left, right); !ok {
			return nil
		}
		return left.(float64) * right.(float64)
	}

	return i.unreachable()
}

// VisitGrouping implements parser.Visitor.
func (i *interpreter) VisitGrouping(expr *parser.Grouping) any {
	if v, err := i.evaluate(expr.Expression); err == nil {
		return v
	}
	return nil
}

// VisitLiteral implements parser.Visitor.
func (i *interpreter) VisitLiteral(expr *parser.Literal) any {
	return expr.Value
}

// VisitUnary implements parser.Visitor.
func (i *interpreter) VisitUnary(expr *parser.Unary) any {
	right, err := i.evaluate(expr.Right)
	if err != nil {
		return nil
	}

	switch expr.Operator.Type {
	case token.MINUS:
		if ok := i.checkNumberOperand(expr.Operator, right); !ok {
			return nil
		}
		return -right.(float64)
	case token.BANG:
		return !i.isTruthy(right)
	}

	return i.unreachable()
}

func (i *interpreter) execute(stmt parser.Stmt) (any, error) {
	if i.hasErr() {
		return nil, i.err
	}

	value := stmt.Accept(i)

	return value, i.err
}

func (i *interpreter) evaluate(expr parser.Expr) (any, error) {
	if i.hasErr() {
		return nil, i.err
	}

	value := expr.Accept(i)

	return value, i.err
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

func (i *interpreter) unreachable() any {
	panic("unreachable")
}

func (i *interpreter) hasErr() bool {
	return i.err != nil
}

func (i *interpreter) checkNumberOperands(tok *token.Token, left, right any) bool {
	if _, ok := left.(float64); !ok {
		i.reportError(tok, loxerrors.ErrRuntimeOperandsMustBeNumbers)
	} else if _, ok := right.(float64); !ok {
		i.reportError(tok, loxerrors.ErrRuntimeOperandsMustBeNumbers)
	}
	return !i.hasErr()
}

func (i *interpreter) checkNumberOperand(tok *token.Token, val any) bool {
	if _, ok := val.(float64); !ok {
		i.reportError(tok, loxerrors.ErrRuntimeOperandMustBeNumber)
	}

	return !i.hasErr()
}

func (i *interpreter) reportError(tok *token.Token, err error) any {
	i.err = loxerrors.NewRuntimeError(tok, err)
	return nil
}

func (i *interpreter) resetError() {
	i.err = nil
}

var _ parser.ExprVisitor = (*interpreter)(nil)
var _ parser.StmtVisitor = (*interpreter)(nil)
var _ Interpreter = (*interpreter)(nil)
