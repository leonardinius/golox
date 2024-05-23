package interpreter

import (
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
	Interpret(stmts []parser.Stmt) (string, error)

	// Evaluate evaluates the given statement.
	// Returns an error if any.
	// The error is nil if the statement is valid.
	//
	// Not thread safe.
	Evaluate(stmt parser.Stmt) (any, error)
}

type interpreter struct {
	Globals     *environment
	Env         *environment
	Stdin       io.Reader
	Stdout      io.Writer
	Stderr      io.Writer
	ErrReporter loxerrors.ErrReporter
	Locals      map[parser.Expr]int
}

func NewInterpreter(options ...InterpreterOption) *interpreter {
	opts := newInterpreterOpts(options...)
	globals := opts.globals
	globals.Define("Array", NativeFunction1(StdFnCreateArray))
	globals.Define("clock", NativeFunction0(StdFnTime))
	globals.Define("pprint", NativeFunctionVarArgs(StdFnPPrint))

	return &interpreter{
		Globals:     globals,
		Env:         globals,
		Stdin:       opts.stdin,
		Stdout:      opts.stdout,
		Stderr:      opts.stderr,
		ErrReporter: opts.reporter,
		Locals:      make(map[parser.Expr]int),
	}
}

// Interpret implements Interpreter.
func (i *interpreter) Interpret(stmts []parser.Stmt) (string, error) {
	var v any
	var err error

	for _, stmt := range stmts {
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
func (i *interpreter) VisitStmtExpression(expr *parser.StmtExpression) (any, error) {
	return i.evaluate(expr.Expression)
}

// VisitStmtFunction implements parser.StmtVisitor.
func (i *interpreter) VisitStmtFunction(stmtFunction *parser.StmtFunction) (any, error) {
	function := NewLoxFunction(stmtFunction.Name, stmtFunction.Fn, i.Env, false)
	i.Env.Define(stmtFunction.Name.Lexeme, function)

	return nil, errNilnil
}

// VisitStmtIf implements parser.StmtVisitor.
func (i *interpreter) VisitStmtIf(stmtIf *parser.StmtIf) (any, error) {
	condition, err := i.evaluate(stmtIf.Condition)
	if err != nil {
		return nil, err
	}

	if i.isTruthy(condition) {
		return i.execute(stmtIf.ThenBranch)
	} else if stmtIf.ElseBranch != nil {
		return i.execute(stmtIf.ElseBranch)
	}

	return nil, errNilnil
}

// VisitPrint implements parser.StmtVisitor.
func (i *interpreter) VisitStmtPrint(expr *parser.StmtPrint) (any, error) {
	value, err := i.evaluate(expr.Expression)
	if err == nil {
		i.print(value)
	}
	return nil, err
}

// VisitStmtReturn implements parser.StmtVisitor.
func (i *interpreter) VisitStmtReturn(stmtReturn *parser.StmtReturn) (value any, err error) {
	if stmtReturn.Value != nil {
		if value, err = i.evaluate(stmtReturn.Value); err != nil {
			return nil, err
		}
	}

	return nil, &ReturnValueError{Value: value}
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

	i.Env.Define(stmt.Name.Lexeme, value)

	return nil, errNilnil
}

// VisitStmtWhile implements parser.StmtVisitor.
func (i *interpreter) VisitStmtWhile(stmtWhile *parser.StmtWhile) (any, error) {
	var condition any
	var value any
	var err error

	for err == nil {
		if condition, err = i.evaluate(stmtWhile.Condition); err != nil {
			break
		}

		if !i.isTruthy(condition) {
			break
		}

		if value, err = i.execute(stmtWhile.Body); err != nil {
			switch {
			case err == errBreak:
				// returns immediately
				return nil, errNilnil
			case err == errContinue:
				// continue to next iteration
				err = nil
			}
		}
	}

	return value, err
}

// VisitStmtFor implements parser.StmtVisitor.
func (i *interpreter) VisitStmtFor(stmtFor *parser.StmtFor) (any, error) {
	var condition any
	var value any
	var err error

	if stmtFor.Initializer != nil {
		_, err = i.execute(stmtFor.Initializer)
	}

	for err == nil {
		if condition, err = i.evaluate(stmtFor.Condition); err != nil {
			break
		}

		if !i.isTruthy(condition) {
			break
		}

		if value, err = i.execute(stmtFor.Body); err != nil {
			switch {
			case err == errBreak:
				// returns immediately
				return nil, errNilnil
			case err == errContinue:
				// continue to next iteration
				err = nil
			}
		}

		if err == nil && stmtFor.Increment != nil {
			_, err = i.evaluate(stmtFor.Increment)
		}
	}

	return value, err
}

// VisitStmtBreak implements parser.StmtVisitor.
func (*interpreter) VisitStmtBreak(stmtBreak *parser.StmtBreak) (any, error) {
	return nil, errBreak
}

// VisitStmtContinue implements parser.StmtVisitor.
func (*interpreter) VisitStmtContinue(stmtContinue *parser.StmtContinue) (any, error) {
	return nil, errContinue
}

// VisitStmtBlock implements parser.StmtVisitor.
func (i *interpreter) VisitStmtBlock(block *parser.StmtBlock) (any, error) {
	newEnv := i.Env.Nest()
	return i.executeBlock(newEnv, block.Statements)
}

// VisitStmtClass implements parser.StmtVisitor.
func (i *interpreter) VisitStmtClass(stmtClass *parser.StmtClass) (any, error) {
	var superClass *LoxClass
	if stmtClass.SuperClass != nil {
		if superClassValue, err := i.evaluate(stmtClass.SuperClass); err != nil {
			return nil, err
		} else {
			if cast, ok := superClassValue.(*LoxClass); ok {
				superClass = cast
			}
		}
		if superClass == nil {
			return i.returnRuntimeError(stmtClass.SuperClass.Name, loxerrors.ErrRuntimeSuperClassMustBeClass)
		}
	}
	env := i.Env
	env.Define(stmtClass.Name.Lexeme, nil)

	if superClass != nil {
		env = env.Nest()
		env.Define("super", superClass)
	}

	classMethods := make(map[string]*LoxFunction)
	methods := make(map[string]*LoxFunction)
	for _, method := range stmtClass.ClassMethods {
		function := NewLoxFunction(method.Name, method.Fn, env, false)
		classMethods[method.Name.Lexeme] = function
	}
	for _, method := range stmtClass.Methods {
		function := NewLoxFunction(method.Name, method.Fn, env, method.Name.Lexeme == "init")
		methods[method.Name.Lexeme] = function
	}

	class := NewLoxClass(stmtClass.Name.Lexeme, superClass, methods, classMethods)
	if superClass != nil {
		env = env.Enclosing()
	}
	return nil, env.Assign(stmtClass.Name, class)
}

// VisitExprGet implements parser.ExprVisitor.
func (i *interpreter) VisitExprGet(exprGet *parser.ExprGet) (any, error) {
	var instance any
	var err error
	if instance, err = i.evaluate(exprGet.Instance); err == nil {
		if _, ok := instance.(LoxInstance); !ok {
			err = i.runtimeError(exprGet.Name, loxerrors.ErrRuntimeOnlyInstancesHaveProperties)
		}
	}
	if err != nil {
		return nil, err
	}

	return instance.(LoxInstance).Get(exprGet.Name)
}

// VisitVariable implements parser.ExprVisitor.
func (i *interpreter) VisitExprVariable(expr *parser.ExprVariable) (any, error) {
	return i.lookupVariable(expr.Name, expr)
}

// VisitExprAssign implements parser.ExprVisitor.
func (i *interpreter) VisitExprAssign(assign *parser.ExprAssign) (any, error) {
	value, err := i.evaluate(assign.Value)
	if err != nil {
		return nil, err
	}

	return i.assignVariable(assign, value)
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
		if err := i.checkNumberOperands(expr.Operator, left, right); err != nil {
			return nil, err
		}
		return left.(float64) > right.(float64), nil
	case token.GREATER_EQUAL:
		if err := i.checkNumberOperands(expr.Operator, left, right); err != nil {
			return nil, err
		}
		return left.(float64) >= right.(float64), nil
	case token.LESS:
		if err := i.checkNumberOperands(expr.Operator, left, right); err != nil {
			return nil, err
		}
		return left.(float64) < right.(float64), nil
	case token.LESS_EQUAL:
		if err := i.checkNumberOperands(expr.Operator, left, right); err != nil {
			return nil, err
		}
		return left.(float64) <= right.(float64), nil
	case token.BANG_EQUAL:
		return !i.isEqual(left, right), nil
	case token.EQUAL_EQUAL:
		return i.isEqual(left, right), nil
	case token.MINUS:
		if err := i.checkNumberOperands(expr.Operator, left, right); err != nil {
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
		if err := i.checkNumberOperands(expr.Operator, left, right); err != nil {
			return nil, err
		}
		return left.(float64) / right.(float64), nil
	case token.STAR:
		if err := i.checkNumberOperands(expr.Operator, left, right); err != nil {
			return nil, err
		}
		return left.(float64) * right.(float64), nil
	}

	return i.unreachable()
}

// VisitExprFunction implements parser.ExprVisitor.
func (i *interpreter) VisitExprFunction(exprFunction *parser.ExprFunction) (any, error) {
	fn := NewLoxFunction(nil, exprFunction, i.Env, false)
	return fn, nil
}

// VisitExprCall implements parser.ExprVisitor.
func (i *interpreter) VisitExprCall(exprCall *parser.ExprCall) (any, error) {
	callee, err := i.evaluate(exprCall.Callee)
	if err != nil {
		return nil, err
	}
	callable, ok := callee.(Callable)
	if !ok {
		return i.returnRuntimeError(exprCall.CloseParen, loxerrors.ErrRuntimeCalleeMustBeCallable)
	}

	args := make([]any, len(exprCall.Arguments))
	for index, arg := range exprCall.Arguments {
		argValue, err := i.evaluate(arg)
		if err != nil {
			return nil, err
		}
		args[index] = argValue
	}

	if !callable.Arity().IsVarArgs() && len(args) != int(callable.Arity()) {
		return i.returnRuntimeError(exprCall.CloseParen,
			loxerrors.ErrRuntimeCalleeArityError(
				int(callable.Arity()),
				len(args),
			))
	}

	return callable.Call(i, args)
}

// VisitGrouping implements parser.Visitor.
func (i *interpreter) VisitExprGrouping(expr *parser.ExprGrouping) (any, error) {
	return i.evaluate(expr.Expression)
}

// VisitLiteral implements parser.Visitor.
func (i *interpreter) VisitExprLiteral(expr *parser.ExprLiteral) (any, error) {
	return expr.Value, nil
}

// VisitExprLogical implements parser.ExprVisitor.
func (i *interpreter) VisitExprLogical(exprLogical *parser.ExprLogical) (any, error) {
	switch exprLogical.Operator.Type {
	case token.AND:
		return i.evalLogicalAnd(exprLogical.Left, exprLogical.Right)
	case token.OR:
		return i.evalLogicalOr(exprLogical.Left, exprLogical.Right)
	default:
		return i.unreachable()
	}
}

// VisitExprSet implements parser.ExprVisitor.
func (i *interpreter) VisitExprSet(exprSet *parser.ExprSet) (any, error) {
	var instance any
	var err error
	if instance, err = i.evaluate(exprSet.Instance); err == nil {
		if _, ok := instance.(LoxInstance); !ok {
			err = i.runtimeError(exprSet.Name, loxerrors.ErrRuntimeOnlyInstancesHaveFields)
		}
	}
	if err != nil {
		return nil, err
	}

	value, err := i.evaluate(exprSet.Value)
	if err != nil {
		return nil, err
	}

	return instance.(LoxInstance).Set(exprSet.Name, value)
}

// VisitExprSuper implements parser.ExprVisitor.
func (i *interpreter) VisitExprSuper(exprSuper *parser.ExprSuper) (any, error) {
	var distance int
	if depth, ok := i.Locals[exprSuper]; !ok {
		return i.unreachable()
	} else {
		distance = depth
	}

	var superClass *LoxClass
	if _superClass, err := i.Env.GetAt(distance, "super"); err != nil {
		return nil, err
	} else if _superClass, ok := _superClass.(*LoxClass); !ok {
		return i.unreachable()
	} else {
		superClass = _superClass
	}

	var instance LoxInstance
	if _instance, err := i.Env.GetAt(distance-1, "this"); err != nil {
		return nil, err
	} else if _instance, ok := _instance.(LoxInstance); !ok {
		return i.unreachable()
	} else {
		instance = _instance
	}

	method := superClass.FindMethod(exprSuper.Method.Lexeme)
	if method == nil {
		return i.returnRuntimeError(exprSuper.Method, loxerrors.ErrRuntimeUndefinedProperty(exprSuper.Method.Lexeme))
	}
	return method.Bind(instance), nil
}

// VisitExprThis implements parser.ExprVisitor.
func (i *interpreter) VisitExprThis(exprThis *parser.ExprThis) (any, error) {
	return i.lookupVariable(exprThis.Keyword, exprThis)
}

func (i *interpreter) evalLogicalAnd(left, right parser.Expr) (any, error) {
	if leftValue, err := i.evaluate(left); err != nil || !i.isTruthy(leftValue) {
		return leftValue, err
	}

	return i.evaluate(right)
}

func (i *interpreter) evalLogicalOr(left, right parser.Expr) (any, error) {
	if leftValue, err := i.evaluate(left); err != nil || i.isTruthy(leftValue) {
		return leftValue, err
	}

	return i.evaluate(right)
}

// VisitUnary implements parser.Visitor.
func (i *interpreter) VisitExprUnary(expr *parser.ExprUnary) (any, error) {
	right, err := i.evaluate(expr.Right)
	if err != nil {
		return nil, err
	}

	switch expr.Operator.Type {
	case token.MINUS:
		if err := i.checkNumberOperand(expr.Operator, right); err != nil {
			return nil, err
		}
		return -right.(float64), nil
	case token.BANG:
		return !i.isTruthy(right), nil
	}

	return i.unreachable()
}

func (i *interpreter) execute(stmt parser.Stmt) (any, error) {
	value, err := stmt.Accept(i)
	return value, err
}

func (i *interpreter) executeBlock(env *environment, stmt []parser.Stmt) (value any, err error) {
	oldEnv := i.setEnv(env)
	defer i.setEnv(oldEnv)

	for _, stmt := range stmt {
		if _, err = i.execute(stmt); err != nil {
			return nil, err
		}
	}

	return nil, errNilnil
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

func (i *interpreter) returnRuntimeError(tok *token.Token, err error) (any, error) {
	return nil, i.runtimeError(tok, err)
}

func (i *interpreter) runtimeError(tok *token.Token, err error) error {
	return loxerrors.NewRuntimeError(tok, err)
}

func (i *interpreter) resolve(expr parser.Expr, depth int) {
	i.Locals[expr] = depth
}

func (i *interpreter) lookupVariable(name *token.Token, expr parser.Expr) (any, error) {
	if distance, ok := i.Locals[expr]; ok {
		return i.Env.GetAt(distance, name.Lexeme)
	}

	return i.Globals.Get(name)
}

func (i *interpreter) assignVariable(expr *parser.ExprAssign, value any) (any, error) {
	if distance, ok := i.Locals[expr]; ok {
		return i.Env.AssignAt(distance, expr.Name, value)
	}

	return value, i.Globals.Assign(expr.Name, value)
}

func (i *interpreter) setEnv(env *environment) *environment {
	oldEnv := i.Env
	i.Env = env
	return oldEnv
}

func (i *interpreter) unreachable() (any, error) {
	panic("unreachable")
}

var (
	_ parser.ExprVisitor = (*interpreter)(nil)
	_ parser.StmtVisitor = (*interpreter)(nil)
	_ Interpreter        = (*interpreter)(nil)
)
