package scanner

import (
	"errors"
	"strconv"

	"github.com/leonardinius/golox/internal/grammar"
)

// Token represents a lexical to
type Scanner interface {
	Scan() ([]Token, error)
}

var reservedKeywords = map[string]grammar.TokenType{
	"and":    grammar.AND,
	"class":  grammar.CLASS,
	"else":   grammar.ELSE,
	"false":  grammar.FALSE,
	"for":    grammar.FOR,
	"fun":    grammar.FUN,
	"if":     grammar.IF,
	"nil":    grammar.NIL,
	"or":     grammar.OR,
	"print":  grammar.PRINT,
	"return": grammar.RETURN,
	"super":  grammar.SUPER,
	"this":   grammar.THIS,
	"true":   grammar.TRUE,
	"var":    grammar.VAR,
	"while":  grammar.WHILE,
}

var (
	errUnexpectedCharacter = errors.New("Unexpected character.")
	errUnterminatedString  = errors.New("Unterminated string.")
	errUnterminatedComment = errors.New("Unterminated comment.")
)

type scanner struct {
	source               []rune
	tokens               []Token
	start, current, line int
	err                  error
}

// NewScanner returns a new Scanner.
func NewScanner(input string) Scanner {
	return &scanner{source: []rune(input), start: 0, current: 0, line: 1}
}

// Scan implements Scanner.
func (s *scanner) Scan() ([]Token, error) {
	// return tokens;”

	for !s.isDone() {
		// We are at the beginning of the next lexeme.
		s.start = s.current
		s.scanToken()
	}

	s.tokens = append(s.tokens, Token{grammar.EOF, "", nil, s.line})

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
		s.addToken(grammar.LEFT_PAREN)
	case ')':
		s.addToken(grammar.RIGHT_PAREN)
	case '{':
		s.addToken(grammar.LEFT_BRACE)
	case '}':
		s.addToken(grammar.RIGHT_BRACE)
	case ',':
		s.addToken(grammar.COMMA)
	case '.':
		s.addToken(grammar.DOT)
	case '-':
		s.addToken(grammar.MINUS)
	case '+':
		s.addToken(grammar.PLUS)
	case ';':
		s.addToken(grammar.SEMICOLON)
	case '*':
		s.addToken(grammar.STAR)
	case '!':
		s.addMatchToken('=', grammar.BANG_EQUAL, grammar.BANG)
	case '=':
		s.addMatchToken('=', grammar.EQUAL_EQUAL, grammar.EQUAL)
	case '<':
		s.addMatchToken('=', grammar.LESS_EQUAL, grammar.LESS)
	case '>':
		s.addMatchToken('=', grammar.GREATER_EQUAL, grammar.GREATER)
	case '/':
		if s.match('/') {
			s.comment()
		} else if s.match('*') {
			s.blockComment()
		} else {
			s.addToken(grammar.SLASH)
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

func (s *scanner) addMatchToken(lookAhead rune, ifMatch, ifNotMatched grammar.TokenType) {
	if s.match(lookAhead) {
		s.addToken(ifMatch)
	} else {
		s.addToken(ifNotMatched)
	}
}

func (s *scanner) addToken(t grammar.TokenType) {
	s.addTokenLiteral(t, nil)
}

func (s *scanner) addTokenLiteral(t grammar.TokenType, literal any) {
	s.tokens = append(s.tokens, Token{t, string(s.source[s.start:s.current]), literal, s.line})
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
		s.reportError(errUnterminatedComment)
	}
}

func (s *scanner) string() {
	for !s.isAtEnd() && s.peek() != '"' {
		s.advance()
	}

	if s.isAtEnd() {
		s.reportError(errUnterminatedString)
		return
	}

	// The closing ".
	s.advance()

	value := s.source[s.start+1 : s.current-1]
	s.addTokenLiteral(grammar.STRING, string(value))
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
	s.addTokenLiteral(grammar.NUMBER, value)
}

func (s *scanner) reservedOrIdentifier() {
	for s.isAlphaNumeric(s.peek()) {
		s.advance()
	}

	tokenType := grammar.IDENTIFIER
	name := string(s.source[s.start:s.current])
	if _type, ok := s.reserved(name); ok {
		tokenType = _type
	}
	s.addToken(tokenType)
}

func (s *scanner) reserved(identifier string) (tokenType grammar.TokenType, ok bool) {
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
	s.err = errors.Join(NewScanError(s.line, "", errUnexpectedCharacter.Error(), strconv.QuoteRune(c)),
		errUnexpectedCharacter,
	)
}

func (s *scanner) reportError(err error) {
	s.err = errors.Join(NewScanError(s.line, "", err.Error(), ""),
		err,
	)
}

var _ Scanner = (*scanner)(nil)
