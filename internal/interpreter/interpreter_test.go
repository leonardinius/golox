package interpreter_test

import (
	"context"
	"strings"
	"testing"

	"github.com/leonardinius/golox/internal/interpreter"
	"github.com/leonardinius/golox/internal/parser"
	"github.com/leonardinius/golox/internal/scanner"
	"github.com/stretchr/testify/assert"
)

func TestInterpret(t *testing.T) {
	testcases := []struct {
		name          string
		input         string
		expectedEval  string
		expectedOut   string
		expectedError string
	}{
		{name: `simple expression`, input: `1 + 2;`, expectedEval: `3`},
		{name: `grouped`, input: `(1 + 2);`, expectedEval: `3`},
		{name: `nested`, input: `(1 + (2 + 3));`, expectedEval: `6`},
		{name: `precedence asterix`, input: `1 + 2 * 3;`, expectedEval: `7`},
		{name: `precedence slash`, input: `1 + 9 / 3;`, expectedEval: `4`},
		{name: `precedence asterix slash`, input: `1 + 2 * 6 / 4;`, expectedEval: `4`},
		{name: `grouping nested precedence`, input: `((1 + 2) * 3)/2;`, expectedEval: `4.5`},
		{name: `strings`, input: `"a" + "b";`, expectedEval: `"ab"`},
		{name: `boolean t`, input: `true;`, expectedEval: `true`},
		{name: `boolean f`, input: `false;`, expectedEval: `false`},
		{name: `bang`, input: `!false;`, expectedEval: `true`},
		{name: `bang bang`, input: `!!false;`, expectedEval: `false`},
		{name: `eqeq number`, input: `1 == 1;`, expectedEval: `true`},
		{name: `eqeq number`, input: `1 == 2;`, expectedEval: `false`},
		{name: `eqeq string`, input: `"a" == "a";`, expectedEval: `true`},
		{name: `eqeq string`, input: `"a" == "b";`, expectedEval: `false`},
		{name: `bangeq number`, input: `1 != 1;`, expectedEval: `false`},
		{name: `bangeq number`, input: `1 != 2;`, expectedEval: `true`},
		{name: `bangeq string`, input: `"a" != "a";`, expectedEval: `false`},
		{name: `bangeq string`, input: `"a" != "b";`, expectedEval: `true`},
		{name: `lt number`, input: `1 < 2;`, expectedEval: `true`},
		{name: `lt number`, input: `1 < 1;`, expectedEval: `false`},
		{name: `lte number`, input: `2 <= 1;`, expectedEval: `false`},
		{name: `lte number`, input: `1 <= 1;`, expectedEval: `true`},
		{name: `gt number`, input: `2 > 1;`, expectedEval: `true`},
		{name: `gt number`, input: `1 > 1;`, expectedEval: `false`},
		{name: `gte number`, input: `1 >= 2;`, expectedEval: `false`},
		{name: `gte number`, input: `1 >= 1;`, expectedEval: `true`},
		{name: `invalid expression`, input: `1 + 2 +;`, expectedError: `parse error at ';': expected expression`},
		{name: `invalid expression sum`, input: `"a" + 0;`, expectedError: `at +: operands must be two numbers or two strings.`},
		{name: `invalid expression minus`, input: `0 - "";`, expectedError: `at -: operands must be numbers.`},
		{name: `invalid expression minus string`, input: `-"a";`, expectedError: `at -: operand must be a number.`},
		{name: `bang as boolean`, input: `!"a";`, expectedEval: `false`},
		{name: `emty var`, input: `var a;`, expectedEval: `nil`},
		{name: `emty var eval`, input: `var a;a;`, expectedEval: `nil`},
		{name: `var init`, input: `var a =1;a;`, expectedEval: `1`},
		{name: `var assign`, input: `var a =1;a=2;`, expectedEval: `2`},
		{name: `var multiple var math`, input: `var a =1;var b=2;a+b;`, expectedEval: `3`},
		{name: `var syntax error 1`, input: `var print;`, expectedError: `parse error at end: expect ';' after value.`},
		{name: `var syntax error 2`, input: `var a print;`, expectedError: `parse error at end: expect ';' after value.`},
		{name: `var assign error`, input: `var a;(a)=1;`, expectedError: `parse error at '=': invalid assignment target.`},
		{name: `var assign error unrecognized var`, input: `b=1;`, expectedError: `at b: undefined variable 'b'.`},
		{name: `var scope top level`, input: `var a=1;{a=2; print a; {a=3; print a;{a=4; print a; }}}print a;a;`, expectedEval: `4`, expectedOut: "2\n3\n4\n4\n"},
		{name: `var scope nested`, input: `var a=1;{var a=2; print a; {var a=3; print a;{var a=4; print a; }}}print a;a;`, expectedEval: `1`, expectedOut: "2\n3\n4\n1\n"},
		{name: `var scope multiple`, input: `var a=1;var b=2;{var a=2; print a; var b=4; print b;{var a=3; print a; var b=6; print b;{var a=4; print a; var b=8; print b;}}}print a;print b; a+b;`, expectedEval: `3`, expectedOut: "2\n4\n3\n6\n4\n8\n1\n2\n"},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			output, stdout, err := evaluate(tc.input)
			if tc.expectedError != "" {
				assert.ErrorContains(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedEval, output)
				assert.Equal(t, tc.expectedOut, stdout)
			}
		})
	}
}

func evaluate(script string) (string, string, error) {
	stdin := strings.NewReader("")
	stdout := strings.Builder{}

	eval := interpreter.NewInterpreter(
		interpreter.WithStdin(stdin),
		interpreter.WithStdout(&stdout),
		interpreter.WithStderr(&stdout),
	)
	scan := scanner.NewScanner(script)

	tokens, err := scan.Scan()
	if err != nil {
		return "", stdout.String(), err
	}

	p := parser.NewParser(tokens)
	stmts, err := p.Parse()
	if err != nil {
		return "", stdout.String(), err
	}

	svalue, err := eval.Interpret(context.TODO(), stmts)
	return svalue, stdout.String(), err
}
