# (Go) Lox (Crafting interpreters)

This is an implementation of [The Lox Programming Language](https://www.craftinginterpreters.com/the-lox-language.html) implemented in Go.

Current status:

- Feature Complete as passing original test suite.

> ```raw
> runner_test.go:69: Suite golox: Tests=254, Passed=238, Failed=0, Skipped=16, Expectactions: 557
> ```

Excluded tests (re [eXtra Features](#extra-features)):

- `test/field/get_on_class.lox`.
- `test/field/set_on_class.lox`.

- See benchmarks against jlox (java) below.

---

## eXtra Features

- REPL expression output; readline support.
- block comments.
- `continue`, `break` statements.
- closures and anynymous functions.
- native functions: `Array`, `pprint(...)` varargs function.
- profiles: `-profile=non-strict` (test compliance) and `-profile=strict` **[default]** to report unused variables.
- Static `class` methods, and class properites (metaclass).

## How-To

Use provided make commands

```shell
$ make help
Usage: make <target>
 Default
        help                  Display this help
 Build/Run
        all                   ALL, builds the world
        clean                 Clean-up build artifacts
        gen                   Runs all codegen
        test                  Runs all tests
        lint                  Runs all linters
        run                   Runs golox. Use ARGS="" make run to pass arguments
```

## Benchmarks

Command used:

```shell
go test -v -run=XXX -bench=. -benchtime=30s -timeout=30m ./test/.
```

The benchmark work as follows:

- executes `bin/golox` and [`jlox`](https://github.com/munificent/craftinginterpreters/blob/master/jlox) from original project.
- benchmarks reports execution time and becnhmark elapsed time as printed by bechnmark script.

Please find numbers below:

```raw
goos: darwin
goarch: arm64
pkg: github.com/leonardinius/golox/test
BenchmarkAll
BenchmarkAll/GO/test/benchmark/equality.lox-8                 6 5009233757 ns/op          0.3980 elapsed/op
BenchmarkAll/JAVA/test/benchmark/equality.lox-8               8 4237118536 ns/op          0.7470 elapsed/op
BenchmarkAll/GO/test/benchmark/binary_trees.lox-8             4 9035542448 ns/op          8.914 elapsed/op
BenchmarkAll/JAVA/test/benchmark/binary_trees.lox-8           6 5911415000 ns/op          5.846 elapsed/op
BenchmarkAll/GO/test/benchmark/properties.lox-8               7 4927545500 ns/op          5.334 elapsed/op
BenchmarkAll/JAVA/test/benchmark/properties.lox-8             7 4493170381 ns/op          4.385 elapsed/op
BenchmarkAll/GO/test/benchmark/invocation.lox-8              10 3288547812 ns/op          3.097 elapsed/op
BenchmarkAll/JAVA/test/benchmark/invocation.lox-8            37  867685619 ns/op          0.7800 elapsed/op
BenchmarkAll/GO/test/benchmark/fib.lox-8                      4 8474984490 ns/op          8.397 elapsed/op
BenchmarkAll/JAVA/test/benchmark/fib.lox-8                    4 7828719094 ns/op          7.744 elapsed/op
BenchmarkAll/GO/test/benchmark/trees.lox-8                    2 23599520333 ns/op         21.59 elapsed/op
BenchmarkAll/JAVA/test/benchmark/trees.lox-8                  2 21246922084 ns/op         20.97 elapsed/op
BenchmarkAll/GO/test/benchmark/string_equality.lox-8          4 8742849177 ns/op          4.145 elapsed/op
BenchmarkAll/JAVA/test/benchmark/string_equality.lox-8        7 4765044637 ns/op          2.694 elapsed/op
BenchmarkAll/GO/test/benchmark/instantiation.lox-8           12 3103455396 ns/op          3.037 elapsed/op
BenchmarkAll/JAVA/test/benchmark/instantiation.lox-8         37  879199985 ns/op          0.8210 elapsed/op
BenchmarkAll/GO/test/benchmark/zoo_batch.lox-8                3 10016985805 ns/op         10.02 elapsed/op
BenchmarkAll/JAVA/test/benchmark/zoo_batch.lox-8              3 10068449986 ns/op         10.00 elapsed/op
BenchmarkAll/GO/test/benchmark/method_call.lox-8             22 1432509085 ns/op          1.429 elapsed/op
BenchmarkAll/JAVA/test/benchmark/method_call.lox-8           22 1625271705 ns/op          1.391 elapsed/op
BenchmarkAll/GO/test/benchmark/zoo.lox-8                     10 3097766238 ns/op          3.134 elapsed/op
BenchmarkAll/JAVA/test/benchmark/zoo.lox-8                    9 3355564380 ns/op          2.831 elapsed/op
PASS
ok   github.com/leonardinius/golox/test 1144.323s
```

Please find benchmark code [here](./test/gobenchmark_test.go).
