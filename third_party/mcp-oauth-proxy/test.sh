#!/bin/bash

# Test script for OAuth Proxy
# This script tests the OAuth proxy endpoints

BASE_URL="http://localhost:8080"

echo "Testing OAuth Proxy..."

# Test 1: OAuth Metadata
echo "1. Testing OAuth Metadata..."
curl -s "$BASE_URL/.well-known/oauth-authorization-server" | jq .
echo ""

# Test 2: Protected Resource Metadata
echo "2. Testing Protected Resource Metadata..."
curl -s "$BASE_URL/.well-known/oauth-protected-resource" | jq .
echo ""

# Test 3: Authorization Endpoint (should redirect to Google)
echo "3. Testing Authorization Endpoint..."
echo "This will redirect to Google OAuth. Check the URL:"
AUTH_URL="$BASE_URL/authorize?response_type=code&client_id=test-client&redirect_uri=http://localhost:3000/callback&scope=openid%20profile%20email&state=test-state&code_challenge=test-challenge&code_challenge_method=S256&provider=google"
echo "$AUTH_URL"
echo ""

# Test 4: Token Endpoint (will fail without valid code)
echo "4. Testing Token Endpoint (should fail without valid code)..."
curl -s -X POST "$BASE_URL/token" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=authorization_code&client_id=test-client&client_secret=test-secret&code=invalid-code&code_verifier=test-verifier&redirect_uri=http://localhost:3000/callback&provider=google" | jq .
echo ""

# Test 5: MCP Proxy (will fail without valid token)
echo "5. Testing MCP Proxy (should fail without valid token)..."
curl -s "$BASE_URL/mcp/test-endpoint" | jq .
echo ""

# Test 6: MCP Proxy with invalid token
echo "6. Testing MCP Proxy with invalid token..."
curl -s -H "Authorization: Bearer invalid-token" "$BASE_URL/mcp/test-endpoint" | jq .
echo ""

echo "Tests completed!"
echo ""
echo "OAuth Flow:"
echo "1. Client → /authorize → Redirects to Google OAuth"
echo "2. User authenticates with Google"
echo "3. Google redirects back to client with authorization code"
echo "4. Client → /token with Google's code → Gets our JWT token"
echo "5. Client uses our JWT token to access MCP proxy"
echo ""
echo "To test the full flow:"
echo "1. Update config.yaml with your Google OAuth credentials"
echo "2. Start the proxy server: go run main.go"
echo "3. Visit the authorization URL to start the OAuth flow"
echo "4. After getting authorization code from Google, exchange it for our JWT token"
echo "5. Use our JWT token to test the MCP proxy endpoint"