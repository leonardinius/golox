package parser

import (
	"fmt"

	"github.com/leonardinius/golox/internal/loxerrors"
	"github.com/leonardinius/golox/internal/token"
)

var (
	nilExpr       Expr   = nil
	nilStmt       Stmt   = nil
	nilStatements []Stmt = nil
	maxArguments         = 255
)

type Parser interface {
	Parse() ([]Stmt, error)
}

type parser struct {
	tokens    []token.Token
	current   int
	reporter  loxerrors.ErrReporter
	loopDepth int
	funcDepth int
	panic     error
	err       error
	recover   bool
}

func NewParser(tokens []token.Token, reporter loxerrors.ErrReporter) Parser {
	if len(tokens) == 0 {
		panic("tokens cannot be empty")
	}
	if tokens[len(tokens)-1].Type != token.EOF {
		panic("tokens must end with EOF")
	}

	return &parser{
		tokens:   tokens,
		current:  0,
		reporter: reporter,
	}
}

// GoString implements fmt.GoStringer.
func (p *parser) GoString() string {
	return fmt.Sprintf("parser{tokens: %#v, current: %d, err: %#v}", p.tokens, p.current, p.panic)
}

// String implements fmt.Stringer.
func (p *parser) String() string {
	return fmt.Sprintf("parser{tokens: %d, err: %v}", len(p.tokens), p.panic)
}

// Parse implements Parser.
func (p *parser) Parse() (statements []Stmt, err error) {
	var stmt Stmt
	for !p.isDone() {
		stmt, err = p.declaration(), p.panic
		if err != nil {
			break
		}
		statements = append(statements, stmt)
	}

	if err == nil && p.err == nil {
		return statements, nil
	}

	// if we are at error state, we do not return invalid ast tree
	// return nil, err instead
	for !p.isAtEnd() {
		p.synchronize()
		p.panic = nil
		p.recover = true
		_ = p.declaration()
	}

	return nilStatements, loxerrors.ErrParseError
}

func (p *parser) declaration() Stmt {
	if p.match(token.CLASS) {
		return p.classDeclaration()
	}

	if p.check(token.FUN) && p.checkNext(token.IDENTIFIER) {
		p.advance()
		return p.funDeclaration("function")
	}

	if p.match(token.VAR) {
		return p.varDeclaration()
	}

	return p.statement()
}

func (p *parser) classDeclaration() Stmt {
	if !p.match(token.IDENTIFIER) {
		return p.reportFatalErrorStmt(loxerrors.ErrParseExpectClassName)
	}
	name := p.previous()

	var superClass *ExprVariable
	if p.match(token.LESS) {
		if !p.match(token.IDENTIFIER) {
			return p.reportFatalErrorStmt(loxerrors.ErrParseExpectSuperClassName)
		}
		superClass = &ExprVariable{Name: p.previous()}
	}

	if !p.match(token.LEFT_BRACE) {
		return p.reportFatalErrorStmt(loxerrors.ErrParseExpectLeftCurlyBeforeClassBody)
	}

	var methods []*StmtFunction
	var classMethods []*StmtFunction
	for !p.check(token.RIGHT_BRACE) && !p.isDone() {
		if p.match(token.CLASS) {
			classMethods = append(classMethods, p.funDeclaration("method"))
		} else {
			methods = append(methods, p.funDeclaration("method"))
		}
	}

	if !p.match(token.RIGHT_BRACE) {
		return p.reportFatalErrorStmt(loxerrors.ErrParseExpectRightCurlyAfterClassBody)
	}

	return &StmtClass{Name: name, SuperClass: superClass, Methods: methods, ClassMethods: classMethods}
}

func (p *parser) funDeclaration(kind string) *StmtFunction {
	// function name
	if !p.match(token.IDENTIFIER) {
		p.reportFatalErrorStmt(loxerrors.ErrParseExpectedIdentifierKindError(kind))
		return nil
	}
	name := p.previous()
	if fn, ok := p.functionBody(kind).(*ExprFunction); ok {
		return &StmtFunction{Name: name, Fn: fn}
	}

	return nil
}

func (p *parser) functionBody(kind string) Expr {
	if !p.match(token.LEFT_PAREN) {
		return p.reportFatalErrorExpr(loxerrors.ErrParseExpectedLeftParenError(kind))
	}

	var params []*token.Token
	if !p.check(token.RIGHT_PAREN) {
		for {
			if len(params) >= maxArguments {
				switch kind {
				case "method":
					p.reportErrorExpr(loxerrors.ErrParseTooManyParameters)
				default:
					p.reportErrorExpr(loxerrors.ErrParseTooManyArguments)
				}

			}

			if !p.match(token.IDENTIFIER) {
				return p.reportFatalErrorExpr(loxerrors.ErrParseUnexpectedParameterName)
			}
			params = append(params, p.previous())

			if !p.match(token.COMMA) {
				break
			}
		}
	}

	// function body
	if !p.match(token.RIGHT_PAREN) {
		return p.reportFatalErrorExpr(loxerrors.ErrParseExpectedRightParentFunToken)
	}
	if !p.match(token.LEFT_BRACE) {
		return p.reportFatalErrorExpr(loxerrors.ErrParseExpectedLeftBraceFunToken(kind))
	}

	p.funcDepth++
	defer func() { p.funcDepth-- }()
	body := p.blockStatement()

	return &ExprFunction{Parameters: params, Body: body}
}

func (p *parser) varDeclaration() Stmt {

	if !p.match(token.IDENTIFIER) {
		return p.reportFatalErrorStmt(loxerrors.ErrParseUnexpectedVariableName)
	}
	name := p.previous()

	var initializer Expr = nilExpr
	if p.match(token.EQUAL) {
		initializer = p.expression()
	}

	if !p.match(token.SEMICOLON) {
		return p.reportFatalErrorStmt(loxerrors.ErrParseExpectedSemicolonTokenAfterVar)
	}

	return &StmtVar{Name: name, Initializer: initializer}
}

func (p *parser) statement() Stmt {
	if p.match(token.FOR) {
		return p.forStatement()
	}

	if p.match(token.IF) {
		return p.ifStatement()
	}

	if p.match(token.PRINT) {
		return p.printStatement()
	}

	if p.match(token.RETURN) {
		return p.returnStatement()
	}

	if p.match(token.WHILE) {
		return p.whileStatement()
	}

	if p.match(token.BREAK) {
		return p.breakStatement()
	}

	if p.match(token.LEFT_BRACE) {
		block := p.blockStatement()
		return &StmtBlock{Statements: block}
	}

	if p.match(token.CONTINUE) {
		return p.continueStatement()
	}

	return p.expressionStatement()
}

func (p *parser) ifStatement() Stmt {

	if !p.match(token.LEFT_PAREN) {
		return p.reportFatalErrorStmt(loxerrors.ErrParseExpectedLeftParentIfToken)
	}

	condition := p.expression()

	if !p.match(token.RIGHT_PAREN) {
		return p.reportFatalErrorStmt(loxerrors.ErrParseExpectedRightParentIfToken)
	}

	thenBranch := p.statement()
	var elseBranch Stmt
	if p.match(token.ELSE) {
		elseBranch = p.statement()
	}

	return &StmtIf{Condition: condition, ThenBranch: thenBranch, ElseBranch: elseBranch}
}

func (p *parser) printStatement() Stmt {

	expr := p.expression()

	if !p.match(token.SEMICOLON) {
		return p.reportFatalErrorStmt(loxerrors.ErrParseExpectedSemicolonTokenAfterPrintValue)
	}

	return &StmtPrint{Expression: expr}
}

func (p *parser) returnStatement() Stmt {
	tok := p.previous()

	if p.funcDepth == 0 {
		return p.reportFatalErrorStmtToken(tok, loxerrors.ErrParseReturnOutsideFunction)
	}

	var value Expr = nilExpr
	if !p.check(token.SEMICOLON) {
		value = p.expression()
	}

	if !p.match(token.SEMICOLON) {
		return p.reportFatalErrorStmt(loxerrors.ErrParseExpectedSemicolonTokenAfterReturn)
	}

	return &StmtReturn{Keyword: tok, Value: value}
}

func (p *parser) whileStatement() Stmt {

	if !p.match(token.LEFT_PAREN) {
		return p.reportFatalErrorStmt(loxerrors.ErrParseExpectedLeftParentWhileToken)
	}
	condition := p.expression()
	if !p.match(token.RIGHT_PAREN) {
		return p.reportFatalErrorStmt(loxerrors.ErrParseExpectedRightParentWhileToken)
	}

	p.loopDepth++
	defer func() { p.loopDepth-- }()
	body := p.statement()

	return &StmtWhile{Condition: condition, Body: body}
}

func (p *parser) forStatement() Stmt {
	if !p.match(token.LEFT_PAREN) {
		return p.reportFatalErrorStmt(loxerrors.ErrParseExpectedLeftParentForToken)
	}

	var initializer Stmt
	if p.match(token.SEMICOLON) {
		initializer = nilStmt
	} else if p.match(token.VAR) {
		initializer = p.varDeclaration()
	} else {
		initializer = p.expressionStatement()
	}

	var condition Expr
	if !p.check(token.SEMICOLON) {
		condition = p.expression()
	}
	if !p.match(token.SEMICOLON) {
		return p.reportFatalErrorStmt(loxerrors.ErrParseExpectedSemicolonAfterForLoopCond)
	}

	var increment Expr
	if !p.check(token.RIGHT_PAREN) {
		increment = p.expression()
	}
	if !p.match(token.RIGHT_PAREN) {
		return p.reportFatalErrorStmt(loxerrors.ErrParseExpectedRightParentForToken)
	}

	p.loopDepth++
	defer func() { p.loopDepth-- }()
	body := p.statement()

	if condition == nilExpr {
		condition = &ExprLiteral{Value: true}
	}

	return &StmtFor{Initializer: initializer, Condition: condition, Increment: increment, Body: body}
}

func (p *parser) breakStatement() Stmt {
	if p.loopDepth == 0 {
		return p.reportFatalErrorStmt(loxerrors.ErrParseBreakOutsideLoop)
	}
	if !p.match(token.SEMICOLON) {
		return p.reportFatalErrorStmt(loxerrors.ErrParseExpectedSemicolonTokenAfterBreak)
	}
	return &StmtBreak{}
}

func (p *parser) continueStatement() Stmt {
	if p.loopDepth == 0 {
		return p.reportFatalErrorStmt(loxerrors.ErrParseContinueOutsideLoop)
	}
	if !p.match(token.SEMICOLON) {
		return p.reportFatalErrorStmt(loxerrors.ErrParseExpectedSemicolonTokenAfterContinue)
	}
	return &StmtContinue{}
}

func (p *parser) blockStatement() []Stmt {
	var stmts []Stmt
	for !p.check(token.RIGHT_BRACE) && !p.isDone() {
		stmts = append(stmts, p.declaration())
	}

	if !p.match(token.RIGHT_BRACE) {
		return p.reportFatalErrorStmtList(loxerrors.ErrParseExpectedRightCurlyBlockToken)
	}

	return stmts
}

func (p *parser) expressionStatement() Stmt {
	expr := p.expression()
	if !p.match(token.SEMICOLON) {
		return p.reportFatalErrorStmt(loxerrors.ErrParseExpectedRightParenToken)
	}
	return &StmtExpression{Expression: expr}
}

func (p *parser) expression() Expr {
	return p.assignment()
}

func (p *parser) assignment() Expr {
	expr := p.logicOr()

	if p.match(token.EQUAL) {
		equals := p.previous()
		value := p.assignment()

		if v, ok := expr.(*ExprVariable); ok {
			return &ExprAssign{Name: v.Name, Value: value}
		} else if v, ok := expr.(*ExprGet); ok {
			return &ExprSet{Instance: v.Instance, Name: v.Name, Value: value}
		}

		p.reportErrorExprToken(equals, loxerrors.ErrParseInvalidAssignmentTarget)
	}

	return expr
}

func (p *parser) logicOr() Expr {
	expr := p.logicAnd()

	for p.match(token.OR) {
		operator := p.previous()
		right := p.logicAnd()
		return &ExprLogical{Left: expr, Operator: operator, Right: right}
	}

	return expr
}

func (p *parser) logicAnd() Expr {
	expr := p.equality()

	for p.match(token.AND) {
		operator := p.previous()
		right := p.equality()
		return &ExprLogical{Left: expr, Operator: operator, Right: right}
	}

	return expr
}

func (p *parser) equality() Expr {
	expr := p.comparison()

	for p.anyMatch(token.BANG_EQUAL, token.EQUAL_EQUAL) {
		operator := p.previous()
		right := p.comparison()
		expr = &ExprBinary{Left: expr, Operator: operator, Right: right}
	}

	return expr
}

func (p *parser) comparison() Expr {
	expr := p.term()

	for p.anyMatch(token.GREATER, token.GREATER_EQUAL, token.LESS, token.LESS_EQUAL) {
		operator := p.previous()
		right := p.term()
		expr = &ExprBinary{Left: expr, Operator: operator, Right: right}
	}

	return expr
}

func (p *parser) term() Expr {
	expr := p.factor()

	for p.anyMatch(token.MINUS, token.PLUS) {
		operator := p.previous()
		right := p.factor()
		expr = &ExprBinary{Left: expr, Operator: operator, Right: right}
	}

	return expr
}

func (p *parser) factor() Expr {
	expr := p.unary()

	for p.anyMatch(token.SLASH, token.STAR) {
		operator := p.previous()
		right := p.unary()
		expr = &ExprBinary{Left: expr, Operator: operator, Right: right}
	}

	return expr
}

func (p *parser) unary() Expr {
	if p.anyMatch(token.BANG, token.MINUS) {
		operator := p.previous()
		right := p.unary()
		return &ExprUnary{Operator: operator, Right: right}
	}

	return p.call()
}

func (p *parser) call() Expr {
	expr := p.primary()

	for {
		if p.match(token.LEFT_PAREN) {
			expr = p.finishCall(expr)
		} else if p.match(token.DOT) {
			if !p.match(token.IDENTIFIER) {
				return p.reportFatalErrorExpr(loxerrors.ErrParseExpectedPropertyNameAfterDot)
			}
			name := p.previous()
			expr = &ExprGet{Instance: expr, Name: name}
		} else {
			break
		}
	}

	return expr
}

func (p *parser) finishCall(callee Expr) Expr {
	var args []Expr
	if !p.check(token.RIGHT_PAREN) {
		for {
			if len(args) >= maxArguments {
				p.reportErrorExpr(loxerrors.ErrParseTooManyArguments)
			}
			args = append(args, p.expression())
			if !p.match(token.COMMA) {
				break
			}
		}
	}

	if !p.match(token.RIGHT_PAREN) {
		return p.reportFatalErrorExpr(loxerrors.ErrParseExpectedRightParenToken)
	}
	paren := p.previous()

	return &ExprCall{Callee: callee, CloseParen: paren, Arguments: args}

}

func (p *parser) primary() Expr {
	if p.match(token.FALSE) {
		return &ExprLiteral{Value: false}
	}
	if p.match(token.TRUE) {
		return &ExprLiteral{Value: true}
	}
	if p.match(token.NIL) {
		return &ExprLiteral{Value: nil}
	}
	if p.match(token.FUN) {
		return p.functionBody("function")
	}

	if p.anyMatch(token.NUMBER, token.STRING) {
		tok := p.previous()
		return &ExprLiteral{Value: tok.Literal}
	}

	if p.match(token.SUPER) {
		tok := p.previous()
		if !p.match(token.DOT) {
			return p.reportFatalErrorExpr(loxerrors.ErrParseExpectedDotAfterSuper)
		}

		if !p.match(token.IDENTIFIER) {
			return p.reportFatalErrorExpr(loxerrors.ErrParseExpectedSuperClassMethodName)
		}
		method := p.previous()

		return &ExprSuper{Keyword: tok, Method: method}
	}

	if p.match(token.THIS) {
		tok := p.previous()
		return &ExprThis{Keyword: tok}
	}

	if p.match(token.IDENTIFIER) {
		tok := p.previous()
		return &ExprVariable{Name: tok}
	}

	return p.grouping()
}

func (p *parser) grouping() Expr {
	if p.match(token.LEFT_PAREN) {
		expr := p.expression()
		if !p.match(token.RIGHT_PAREN) {
			return p.reportFatalErrorExpr(loxerrors.ErrParseExpectedRightParenToken)
		}
		return &ExprGrouping{Expression: expr}
	}

	return p.reportFatalErrorExpr(loxerrors.ErrParseUnexpectedToken)
}

func (p *parser) anyMatch(types ...token.TokenType) bool {
	for _, t := range types {
		if p.check(t) {
			p.advance()
			return true
		}
	}
	return false
}

func (p *parser) match(tokType token.TokenType) bool {
	if p.check(tokType) {
		p.advance()
		return true
	}
	return false
}

func (p *parser) check(tokenType token.TokenType) bool {
	return !p.isDone() && p.peek().Type == tokenType
}

func (p *parser) checkNext(tokenType token.TokenType) bool {
	if p.isDone() {
		return false
	}
	if p.peek().Type == token.EOF {
		return false
	}
	return p.tokens[p.current+1].Type == tokenType
}

func (p *parser) peek() *token.Token {
	return &p.tokens[p.current]
}

func (p *parser) previous() *token.Token {
	return &p.tokens[p.current-1]
}

func (p *parser) advance() *token.Token {
	if !p.isAtEnd() {
		p.current++
	}
	return p.previous()
}

// Be carefull with isAtEnd, it does not check for parse errors.
// Use isDone instead.
// isAtEnd is used from top level Parse, synchronize and advance ony.
func (p *parser) isAtEnd() bool {
	return p.peek().Type == token.EOF
}

func (p *parser) isDone() bool {
	// is at the end, OR, have errors
	return p.isAtEnd() || p.panic != nil
}

func (p *parser) isSkipRecoverError(tok *token.Token) bool {
	return p.recover && tok.Type == token.EOF
}

func (p *parser) reportFatalErrorStmt(err error) Stmt {
	return p.reportFatalErrorStmtToken(p.peek(), err)
}

func (p *parser) reportFatalErrorStmtToken(tok *token.Token, err error) Stmt {
	// do not overwrite present error.
	// preserves the original error and bubbles up to return in Parse() with .err
	if p.panic == nil && !p.isSkipRecoverError(tok) {
		p.panic = loxerrors.NewParseError(tok, err)
		p.fatal(p.panic)
	}
	return nilStmt
}

func (p *parser) reportFatalErrorStmtList(err error) []Stmt {
	// do not overwrite present error.
	// preserves the original error and bubbles up to return in Parse() with .err
	if p.panic == nil && !p.isSkipRecoverError(p.peek()) {
		p.panic = loxerrors.NewParseError(p.peek(), err)
		p.fatal(p.panic)
	}
	return nilStatements
}

func (p *parser) reportFatalErrorExpr(err error) Expr {
	return p.reportFatalErrorExprToken(p.peek(), err)
}

func (p *parser) reportFatalErrorExprToken(tok *token.Token, err error) Expr {
	// do not overwrite present error.
	// preserves the original error and bubbles up to return in Parse() with .err
	if p.panic == nil && !p.isSkipRecoverError(tok) {
		p.panic = loxerrors.NewParseError(tok, err)
		p.fatal(p.panic)
	}
	return nilExpr
}

func (p *parser) reportErrorExpr(err error) {
	p.reportErrorExprToken(p.peek(), err)
}

func (p *parser) reportErrorExprToken(tok *token.Token, err error) {
	if !p.isSkipRecoverError(tok) {
		p.error(loxerrors.NewParseError(tok, err))
	}
}

func (p *parser) fatal(err error) {
	p.reporter.ReportPanic(err)
}

func (p *parser) error(err error) {
	p.err = err
	p.reporter.ReportError(err)
}

func (p *parser) synchronize() {
	p.advance()

	for !p.isAtEnd() {
		if p.previous().Type == token.SEMICOLON {
			return
		}

		switch p.peek().Type {
		case token.CLASS,
			token.FUN,
			token.VAR,
			token.FOR,
			token.IF,
			token.WHILE,
			token.PRINT,
			token.RETURN:
			return
		}

		p.advance()
	}
}

var _ Parser = (*parser)(nil)
var _ fmt.Stringer = (*parser)(nil)
var _ fmt.GoStringer = (*parser)(nil)
