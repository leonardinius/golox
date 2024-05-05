package scanner

import (
	"strconv"

	"github.com/leonardinius/golox/internal/grammar"
)

// Token represents a lexical to
type Scanner interface {
	Scan() ([]Token, error)
}

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
		} else {
			s.addToken(grammar.SLASH)
		}
	case ' ', '\r', '\t':
		// Ignore whitespace.
	case '\n':
		s.line++
	default:
		s.unrecognizedCharacter(c)
	}
}

func (s *scanner) peek() rune {
	if s.isAtEnd() {
		return '\000'
	}
	return s.source[s.current]
}

func (s *scanner) advance() rune {
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

func (s *scanner) addTokenLiteral(t grammar.TokenType, literal interface{}) {
	s.tokens = append(s.tokens, Token{t, string(s.source[s.start:s.current]), literal, s.line})
}

func (s *scanner) comment() {
	for s.peek() != '\n' && !s.isAtEnd() {
		s.advance()
	}
}

func (s *scanner) unrecognizedCharacter(c rune) {
	s.err = NewScanError(s.line, "", "Unrecognized character.", strconv.QuoteRune(c))
}

var _ Scanner = (*scanner)(nil)
