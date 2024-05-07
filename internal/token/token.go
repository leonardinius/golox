package token

import (
	"fmt"
)

type DoubleNumber float64

// Token represents a lexical token.
type Token struct {
	Type    TokenType
	Lexeme  string
	Literal any
	Line    int
}

func NewToken(t TokenType, lexeme string, literal any, line int) Token {
	return Token{
		Type:    t,
		Lexeme:  lexeme,
		Literal: literal,
		Line:    line,
	}
}

func NewTokenHeap(t TokenType, lexeme string, literal any, line int) *Token {
	tt := NewToken(t, lexeme, literal, line)
	return &tt
}

// String implements fmt.Stringer.
func (t Token) String() string {
	return fmt.Sprintf("%s %s %v", t.Type, t.Lexeme, t.Literal)
}

// GoString implements fmt.GoStringer.
func (t Token) GoString() string {
	return fmt.Sprintf("{Type: %s, Lexeme: %q, Literal: %#v, Line: %d}", t.Type, t.Lexeme, t.Literal, t.Line)
}

var _ fmt.Stringer = (*Token)(nil)
var _ fmt.GoStringer = (*Token)(nil)
