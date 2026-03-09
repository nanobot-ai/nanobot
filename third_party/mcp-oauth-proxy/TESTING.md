# MCP OAuth Proxy Test Suite

This document describes the comprehensive test suite for the MCP OAuth Proxy, which includes unit tests, integration tests, and mock OAuth providers for testing external authentication flows.

## Overview

The test suite covers all endpoints and functionality of the OAuth proxy:

- **OAuth 2.1 Endpoints**: Authorization, token, revocation, registration
- **MCP Proxy**: Protected resource access
- **Database Operations**: Client, grant, token, and authorization code management
- **Provider Management**: OAuth provider integration and mocking
- **Token Management**: JWT validation and token lifecycle
- **Rate Limiting**: Request throttling functionality
- **Error Handling**: Various error scenarios and edge cases

## Test Structure

```
├── main_test.go              # Main integration tests for all endpoints
├── database/
│   └── database_test.go      # Database operation tests
├── providers/
│   └── provider_test.go      # Provider management and mock provider tests
├── tokens/
│   └── jwt_test.go          # Token validation and management tests
├── run_tests.sh             # Test runner script
└── TESTING.md               # This documentation
```

## Quick Start

### Prerequisites

1. **Go 1.21+**: Required for running tests
2. **PostgreSQL** (optional): For database integration tests
3. **golangci-lint** (optional): For code linting

### Running Tests

#### Option 1: Use the test runner script (Recommended)

```bash
# Run all tests
./run_tests.sh

# Run tests and clean up artifacts
./run_tests.sh --clean
```

#### Option 2: Run tests manually

```bash
# Install dependencies
go mod tidy

# Run unit tests (fast)
go test -v -short ./...

# Run all tests
go test -v ./...

# Run tests with coverage
go test -v -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Run tests with race detection
go test -v -race ./...

# Run specific test packages
go test -v ./database/
go test -v ./providers/
go test -v ./tokens/
```

## Test Categories

### 1. Unit Tests (`-short` flag)

Fast tests that don't require external dependencies:

- Provider interface implementations
- Token validation logic
- Data structure tests
- Utility function tests

### 2. Integration Tests

Tests that require a database connection:

- Database operations (CRUD)
- OAuth flow integration
- End-to-end authorization flows
- Token lifecycle management

### 3. Mock OAuth Provider Tests

Tests using mock OAuth providers to simulate external authentication:

- Authorization URL generation
- Token exchange simulation
- User info retrieval
- Refresh token handling

## Mock OAuth Provider

The test suite includes a comprehensive mock OAuth provider that implements the full OAuth 2.1 flow:

### Features

- **Authorization URL Generation**: Creates proper OAuth authorization URLs
- **PKCE Support**: Implements Proof Key for Code Exchange
- **Token Exchange**: Simulates authorization code to token exchange
- **User Info**: Returns mock user information
- **Token Refresh**: Handles refresh token requests
- **Error Simulation**: Can simulate various error conditions

### Usage in Tests

```go
// Create mock provider
provider := &MockProvider{name: "test"}

// Test authorization URL
authURL := provider.GetAuthorizationURL("client_id", "redirect_uri", "scope", "state")

// Test token exchange
tokenInfo, err := provider.ExchangeCodeForToken(ctx, "code", "client_id", "secret", "redirect_uri")

// Test user info
userInfo, err := provider.GetUserInfo(ctx, "access_token")
```

## Test Configuration

The test suite uses `test_config.yaml` for configuration:

```yaml
server:
  port: 8080
  host: "localhost"

database:
  dsn: "postgres://test:test@localhost:5432/oauth_test?sslmode=disable"

oauth:
  issuer_url: "https://test.example.com"
  # ... other OAuth settings

providers:
  google:
    client_id: "test_google_client_id"
    client_secret: "test_google_client_secret"
```

## Database Setup for Integration Tests

### Option 1: Local PostgreSQL

```bash
# Create test database
createdb oauth_test

# Set environment variable
export TEST_DATABASE_DSN="postgres://username:password@localhost:5432/oauth_test?sslmode=disable"

# Run integration tests
go test -v -run "TestIntegration" ./...
```

### Option 2: Docker PostgreSQL

```bash
# Start PostgreSQL container
docker run --name oauth-test-db \
  -e POSTGRES_DB=oauth_test \
  -e POSTGRES_USER=test \
  -e POSTGRES_PASSWORD=test \
  -p 5432:5432 \
  -d postgres:15

# Set environment variable
export TEST_DATABASE_DSN="postgres://test:test@localhost:5432/oauth_test?sslmode=disable"

# Run tests
go test -v ./...
```

## Test Coverage

The test suite provides comprehensive coverage of:

### OAuth 2.1 Endpoints

- **Authorization Endpoint** (`/oauth/authorize`)

  - Valid authorization requests
  - PKCE support
  - Invalid request handling
  - Client validation

- **Token Endpoint** (`/oauth/token`)

  - Authorization code grant
  - Refresh token grant
  - PKCE validation
  - Client authentication

- **Revocation Endpoint** (`/oauth/revoke`)

  - Token revocation
  - Client authentication

- **Registration Endpoint** (`/oauth/register`)

  - Dynamic client registration
  - Client metadata validation

- **Metadata Endpoints**
  - OAuth authorization server metadata
  - Protected resource metadata

### MCP Proxy

- **Protected Resource Access** (`/mcp`)
  - Token validation
  - Request proxying
  - Unauthorized access handling

### Database Operations

- **Client Management**

  - Client registration and retrieval
  - Client metadata validation

- **Grant Management**

  - Authorization grant creation and retrieval
  - Grant expiration handling

- **Token Management**

  - Access token storage and retrieval
  - Refresh token handling
  - Token revocation

- **Authorization Code Management**
  - Code storage and validation
  - Code expiration and cleanup

### Provider Integration

- **Provider Registration**

  - Provider manager functionality
  - Concurrent access handling

- **Mock Provider**
  - OAuth flow simulation
  - Error condition testing

### Security Features

- **Rate Limiting**

  - Request throttling
  - Rate limit bypass attempts

- **Token Security**
  - Token validation
  - Expired token handling
  - Revoked token detection

## Running Specific Tests

### OAuth Flow Tests

```bash
# Test complete OAuth authorization flow
go test -v -run "TestIntegrationFlow" ./...

# Test authorization endpoint
go test -v -run "TestAuthorizationEndpoint" ./...

# Test token endpoint
go test -v -run "TestTokenEndpoint" ./...
```

### Database Tests

```bash
# Test all database operations
go test -v ./database/

# Test specific database operations
go test -v -run "TestClientOperations" ./database/
go test -v -run "TestTokenOperations" ./database/
```

### Provider Tests

```bash
# Test provider management
go test -v ./providers/

# Test mock provider functionality
go test -v -run "TestMockProvider" ./providers/
```

### Token Tests

```bash
# Test token validation
go test -v ./tokens/

# Test token edge cases
go test -v -run "TestTokenValidationEdgeCases" ./tokens/
```

## Continuous Integration

The project includes comprehensive CI/CD pipelines using GitHub Actions:

### Workflow Files

- **`.github/workflows/test.yml`** - Comprehensive test suite
- **`.github/workflows/build.yml`** - Docker image building and publishing

### Test Workflow Features

The test workflow (`test.yml`) includes:

1. **Matrix Testing**: Tests against multiple Go versions (1.21, 1.22) and PostgreSQL versions (15, 16)
2. **Comprehensive Testing**:
   - Unit tests (`-short` flag)
   - Integration tests with database
   - Race condition detection
   - Code coverage reporting
3. **Code Quality**:
   - golangci-lint with comprehensive rules
   - Security scanning with govulncheck
4. **Performance Testing**:
   - Benchmark tests
   - Performance regression detection

### Build Workflow Features

The build workflow (`build.yml`) includes:

1. **Multi-platform Support**: Builds for linux/amd64 and linux/arm64
2. **Docker Registry Integration**: Publishes to GitHub Container Registry
3. **Caching**: Uses GitHub Actions cache for faster builds
4. **Tagging**: Automatic version tagging based on git tags and branches

### Local Development with Make

Use the provided Makefile for local development:

```bash
# Run all tests locally
make test

# Run only short tests
make test-short

# Run integration tests (requires PostgreSQL)
make test-integration

# Run tests with race detection
make test-race

# Run tests with coverage
make test-coverage

# Run linter
make lint

# Build the application
make build

# Setup test database
make setup-test-db

# Clean up test database
make clean-test-db

# Run all CI checks locally
make ci
```

### CI Configuration

The project uses `.golangci.yml` for consistent linting rules:

- **Code Style**: gofmt, goimports, govet
- **Security**: gosec, govulncheck
- **Code Quality**: staticcheck, gosimple, errcheck
- **Performance**: gocyclo, dupl, goconst
- **Documentation**: godot, goheader

### Coverage Reporting

CI automatically generates and uploads coverage reports:

- **Codecov Integration**: Automatic coverage upload
- **HTML Reports**: Downloadable coverage reports as artifacts
- **Coverage Thresholds**: Enforced minimum coverage levels

### Security Scanning

The CI pipeline includes security checks:

- **Dependency Scanning**: Nancy for vulnerability detection
- **Code Scanning**: govulncheck for Go-specific vulnerabilities
- **Container Scanning**: Built into Docker build process

## Troubleshooting

### Common Issues

1. **Database Connection Errors**

   - Ensure PostgreSQL is running
   - Check connection string format
   - Verify database exists and is accessible

2. **Test Timeouts**

   - Increase test timeout: `go test -v -timeout 5m ./...`
   - Check for slow database queries

3. **Race Conditions**

   - Run with race detection: `go test -v -race ./...`
   - Review concurrent access patterns

4. **Coverage Issues**
   - Ensure all code paths are tested
   - Check for conditional compilation tags

### Debug Mode

Run tests with verbose output and debug information:

```bash
# Enable debug logging
export DEBUG=true
go test -v ./...

# Run specific test with debug
go test -v -run "TestSpecificTest" ./...
```

## Contributing

When adding new tests:

1. Follow the existing test structure
2. Use descriptive test names
3. Include both positive and negative test cases
4. Add appropriate assertions and error checking
5. Update this documentation if needed

### Test Naming Convention

- `TestFunctionName` - Unit tests for specific functions
- `TestComponentName` - Integration tests for components
- `TestFeatureName` - Feature-specific tests
- `TestErrorCondition` - Error handling tests

### Mock Usage

- Use mocks for external dependencies
- Keep mocks simple and predictable
- Document mock behavior
- Test both success and failure scenarios
