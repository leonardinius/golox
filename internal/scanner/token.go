package scanner

import (
	"fmt"

	"github.com/leonardinius/golox/internal/grammar"
)

// Token represents a lexical token.
type Token struct {
	Type    grammar.TokenType
	Lexeme  string
	Literal interface{}
	Line    int
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
