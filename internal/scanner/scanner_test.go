package scanner_test

import (
	"testing"

	"github.com/leonardinius/golox/internal/scanner"
	"github.com/stretchr/testify/assert"
)

func TestScanTokens(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name     string
		input    string
		expected []string
		err      string
	}{
		{"empty", "", []string{`{Type: EOF, Lexeme: "", Literal: <nil>, Line: 1}`}, ""},
		{"syntax error", "⌘", nil, "[line 1] Error: Unexpected character. '⌘'"},
		{
			"basic",
			"(){},*+-;",
			[]string{
				`{Type: LEFT_PAREN, Lexeme: "(", Literal: <nil>, Line: 1}`,
				`{Type: RIGHT_PAREN, Lexeme: ")", Literal: <nil>, Line: 1}`,
				`{Type: LEFT_BRACE, Lexeme: "{", Literal: <nil>, Line: 1}`,
				`{Type: RIGHT_BRACE, Lexeme: "}", Literal: <nil>, Line: 1}`,
				`{Type: COMMA, Lexeme: ",", Literal: <nil>, Line: 1}`,
				`{Type: STAR, Lexeme: "*", Literal: <nil>, Line: 1}`,
				`{Type: PLUS, Lexeme: "+", Literal: <nil>, Line: 1}`,
				`{Type: MINUS, Lexeme: "-", Literal: <nil>, Line: 1}`,
				`{Type: SEMICOLON, Lexeme: ";", Literal: <nil>, Line: 1}`,
				`{Type: EOF, Lexeme: "", Literal: <nil>, Line: 1}`,
			},
			"",
		},
		{
			"bang",
			"!",
			[]string{
				`{Type: BANG, Lexeme: "!", Literal: <nil>, Line: 1}`,
				`{Type: EOF, Lexeme: "", Literal: <nil>, Line: 1}`,
			},
			"",
		},
		{
			"bangbang",
			"!!",
			[]string{
				`{Type: BANG, Lexeme: "!", Literal: <nil>, Line: 1}`,
				`{Type: BANG, Lexeme: "!", Literal: <nil>, Line: 1}`,
				`{Type: EOF, Lexeme: "", Literal: <nil>, Line: 1}`,
			},
			"",
		},
		{
			"bangbangeqeqeqeq",
			"!====",
			[]string{
				`{Type: BANG_EQUAL, Lexeme: "!=", Literal: <nil>, Line: 1}`,
				`{Type: EQUAL_EQUAL, Lexeme: "==", Literal: <nil>, Line: 1}`,
				`{Type: EQUAL, Lexeme: "=", Literal: <nil>, Line: 1}`,
				`{Type: EOF, Lexeme: "", Literal: <nil>, Line: 1}`,
			},
			"",
		},
		{
			"lt",
			"<",
			[]string{
				`{Type: LESS, Lexeme: "<", Literal: <nil>, Line: 1}`,
				`{Type: EOF, Lexeme: "", Literal: <nil>, Line: 1}`,
			},
			"",
		},
		{
			"lteq",
			"<=",
			[]string{
				`{Type: LESS_EQUAL, Lexeme: "<=", Literal: <nil>, Line: 1}`,
				`{Type: EOF, Lexeme: "", Literal: <nil>, Line: 1}`,
			},
			"",
		},
		{
			"lteqeqeqeq",
			"<====",
			[]string{
				`{Type: LESS_EQUAL, Lexeme: "<=", Literal: <nil>, Line: 1}`,
				`{Type: EQUAL_EQUAL, Lexeme: "==", Literal: <nil>, Line: 1}`,
				`{Type: EQUAL, Lexeme: "=", Literal: <nil>, Line: 1}`,
				`{Type: EOF, Lexeme: "", Literal: <nil>, Line: 1}`,
			},
			"",
		},
		{
			"gt",
			">",
			[]string{
				`{Type: GREATER, Lexeme: ">", Literal: <nil>, Line: 1}`,
				`{Type: EOF, Lexeme: "", Literal: <nil>, Line: 1}`,
			},
			"",
		},
		{
			"gteq",
			">=",
			[]string{
				`{Type: GREATER_EQUAL, Lexeme: ">=", Literal: <nil>, Line: 1}`,
				`{Type: EOF, Lexeme: "", Literal: <nil>, Line: 1}`,
			},
			"",
		},
		{
			"gteqeqeqeq",
			">====",
			[]string{
				`{Type: GREATER_EQUAL, Lexeme: ">=", Literal: <nil>, Line: 1}`,
				`{Type: EQUAL_EQUAL, Lexeme: "==", Literal: <nil>, Line: 1}`,
				`{Type: EQUAL, Lexeme: "=", Literal: <nil>, Line: 1}`,
				`{Type: EOF, Lexeme: "", Literal: <nil>, Line: 1}`,
			},
			"",
		},
		{
			"comment",
			"//comment",
			[]string{
				`{Type: EOF, Lexeme: "", Literal: <nil>, Line: 1}`,
			},
			"",
		},
		{
			"bangcomment",
			"!//comment",
			[]string{
				`{Type: BANG, Lexeme: "!", Literal: <nil>, Line: 1}`,
				`{Type: EOF, Lexeme: "", Literal: <nil>, Line: 1}`,
			},
			"",
		},
		{
			"spaces",
			"! \r\t=",
			[]string{
				`{Type: BANG, Lexeme: "!", Literal: <nil>, Line: 1}`,
				`{Type: EQUAL, Lexeme: "=", Literal: <nil>, Line: 1}`,
				`{Type: EOF, Lexeme: "", Literal: <nil>, Line: 1}`,
			},
			"",
		},
		{
			"string",
			`"string"`,
			[]string{
				`{Type: STRING, Lexeme: "\"string\"", Literal: "string", Line: 1}`,
				`{Type: EOF, Lexeme: "", Literal: <nil>, Line: 1}`,
			},
			"",
		},
		{
			"empty-string",
			`""`,
			[]string{
				`{Type: STRING, Lexeme: "\"\"", Literal: "", Line: 1}`,
				`{Type: EOF, Lexeme: "", Literal: <nil>, Line: 1}`,
			},
			"",
		},
		{
			"string-nl",
			`"string\nstring"`,
			[]string{
				`{Type: STRING, Lexeme: "\"string\\nstring\"", Literal: "string\\nstring", Line: 1}`,
				`{Type: EOF, Lexeme: "", Literal: <nil>, Line: 1}`,
			},
			"",
		},
		{
			"number-integer",
			`10`,
			[]string{
				`{Type: NUMBER, Lexeme: "10", Literal: 10, Line: 1}`,
				`{Type: EOF, Lexeme: "", Literal: <nil>, Line: 1}`,
			},
			"",
		},
		{
			"number-integer-leading-zeroes",
			`0010`,
			[]string{
				`{Type: NUMBER, Lexeme: "0010", Literal: 10, Line: 1}`,
				`{Type: EOF, Lexeme: "", Literal: <nil>, Line: 1}`,
			},
			"",
		},
		{
			"number-decimal",
			`12.34`,
			[]string{
				`{Type: NUMBER, Lexeme: "12.34", Literal: 12.34, Line: 1}`,
				`{Type: EOF, Lexeme: "", Literal: <nil>, Line: 1}`,
			},
			"",
		},
		{
			"number-decimal-leading-zeroes",
			`0012.34`,
			[]string{
				`{Type: NUMBER, Lexeme: "0012.34", Literal: 12.34, Line: 1}`,
				`{Type: EOF, Lexeme: "", Literal: <nil>, Line: 1}`,
			},
			"",
		},
		{
			"number-dot",
			`12.`,
			[]string{
				`{Type: NUMBER, Lexeme: "12", Literal: 12, Line: 1}`,
				`{Type: DOT, Lexeme: ".", Literal: <nil>, Line: 1}`,
				`{Type: EOF, Lexeme: "", Literal: <nil>, Line: 1}`,
			},
			"",
		},
		{
			"identifier",
			`identifier`,
			[]string{
				`{Type: IDENTIFIER, Lexeme: "identifier", Literal: <nil>, Line: 1}`,
				`{Type: EOF, Lexeme: "", Literal: <nil>, Line: 1}`,
			},
			"",
		},
		{
			"reserved",
			`and class else false for fun if nil or print return super this true var while`,
			[]string{
				`{Type: AND, Lexeme: "and", Literal: <nil>, Line: 1}`,
				`{Type: CLASS, Lexeme: "class", Literal: <nil>, Line: 1}`,
				`{Type: ELSE, Lexeme: "else", Literal: <nil>, Line: 1}`,
				`{Type: FALSE, Lexeme: "false", Literal: <nil>, Line: 1}`,
				`{Type: FOR, Lexeme: "for", Literal: <nil>, Line: 1}`,
				`{Type: FUN, Lexeme: "fun", Literal: <nil>, Line: 1}`,
				`{Type: IF, Lexeme: "if", Literal: <nil>, Line: 1}`,
				`{Type: NIL, Lexeme: "nil", Literal: <nil>, Line: 1}`,
				`{Type: OR, Lexeme: "or", Literal: <nil>, Line: 1}`,
				`{Type: PRINT, Lexeme: "print", Literal: <nil>, Line: 1}`,
				`{Type: RETURN, Lexeme: "return", Literal: <nil>, Line: 1}`,
				`{Type: SUPER, Lexeme: "super", Literal: <nil>, Line: 1}`,
				`{Type: THIS, Lexeme: "this", Literal: <nil>, Line: 1}`,
				`{Type: TRUE, Lexeme: "true", Literal: <nil>, Line: 1}`,
				`{Type: VAR, Lexeme: "var", Literal: <nil>, Line: 1}`,
				`{Type: WHILE, Lexeme: "while", Literal: <nil>, Line: 1}`,
				`{Type: EOF, Lexeme: "", Literal: <nil>, Line: 1}`,
			},
			"",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(tt *testing.T) {
			s := scanner.NewScanner(tc.input)
			tokens, err := s.Scan()
			if tc.err != "" {
				assert.ErrorContainsf(tt, err, tc.err, "expected error %v, got %v", tc.err, err)
			} else {
				tokensAsStrings := make([]string, len(tokens))
				for i, token := range tokens {
					tokensAsStrings[i] = token.GoString()
				}
				assert.Equal(tt, tc.expected, tokensAsStrings)
			}
		})
	}
}
