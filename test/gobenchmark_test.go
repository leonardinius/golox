package runner_test

import (
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

const benchProjectHomeDir = ".."

func BenchmarkAll(b *testing.B) {
	jloxBin := "/Users/leo/src/craftinginterpreters/jlox"

	workDir, err := filepath.Abs(benchProjectHomeDir)
	if err != nil {
		b.Fatalf("Failed to get absolute path: %v", err)
	}
	mainGo := workDir + "/main.go"
	goloxBin := workDir + "/bin/golox"
	cmd := exec.Command("go", "build", "-o", goloxBin, mainGo)
	if outbytes, err := cmd.CombinedOutput(); err != nil {
		out := string(outbytes)
		b.Fatalf("go build failed with %v: %#v\n", err, out)
	}

	benchmarks := []string{
		"test/benchmark/binary_trees.lox",
		"test/benchmark/equality.lox",
		"test/benchmark/fib.lox",
		"test/benchmark/instantiation.lox",
		"test/benchmark/invocation.lox",
		"test/benchmark/method_call.lox",
		"test/benchmark/properties.lox",
		"test/benchmark/string_equality.lox",
		"test/benchmark/trees.lox",
		"test/benchmark/zoo_batch.lox",
		"test/benchmark/zoo.lox",
	}

	for _, bench := range benchmarks {
		b.Run("GO/"+bench, func(b *testing.B) {
			runBenchN(b, workDir, bench, goloxBin)
		})
		b.Run("JAVA/"+bench, func(b *testing.B) {
			runBenchN(b, workDir, bench, jloxBin)
		})
	}
}

func runBenchN(b *testing.B, workDir, bench string, cli ...string) {
	b.Helper()
	for n := 0; n < b.N; n++ {
		runBench(b, workDir, bench, cli...)
	}
}

func runBench(b *testing.B, workDir, bench string, cli ...string) {
	b.Helper()
	cmd := exec.Command(cli[0], append(cli[1:], bench)...)
	cmd.Dir = workDir
	stdout := new(strings.Builder)
	stderr := new(strings.Builder)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	func() {
		defer func() {
			if err := recover(); err != nil {
				b.Errorf("Execute error %v: %v", cmd, err)
				return
			}
		}()
		if err := cmd.Run(); err != nil {
			b.Errorf("Execute error %v: %v", cmd, err)
			return
		}
	}()

	exitCode := cmd.ProcessState.ExitCode()
	outputLines := strings.Split(stdout.String(), "\n")
	errorLines := strings.Split(stderr.String(), "\n")
	for len(outputLines) > 0 && outputLines[len(outputLines)-1] == "" {
		outputLines = outputLines[:len(outputLines)-1]
	}
	for len(errorLines) > 0 && errorLines[len(errorLines)-1] == "" {
		errorLines = errorLines[:len(errorLines)-1]
	}

	if exitCode != 0 || len(errorLines) > 0 {
		b.Errorf("Command %v exited with code %v and error %v", cmd, exitCode, errorLines)
		return
	}

	elapsedTimeString := outputLines[len(outputLines)-1]
	elapsedTimeSeconds, err := strconv.ParseFloat(elapsedTimeString, 64)
	if err != nil {
		b.Errorf("Failed to parse elapsed time %v", elapsedTimeString)
		return
	}
	b.ReportMetric(elapsedTimeSeconds, "elapsed/op")
}
