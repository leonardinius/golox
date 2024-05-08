package interpreter_test

import (
	"testing"

	"github.com/leonardinius/golox/internal/interpreter"
	"github.com/leonardinius/golox/internal/parser"
	"github.com/leonardinius/golox/internal/scanner"
	"github.com/stretchr/testify/assert"
)

func TestInterpret(t *testing.T) {
	testcases := []struct {
		name           string
		input          string
		expectedOutput string
		expectedError  string
	}{
		{name: `simple expression`, input: `1 + 2`, expectedOutput: `3`},
		{name: `grouped`, input: `(1 + 2)`, expectedOutput: `3`},
		{name: `nested`, input: `(1 + (2 + 3))`, expectedOutput: `6`},
		{name: `precedence asterix`, input: `1 + 2 * 3`, expectedOutput: `7`},
		{name: `precedence slash`, input: `1 + 9 / 3`, expectedOutput: `4`},
		{name: `precedence asterix slash`, input: `1 + 2 * 6 / 4`, expectedOutput: `4`},
		{name: `grouping nested precedence`, input: `((1 + 2) * 3)/2`, expectedOutput: `4.5`},
		{name: `strings`, input: `"a" + "b"`, expectedOutput: `"ab"`},
		{name: `boolean t`, input: `true`, expectedOutput: `true`},
		{name: `boolean f`, input: `false`, expectedOutput: `false`},
		{name: `bang`, input: `!false`, expectedOutput: `true`},
		{name: `bang bang`, input: `!!false`, expectedOutput: `false`},
		{name: `eqeq number`, input: `1 == 1`, expectedOutput: `true`},
		{name: `eqeq number`, input: `1 == 2`, expectedOutput: `false`},
		{name: `eqeq string`, input: `"a" == "a"`, expectedOutput: `true`},
		{name: `eqeq string`, input: `"a" == "b"`, expectedOutput: `false`},
		{name: `bangeq number`, input: `1 != 1`, expectedOutput: `false`},
		{name: `bangeq number`, input: `1 != 2`, expectedOutput: `true`},
		{name: `bangeq string`, input: `"a" != "a"`, expectedOutput: `false`},
		{name: `bangeq string`, input: `"a" != "b"`, expectedOutput: `true`},
		{name: `lt number`, input: `1 < 2`, expectedOutput: `true`},
		{name: `lt number`, input: `1 < 1`, expectedOutput: `false`},
		{name: `lte number`, input: `2 <= 1`, expectedOutput: `false`},
		{name: `lte number`, input: `1 <= 1`, expectedOutput: `true`},
		{name: `gt number`, input: `2 > 1`, expectedOutput: `true`},
		{name: `gt number`, input: `1 > 1`, expectedOutput: `false`},
		{name: `gte number`, input: `1 >= 2`, expectedOutput: `false`},
		{name: `gte number`, input: `1 >= 1`, expectedOutput: `true`},
		{name: `invalid expression`, input: `1 + 2 +`, expectedError: `parse error at end: expected expression`},
		{name: `invalid expression sum`, input: `"a" + 0`, expectedError: `at +: Operands must be two numbers or two strings.`},
		{name: `invalid expression minus`, input: `0 - ""`, expectedError: `at -: Operands must be numbers.`},
		{name: `invalid expression minus string`, input: `-"a"`, expectedError: `at -: Operand must be a number.`},
		{name: `bang as boolean`, input: `!"a"`, expectedOutput: `false`},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			output, err := evaluate(tc.input)
			if tc.expectedError != "" {
				assert.ErrorContains(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedOutput, output)
			}
		})
	}
}

func evaluate(script string) (string, error) {
	eval := interpreter.NewInterpreter()
	scan := scanner.NewScanner(script)

	tokens, err := scan.Scan()
	if err != nil {
		return "", err
	}

	parse := parser.NewParser(tokens)
	expr, err := parse.Parse()
	if err != nil {
		return "", err
	}

	return eval.Interpret(expr)
}
