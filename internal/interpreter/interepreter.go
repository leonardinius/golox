package interpreter

import (
	"fmt"

	"github.com/leonardinius/golox/internal/loxerrors"
	"github.com/leonardinius/golox/internal/parser"
	"github.com/leonardinius/golox/internal/token"
)

type Interpreter interface {
	// Interpret interprets the given expression.
	// Returns the stringified result of the expression and an error if any.
	// The error is nil if the expression is valid.
	//
	// Not thread safe.
	// Resets internal state on Interpret.
	Interpret(expr parser.Expr) (string, error)

	// Evaluate evaluates the given expression.
	// Returns the result of the expression and an error if any.
	// The error is nil if the expression is valid.
	//
	// Not thread safe.
	// Resets internal state on Evaluate.
	Evaluate(expr parser.Expr) (any, error)
}

type interpreter struct {
	err error
}

func NewInterpreter() Interpreter {
	return &interpreter{}
}

// Interpret implements Interpreter.
func (i *interpreter) Interpret(expr parser.Expr) (string, error) {
	if value, err := i.Evaluate(expr); err != nil {
		return "", err
	} else {
		return i.stringify(value), nil
	}
}

// Evaluate implements Interpreter.
func (i *interpreter) Evaluate(expr parser.Expr) (any, error) {
	i.reset()

	return i.evaluate(expr)
}

func (i *interpreter) stringify(v any) string {
	return fmt.Sprintf("%#v", v)
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
		return left.(token.DoubleNumber) > right.(token.DoubleNumber)
	case token.GREATER_EQUAL:
		if ok := i.checkNumberOperands(expr.Operator, left, right); !ok {
			return nil
		}
		return left.(token.DoubleNumber) >= right.(token.DoubleNumber)
	case token.LESS:
		if ok := i.checkNumberOperands(expr.Operator, left, right); !ok {
			return nil
		}
		return left.(token.DoubleNumber) < right.(token.DoubleNumber)
	case token.LESS_EQUAL:
		if ok := i.checkNumberOperands(expr.Operator, left, right); !ok {
			return nil
		}
		return left.(token.DoubleNumber) <= right.(token.DoubleNumber)
	case token.BANG_EQUAL:
		return !i.isEqual(left, right)
	case token.EQUAL_EQUAL:
		return i.isEqual(left, right)
	case token.MINUS:
		if ok := i.checkNumberOperands(expr.Operator, left, right); !ok {
			return nil
		}
		return left.(token.DoubleNumber) - right.(token.DoubleNumber)
	case token.PLUS:
		if left, ok := left.(string); ok {
			if right, ok := right.(string); ok {
				return left + right
			}
		}
		if left, ok := left.(token.DoubleNumber); ok {
			if right, ok := right.(token.DoubleNumber); ok {
				return left + right
			}
		}
		return i.reportError(expr.Operator, "Operands must be two numbers or two strings.")
	case token.SLASH:
		if ok := i.checkNumberOperands(expr.Operator, left, right); !ok {
			return nil
		}
		return left.(token.DoubleNumber) / right.(token.DoubleNumber)
	case token.STAR:
		if ok := i.checkNumberOperands(expr.Operator, left, right); !ok {
			return nil
		}
		return left.(token.DoubleNumber) * right.(token.DoubleNumber)
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
		return -right.(token.DoubleNumber)
	case token.BANG:
		return !i.isTruthy(right)
	}

	return i.unreachable()
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
	if _, ok := left.(token.DoubleNumber); !ok {
		i.reportError(tok, "Operands must be numbers.")
	} else if _, ok := right.(token.DoubleNumber); !ok {
		i.reportError(tok, "Operands must be numbers.")
	}
	return !i.hasErr()
}

func (i *interpreter) checkNumberOperand(tok *token.Token, val any) bool {
	if _, ok := val.(token.DoubleNumber); !ok {
		i.reportError(tok, "Operand must be a number.")
	}

	return !i.hasErr()
}

func (i *interpreter) reportError(tok *token.Token, msg string) any {
	i.err = loxerrors.NewRuntimeError(tok, msg)
	return nil
}

func (i *interpreter) reset() {
	i.err = nil
}

var _ parser.Visitor = (*interpreter)(nil)
var _ Interpreter = (*interpreter)(nil)
