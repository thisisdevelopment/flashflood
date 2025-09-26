GOFILES = $(shell find . -name '*.go' -not -path './vendor/*')
GOPACKAGES = $(shell go list ./...  | grep -v /vendor/)

default: test

workdir:
	mkdir -p workdir

# For library projects, build means compile-check all packages
build: $(GOFILES)
	go build ./...

# Native build also just checks compilation for libraries
build-native: $(GOFILES)
	go build ./...

# Run tests
test: test-all

test-all:
	@go test -v $(GOPACKAGES)

# Run tests with coverage
test-coverage:
	@go test -race -coverprofile=coverage.out -covermode=atomic $(GOPACKAGES)

# Format code
fmt:
	@go fmt $(GOPACKAGES)

# Lint code
lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed, skipping linting"; \
	fi

# Run all quality checks
check: fmt lint test-coverage

# Clean build artifacts
clean:
	rm -rf workdir coverage.out coverage.html coverage.txt

.PHONY: default build build-native test test-all test-coverage fmt lint check clean