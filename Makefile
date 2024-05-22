############################################################################################################
# Few rules to follow when editing this file:
#
# 1. Shell commands must be indented with a tab
# 2. Before each target add ".PHONY: target_name" to disable default file target
# 3. Add target description prefixed with "##" on the same line as target definition for "help" target to work
# 4. Be aware that each make command is executed in separate shell
#
# Tips:
#
# * Add an @ sign to suppress output of the command that is executed
# * Define variable like: VAR := value
# * Reference variable like: $(VAR)
# * Reference environment variables like: $(ENV_VARIABLE)
#
#############################################################################################################
.DELETE_ON_ERROR:
.SHELLFLAGS 	:= -eu -o pipefail -c
SHELL			= /bin/bash
BIN 			:= .bin
BUILDOUT		= ./bin
MAKEFLAGS 		+= --warn-undefined-variables
MAKEFLAGS 		+= --no-builtin-rules
MAKEFLAGS		+= --no-print-directory
GOPATH			?= ${shell go env GOPATH}
export GOBIN 	:= $(abspath $(BIN))

BOLD = \033[1m
CLEAR = \033[0m
CYAN = \033[36m
GREEN = \033[32m

##@: Default
.PHONY: help
help: ## Display this help
	@awk '\
		BEGIN {FS = ":.*##"; printf "Usage: make $(CYAN)<target>$(CLEAR)\n"} \
		/^[0-9a-zA-Z_\-\/]+?:[^#]*?## .*$$/ { printf "\t$(CYAN)%-20s$(CLEAR) %s\n", $$1, $$2 } \
		/^##@/ { printf "$(BOLD)%s$(CLEAR)\n", substr($$0, 5); }' \
		$(MAKEFILE_LIST)

##@: Build/Run

all: clean go/gen go/tidy go/format lint test ## ALL, builds the world

.PHONY: clean
clean: ## Clean-up build artifacts
	@echo -e "$(CYAN)--- clean...$(CLEAR)"
	@go clean
	@rm -rf ${BUILDOUT}

.PHONY: gen
gen: go/gen ## Runs all codegen

.PHONY: test
test: clean go/test ## Runs all tests

.PHONY: lint
lint: go/lint ## Runs all linters

.PHONY: run
run: ## Runs golox. Use ARGS="" make run to pass arguments
	@echo -e "$(CYAN)--- run ...$(CLEAR)"
	go run ./main.go $(ARGS)

###@: Go
.PHONY: go/format
go/format: $(BIN)/gofumpt ### Format all go files
	@echo -e "$(CYAN)--- format go files...$(CLEAR)"
	$(BIN)/gofumpt -w .

go/tidy: go.mod go.sum ### Tidy all Go dependencies
	@echo -e "$(CYAN)--- tidy go dependencies...$(CLEAR)"
	go mod tidy -v -x

.PHONY: go/lint
go/lint: $(BIN)/golangci-lint ### Lints the codebase using golangci-lint
	@echo -e "$(CYAN)--- lint codebase...$(CLEAR)"
	$(BIN)/golangci-lint run --modules-download-mode=readonly --config .golangci.yml

.PHONY: go/test
go/test: ### Runs all tests
	@echo -e "$(CYAN)--- go test ...$(CLEAR)"
	go test -shuffle=on -race -cover -timeout=60s -count 1 -parallel 3 -v ./...

.PHONY: go/gen
go/gen: ### Generate go code from API spec
	@echo -e "$(CYAN)--- go/gen...$(CLEAR)"
	@go generate ./internal/...

# TOOLS
$(BIN)/golangci-lint: Makefile
	@mkdir -p $(@D)
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.57.2

$(BIN)/gofumpt: Makefile
	@mkdir -p $(@D)
	go install mvdan.cc/gofumpt@v0.6.0
