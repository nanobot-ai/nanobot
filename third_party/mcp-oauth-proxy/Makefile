.PHONY: help test test-short test-integration test-race test-coverage lint build clean docker-build docker-run demo-sqlite test-sqlite

# Default target
help:
	@echo "Available commands:"
	@echo "  test           - Run all tests"
	@echo "  test-short     - Run only short tests"
	@echo "  test-integration - Run integration tests"
	@echo "  test-race      - Run tests with race detection"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  test-sqlite    - Run SQLite database tests"
	@echo "  lint           - Run linter"
	@echo "  build          - Build the application"
	@echo "  clean          - Clean build artifacts"
	@echo "  demo-sqlite    - Run SQLite demo"
	@echo "  docker-build   - Build Docker image"
	@echo "  docker-run     - Run Docker container"

# Test targets
test:
	go test -v ./...

test-short:
	go test -v -short ./...

test-integration:
	TEST_DATABASE_DSN="postgres://test:test@localhost:5432/oauth_test?sslmode=disable" go test -v -run "TestIntegration" ./...

test-race:
	go test -v -race ./...

test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Linting
lint:
	golangci-lint run --tests=false

# Build targets
build:
	go build -o bin/mcp-oauth-proxy .

clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

# Docker targets
docker-build:
	docker build -t mcp-oauth-proxy .

docker-run:
	docker run -p 8080:8080 --env-file .env mcp-oauth-proxy

# Development helpers
deps:
	go mod download
	go mod tidy

fmt:
	go fmt ./...

vet:
	go vet ./...

# Database setup for testing
setup-test-db:
	docker run --name postgres-test -e POSTGRES_DB=oauth_test -e POSTGRES_USER=test -e POSTGRES_PASSWORD=test -p 5432:5432 -d postgres:15

clean-test-db:
	docker stop postgres-test || true
	docker rm postgres-test || true

# SQLite targets
test-sqlite:
	go test -v ./database -run TestSQLiteDatabase
	go test -v ./database
	DATABASE_DSN="" go test -v ./database -run TestSQLiteDatabase
	DATABASE_DSN="" go test -v -run "TestIntegration" ./...

demo-sqlite:
	./demo_sqlite.sh

# Run all checks (for CI)
ci: deps fmt vet lint test-short test-race test-coverage

# Run all checks including SQLite (for CI)
ci-full: deps fmt vet lint test-short test-race test-coverage test-sqlite 