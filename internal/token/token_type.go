package token

import "fmt"

type TokenType byte

const (
	// EOF
	EOF TokenType = iota

	// Single-character tokens.
	LEFT_PAREN
	RIGHT_PAREN
	LEFT_BRACE
	RIGHT_BRACE
	COMMA
	DOT
	MINUS
	PLUS
	SEMICOLON
	SLASH
	STAR

	// One or two character tokens.
	BANG
	BANG_EQUAL
	EQUAL
	EQUAL_EQUAL
	GREATER
	GREATER_EQUAL
	LESS
	LESS_EQUAL

	// Literals.
	IDENTIFIER
	STRING
	NUMBER

	// Keywords.
	AND
	BREAK
	CONTINUE
	CLASS
	ELSE
	FALSE
	FUN
	FOR
	IF
	NIL
	OR
	PRINT
	RETURN
	SUPER
	THIS
	TRUE
	VAR
	WHILE
)

var tokenTypeStrings = map[TokenType]string{
	EOF: "EOF",

	// Single-character tokens.
	LEFT_PAREN:  "LEFT_PAREN",
	RIGHT_PAREN: "RIGHT_PAREN",
	LEFT_BRACE:  "LEFT_BRACE",
	RIGHT_BRACE: "RIGHT_BRACE",
	COMMA:       "COMMA",
	DOT:         "DOT",
	MINUS:       "MINUS",
	PLUS:        "PLUS",
	SEMICOLON:   "SEMICOLON",
	SLASH:       "SLASH",
	STAR:        "STAR",

	// One or two character tokens.
	BANG:          "BANG",
	BANG_EQUAL:    "BANG_EQUAL",
	EQUAL:         "EQUAL",
	EQUAL_EQUAL:   "EQUAL_EQUAL",
	GREATER:       "GREATER",
	GREATER_EQUAL: "GREATER_EQUAL",
	LESS:          "LESS",
	LESS_EQUAL:    "LESS_EQUAL",

	// Literals.
	IDENTIFIER: "IDENTIFIER",
	STRING:     "STRING",
	NUMBER:     "NUMBER",

	// Keywords.
	AND:      "AND",
	BREAK:    "BREAK",
	CONTINUE: "CONTINUE",
	CLASS:    "CLASS",
	ELSE:     "ELSE",
	FALSE:    "FALSE",
	FUN:      "FUN",
	FOR:      "FOR",
	IF:       "IF",
	NIL:      "NIL",
	OR:       "OR",
	PRINT:    "PRINT",
	RETURN:   "RETURN",
	SUPER:    "SUPER",
	THIS:     "THIS",
	TRUE:     "TRUE",
	VAR:      "VAR",
	WHILE:    "WHILE",
}

func (t TokenType) String() string {
	if v, ok := tokenTypeStrings[t]; ok {
		return v
	}
	return fmt.Sprintf("<Unrecognized: %d>", t)
}

var _ fmt.Stringer = EOF
