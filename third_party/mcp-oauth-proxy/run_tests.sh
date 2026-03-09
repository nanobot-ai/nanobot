#!/bin/bash

# Test runner script for MCP OAuth Proxy
set -e

echo "🚀 Starting MCP OAuth Proxy Test Suite"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if Go is installed
if ! command -v go &> /dev/null; then
    print_error "Go is not installed. Please install Go 1.21 or later."
    exit 1
fi

print_status "Go version: $(go version)"

# Check if PostgreSQL is running (optional)
if command -v pg_isready &> /dev/null; then
    if pg_isready -q; then
        print_success "PostgreSQL is running"
    else
        print_warning "PostgreSQL is not running. Database tests will be skipped."
        export TEST_DATABASE_DSN=""
    fi
else
    print_warning "PostgreSQL client not found. Database tests will be skipped."
    export TEST_DATABASE_DSN=""
fi

# Install dependencies
print_status "Installing dependencies..."
go mod tidy
go mod download

# Run linter
print_status "Running linter..."
if command -v golangci-lint &> /dev/null; then
    golangci-lint run
    print_success "Linter passed"
else
    print_warning "golangci-lint not found. Skipping linting."
fi

# Run tests with different options
print_status "Running unit tests..."
go test -v -short ./...

print_status "Running integration tests..."
if [ -n "$TEST_DATABASE_DSN" ]; then
    go test -v -run "TestIntegration" ./...
else
    print_warning "Skipping integration tests (no database)"
fi

print_status "Running all tests with race detection..."
go test -v -race ./...

print_status "Running tests with coverage..."
go test -v -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

print_status "Running benchmarks..."
go test -v -bench=. -benchmem ./...

# Show test results
print_status "Test Results:"
echo "=================="

# Count test files
TEST_FILES=$(find . -name "*_test.go" | wc -l)
print_status "Found $TEST_FILES test files"

# Show coverage summary
if [ -f coverage.out ]; then
    COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}')
    print_status "Code coverage: $COVERAGE"
    print_status "Coverage report saved to: coverage.html"
fi

print_success "Test suite completed successfully!"

# Optional: Clean up
if [ "$1" = "--clean" ]; then
    print_status "Cleaning up test artifacts..."
    rm -f coverage.out coverage.html
    print_success "Cleanup completed"
fi

echo ""
print_status "To run specific tests:"
echo "  go test -v ./database/     # Database tests only"
echo "  go test -v ./providers/    # Provider tests only"
echo "  go test -v ./tokens/       # Token tests only"
echo "  go test -v -run TestOAuth  # OAuth endpoint tests only"
echo ""
print_status "To run with database:"
echo "  export TEST_DATABASE_DSN='postgres://user:pass@localhost:5432/oauth_test?sslmode=disable'"
echo "  go test -v ./..." 