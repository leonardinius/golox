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

var testDir = "../test/"
var projectDir = "/Users/leo/src/golox"

var expectedOutputPattern = regexp.MustCompile(`// expect: ?(.*)`)
var expectedErrorPattern = regexp.MustCompile(`// (Error.*)`)
var errorLinePattern = regexp.MustCompile(`// \[((java|c|go) )?line (\d+)\] (Error.*)`)
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

// var cSuites = []string{}
var goSuites = []string{}

func TestAll(t *testing.T) {
	defineTestSuites(t)

	runSuites(t, maps.Keys(allSuites))
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
		failures := test.run()
		if len(failures) > 0 {
			t.Fatalf("%s\n%s", path, strings.Join(failures, "\n"))
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
			msg := fmt.Sprintf("[line %d] %s", lineNum, match[1])
			t.expectedErrors[msg] = msg
			t.expectedExitCode = 65
			continue
		}

		match = errorLinePattern.FindStringSubmatch(line)
		if match != nil {
			language := match[2]
			if language == "" || language == t.suite.language {
				msg := fmt.Sprintf("[line %s] %s", match[3], match[4])
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

func (t *Test) run() []string {
	args := []string{}
	args = append(args, t.suite.args...)
	args = append(args, t.path)

	cmd := exec.Command(t.suite.executable, args...)
	cmd.Dir = projectDir
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
		t.Errorf("Expected runtime error '%s' and got: %s", t.expectedRuntimeError, errorLines[0])
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
			errorMsg := fmt.Sprintf("[line %s] %s", match[1], match[2])
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

	t.Errorf("Expected return code %d and got %d. Stderr: %s", t.expectedExitCode, exitCode, strings.Join(errorLines, "\n"))
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
	// t.t.Errorf(format, args...)
}

func defineTestSuites(t *testing.T) {
	func() {
		// Build go lox
		cmd := exec.Command("go", "build", "-o", projectDir+"/bin/golox", projectDir+"/main.go")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("go build failed %v: %#v\n", err, out)
		}
	}()

	golox := func(name string, tests ...map[string]string) {
		suiteTests := map[string]string{}
		for _, test := range tests {
			maps.Copy(suiteTests, test)
		}
		allSuites[name] = &Suite{
			name:       name,
			language:   "go",
			executable: projectDir + "/bin/golox",
			tests:      suiteTests,
			args:       []string{"-profile=non-strict"},
		}
		goSuites = append(goSuites, name)
	}

	// These are just for earlier chapters.
	var earlyChapters = map[string]string{
		"test/scanning":    "skip",
		"test/expressions": "skip",
	}

	// Go doesn't correctly implement IEEE equality on boxed doubles.
	var goNaNEquality = map[string]string{
		// "test/number/nan_equality.lox": "skip",
	}

	// No hardcoded limits.
	var noGoLimits = map[string]string{
		"test/limit/loop_too_large.lox":     "skip",
		"test/limit/no_reuse_constants.lox": "skip",
		"test/limit/too_many_constants.lox": "skip",
		"test/limit/too_many_locals.lox":    "skip",
		"test/limit/too_many_upvalues.lox":  "skip",
		// Rely on Go for stack overflow checking.
		"test/limit/stack_overflow.lox": "skip",
	}

	var goloxClassAttributesAccessErrors = map[string]string{
		"test/field/get_on_class.lox": "skip",
		"test/field/set_on_class.lox": "skip",
	}

	golox("golox",
		map[string]string{"test": "pass"},
		earlyChapters,
		goNaNEquality,
		noGoLimits,
		goloxClassAttributesAccessErrors,
	)
}
