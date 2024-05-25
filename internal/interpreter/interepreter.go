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
	Evaluate(stmt parser.Stmt) error
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
	globals.Define("Array", ValueCallable{NativeFunction1(StdFnCreateArray)})
	globals.Define("clock", ValueCallable{NativeFunction0(StdFnTime)})
	globals.Define("pprint", ValueCallable{NativeFunctionVarArgs(StdFnPPrint)})

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

	for _, stmt := range stmts {
		if err := i.Evaluate(stmt); err != nil {
			return "", err
		}
	}

	return i.stringify(v), nil
}

// Evaluate implements Interpreter.
func (i *interpreter) Evaluate(stmt parser.Stmt) error {
	return i.execute(stmt)
}

func (i *interpreter) print(v ...Value) {
	vv := make([]any, len(v))
	for index, vvv := range v {
		vv[index] = i.stringify(vvv)
	}

	_, _ = fmt.Fprintln(i.Stdout, vv...)
}

func (i *interpreter) stringify(v any) string {
	if v == nil {
		return "nil"
	}
	return fmt.Sprintf("%#v", v)
}

// VisitExpression implements parser.StmtVisitor.
func (i *interpreter) VisitStmtExpression(expr *parser.StmtExpression) error {
	_, err := i.evaluate(expr.Expression)
	return err
}

// VisitStmtFunction implements parser.StmtVisitor.
func (i *interpreter) VisitStmtFunction(stmtFunction *parser.StmtFunction) error {
	function := NewLoxFunction(stmtFunction.Name, stmtFunction.Fn, i.Env, false)
	i.Env.Define(stmtFunction.Name.Lexeme, ValueCallable{function})

	return ErrNilNil
}

// VisitStmtIf implements parser.StmtVisitor.
func (i *interpreter) VisitStmtIf(stmtIf *parser.StmtIf) error {
	condition, err := i.evaluate(stmtIf.Condition)
	if err != nil {
		return err
	}

	if i.isTruthy(condition) {
		return i.execute(stmtIf.ThenBranch)
	} else if stmtIf.ElseBranch != nil {
		return i.execute(stmtIf.ElseBranch)
	}

	return ErrNilNil
}

// VisitPrint implements parser.StmtVisitor.
func (i *interpreter) VisitStmtPrint(expr *parser.StmtPrint) error {
	value, err := i.evaluate(expr.Expression)
	if err == nil {
		i.print(value)
	}
	return err
}

// VisitStmtReturn implements parser.StmtVisitor.
func (i *interpreter) VisitStmtReturn(stmtReturn *parser.StmtReturn) error {
	var value Value
	var err error
	if stmtReturn.Value != nil {
		if value, err = i.evaluate(stmtReturn.Value); err != nil {
			return err
		}
	}

	return &ReturnValueError{Value: value}
}

// VisitVar implements parser.StmtVisitor.
func (i *interpreter) VisitStmtVar(stmt *parser.StmtVar) error {
	var value Value
	var err error
	if stmt.Initializer != nil {
		if value, err = i.evaluate(stmt.Initializer); err != nil {
			return err
		}
	}

	i.Env.Define(stmt.Name.Lexeme, value)

	return ErrNilNil
}

// VisitStmtWhile implements parser.StmtVisitor.
func (i *interpreter) VisitStmtWhile(stmtWhile *parser.StmtWhile) error {
	var condition Value
	var err error

	for err == nil {
		if condition, err = i.evaluate(stmtWhile.Condition); err != nil {
			break
		}

		if !i.isTruthy(condition) {
			break
		}

		if err = i.execute(stmtWhile.Body); err != nil {
			switch {
			case err == errBreak:
				// returns immediately
				return ErrNilNil
			case err == errContinue:
				// continue to next iteration
				err = nil
			}
		}
	}

	return err
}

// VisitStmtFor implements parser.StmtVisitor.
func (i *interpreter) VisitStmtFor(stmtFor *parser.StmtFor) error {
	var condition Value
	var err error

	if stmtFor.Initializer != nil {
		err = i.execute(stmtFor.Initializer)
	}

	for err == nil {
		if condition, err = i.evaluate(stmtFor.Condition); err != nil {
			break
		}

		if !i.isTruthy(condition) {
			break
		}

		if err = i.execute(stmtFor.Body); err != nil {
			switch {
			case err == errBreak:
				// returns immediately
				return ErrNilNil
			case err == errContinue:
				// continue to next iteration
				err = nil
			}
		}

		if err == nil && stmtFor.Increment != nil {
			_, err = i.evaluate(stmtFor.Increment)
		}
	}

	return err
}

// VisitStmtBreak implements parser.StmtVisitor.
func (*interpreter) VisitStmtBreak(stmtBreak *parser.StmtBreak) error {
	return errBreak
}

// VisitStmtContinue implements parser.StmtVisitor.
func (*interpreter) VisitStmtContinue(stmtContinue *parser.StmtContinue) error {
	return errContinue
}

// VisitStmtBlock implements parser.StmtVisitor.
func (i *interpreter) VisitStmtBlock(block *parser.StmtBlock) error {
	newEnv := i.Env.Nest()
	return i.executeBlock(newEnv, block.Statements)
}

// VisitStmtClass implements parser.StmtVisitor.
func (i *interpreter) VisitStmtClass(stmtClass *parser.StmtClass) error {
	var superClass *LoxClass
	if stmtClass.SuperClass != nil {
		if superClassValue, err := i.evaluate(stmtClass.SuperClass); err != nil {
			return err
		} else {
			if cast, ok := i.asLoxClass(superClassValue); ok {
				superClass = cast
			}
		}
		if superClass == nil {
			return i.runtimeError(stmtClass.SuperClass.Name, loxerrors.ErrRuntimeSuperClassMustBeClass)
		}
	}
	env := i.Env
	env.Define(stmtClass.Name.Lexeme, nil)

	if superClass != nil {
		env = env.Nest()
		env.Define("super", ValueClass{superClass})
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
	return env.Assign(stmtClass.Name, ValueClass{class})
}

// VisitExprGet implements parser.ExprVisitor.
func (i *interpreter) VisitExprGet(exprGet *parser.ExprGet) (Value, error) {
	var eval parser.Value
	var instance LoxObject
	var err error
	if eval, err = i.evaluate(exprGet.Instance); err == nil {
		var ok bool
		if instance, ok = i.asLoxInstance(eval); !ok {
			err = i.runtimeError(exprGet.Name, loxerrors.ErrRuntimeOnlyInstancesHaveProperties)
		}
	}
	if err != nil {
		return NilValue, err
	}

	return instance.Get(exprGet.Name)
}

// VisitVariable implements parser.ExprVisitor.
func (i *interpreter) VisitExprVariable(expr *parser.ExprVariable) (Value, error) {
	return i.lookupVariable(expr.Name, expr)
}

// VisitExprAssign implements parser.ExprVisitor.
func (i *interpreter) VisitExprAssign(assign *parser.ExprAssign) (Value, error) {
	value, err := i.evaluate(assign.Value)
	if err != nil {
		return NilValue, err
	}

	return i.assignVariable(assign, value)
}

// VisitBinary implements parser.Visitor.
func (i *interpreter) VisitExprBinary(expr *parser.ExprBinary) (Value, error) {
	left, err := i.evaluate(expr.Left)
	if err != nil {
		return NilValue, err
	}
	right, err := i.evaluate(expr.Right)
	if err != nil {
		return NilValue, err
	}

	switch expr.Operator.Type {
	case token.GREATER:
		if err := i.checkNumberOperands(expr.Operator, left, right); err != nil {
			return NilValue, err
		}
		return ValueBool(float64(left.(ValueFloat)) > float64(right.(ValueFloat))), nil
	case token.GREATER_EQUAL:
		if err := i.checkNumberOperands(expr.Operator, left, right); err != nil {
			return NilValue, err
		}
		return ValueBool(float64(left.(ValueFloat)) >= float64(right.(ValueFloat))), nil
	case token.LESS:
		if err := i.checkNumberOperands(expr.Operator, left, right); err != nil {
			return NilValue, err
		}
		return ValueBool(float64(left.(ValueFloat)) < float64(right.(ValueFloat))), nil
	case token.LESS_EQUAL:
		if err := i.checkNumberOperands(expr.Operator, left, right); err != nil {
			return NilValue, err
		}
		return ValueBool(float64(left.(ValueFloat)) <= float64(right.(ValueFloat))), nil
	case token.BANG_EQUAL:
		return !i.isEqual(left, right), nil
	case token.EQUAL_EQUAL:
		return i.isEqual(left, right), nil
	case token.MINUS:
		if err := i.checkNumberOperands(expr.Operator, left, right); err != nil {
			return NilValue, err
		}
		return ValueFloat(float64(left.(ValueFloat)) - float64(right.(ValueFloat))), nil
	case token.PLUS:
		if left, ok := left.(ValueFloat); ok {
			if right, ok := right.(ValueFloat); ok {
				return ValueFloat(left + right), nil
			}
		}
		if left, ok := left.(ValueString); ok {
			if right, ok := right.(ValueString); ok {
				return left + right, nil
			}
		}
		return i.returnRuntimeError(expr.Operator, loxerrors.ErrRuntimeOperandsMustNumbersOrStrings)
	case token.SLASH:
		if err := i.checkNumberOperands(expr.Operator, left, right); err != nil {
			return NilValue, err
		}
		return ValueFloat(float64(left.(ValueFloat)) / float64(right.(ValueFloat))), nil
	case token.STAR:
		if err := i.checkNumberOperands(expr.Operator, left, right); err != nil {
			return NilValue, err
		}
		return ValueFloat(float64(left.(ValueFloat)) + float64(right.(ValueFloat))), nil
	}

	return i.unreachable()
}

// VisitExprFunction implements parser.ExprVisitor.
func (i *interpreter) VisitExprFunction(exprFunction *parser.ExprFunction) (Value, error) {
	fn := NewLoxFunction(nil, exprFunction, i.Env, false)
	return ValueCallable{fn}, nil
}

// VisitExprCall implements parser.ExprVisitor.
func (i *interpreter) VisitExprCall(exprCall *parser.ExprCall) (Value, error) {
	callee, err := i.evaluate(exprCall.Callee)
	if err != nil {
		return NilValue, err
	}
	callable, ok := i.asCallable(callee)
	if !ok {
		return i.returnRuntimeError(exprCall.CloseParen, loxerrors.ErrRuntimeCalleeMustBeCallable)
	}

	args := make([]Value, len(exprCall.Arguments))
	for index, arg := range exprCall.Arguments {
		argValue, err := i.evaluate(arg)
		if err != nil {
			return NilValue, err
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
func (i *interpreter) VisitExprGrouping(expr *parser.ExprGrouping) (Value, error) {
	return i.evaluate(expr.Expression)
}

// VisitLiteral implements parser.Visitor.
func (i *interpreter) VisitExprLiteral(expr *parser.ExprLiteral) (Value, error) {
	if expr.Value == nil {
		return NilValue, nil
	}
	switch expr.Value.(type) { //nolint:gocritic // ok
	case bool:
		return ValueBool(expr.Value.(bool)), nil
	case float64:
		return ValueFloat(expr.Value.(float64)), nil
	case string:
		return ValueString(expr.Value.(string)), nil
	case nil:
		return NilValue, nil
	default:
		return i.unreachable()
	}
}

// VisitExprLogical implements parser.ExprVisitor.
func (i *interpreter) VisitExprLogical(exprLogical *parser.ExprLogical) (Value, error) {
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
func (i *interpreter) VisitExprSet(exprSet *parser.ExprSet) (Value, error) {
	var eval parser.Value
	var instance LoxObject
	var err error
	if eval, err = i.evaluate(exprSet.Instance); err == nil {
		var ok bool
		if instance, ok = i.asLoxInstance(eval); !ok {
			err = i.runtimeError(exprSet.Name, loxerrors.ErrRuntimeOnlyInstancesHaveFields)
		}
	}
	if err != nil {
		return NilValue, err
	}

	value, err := i.evaluate(exprSet.Value)
	if err != nil {
		return NilValue, err
	}

	return instance.Set(exprSet.Name, value)
}

// VisitExprSuper implements parser.ExprVisitor.
func (i *interpreter) VisitExprSuper(exprSuper *parser.ExprSuper) (Value, error) {
	var distance int
	if depth, ok := i.Locals[exprSuper]; !ok {
		return i.unreachable()
	} else {
		distance = depth
	}

	var superClass *LoxClass
	if _superClass, err := i.Env.GetAt(distance, "super"); err != nil {
		return NilValue, err
	} else if _superClass, ok := i.asLoxClass(_superClass); !ok {
		return i.unreachable()
	} else {
		superClass = _superClass
	}

	var instance LoxObject
	if _instance, err := i.Env.GetAt(distance-1, "this"); err != nil {
		return NilValue, err
	} else if _instance, ok := i.asLoxInstance(_instance); !ok {
		return i.unreachable()
	} else {
		instance = _instance
	}

	method := superClass.FindMethod(exprSuper.Method.Lexeme)
	if method == nil {
		return i.returnRuntimeError(exprSuper.Method, loxerrors.ErrRuntimeUndefinedProperty(exprSuper.Method.Lexeme))
	}
	return ValueCallable{method.Bind(instance)}, nil
}

// VisitExprThis implements parser.ExprVisitor.
func (i *interpreter) VisitExprThis(exprThis *parser.ExprThis) (Value, error) {
	return i.lookupVariable(exprThis.Keyword, exprThis)
}

func (i *interpreter) evalLogicalAnd(left, right parser.Expr) (Value, error) {
	if leftValue, err := i.evaluate(left); err != nil || !i.isTruthy(leftValue) {
		return leftValue, err
	}

	return i.evaluate(right)
}

func (i *interpreter) evalLogicalOr(left, right parser.Expr) (Value, error) {
	if leftValue, err := i.evaluate(left); err != nil || i.isTruthy(leftValue) {
		return leftValue, err
	}

	return i.evaluate(right)
}

// VisitUnary implements parser.Visitor.
func (i *interpreter) VisitExprUnary(expr *parser.ExprUnary) (Value, error) {
	right, err := i.evaluate(expr.Right)
	if err != nil {
		return NilValue, err
	}

	switch expr.Operator.Type {
	case token.MINUS:
		if err := i.checkNumberOperand(expr.Operator, right); err != nil {
			return NilValue, err
		}
		return ValueFloat(-(right.(ValueFloat))), nil
	case token.BANG:
		return ValueBool(!i.isTruthy(right)), nil
	}

	return i.unreachable()
}

func (i *interpreter) execute(stmt parser.Stmt) error {
	return stmt.Accept(i)
}

func (i *interpreter) executeBlock(env *environment, stmt []parser.Stmt) error {
	oldEnv := i.setEnv(env)
	defer i.setEnv(oldEnv)

	for _, stmt := range stmt {
		if err := i.execute(stmt); err != nil {
			return err
		}
	}

	return ErrNilNil
}

func (i *interpreter) evaluate(expr parser.Expr) (Value, error) {
	return expr.Accept(i)
}

func (i *interpreter) isTruthy(value Value) bool {
	if value == nil || value == NilValue {
		return false
	}
	if value.Type() == parser.ValueBoolType {
		return bool(value.(ValueBool))
	}

	return true
}

func (i *interpreter) isEqual(left, right Value) ValueBool {
	if left == nil && right == nil {
		return true
	}
	return ValueBool(left == right)
}

func (i *interpreter) checkNumberOperands(tok *token.Token, left, right Value) error {
	if left.Type() != parser.ValueFloatType {
		return i.runtimeError(tok, loxerrors.ErrRuntimeOperandsMustBeNumbers)
	} else if right.Type() != parser.ValueFloatType {
		return i.runtimeError(tok, loxerrors.ErrRuntimeOperandsMustBeNumbers)
	}
	return nil
}

func (i *interpreter) checkNumberOperand(tok *token.Token, val Value) error {
	if val.Type() != parser.ValueFloatType {
		return i.runtimeError(tok, loxerrors.ErrRuntimeOperandMustBeNumber)
	}

	return nil
}

func (i *interpreter) returnRuntimeError(tok *token.Token, err error) (Value, error) {
	return NilValue, i.runtimeError(tok, err)
}

func (i *interpreter) runtimeError(tok *token.Token, err error) error {
	return loxerrors.NewRuntimeError(tok, err)
}

func (i *interpreter) resolve(expr parser.Expr, depth int) {
	i.Locals[expr] = depth
}

func (i *interpreter) lookupVariable(name *token.Token, expr parser.Expr) (Value, error) {
	if distance, ok := i.Locals[expr]; ok {
		return i.Env.GetAt(distance, name.Lexeme)
	}

	return i.Globals.Get(name)
}

func (i *interpreter) assignVariable(expr *parser.ExprAssign, value Value) (Value, error) {
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

func (i *interpreter) unreachable() (Value, error) {
	panic("unreachable")
}

func (i *interpreter) asCallable(value Value) (Callable, bool) {
	if value.Type() == parser.ValueCallableType {
		return value.(Callable), true
	}
	if value.Type() == parser.ValueClassType {
		v := value.(ValueClass)
		return Callable(v.LoxClass), true
	}
	return nil, false
}

func (i *interpreter) asLoxClass(value Value) (*LoxClass, bool) {
	if value.Type() == parser.ValueClassType {
		return value.(ValueClass).LoxClass, true
	}
	return nil, false
}

func (i *interpreter) asLoxInstance(value Value) (LoxObject, bool) {
	if value.Type() == parser.ValueObjectType {
		if vc, ok := value.(ValueObject); ok {
			return vc.LoxObject, true
		}
	}
	return nil, false
}

var (
	_ parser.ExprVisitor = (*interpreter)(nil)
	_ parser.StmtVisitor = (*interpreter)(nil)
	_ Interpreter        = (*interpreter)(nil)
)
