# (Go) Lox (Crafting interpreters)

This is an implementation of [The Lox Programming Language](https://www.craftinginterpreters.com/the-lox-language.html) implemented in Go.

Current status: Feature Complete as passing original test suite

> ```raw
> runner_test.go:69: Suite golox: Tests=254, Passed=238, Failed=0, Skipped=16, Expectactions: 557
> ```

The tests excluded: [GH](https://github.com/leonardinius/golox/blob/dbff2b805a05c01eb33589c503703f5ade571eb0/test/runner_test.go#L412-L415)

> ```go
> var goloxClassAttributesAccessErrors = map[string]string{
>   "test/field/get_on_class.lox": "skip",
>   "test/field/set_on_class.lox": "skip",
> }
> ```

---

## eXtra Features

- REPL expression output; readline support.
- `continue`, `break` statement.
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
        clean                 Format all go files
        gen                   Runs all codegen
        test                  Runs all tests
        lint                  Runs all linters
        run                   Runs golox. Use ARGS="" make run to pass arguments
```
