#!/bin/bash

# Test OAuth Proxy
echo "Testing OAuth Proxy..."

# Test 1: OAuth Metadata Discovery
echo "1. Testing OAuth Metadata Discovery..."
curl -s http://localhost:8080/.well-known/oauth-authorization-server | jq .

# Test 2: Protected Resource Metadata
echo -e "\n2. Testing Protected Resource Metadata..."
curl -s http://localhost:8080/.well-known/oauth-protected-resource | jq .

# Test 3: Authorization Request (this will redirect to approval page)
echo -e "\n3. Testing Authorization Request..."
echo "Visit this URL in your browser to start the OAuth flow:"
echo "http://localhost:8080/authorize?response_type=code&client_id=test-client&redirect_uri=http://localhost:3000/callback&scope=openid%20profile%20email&state=test-state"

# Test 4: Token Exchange (requires authorization code from step 3)
echo -e "\n4. Testing Token Exchange (requires authorization code)..."
echo "After getting the authorization code from step 3, run:"
echo "curl -X POST http://localhost:8080/token \\"
echo "  -H 'Content-Type: application/x-www-form-urlencoded' \\"
echo "  -d 'grant_type=authorization_code' \\"
echo "  -d 'client_id=test-client' \\"
echo "  -d 'client_secret=test-secret' \\"
echo "  -d 'code=YOUR_AUTH_CODE' \\"
echo "  -d 'redirect_uri=http://localhost:3000/callback'"

# Test 5: MCP Proxy with Token (requires access token from step 4)
echo -e "\n5. Testing MCP Proxy with Token (requires access token)..."
echo "After getting the access token from step 4, run:"
echo "curl -X GET http://localhost:8080/mcp/test \\"
echo "  -H 'Authorization: Bearer YOUR_ACCESS_TOKEN'"

# Test 6: Token Revocation
echo -e "\n6. Testing Token Revocation (requires access token)..."
echo "To revoke a token, run:"
echo "curl -X POST http://localhost:8080/revoke \\"
echo "  -H 'Content-Type: application/x-www-form-urlencoded' \\"
echo "  -d 'token=YOUR_ACCESS_TOKEN' \\"
echo "  -d 'token_type_hint=access_token' \\"
echo "  -d 'client_id=test-client' \\"
echo "  -d 'client_secret=test-secret'"

echo -e "\nOAuth Proxy test instructions completed!" 