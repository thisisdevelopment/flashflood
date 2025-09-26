#!/bin/bash

# FlashFlood v2 Test Runner with Coverage Reporting
# Usage: ./test-runner.sh [-v] [-c] [-r] [-b] [-s] [-h]
#   -v: Verbose output
#   -c: Generate coverage report
#   -r: Open coverage report in browser
#   -b: Run benchmarks
#   -s: Short tests (CI-friendly, smaller datasets)
#   -h: Show help

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default options
VERBOSE=false
COVERAGE=false
OPEN_REPORT=false
RUN_BENCHMARKS=false
SHORT_TESTS=false

# Parse command line arguments
while getopts "vcrbsh" opt; do
    case $opt in
        v)
            VERBOSE=true
            ;;
        c)
            COVERAGE=true
            ;;
        r)
            COVERAGE=true
            OPEN_REPORT=true
            ;;
        b)
            RUN_BENCHMARKS=true
            ;;
        s)
            SHORT_TESTS=true
            ;;
        h)
            echo "FlashFlood v2 Test Runner"
            echo "Usage: $0 [-v] [-c] [-r] [-b] [-s] [-h]"
            echo "  -v: Verbose output"
            echo "  -c: Generate coverage report"
            echo "  -r: Open coverage report in browser"
            echo "  -b: Run benchmarks"
            echo "  -s: Short tests (CI-friendly, smaller datasets)"
            echo "  -h: Show help"
            exit 0
            ;;
        \?)
            echo "Invalid option: -$OPTARG" >&2
            exit 1
            ;;
    esac
done

echo -e "${BLUE}===========================================${NC}"
echo -e "${BLUE}    FlashFlood v2 Test Runner${NC}"
echo -e "${BLUE}===========================================${NC}"

# Check if we're in the right directory
if [ ! -f "go.mod" ]; then
    echo -e "${RED}Error: go.mod not found. Please run this script from the project root.${NC}"
    exit 1
fi

# Verify this is the FlashFlood v2 project
if ! grep -q "github.com/thisisdevelopment/flashflood/v2" go.mod; then
    echo -e "${YELLOW}Warning: This doesn't appear to be FlashFlood v2 project${NC}"
fi

echo -e "\n${YELLOW}🔧 Environment Check${NC}"
echo "Go version: $(go version)"
echo "Project: $(grep '^module' go.mod | cut -d' ' -f2)"

# Clean any previous test artifacts
echo -e "\n${YELLOW}🧹 Cleaning previous test artifacts${NC}"
rm -f coverage.out coverage.html coverage.txt
go clean -testcache

# Format code
echo -e "\n${YELLOW}📐 Formatting code${NC}"
go fmt ./...

# Download dependencies
echo -e "\n${YELLOW}📦 Downloading dependencies${NC}"
go mod download
go mod tidy

# Build to check for compile errors
echo -e "\n${YELLOW}🔨 Building project${NC}"
if ! go build ./...; then
    echo -e "${RED}❌ Build failed${NC}"
    exit 1
fi
echo -e "${GREEN}✅ Build successful${NC}"

# Run tests
echo -e "\n${YELLOW}🧪 Running tests${NC}"

if [ "$COVERAGE" = true ]; then
    echo "Generating coverage report..."

    # Build test flags
    TEST_FLAGS="-race -coverprofile=coverage.out -covermode=atomic"
    if [ "$VERBOSE" = true ]; then
        TEST_FLAGS="-v $TEST_FLAGS"
    fi
    if [ "$SHORT_TESTS" = true ]; then
        TEST_FLAGS="$TEST_FLAGS -short"
    fi

    go test $TEST_FLAGS ./...

    # Check if tests passed
    if [ $? -ne 0 ]; then
        echo -e "${RED}❌ Tests failed${NC}"
        exit 1
    fi

    # Generate coverage report
    echo -e "\n${YELLOW}📊 Generating coverage report${NC}"
    go tool cover -html=coverage.out -o coverage.html

    # Get coverage percentage
    COVERAGE_PERCENT=$(go tool cover -func=coverage.out | grep "total:" | awk '{print $3}')

    # Coverage summary
    echo -e "\n${BLUE}===========================================${NC}"
    echo -e "${BLUE}         Coverage Report${NC}"
    echo -e "${BLUE}===========================================${NC}"
    go tool cover -func=coverage.out | tail -20
    echo -e "${BLUE}===========================================${NC}"
    echo -e "${GREEN}Overall Coverage: ${COVERAGE_PERCENT}${NC}"
    echo -e "${BLUE}===========================================${NC}"

    # Save coverage to text file for CI
    go tool cover -func=coverage.out > coverage.txt

    # Open report in browser if requested
    if [ "$OPEN_REPORT" = true ]; then
        echo -e "\n${YELLOW}🌐 Opening coverage report in browser${NC}"
        case "$(uname -s)" in
            Darwin)
                open coverage.html
                ;;
            Linux)
                if command -v xdg-open > /dev/null; then
                    xdg-open coverage.html
                elif command -v firefox > /dev/null; then
                    firefox coverage.html
                elif command -v google-chrome > /dev/null; then
                    google-chrome coverage.html
                else
                    echo "Please open coverage.html manually in your browser"
                fi
                ;;
            *)
                echo "Please open coverage.html manually in your browser"
                ;;
        esac
    fi

    echo -e "\n${GREEN}📄 Coverage files generated:${NC}"
    echo "  - coverage.html (detailed HTML report)"
    echo "  - coverage.out (raw coverage data)"
    echo "  - coverage.txt (summary for CI)"

else
    # Run tests without coverage
    # Build test flags
    TEST_FLAGS="-race"
    if [ "$VERBOSE" = true ]; then
        TEST_FLAGS="-v $TEST_FLAGS"
    fi
    if [ "$SHORT_TESTS" = true ]; then
        TEST_FLAGS="$TEST_FLAGS -short"
    fi

    go test $TEST_FLAGS ./...

    # Check if tests passed
    if [ $? -ne 0 ]; then
        echo -e "${RED}❌ Tests failed${NC}"
        exit 1
    fi
fi

# Run benchmarks if requested
if [ "$RUN_BENCHMARKS" = true ]; then
    echo -e "\n${YELLOW}⚡ Running benchmarks (quick)${NC}"
    go test -bench=. -benchtime=1s ./... | grep -E "(Benchmark|ok|PASS)"
fi

# Final summary
echo -e "\n${BLUE}===========================================${NC}"
echo -e "${GREEN}✅ All tests passed successfully!${NC}"

if [ "$COVERAGE" = true ]; then
    echo -e "${GREEN}📊 Coverage: ${COVERAGE_PERCENT}${NC}"
fi

echo -e "${BLUE}===========================================${NC}"
echo -e "\n${YELLOW}💡 Usage examples:${NC}"
echo "  ./test-runner.sh           # Run tests only"
echo "  ./test-runner.sh -v        # Verbose output"
echo "  ./test-runner.sh -c        # Generate coverage"
echo "  ./test-runner.sh -s        # Short tests (CI-friendly)"
echo "  ./test-runner.sh -b        # Run benchmarks"
echo "  ./test-runner.sh -v -c -r  # Verbose + coverage + open report"
echo "  ./test-runner.sh -v -c -s  # Verbose + coverage + short tests"
echo "  ./test-runner.sh -v -c -b  # Verbose + coverage + benchmarks"

echo -e "\n${GREEN}🎉 Test run completed!${NC}"