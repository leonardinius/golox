package runner_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/exp/maps"
)

var testDir = "/Users/leo/src/craftinginterpreters/test/"
var binDir = "/Users/leo/src/craftinginterpreters/"

var expectedOutputPattern = regexp.MustCompile(`// expect: ?(.*)`)
var expectedErrorPattern = regexp.MustCompile(`// (Error.*)`)
var errorLinePattern = regexp.MustCompile(`// \[((java|c) )?line (\d+)\] (Error.*)`)
var expectedRuntimeErrorPattern = regexp.MustCompile(`// expect runtime error: (.+)`)
var syntaxErrorPattern = regexp.MustCompile(`\[.*line (\d+)\] (Error.+)`)
var stackTracePattern = regexp.MustCompile(`\[line (\d+)\]`)
var nonTestPattern = regexp.MustCompile(`// nontest`)

type Suite struct {
	name       string
	language   string
	executable string
	args       []string
	tests      map[string]string
}

var allSuites = map[string]*Suite{}
var cSuites = []string{}
var javaSuites = []string{}

func init() {
	defineTestSuites()
}

func TestAll(t *testing.T) {
	runSuites(t, maps.Keys(allSuites))
}

func TestJavaSuites(t *testing.T) {
	runSuites(t, javaSuites)
}

func TestCSuites(t *testing.T) {
	runSuites(t, cSuites)
}

func runSuites(t *testing.T, names []string) {
	t.Helper()
	t.Parallel()
	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			runSuite(t, allSuites[name])
		})
	}
}

func runSuite(t *testing.T, suite *Suite) {
	t.Helper()
	require.DirExists(t, testDir)

	var files []string
	err := filepath.Walk(testDir, func(path string, f os.FileInfo, _ error) error {
		if f.IsDir() || filepath.Ext(path) != ".lox" {
			return nil
		}

		relPath, err := filepath.Rel(filepath.Join(testDir, ".."), path)
		if err == nil {
			files = append(files, relPath)
		}

		return err
	})
	require.NoError(t, err)

	for _, file := range files {
		runTest(t, suite, file)
	}
}

func runTest(t *testing.T, suite *Suite, path string) {
	if strings.Contains(path, "benchmark") {
		return
	}

	test := &Test{path: path, suite: suite, expectedErrors: make(map[string]string)}

	t.Run(path, func(t *testing.T) {
		test.t = t
		if !test.parse() {
			return
		}
		failures := test.run(suite.executable, suite.args)
		if len(failures) > 0 {
			t.Fatalf("Test failed:\n%s", strings.Join(failures, "\n"))
		}
	})
}

type ExpectedOutput struct {
	line   int
	output string
}

type Test struct {
	t                    *testing.T
	path                 string
	suite                *Suite
	expectedOutput       []ExpectedOutput
	expectedErrors       map[string]string
	expectedRuntimeError string
	runtimeErrorLine     int
	expectedExitCode     int
	failures             []string
}

func (t *Test) parse() bool {
	// Get the path components.
	parts := strings.Split(t.path, "/")
	var subpath string
	var state string

	// Figure out the state of the test. We don't break out of this loop because
	// we want lines for more specific paths to override more general ones.
	for _, part := range parts {
		if subpath != "" {
			subpath += "/"
		}
		subpath += part

		if val, ok := t.suite.tests[subpath]; ok {
			state = val
		}
	}

	require.NotEmptyf(t.t, state, "Unknown test state for '%s'", t.path)
	if state == "skip" {
		return false
	}

	lines, err := os.ReadFile(filepath.Join(testDir, "..", t.path))
	require.NoError(t.t, err)

	for lineNum, line := range strings.Split(string(lines), "\n") {
		lineNum++

		if nonTestPattern.MatchString(line) {
			return false
		}

		match := expectedOutputPattern.FindStringSubmatch(line)
		if match != nil {
			t.expectedOutput = append(t.expectedOutput, ExpectedOutput{line: lineNum, output: match[1]})
			continue
		}

		match = expectedErrorPattern.FindStringSubmatch(line)
		if match != nil {
			msg := fmt.Sprintf("[%d] %s", lineNum, match[1])
			t.expectedErrors[msg] = msg
			t.expectedExitCode = 65
			continue
		}

		match = errorLinePattern.FindStringSubmatch(line)
		if match != nil {
			language := match[2]
			if language == "" || language == t.suite.language {
				msg := fmt.Sprintf("[%s] %s", match[3], match[4])
				t.expectedErrors[msg] = msg
				t.expectedExitCode = 65
			}
			continue
		}

		match = expectedRuntimeErrorPattern.FindStringSubmatch(line)
		if match != nil {
			t.runtimeErrorLine = lineNum
			t.expectedRuntimeError = match[1]
			t.expectedExitCode = 70
		}
	}

	if len(t.expectedErrors) > 0 && t.expectedRuntimeError != "" {
		require.Fail(t.t, "parse", "TEST ERROR %s\nCannot expect both compile and runtime errors.", t.path)
		return false
	}

	return true
}

func (t *Test) run(customInterpreter string, customArguments []string) []string {
	args := []string{}
	if customInterpreter != "" {
		args = append(args, customArguments...)
	} else {
		args = append(args, t.suite.args...)
	}
	args = append(args, t.path)

	cmd := exec.Command(customInterpreter, args...)
	cmd.Dir = binDir
	stdout := new(strings.Builder)
	stderr := new(strings.Builder)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	func() {
		if err := recover(); err != nil {
			t.Errorf("Execute error %v: %#v", cmd, err)
		}
		_ = cmd.Run()
	}()

	outputLines := strings.Split(stdout.String(), "\n")
	errorLines := strings.Split(stderr.String(), "\n")

	if t.expectedRuntimeError != "" {
		t.validateRuntimeError(errorLines)
	} else {
		t.validateCompileErrors(errorLines)
	}
	t.validateExitCode(cmd.ProcessState.ExitCode(), errorLines)
	t.validateOutput(outputLines)

	return t.failures
}

func (t *Test) validateRuntimeError(errorLines []string) {

	if len(errorLines) < 2 {
		t.Errorf("Expected runtime error '%s' and got none.", t.expectedRuntimeError)
		return
	}

	if errorLines[0] != t.expectedRuntimeError {
		t.Errorf("Expected runtime error '%s' and got: %s", errorLines[0], t.expectedRuntimeError)
		return
	}

	var stackLine int
	for _, line := range errorLines[1:] {
		match := stackTracePattern.FindStringSubmatch(line)
		if match != nil {
			stackLine, _ = strconv.Atoi(match[1])
			break
		}
	}

	if stackLine == 0 {
		t.Errorf("Expected stack trace and got: %s", errorLines[1:])
	} else if stackLine != t.runtimeErrorLine {
		t.Errorf("Expected runtime error on line %d but was on line %d.", t.runtimeErrorLine, stackLine)
	}
}

func (t *Test) validateCompileErrors(errorLines []string) {
	foundErrors := map[string]bool{}
	unexpectedCount := 0

	for _, line := range errorLines {
		match := syntaxErrorPattern.FindStringSubmatch(line)
		if match != nil {
			errorMsg := fmt.Sprintf("[%s] %s", match[1], match[2])
			if _, ok := t.expectedErrors[errorMsg]; ok {
				foundErrors[errorMsg] = true
			} else {
				if unexpectedCount < 10 {
					t.Errorf("Unexpected error: %s", line)
				}
				unexpectedCount++
			}
		} else if line != "" {
			if unexpectedCount < 10 {
				t.Errorf("Unexpected output on stderr: %s", line)
			}
			unexpectedCount++
		}
	}

	if unexpectedCount > 10 {
		t.Errorf("(truncated %d more...)", unexpectedCount-10)
	}

	for errorMsg := range t.expectedErrors {
		if _, ok := foundErrors[errorMsg]; !ok {
			t.Errorf("Missing expected error: %s", errorMsg)
		}
	}
}

func (t *Test) validateExitCode(exitCode int, errorLines []string) {
	if exitCode == t.expectedExitCode {
		return
	}

	if len(errorLines) > 10 {
		errorLines = errorLines[:10]
		errorLines = append(errorLines, "(truncated...)")
	}

	t.Errorf("Expected return code %d and got %d. Stderr: %v", t.expectedExitCode, exitCode, errorLines)
}

func (t *Test) validateOutput(outputLines []string) {
	if len(outputLines) > 0 && outputLines[len(outputLines)-1] == "" {
		outputLines = outputLines[:len(outputLines)-1]
	}

	if len(outputLines) > len(t.expectedOutput) {
		t.Errorf("Got output '%s' when none was expected.", outputLines[len(t.expectedOutput)])
		return
	}

	for i, line := range outputLines {
		expected := t.expectedOutput[i]
		if expected.output != line {
			t.Errorf("Expected output '%s' on line %d and got '%s'.", expected.output, expected.line, line)
		}
	}

	for i := len(outputLines); i < len(t.expectedOutput); i++ {
		expected := t.expectedOutput[i]
		t.Errorf("Missing expected output '%s' on line %d.", expected.output, expected.line)
	}
}

func (t *Test) Errorf(format string, args ...interface{}) {
	t.t.Helper()
	t.failures = append(t.failures, fmt.Sprintf(format, args...))
	t.t.Errorf(format, args...)
}

func defineTestSuites() {
	c := func(name string, tests ...map[string]string) {
		executable := "build/" + name
		if name == "clox" {
			executable = "build/cloxd"
		}

		suiteTests := map[string]string{}
		for _, test := range tests {
			maps.Copy(suiteTests, test)
		}

		allSuites[name] = &Suite{name: name, language: "c", executable: executable, args: nil, tests: suiteTests}
		cSuites = append(cSuites, name)
	}

	java := func(name string, tests ...map[string]string) {
		dir := "build/gen/" + name
		if name == "jlox" {
			dir = "build/java"
		}

		suiteTests := map[string]string{}
		for _, test := range tests {
			maps.Copy(suiteTests, test)
		}
		allSuites[name] = &Suite{name: name, language: "java", executable: "java", args: []string{"-cp", dir, "com.craftinginterpreters.lox.Lox"}, tests: suiteTests}
		javaSuites = append(javaSuites, name)
	}

	// These are just for earlier chapters.
	var earlyChapters = map[string]string{
		"test/scanning":    "skip",
		"test/expressions": "skip",
	}

	// JVM doesn't correctly implement IEEE equality on boxed doubles.
	var javaNaNEquality = map[string]string{
		"test/number/nan_equality.lox": "skip",
	}

	// No hardcoded limits in jlox.
	var noJavaLimits = map[string]string{
		"test/limit/loop_too_large.lox":     "skip",
		"test/limit/no_reuse_constants.lox": "skip",
		"test/limit/too_many_constants.lox": "skip",
		"test/limit/too_many_locals.lox":    "skip",
		"test/limit/too_many_upvalues.lox":  "skip",

		// Rely on JVM for stack overflow checking.
		"test/limit/stack_overflow.lox": "skip",
	}

	// No classes in Java yet.
	var noJavaClasses = map[string]string{
		"test/assignment/to_this.lox":                  "skip",
		"test/call/object.lox":                         "skip",
		"test/class":                                   "skip",
		"test/closure/close_over_method_parameter.lox": "skip",
		"test/constructor":                             "skip",
		"test/field":                                   "skip",
		"test/inheritance":                             "skip",
		"test/method":                                  "skip",
		"test/number/decimal_point_at_eof.lox":         "skip",
		"test/number/trailing_dot.lox":                 "skip",
		"test/operator/equals_class.lox":               "skip",
		"test/operator/equals_method.lox":              "skip",
		"test/operator/not_class.lox":                  "skip",
		"test/regression/394.lox":                      "skip",
		"test/super":                                   "skip",
		"test/this":                                    "skip",
		"test/return/in_method.lox":                    "skip",
		"test/variable/local_from_method.lox":          "skip",
	}

	// No functions in Java yet.
	var noJavaFunctions = map[string]string{
		"test/call":                      "skip",
		"test/closure":                   "skip",
		"test/for/closure_in_body.lox":   "skip",
		"test/for/return_closure.lox":    "skip",
		"test/for/return_inside.lox":     "skip",
		"test/for/syntax.lox":            "skip",
		"test/function":                  "skip",
		"test/operator/not.lox":          "skip",
		"test/regression/40.lox":         "skip",
		"test/return":                    "skip",
		"test/unexpected_character.lox":  "skip",
		"test/while/closure_in_body.lox": "skip",
		"test/while/return_closure.lox":  "skip",
		"test/while/return_inside.lox":   "skip",
	}

	// No resolution in Java yet.
	var noJavaResolution = map[string]string{
		"test/closure/assign_to_shadowed_later.lox": "skip",
		"test/function/local_mutual_recursion.lox":  "skip",
		"test/variable/collide_with_parameter.lox":  "skip",
		"test/variable/duplicate_local.lox":         "skip",
		"test/variable/duplicate_parameter.lox":     "skip",
		"test/variable/early_bound.lox":             "skip",

		// Broken because we haven"t fixed it yet by detecting the error.
		"test/return/at_top_level.lox":               "skip",
		"test/variable/use_local_in_initializer.lox": "skip",
	}

	// No control flow in C yet.
	var noCControlFlow = map[string]string{
		"test/block/empty.lox":                  "skip",
		"test/for":                              "skip",
		"test/if":                               "skip",
		"test/limit/loop_too_large.lox":         "skip",
		"test/logical_operator":                 "skip",
		"test/variable/unreached_undefined.lox": "skip",
		"test/while":                            "skip",
	}

	// No functions in C yet.
	var noCFunctions = map[string]string{
		"test/call":                                "skip",
		"test/closure":                             "skip",
		"test/for/closure_in_body.lox":             "skip",
		"test/for/return_closure.lox":              "skip",
		"test/for/return_inside.lox":               "skip",
		"test/for/syntax.lox":                      "skip",
		"test/function":                            "skip",
		"test/limit/no_reuse_constants.lox":        "skip",
		"test/limit/stack_overflow.lox":            "skip",
		"test/limit/too_many_constants.lox":        "skip",
		"test/limit/too_many_locals.lox":           "skip",
		"test/limit/too_many_upvalues.lox":         "skip",
		"test/regression/40.lox":                   "skip",
		"test/return":                              "skip",
		"test/unexpected_character.lox":            "skip",
		"test/variable/collide_with_parameter.lox": "skip",
		"test/variable/duplicate_parameter.lox":    "skip",
		"test/variable/early_bound.lox":            "skip",
		"test/while/closure_in_body.lox":           "skip",
		"test/while/return_closure.lox":            "skip",
		"test/while/return_inside.lox":             "skip",
	}

	// No classes in C yet.
	var noCClasses = map[string]string{
		"test/assignment/to_this.lox":                  "skip",
		"test/call/object.lox":                         "skip",
		"test/class":                                   "skip",
		"test/closure/close_over_method_parameter.lox": "skip",
		"test/constructor":                             "skip",
		"test/field":                                   "skip",
		"test/inheritance":                             "skip",
		"test/method":                                  "skip",
		"test/number/decimal_point_at_eof.lox":         "skip",
		"test/number/trailing_dot.lox":                 "skip",
		"test/operator/equals_class.lox":               "skip",
		"test/operator/equals_method.lox":              "skip",
		"test/operator/not.lox":                        "skip",
		"test/operator/not_class.lox":                  "skip",
		"test/regression/394.lox":                      "skip",
		"test/return/in_method.lox":                    "skip",
		"test/super":                                   "skip",
		"test/this":                                    "skip",
		"test/variable/local_from_method.lox":          "skip",
	}

	// No inheritance in C yet.
	var noCInheritance = map[string]string{
		"test/class/local_inherit_other.lox": "skip",
		"test/class/local_inherit_self.lox":  "skip",
		"test/class/inherit_self.lox":        "skip",
		"test/class/inherited_method.lox":    "skip",
		"test/inheritance":                   "skip",
		"test/regression/394.lox":            "skip",
		"test/super":                         "skip",
	}

	java("jlox",
		map[string]string{"test": "pass"},
		earlyChapters,
		javaNaNEquality,
		noJavaLimits,
	)

	java("chap04_scanning", map[string]string{
		// No interpreter yet.
		"test":          "skip",
		"test/scanning": "pass",
	})

	// No test for chapter 5. It just has a hardcoded main() in AstPrinter.

	java("chap06_parsing", map[string]string{
		// No real interpreter yet.
		"test":                       "skip",
		"test/expressions/parse.lox": "pass",
	})

	java("chap07_evaluating", map[string]string{
		// No real interpreter yet.
		"test":                          "skip",
		"test/expressions/evaluate.lox": "pass",
	})

	java("chap08_statements",
		map[string]string{"test": "pass"},
		earlyChapters,
		javaNaNEquality,
		noJavaLimits,
		noJavaFunctions,
		noJavaResolution,
		noJavaClasses,
		map[string]string{
			// No control flow.
			"test/block/empty.lox":                  "skip",
			"test/for":                              "skip",
			"test/if":                               "skip",
			"test/logical_operator":                 "skip",
			"test/while":                            "skip",
			"test/variable/unreached_undefined.lox": "skip",
		})

	java("chap09_control",
		map[string]string{"test": "pass"},
		earlyChapters,
		javaNaNEquality,
		noJavaLimits,
		noJavaFunctions,
		noJavaResolution,
		noJavaClasses,
	)

	java("chap10_functions",
		map[string]string{"test": "pass"},
		earlyChapters,
		javaNaNEquality,
		noJavaLimits,
		noJavaResolution,
		noJavaClasses,
	)

	java("chap11_resolving",
		map[string]string{"test": "pass"},
		earlyChapters,
		javaNaNEquality,
		noJavaLimits,
		noJavaClasses,
	)

	java("chap12_classes",
		map[string]string{"test": "pass"},
		earlyChapters,
		noJavaLimits,
		javaNaNEquality,

		map[string]string{
			// No inheritance.
			"test/class/local_inherit_other.lox": "skip",
			"test/class/local_inherit_self.lox":  "skip",
			"test/class/inherit_self.lox":        "skip",
			"test/class/inherited_method.lox":    "skip",
			"test/inheritance":                   "skip",
			"test/regression/394.lox":            "skip",
			"test/super":                         "skip",
		})

	java("chap13_inheritance",
		map[string]string{"test": "pass"},
		earlyChapters,
		javaNaNEquality,
		noJavaLimits,
	)

	c("clox",
		map[string]string{"test": "pass"},
		earlyChapters,
	)

	c("chap17_compiling", map[string]string{
		// No real interpreter yet.
		"test":                          "skip",
		"test/expressions/evaluate.lox": "pass",
	})

	c("chap18_types", map[string]string{
		// No real interpreter yet.
		"test":                          "skip",
		"test/expressions/evaluate.lox": "pass",
	})

	c("chap19_strings", map[string]string{
		// No real interpreter yet.
		"test":                          "skip",
		"test/expressions/evaluate.lox": "pass",
	})

	c("chap20_hash", map[string]string{
		// No real interpreter yet.
		"test":                          "skip",
		"test/expressions/evaluate.lox": "pass",
	})

	c("chap21_global",
		map[string]string{"test": "pass"},
		earlyChapters,
		noCControlFlow,
		noCFunctions,
		noCClasses,

		map[string]string{
			// No blocks.
			"test/assignment/local.lox":                         "skip",
			"test/variable/in_middle_of_block.lox":              "skip",
			"test/variable/in_nested_block.lox":                 "skip",
			"test/variable/scope_reuse_in_different_blocks.lox": "skip",
			"test/variable/shadow_and_local.lox":                "skip",
			"test/variable/undefined_local.lox":                 "skip",

			// No local variables.
			"test/block/scope.lox":                       "skip",
			"test/variable/duplicate_local.lox":          "skip",
			"test/variable/shadow_global.lox":            "skip",
			"test/variable/shadow_local.lox":             "skip",
			"test/variable/use_local_in_initializer.lox": "skip",
		})

	c("chap22_local",
		map[string]string{"test": "pass"},
		earlyChapters,
		noCControlFlow,
		noCFunctions,
		noCClasses,
	)

	c("chap23_jumping",
		map[string]string{"test": "pass"},
		earlyChapters,
		noCFunctions,
		noCClasses,
	)

	c("chap24_calls",
		map[string]string{"test": "pass"},
		earlyChapters,
		noCClasses,

		map[string]string{
			// No closures.
			"test/closure":                      "skip",
			"test/for/closure_in_body.lox":      "skip",
			"test/for/return_closure.lox":       "skip",
			"test/function/local_recursion.lox": "skip",
			"test/limit/too_many_upvalues.lox":  "skip",
			"test/regression/40.lox":            "skip",
			"test/while/closure_in_body.lox":    "skip",
			"test/while/return_closure.lox":     "skip",
		})

	c("chap25_closures",
		map[string]string{"test": "pass"},
		earlyChapters,
		noCClasses,
	)

	c("chap26_garbage",
		map[string]string{"test": "pass"},
		earlyChapters,
		noCClasses,
	)

	c("chap27_classes",
		map[string]string{"test": "pass"},
		earlyChapters,
		noCInheritance,

		map[string]string{
			// No methods.
			"test/assignment/to_this.lox":                  "skip",
			"test/class/local_reference_self.lox":          "skip",
			"test/class/reference_self.lox":                "skip",
			"test/closure/close_over_method_parameter.lox": "skip",
			"test/constructor":                             "skip",
			"test/field/get_and_set_method.lox":            "skip",
			"test/field/method.lox":                        "skip",
			"test/field/method_binds_this.lox":             "skip",
			"test/method":                                  "skip",
			"test/operator/equals_class.lox":               "skip",
			"test/operator/equals_method.lox":              "skip",
			"test/return/in_method.lox":                    "skip",
			"test/this":                                    "skip",
			"test/variable/local_from_method.lox":          "skip",
		})

	c("chap28_methods",
		map[string]string{"test": "pass"},
		earlyChapters,
		noCInheritance,
	)

	c("chap29_superclasses",
		map[string]string{"test": "pass"},
		earlyChapters,
	)

	c("chap30_optimization",
		map[string]string{"test": "pass"},
		earlyChapters,
	)

}
