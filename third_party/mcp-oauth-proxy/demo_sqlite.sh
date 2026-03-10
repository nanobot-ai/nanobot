#!/bin/bash

# Demo script for MCP OAuth Proxy with SQLite
set -e

echo "🚀 MCP OAuth Proxy SQLite Demo"
echo "================================"

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.21 or later."
    exit 1
fi

print_info "Go version: $(go version)"

# Build the application
print_info "Building the application..."
go build -o bin/mcp-oauth-proxy .

print_success "Application built successfully"

# Set up environment variables for SQLite demo
export SCOPES_SUPPORTED="openid,profile,email"
export OAUTH_CLIENT_ID="demo_client_id"
export OAUTH_CLIENT_SECRET="demo_client_secret"
export OAUTH_AUTHORIZE_URL="https://demo.example.com/oauth/authorize"
export MCP_SERVER_URL="http://localhost:3000"

# Clear DATABASE_DSN to use SQLite
unset DATABASE_DSN

print_info "Environment variables set:"
print_info "  SCOPES_SUPPORTED: $SCOPES_SUPPORTED"
print_info "  OAUTH_CLIENT_ID: $OAUTH_CLIENT_ID"
print_info "  OAUTH_CLIENT_SECRET: $OAUTH_CLIENT_SECRET"
print_info "  OAUTH_AUTHORIZE_URL: $OAUTH_AUTHORIZE_URL"
print_info "  MCP_SERVER_URL: $MCP_SERVER_URL"
print_info "  DATABASE_DSN: (not set - will use SQLite)"

print_info "Starting OAuth proxy with SQLite database..."
print_info "The database will be created at: data/oauth_proxy.db"

echo ""
print_warning "Press Ctrl+C to stop the server"
echo ""

# Run the application
./bin/mcp-oauth-proxy 