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
		{"syntax error", "⌘", nil, "[line 1] Error: Unrecognized character. '⌘'"},
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
				assert.Equalf(tt, tc.expected, tokensAsStrings, "expected tokens %v, got %v", tc.expected, tokens)
			}
		})
	}
}
