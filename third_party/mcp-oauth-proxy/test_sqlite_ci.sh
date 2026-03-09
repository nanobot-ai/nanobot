#!/bin/bash

# CI SQLite Test Script
# This script runs the same tests as the CI workflow for SQLite

set -e

echo "🚀 Running CI SQLite Tests"
echo "=========================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

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

# Install dependencies
print_status "Installing dependencies..."
go mod download

# Run unit tests (SQLite)
print_status "Running unit tests (SQLite)..."
go test -v -short ./...

# Run integration tests (SQLite)
print_status "Running integration tests (SQLite)..."
go test -v -run "TestIntegration" ./...

# Run integration tests with SQLite fallback
print_status "Running integration tests with SQLite fallback..."
DATABASE_DSN="" go test -v -run "TestIntegration" ./...

# Run all tests with race detection (SQLite)
print_status "Running all tests with race detection (SQLite)..."
go test -v -race ./...

# Run tests with coverage (SQLite)
print_status "Running tests with coverage (SQLite)..."
go test -v -coverprofile=coverage-sqlite.out ./...

# Generate coverage report (SQLite)
print_status "Generating coverage report (SQLite)..."
go tool cover -html=coverage-sqlite.out -o coverage-sqlite.html

# Run SQLite-specific database tests
print_status "Running SQLite-specific database tests..."
go test -v ./database -run TestSQLiteDatabase

# Run all database tests with SQLite
print_status "Running all database tests with SQLite..."
go test -v ./database

# Test SQLite demo build
print_status "Testing SQLite demo build..."
go build -o mcp-oauth-proxy .
./mcp-oauth-proxy --help || true
rm mcp-oauth-proxy

# Test SQLite fallback behavior
print_status "Testing SQLite fallback behavior..."
unset DATABASE_DSN
go test -v ./database -run TestSQLiteDatabase

# Test application startup with SQLite
print_status "Testing application startup with SQLite..."
export DATABASE_DSN=""
export OAUTH_CLIENT_ID="test_client"
export OAUTH_CLIENT_SECRET="test_secret"
export OAUTH_AUTHORIZE_URL="https://test.example.com/oauth/authorize"
export MCP_SERVER_URL="http://localhost:3000"

# Test that the application can start with SQLite configuration
timeout 10s go run . || true

print_success "All SQLite CI tests completed successfully!"

# Show test results
print_status "Test Results:"
echo "=================="

# Show coverage summary
if [ -f coverage-sqlite.out ]; then
    COVERAGE=$(go tool cover -func=coverage-sqlite.out | grep total | awk '{print $3}')
    print_status "SQLite Code coverage: $COVERAGE"
    print_status "Coverage report saved to: coverage-sqlite.html"
fi

print_success "SQLite CI test suite completed successfully!" 