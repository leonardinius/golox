package interpreter_test

import (
	"context"
	"strings"
	"testing"

	"github.com/leonardinius/golox/internal/interpreter"
	"github.com/leonardinius/golox/internal/loxerrors"
	"github.com/leonardinius/golox/internal/parser"
	"github.com/leonardinius/golox/internal/scanner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInterpret(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		name string
		in   string // Input
		eval string // Expected eval
		out  string // Expected output
		err  string // Expected error
	}{
		{name: `simple expression`, in: `1 + 2;`, eval: `3`},
		{name: `grouped`, in: `(1 + 2);`, eval: `3`},
		{name: `nested`, in: `(1 + (2 + 3));`, eval: `6`},
		{name: `precedence asterix`, in: `1 + 2 * 3;`, eval: `7`},
		{name: `precedence slash`, in: `1 + 9 / 3;`, eval: `4`},
		{name: `precedence asterix slash`, in: `1 + 2 * 6 / 4;`, eval: `4`},
		{name: `grouping nested precedence`, in: `((1 + 2) * 3)/2;`, eval: `4.5`},
		{name: `strings`, in: `"a" + "b";`, eval: `"ab"`},
		{name: `boolean t`, in: `true;`, eval: `true`},
		{name: `boolean f`, in: `false;`, eval: `false`},
		{name: `bang`, in: `!false;`, eval: `true`},
		{name: `bang bang1`, in: `!!false;`, eval: `false`},
		{name: `eqeq number`, in: `1 == 1;`, eval: `true`},
		{name: `eqeq number`, in: `1 == 2;`, eval: `false`},
		{name: `eqeq string`, in: `"a" == "a";`, eval: `true`},
		{name: `eqeq string`, in: `"a" == "b";`, eval: `false`},
		{name: `bangeq number`, in: `1 != 1;`, eval: `false`},
		{name: `bangeq number`, in: `1 != 2;`, eval: `true`},
		{name: `bangeq string`, in: `"a" != "a";`, eval: `false`},
		{name: `bangeq string`, in: `"a" != "b";`, eval: `true`},
		{name: `lt number`, in: `1 < 2;`, eval: `true`},
		{name: `lt number`, in: `1 < 1;`, eval: `false`},
		{name: `lte number`, in: `2 <= 1;`, eval: `false`},
		{name: `lte number`, in: `1 <= 1;`, eval: `true`},
		{name: `gt number`, in: `2 > 1;`, eval: `true`},
		{name: `gt number`, in: `1 > 1;`, eval: `false`},
		{name: `gte number`, in: `1 >= 2;`, eval: `false`},
		{name: `gte number`, in: `1 >= 1;`, eval: `true`},
		{name: `invalid expression`, in: `1 + 2 +;`, err: `Parse error.`, out: `parse error at ';': expected expression`},
		{name: `invalid expression sum`, in: `"a" + 0;`, err: `Operands must be two numbers or two strings.`},
		{name: `invalid expression minus`, in: `0 - "";`, err: `Operands must be numbers.`},
		{name: `invalid expression minus string`, in: `-"a";`, err: `Operand must be a number.`},
		{name: `bang as boolean`, in: `!"a";`, eval: `false`},
		{name: `emty var`, in: `var a;`, eval: `nil`},
		{name: `emty var eval`, in: `var a;a;`, eval: `nil`},
		{name: `var init`, in: `var a =1;a;`, eval: `1`},
		{name: `var assign`, in: `var a =1;a=2;`, eval: `2`},
		{name: `var multiple var math`, in: `var a =1;var b=2;a+b;`, eval: `3`},
		{name: `var syntax error 1`, in: `var print;`, err: `Parse error.`, out: `parse error at 'print': expect variable name.`},
		{name: `var syntax error 2`, in: `var a`, err: `Parse error.`, out: `parse error at end: expect ';' after variable declaration.`},
		{name: `var assign error`, in: `var a;(a)=1;`, err: `Parse error.`, out: `parse error at '=': invalid assignment target.`},
		{name: `var assign error unrecognized var`, in: `b=1;`, err: `Undefined variable 'b'.`},
		{name: `var scope top level`, in: `var a=1;{a=2;print a;{a=3;print a;{a=4;print a;}}}print a;a;`, eval: `4`, out: "2\n3\n4\n4\n"},
		{name: `var scope nested`, in: `var a=1;{var a=2;print a;{var a=3;print a;{var a=4;print a;}}}print a;a;`, eval: `1`, out: "2\n3\n4\n1\n"},
		{name: `var scope multiple`, in: `var a=1;var b=2;{var a=2;print a;var b=4;print b;{var a=3;print a;var b=6;print b;{var a=4;print a;var b=8;print b;}}}print a;print b;a+b;`, eval: `3`, out: "2\n4\n3\n6\n4\n8\n1\n2\n"},
		{name: `logic and 1`, in: `1 and 2;`, eval: `2`},
		{name: `logic and 2`, in: `nil and 1;`, eval: `nil`},
		{name: `logic and 3`, in: `1 and nil;`, eval: `nil`},
		{name: `logic and shortcuit`, in: `nil and Unknown;`, eval: `nil`},
		{name: `logic or 1`, in: `1 or 2;`, eval: `1`},
		{name: `logic or 2`, in: `nil or 1;`, eval: `1`},
		{name: `logic or 3`, in: `1 or nil;`, eval: `1`},
		{name: `logic or short circuit`, in: `1 or Unknown;`, eval: `1`},
		{name: `while loop`, in: `var a=1;while(a<10){print a;a=a+1;}`, eval: `nil`, out: "1\n2\n3\n4\n5\n6\n7\n8\n9\n"},
		{name: `for loop`, in: `for(var a=1;a<10;a=a+1){print a;}`, eval: `nil`, out: "1\n2\n3\n4\n5\n6\n7\n8\n9\n"},
		{name: `break invalid syntax`, in: `break;1;`, err: `Parse error.`, out: `parse error at ';': must be inside a loop to use 'break'`},
		{name: `continue invalid syntax`, in: `continue;1;`, err: `Parse error.`, out: `parse error at ';': must be inside a loop to use 'continue'`},
		{name: `for loop`, in: `for(var a=1;a<10;a=a+1){print a;}`, eval: `nil`, out: "1\n2\n3\n4\n5\n6\n7\n8\n9\n"},
		{name: `while break`, in: `var a=0;while(true){if(a>3)break;a=a+1;print a;}`, eval: `nil`, out: "1\n2\n3\n4\n"},
		{name: `for break`, in: `for(var a=0;a<10;a=a+1){if(a>3)break;print a;}`, eval: `nil`, out: "0\n1\n2\n3\n"},
		{name: `while continue`, in: `var a=0;while(a<10){a=a+1;if(a<5)continue;print a;}`, eval: `nil`, out: "5\n6\n7\n8\n9\n10\n"},
		{name: `for continue`, in: `for(var a=0;a<10;a=a+1){if(a<5)continue;print a;}`, eval: `nil`, out: "5\n6\n7\n8\n9\n"},
		{name: `built in pprint`, in: `pprint();`, eval: `nil`, out: "\n"},
		{name: `built in pprint varargs`, in: `pprint(1,2,nil,3,4);`, eval: `nil`, out: "1 2 nil 3 4\n"},
		{name: `built in time`, in: `clock(1,2);`, eval: `nil`, err: "Expected 0 arguments but got 2."},
		{name: `call non function`, in: `"non function"();`, eval: `nil`, err: "Can only call functions and classes."},
		{name: `define fun add`, in: `fun add(a,b){return a+b;}add(1,2);`, eval: `3`},
		{name: `define fun error 1`, in: `fun add(a,b){return a+b;};add(1,2);`, err: "Parse error.", out: "[line 1] parse error at ';': expected expression.\n"},
		{name: `recursive fun`, in: `fun a(i){if (i==0) return "Exit"; else {print(i);return a(i-1);}} a(3);`, eval: `"Exit"`, out: "3\n2\n1\n"},
		{name: `anon fun`, in: `var a=fun (i){return i;};a(1);`, eval: `1`},
		{name: `closures`, in: `var a="global";{fun showA(){pprint(a);}showA();var a="block";showA();print a;}`, eval: `nil`, out: "global\nglobal\nblock\n"},
		{name: `oop class`, in: `class A{} print A;`, eval: `nil`, out: "A\n"},
		{name: `oop class method decl`, in: `class A{a(){}}`, eval: `nil`},
		{name: `oop class fields decl`, in: `class A{} var a = A();a.a = 1; a.a;`, eval: `1`},
		{name: `oop class method call`, in: `class Bacon{eat(){print "bacon";return 1;}} Bacon().eat();`, eval: `1`, out: "bacon\n"},
		{name: `oop class this bind`, in: `class Thing { getCallback() { this.a = 1; fun localFunction() { return this.a + 2; } return localFunction; } } var callback = Thing().getCallback(); callback();`, eval: `3`},
		{
			name: `oop constructor multi test`, in: `
		class A {
			init (a,b){
				this.a= a;
				this.b = b;
			}
		}

		fun theprint(self){ print(self.a);print(self.b);}

		var a = A(1,2);
		a.method = fun(){ return theprint(a); };
		a.method();`,
			eval: `nil`, out: "1\n2\n",
		},
		{
			name: `oop metaclass`, in: `
		class Math {
			class square(n) {
				this.b = 1; return n * n;
				}
			}
			print Math.square(3);
			print Math.b;
			Math.c=2;
			print Math.c;`,
			eval: `nil`, out: "9\n1\n2\n",
		},
		{
			name: `inheritance with super`, in: `
		class A {
			method() {
			  return "A method";
			}
		  }
		  class B < A {
			method() {
			  return "B method";
			}
			test() {
			  return super.method();
			}
		  }
		  class C < B {}
		  C().test();`,
			eval: `"A method"`,
		},
		{
			name: `inheritance with super`, in: `
		var array = Array(3);
		print array;
		// "length" returns the number of elements.
		print array.length; // "3".
		// "set" sets the element at the given index to the given value.
		array.set(1, "new");
		// "get" returns the element at a given index.
		print array.get(1); // "new".`,
			eval: `nil`, out: "[<nil> <nil> <nil>]\n3\nnew\n",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			evalout, stdout, err := evaluate(tc.in)
			if tc.err != "" {
				require.ErrorContains(t, err, tc.err)
			} else {
				assert.Equal(t, tc.eval, evalout)
				assert.Equal(t, tc.out, stdout)
				require.NoError(t, err)
			}
		})
	}
}

func TestInterpretReplMultiline(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name string
		in   []string // Input
		eval []string // Expected eval
		out  string   // Expected output
		err  string   // Expected error
	}{
		{
			name: `var repl`,
			in:   []string{`var dd;print dd;dd;`, `print dd;dd;`, `dd=5;`, `dd;`},
			eval: []string{`nil`, `nil`, `5`, `5`},
			out:  "nil\nnil\n",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			output, stdout, err := replLineByLine(tc.in...)
			if tc.err != "" {
				require.ErrorContains(t, err, tc.err)
			} else {
				assert.Equal(t, tc.eval, output)
				assert.Equal(t, tc.out, stdout)
				require.NoError(t, err)
			}
		})
	}
}

func evaluate(script string) (_evalout, _stdout string, _err error) {
	stdin := strings.NewReader("")
	stdouterr := strings.Builder{}
	reporter := loxerrors.NewErrReporter(&stdouterr)

	eval := interpreter.NewInterpreter(
		interpreter.WithStdin(stdin),
		interpreter.WithStdout(&stdouterr),
		interpreter.WithStderr(&stdouterr),
		interpreter.WithErrorReporter(reporter),
	)

	scan := scanner.NewScanner(script, reporter)

	tokens, err := scan.Scan()
	if err != nil {
		return "", stdouterr.String(), err
	}

	p := parser.NewParser(tokens, reporter)
	stmts, err := p.Parse()
	if err != nil {
		return "", stdouterr.String(), err
	}

	ctx := context.TODO()
	resolver := interpreter.NewResolver(eval, "default")
	if err := resolver.Resolve(ctx, stmts); err != nil {
		return "", stdouterr.String(), err
	}

	svalue, err := eval.Interpret(ctx, stmts)
	return svalue, stdouterr.String(), err
}

func replLineByLine(script ...string) (_evalout []string, _out string, _err error) {
	stdin := strings.NewReader("")
	stdouterr := strings.Builder{}
	reporter := loxerrors.NewErrReporter(&stdouterr)
	ctx := context.TODO()

	eval := interpreter.NewInterpreter(
		interpreter.WithStdin(stdin),
		interpreter.WithStdout(&stdouterr),
		interpreter.WithStderr(&stdouterr),
		interpreter.WithErrorReporter(reporter),
	)
	resolver := interpreter.NewResolver(eval, "default")

	results := make([]string, len(script))
	for index, s := range script {
		scan := scanner.NewScanner(s, reporter)

		tokens, err := scan.Scan()
		if err != nil {
			return nil, stdouterr.String(), err
		}

		p := parser.NewParser(tokens, reporter)
		stmts, err := p.Parse()
		if err != nil {
			return nil, stdouterr.String(), err
		}

		if err := resolver.Resolve(ctx, stmts); err != nil {
			return nil, stdouterr.String(), err
		}

		svalue, err := eval.Interpret(ctx, stmts)
		if err != nil {
			return nil, stdouterr.String(), err
		}
		results[index] = svalue
	}

	return results, stdouterr.String(), nil
}
