package scanner_test

import (
	"fmt"
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
		{"empty", "", []string{`{Type: EOF, Literal: <nil>, Line: 1}`}, ""},
		{"syntax error", "⌘", nil, "[line 1] syntax error: Unexpected character. '⌘'"},
		{
			"basic",
			"(){},*+-;",
			[]string{
				`{Type: LEFT_PAREN, Literal: <nil>, Line: 1}`,
				`{Type: RIGHT_PAREN, Literal: <nil>, Line: 1}`,
				`{Type: LEFT_BRACE, Literal: <nil>, Line: 1}`,
				`{Type: RIGHT_BRACE, Literal: <nil>, Line: 1}`,
				`{Type: COMMA, Literal: <nil>, Line: 1}`,
				`{Type: STAR, Literal: <nil>, Line: 1}`,
				`{Type: PLUS, Literal: <nil>, Line: 1}`,
				`{Type: MINUS, Literal: <nil>, Line: 1}`,
				`{Type: SEMICOLON, Literal: <nil>, Line: 1}`,
				`{Type: EOF, Literal: <nil>, Line: 1}`,
			},
			"",
		},
		{
			"bang",
			"!",
			[]string{
				`{Type: BANG, Literal: <nil>, Line: 1}`,
				`{Type: EOF, Literal: <nil>, Line: 1}`,
			},
			"",
		},
		{
			"bangbang",
			"!!",
			[]string{
				`{Type: BANG, Literal: <nil>, Line: 1}`,
				`{Type: BANG, Literal: <nil>, Line: 1}`,
				`{Type: EOF, Literal: <nil>, Line: 1}`,
			},
			"",
		},
		{
			"bangbangeqeqeqeq",
			"!====",
			[]string{
				`{Type: BANG_EQUAL, Literal: <nil>, Line: 1}`,
				`{Type: EQUAL_EQUAL, Literal: <nil>, Line: 1}`,
				`{Type: EQUAL, Literal: <nil>, Line: 1}`,
				`{Type: EOF, Literal: <nil>, Line: 1}`,
			},
			"",
		},
		{
			"lt",
			"<",
			[]string{
				`{Type: LESS, Literal: <nil>, Line: 1}`,
				`{Type: EOF, Literal: <nil>, Line: 1}`,
			},
			"",
		},
		{
			"lteq",
			"<=",
			[]string{
				`{Type: LESS_EQUAL, Literal: <nil>, Line: 1}`,
				`{Type: EOF, Literal: <nil>, Line: 1}`,
			},
			"",
		},
		{
			"lteqeqeqeq",
			"<====",
			[]string{
				`{Type: LESS_EQUAL, Literal: <nil>, Line: 1}`,
				`{Type: EQUAL_EQUAL, Literal: <nil>, Line: 1}`,
				`{Type: EQUAL, Literal: <nil>, Line: 1}`,
				`{Type: EOF, Literal: <nil>, Line: 1}`,
			},
			"",
		},
		{
			"gt",
			">",
			[]string{
				`{Type: GREATER, Literal: <nil>, Line: 1}`,
				`{Type: EOF, Literal: <nil>, Line: 1}`,
			},
			"",
		},
		{
			"gteq",
			">=",
			[]string{
				`{Type: GREATER_EQUAL, Literal: <nil>, Line: 1}`,
				`{Type: EOF, Literal: <nil>, Line: 1}`,
			},
			"",
		},
		{
			"gteqeqeqeq",
			">====",
			[]string{
				`{Type: GREATER_EQUAL, Literal: <nil>, Line: 1}`,
				`{Type: EQUAL_EQUAL, Literal: <nil>, Line: 1}`,
				`{Type: EQUAL, Literal: <nil>, Line: 1}`,
				`{Type: EOF, Literal: <nil>, Line: 1}`,
			},
			"",
		},
		{
			"comment",
			"//comment",
			[]string{
				`{Type: EOF, Literal: <nil>, Line: 1}`,
			},
			"",
		},
		{
			"bangcomment",
			"!//comment",
			[]string{
				`{Type: BANG, Literal: <nil>, Line: 1}`,
				`{Type: EOF, Literal: <nil>, Line: 1}`,
			},
			"",
		},
		{
			"spaces",
			"! \r\t=",
			[]string{
				`{Type: BANG, Literal: <nil>, Line: 1}`,
				`{Type: EQUAL, Literal: <nil>, Line: 1}`,
				`{Type: EOF, Literal: <nil>, Line: 1}`,
			},
			"",
		},
		{
			"string",
			`"string"`,
			[]string{
				`{Type: STRING, Literal: "string", Line: 1}`,
				`{Type: EOF, Literal: <nil>, Line: 1}`,
			},
			"",
		},
		{
			"empty-string",
			`""`,
			[]string{
				`{Type: STRING, Literal: "", Line: 1}`,
				`{Type: EOF, Literal: <nil>, Line: 1}`,
			},
			"",
		},
		{
			"string-nl",
			`"string\nstring"`,
			[]string{
				`{Type: STRING, Literal: "string\\nstring", Line: 1}`,
				`{Type: EOF, Literal: <nil>, Line: 1}`,
			},
			"",
		},
		{
			"number-integer",
			`10`,
			[]string{
				`{Type: NUMBER, Literal: 10, Line: 1}`,
				`{Type: EOF, Literal: <nil>, Line: 1}`,
			},
			"",
		},
		{
			"number-integer-leading-zeroes",
			`0010`,
			[]string{
				`{Type: NUMBER, Literal: 10, Line: 1}`,
				`{Type: EOF, Literal: <nil>, Line: 1}`,
			},
			"",
		},
		{
			"number-decimal",
			`12.34`,
			[]string{
				`{Type: NUMBER, Literal: 12.34, Line: 1}`,
				`{Type: EOF, Literal: <nil>, Line: 1}`,
			},
			"",
		},
		{
			"number-decimal-leading-zeroes",
			`0012.34`,
			[]string{
				`{Type: NUMBER, Literal: 12.34, Line: 1}`,
				`{Type: EOF, Literal: <nil>, Line: 1}`,
			},
			"",
		},
		{
			"number-dot",
			`12.`,
			[]string{
				`{Type: NUMBER, Literal: 12, Line: 1}`,
				`{Type: DOT, Literal: <nil>, Line: 1}`,
				`{Type: EOF, Literal: <nil>, Line: 1}`,
			},
			"",
		},
		{
			"identifier",
			`identifier`,
			[]string{
				`{Type: IDENTIFIER, Literal: <nil>, Line: 1}`,
				`{Type: EOF, Literal: <nil>, Line: 1}`,
			},
			"",
		},
		{
			"reserved",
			`and class else false for fun if nil or print return super this true var while`,
			[]string{
				`{Type: AND, Literal: <nil>, Line: 1}`,
				`{Type: CLASS, Literal: <nil>, Line: 1}`,
				`{Type: ELSE, Literal: <nil>, Line: 1}`,
				`{Type: FALSE, Literal: <nil>, Line: 1}`,
				`{Type: FOR, Literal: <nil>, Line: 1}`,
				`{Type: FUN, Literal: <nil>, Line: 1}`,
				`{Type: IF, Literal: <nil>, Line: 1}`,
				`{Type: NIL, Literal: <nil>, Line: 1}`,
				`{Type: OR, Literal: <nil>, Line: 1}`,
				`{Type: PRINT, Literal: <nil>, Line: 1}`,
				`{Type: RETURN, Literal: <nil>, Line: 1}`,
				`{Type: SUPER, Literal: <nil>, Line: 1}`,
				`{Type: THIS, Literal: <nil>, Line: 1}`,
				`{Type: TRUE, Literal: <nil>, Line: 1}`,
				`{Type: VAR, Literal: <nil>, Line: 1}`,
				`{Type: WHILE, Literal: <nil>, Line: 1}`,
				`{Type: EOF, Literal: <nil>, Line: 1}`,
			},
			"",
		},
		{
			"comment-asterix",
			"/**/",
			[]string{
				`{Type: EOF, Literal: <nil>, Line: 1}`,
			},
			"",
		},
		{
			"comment-comment-asterix-bang-bang",
			`/*
			//
			/**/
			/*
			*/
			/*
			/**/
			/*
			*/
			*/
			*/!
			!`,
			[]string{
				`{Type: BANG, Literal: <nil>, Line: 11}`,
				`{Type: BANG, Literal: <nil>, Line: 12}`,
				`{Type: EOF, Literal: <nil>, Line: 12}`,
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
					tokensAsStrings[i] = fmt.Sprintf(`{Type: %s, Literal: %#v, Line: %d}`, token.Type, token.Literal, token.Line)
				}
				assert.Equal(tt, tc.expected, tokensAsStrings)
			}
		})
	}
}
