include ./Makefile.Common

BINARY_NAME := goswaggen
BINARY_PATH := ./bin/$(BINARY_NAME)
GO_PACKAGES := ./...

# -------------------------------------------------------------------
# Build
# -------------------------------------------------------------------
.PHONY: clean
clean:
	rm -rf ./bin coverage.out results.json

.PHONY: build
build:
	go build -o $(BINARY_PATH) -v .

# -------------------------------------------------------------------
# Test
# -------------------------------------------------------------------
.PHONY: gotest
gotest:
	go test -race -coverpkg=$(GO_PACKAGES) -coverprofile=coverage.out $(GO_PACKAGES)

.PHONY: go-cover-html
go-cover-html: gotest
	go tool cover -html=coverage.out

.PHONY: go-cover-func
go-cover-func: gotest
	go tool cover -func=coverage.out

# -------------------------------------------------------------------
# Lint & Format
# -------------------------------------------------------------------
.PHONY: lint
lint:
	golangci-lint run $(GO_PACKAGES)

.PHONY: fmt
fmt:
	gofmt -w .
	goimports -w .

.PHONY: fmt-check
fmt-check:
	@gofmt -l . | grep . && echo "^^^ files need formatting, run 'make fmt'" && exit 1 || true

# -------------------------------------------------------------------
# Security Check
# -------------------------------------------------------------------
.PHONY: vulncheck
vulncheck:
	govulncheck $(GO_PACKAGES)

.PHONY: gosec
gosec:
	gosec -fmt json -out results.json $(GO_PACKAGES)

# -------------------------------------------------------------------
# Tidy
# -------------------------------------------------------------------
.PHONY: tidy
tidy:
	go mod tidy

.PHONY: tidy-check
tidy-check:
	go mod tidy
	git diff --exit-code go.mod go.sum || (echo "go.mod/go.sum is not tidy, run 'make tidy'" && exit 1)

# -------------------------------------------------------------------
# Precommit — run before pushing/committing
# -------------------------------------------------------------------
.PHONY: precommit
precommit: tidy-check fmt-check lint test-all vulncheck
	@echo "✅ All precommit checks passed"
