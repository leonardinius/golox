package parser

import (
	"errors"
	"fmt"

	"github.com/leonardinius/golox/internal/loxerrors"
	"github.com/leonardinius/golox/internal/token"
)

var (
	NilToken      *token.Token = nil
	NilExpr       Expr         = nil
	NilStmt       Stmt         = nil
	NilStatements []Stmt       = nil
)

type Parser interface {
	Parse() ([]Stmt, error)
}

type parser struct {
	tokens  []token.Token
	current int
	err     []error
}

func NewParser(tokens []token.Token) Parser {
	if len(tokens) == 0 {
		panic("tokens cannot be empty")
	}
	if tokens[len(tokens)-1].Type != token.EOF {
		panic("tokens must end with EOF")
	}

	return &parser{
		tokens:  tokens,
		current: 0,
	}
}

// GoString implements fmt.GoStringer.
func (p *parser) GoString() string {
	return fmt.Sprintf("parser{tokens: %#v, current: %d, err: %#v}", p.tokens, p.current, p.err)
}

// String implements fmt.Stringer.
func (p *parser) String() string {
	return fmt.Sprintf("parser{tokens: %d, err: %v}", len(p.tokens), p.err)
}

// Parse implements Parser.
func (p *parser) Parse() (statements []Stmt, err error) {
	var stmt Stmt
	for !p.isAtEnd() {
		stmt, err = p.declaration(), errors.Join(p.err...)
		if err != nil {
			break
		}
		statements = append(statements, stmt)
	}

	if err == nil {
		return statements, nil
	}

	// if we are at error state, we do not return invalid ast tree
	// return nil, err - errors intead
	for !p.isAtEnd() {
		p.synchronize()
		_, err = p.declaration(), errors.Join(p.err...)
	}

	return NilStatements, err
}

func (p *parser) declaration() Stmt {
	if p.anyMatch(token.VAR) {
		return p.varDeclaration()
	}

	return p.statement()
}

func (p *parser) varDeclaration() Stmt {

	name := p.consume(token.IDENTIFIER)
	if name == NilToken {
		return p.reportStmtError(loxerrors.ErrParseUnexpectedVariableName)
	}

	var initializer Expr = NilExpr
	if p.anyMatch(token.EQUAL) {
		initializer = p.expression()
	}

	if tok := p.consume(token.SEMICOLON); tok == NilToken {
		return p.reportStmtError(loxerrors.ErrParseExpectedSemicolonTokenAfterVar)
	}

	return &Var{Name: name, Initializer: initializer}
}

func (p *parser) statement() Stmt {
	if p.anyMatch(token.PRINT) {
		return p.printStatement()
	}

	return p.expressionStatement()
}

func (p *parser) printStatement() Stmt {
	expr := p.expression()
	if tok := p.consume(token.SEMICOLON); tok == NilToken {
		return p.reportStmtError(loxerrors.ErrParseExpectedSemicolonTokenAfterValue)
	}
	return &Print{Expression: expr}
}

func (p *parser) expressionStatement() Stmt {
	expr := p.expression()
	if tok := p.consume(token.SEMICOLON); tok == NilToken {
		return p.reportStmtError(loxerrors.ErrParseExpectedSemicolonTokenAfterExpr)
	}
	return &Expression{Expression: expr}
}

func (p *parser) expression() Expr {
	return p.equality()
}

func (p *parser) equality() Expr {
	expr := p.comparison()

	for p.anyMatch(token.BANG_EQUAL, token.EQUAL_EQUAL) {
		operator := p.previous()
		right := p.comparison()
		expr = &Binary{Left: expr, Operator: operator, Right: right}
	}

	return expr
}

func (p *parser) comparison() Expr {
	expr := p.term()

	for p.anyMatch(token.GREATER, token.GREATER_EQUAL, token.LESS, token.LESS_EQUAL) {
		operator := p.previous()
		right := p.term()
		expr = &Binary{Left: expr, Operator: operator, Right: right}
	}

	return expr
}

func (p *parser) term() Expr {
	expr := p.factor()

	for p.anyMatch(token.MINUS, token.PLUS) {
		operator := p.previous()
		right := p.factor()
		expr = &Binary{Left: expr, Operator: operator, Right: right}
	}

	return expr
}

func (p *parser) factor() Expr {
	expr := p.unary()

	for p.anyMatch(token.SLASH, token.STAR) {
		operator := p.previous()
		right := p.unary()
		expr = &Binary{Left: expr, Operator: operator, Right: right}
	}

	return expr
}

func (p *parser) unary() Expr {
	if p.anyMatch(token.BANG, token.MINUS) {
		operator := p.previous()
		right := p.unary()
		return &Unary{
			Operator: operator,
			Right:    right,
		}
	}

	return p.primary()
}

func (p *parser) primary() Expr {
	if p.anyMatch(token.FALSE) {
		return &Literal{Value: false}
	}
	if p.anyMatch(token.TRUE) {
		return &Literal{Value: true}
	}
	if p.anyMatch(token.NIL) {
		return &Literal{Value: nil}
	}

	if p.anyMatch(token.NUMBER, token.STRING) {
		tok := p.previous()
		return &Literal{Value: tok.Literal}
	}

	if p.anyMatch(token.IDENTIFIER) {
		tok := p.previous()
		return &Variable{Name: tok}
	}

	return p.grouping()
}

func (p *parser) grouping() Expr {
	if p.anyMatch(token.LEFT_PAREN) {
		expr := p.expression()
		if tok := p.consume(token.RIGHT_PAREN); tok == NilToken {
			return p.reportExprError(loxerrors.ErrParseExpectedRightParenToken)
		}
		return &Grouping{Expression: expr}
	}

	return p.reportExprError(loxerrors.ErrParseUnexpectedToken)
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

func (p *parser) consume(tokType token.TokenType) *token.Token {
	if p.check(tokType) {
		return p.advance()
	}
	return NilToken
}

func (p *parser) check(tokenType token.TokenType) bool {
	if p.isAtEnd() {
		return false
	}
	return p.peek().Type == tokenType
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

func (p *parser) isAtEnd() bool {
	return p.peek().Type == token.EOF
}

func (p *parser) reportStmtError(err error) Stmt {
	t := p.peek()
	p.err = append(p.err, loxerrors.NewParseError(t, err))

	return NilStmt
}

func (p *parser) reportExprError(err error) Expr {
	t := p.peek()
	p.err = append(p.err, loxerrors.NewParseError(t, err))

	return NilExpr
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
