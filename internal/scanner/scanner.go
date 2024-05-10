package scanner

import (
	"strconv"

	"github.com/leonardinius/golox/internal/loxerrors"
	"github.com/leonardinius/golox/internal/token"
)

// Token represents a lexical to
type Scanner interface {
	Scan() ([]token.Token, error)
}

var reservedKeywords = map[string]token.TokenType{
	"and":      token.AND,
	"break":    token.BREAK,
	"continue": token.CONTINUE,
	"class":    token.CLASS,
	"else":     token.ELSE,
	"false":    token.FALSE,
	"for":      token.FOR,
	"fun":      token.FUN,
	"if":       token.IF,
	"nil":      token.NIL,
	"or":       token.OR,
	"print":    token.PRINT,
	"return":   token.RETURN,
	"super":    token.SUPER,
	"this":     token.THIS,
	"true":     token.TRUE,
	"var":      token.VAR,
	"while":    token.WHILE,
}

type scanner struct {
	source               []rune
	tokens               []token.Token
	start, current, line int
	err                  error
}

// NewScanner returns a new Scanner.
func NewScanner(input string) Scanner {
	return &scanner{source: []rune(input), start: 0, current: 0, line: 1}
}

// Scan implements Scanner.
func (s *scanner) Scan() ([]token.Token, error) {
	// return tokens;”

	for !s.isDone() {
		// We are at the beginning of the next lexeme.
		s.start = s.current
		s.scanToken()
	}

	s.tokens = append(s.tokens, token.NewToken(token.EOF, "", nil, s.line))

	return s.tokens, s.err
}

func (s *scanner) isAtEnd() bool {
	return s.current >= len(s.source)
}

func (s *scanner) hasErr() bool {
	return s.err != nil
}

func (s *scanner) isDone() bool {
	return s.isAtEnd() || s.hasErr()
}

func (s *scanner) scanToken() {
	var c = s.advance()

	switch c {
	case '(':
		s.addToken(token.LEFT_PAREN)
	case ')':
		s.addToken(token.RIGHT_PAREN)
	case '{':
		s.addToken(token.LEFT_BRACE)
	case '}':
		s.addToken(token.RIGHT_BRACE)
	case ',':
		s.addToken(token.COMMA)
	case '.':
		s.addToken(token.DOT)
	case '-':
		s.addToken(token.MINUS)
	case '+':
		s.addToken(token.PLUS)
	case ';':
		s.addToken(token.SEMICOLON)
	case '*':
		s.addToken(token.STAR)
	case '!':
		s.addMatchToken('=', token.BANG_EQUAL, token.BANG)
	case '=':
		s.addMatchToken('=', token.EQUAL_EQUAL, token.EQUAL)
	case '<':
		s.addMatchToken('=', token.LESS_EQUAL, token.LESS)
	case '>':
		s.addMatchToken('=', token.GREATER_EQUAL, token.GREATER)
	case '/':
		if s.match('/') {
			s.comment()
		} else if s.match('*') {
			s.blockComment()
		} else {
			s.addToken(token.SLASH)
		}
	case ' ', '\r', '\t', '\n':
		// Ignore whitespace.
	case '"':
		s.string()
	default:
		if s.isDigit(c) {
			s.number()
		} else if s.isAlpha(c) {
			s.reservedOrIdentifier()
		} else {
			s.reportUnexpectedCharater(c)
		}
	}
}

func (s *scanner) peek() rune {
	if s.isAtEnd() {
		return '\000'
	}
	return s.source[s.current]
}

func (s *scanner) peekNext() rune {
	if s.current+1 >= len(s.source) {
		return '\000'
	}
	return s.source[s.current+1]
}

func (s *scanner) advance() rune {
	if s.source[s.current] == '\n' {
		s.line++
	}
	s.current++
	return s.source[s.current-1]
}

func (s *scanner) match(expected rune) bool {
	if expected == s.peek() {
		s.advance()
		return true
	}

	return false
}

func (s *scanner) addMatchToken(lookAhead rune, ifMatch, ifNotMatched token.TokenType) {
	if s.match(lookAhead) {
		s.addToken(ifMatch)
	} else {
		s.addToken(ifNotMatched)
	}
}

func (s *scanner) addToken(t token.TokenType) {
	s.addTokenLiteral(t, nil)
}

func (s *scanner) addTokenLiteral(t token.TokenType, literal any) {
	s.tokens = append(s.tokens, token.NewToken(t, string(s.source[s.start:s.current]), literal, s.line))
}

func (s *scanner) comment() {
	for s.peek() != '\n' && !s.isAtEnd() {
		s.advance()
	}
}

func (s *scanner) blockComment() {
	depth := 1

	for !s.isAtEnd() && depth > 0 {

		if s.peek() == '*' && s.peekNext() == '/' {
			depth--
			s.advance()
			s.advance()
		} else if s.peek() == '/' && s.peekNext() == '*' {
			depth++
			s.advance()
			s.advance()
		} else {
			s.advance()
		}
	}

	if depth > 0 {
		s.reportError(loxerrors.ErrScanUnterminatedComment)
	}
}

func (s *scanner) string() {
	for !s.isAtEnd() && s.peek() != '"' {
		s.advance()
	}

	if s.isAtEnd() {
		s.reportError(loxerrors.ErrScanUnterminatedString)
		return
	}

	// The closing ".
	s.advance()

	value := s.source[s.start+1 : s.current-1]
	s.addTokenLiteral(token.STRING, string(value))
}

func (s *scanner) number() {
	for s.isDigit(s.peek()) {
		s.advance()
	}

	if s.peek() == '.' && s.isDigit(s.peekNext()) {
		s.advance()

		for s.isDigit(s.peek()) {
			s.advance()
		}
	}

	svalue := string(s.source[s.start:s.current])
	value, err := strconv.ParseFloat(svalue, 64)
	if err != nil {
		s.reportError(err)
		return
	}
	s.addTokenLiteral(token.NUMBER, float64(value))
}

func (s *scanner) reservedOrIdentifier() {
	for s.isAlphaNumeric(s.peek()) {
		s.advance()
	}

	tokenType := token.IDENTIFIER
	name := string(s.source[s.start:s.current])
	if _type, ok := s.reserved(name); ok {
		tokenType = _type
	}
	s.addToken(tokenType)
}

func (s *scanner) reserved(identifier string) (tokenType token.TokenType, ok bool) {
	tokenType, ok = reservedKeywords[identifier]
	return
}

func (s *scanner) isDigit(c rune) bool {
	return c >= '0' && c <= '9'
}

func (s *scanner) isAlpha(c rune) bool {
	return (c >= 'a' && c <= 'z') ||
		(c >= 'A' && c <= 'Z') ||
		c == '_'
}

func (s *scanner) isAlphaNumeric(c rune) bool {
	return s.isAlpha(c) || s.isDigit(c)
}

func (s *scanner) reportUnexpectedCharater(c rune) {
	s.err = loxerrors.NewScanError(s.line, loxerrors.ErrScanUnexpectedCharacter, strconv.QuoteRune(c))
}

func (s *scanner) reportError(err error) {
	s.err = loxerrors.NewScanError(s.line, err, "")
}

var _ Scanner = (*scanner)(nil)
