package db

import (
	"crypto/rand"
	"encoding/base64"
	"os"
	"testing"
	"time"

	"github.com/obot-platform/mcp-oauth-proxy/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDatabaseOperations tests all database operations
func TestDatabaseOperations(t *testing.T) {
	// Skip if no test database is available
	if testing.Short() {
		t.Skip("Skipping database tests in short mode")
	}

	// Create test database
	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("Skipping database tests: TEST_DATABASE_DSN is not set")
	}

	db, err := New(dsn)
	if err != nil {
		t.Skipf("Skipping database tests: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("Error closing database: %v", err)
		}
	}()

	t.Run("TestClientOperations", func(t *testing.T) {
		testClientOperations(t, db)
	})

	t.Run("TestGrantOperations", func(t *testing.T) {
		testGrantOperations(t, db)
	})

	t.Run("TestTokenOperations", func(t *testing.T) {
		testTokenOperations(t, db)
	})

	t.Run("TestAuthCodeOperations", func(t *testing.T) {
		testAuthCodeOperations(t, db)
	})

	t.Run("TestCleanupOperations", func(t *testing.T) {
		testCleanupOperations(t, db)
	})

	t.Run("TestRefreshTokenExpiration", func(t *testing.T) {
		testRefreshTokenExpiration(t, db)
	})

}

func testClientOperations(t *testing.T, db *Store) {

	clientID, err := generateRandomString(16)
	require.NoError(t, err)

	clientSecret, err := generateRandomString(16)
	require.NoError(t, err)

	// Test storing client
	client := &types.ClientInfo{
		ClientID:                clientID,
		ClientSecret:            clientSecret,
		RedirectUris:            []string{"https://test.example.com/callback"},
		ClientName:              "Test Client DB",
		LogoURI:                 "https://test.example.com/logo.png",
		ClientURI:               "https://test.example.com",
		PolicyURI:               "https://test.example.com/policy",
		TosURI:                  "https://test.example.com/tos",
		JwksURI:                 "https://test.example.com/jwks",
		Contacts:                []string{"admin@test.example.com"},
		GrantTypes:              []string{"authorization_code", "refresh_token"},
		ResponseTypes:           []string{"code"},
		RegistrationDate:        time.Now().Unix(),
		TokenEndpointAuthMethod: "client_secret_basic",
	}

	err = db.StoreClient(client)
	require.NoError(t, err)

	// Test retrieving client
	retrievedClient, err := db.GetClient(clientID)
	require.NoError(t, err)
	assert.Equal(t, client.ClientID, retrievedClient.ClientID)
	assert.Equal(t, client.ClientSecret, retrievedClient.ClientSecret)
	assert.Equal(t, client.RedirectUris, retrievedClient.RedirectUris)
	assert.Equal(t, client.ClientName, retrievedClient.ClientName)
	assert.Equal(t, client.LogoURI, retrievedClient.LogoURI)
	assert.Equal(t, client.ClientURI, retrievedClient.ClientURI)
	assert.Equal(t, client.PolicyURI, retrievedClient.PolicyURI)
	assert.Equal(t, client.TosURI, retrievedClient.TosURI)
	assert.Equal(t, client.JwksURI, retrievedClient.JwksURI)
	assert.Equal(t, client.Contacts, retrievedClient.Contacts)
	assert.Equal(t, client.GrantTypes, retrievedClient.GrantTypes)
	assert.Equal(t, client.ResponseTypes, retrievedClient.ResponseTypes)
	assert.Equal(t, client.RegistrationDate, retrievedClient.RegistrationDate)
	assert.Equal(t, client.TokenEndpointAuthMethod, retrievedClient.TokenEndpointAuthMethod)

	// Test retrieving non-existent client
	_, err = db.GetClient("non_existent_client")
	assert.Error(t, err)
}

func testGrantOperations(t *testing.T, db *Store) {
	clientID, err := generateRandomString(16)
	require.NoError(t, err)

	userID, err := generateRandomString(16)
	require.NoError(t, err)

	grantID, err := generateRandomString(16)
	require.NoError(t, err)

	// Test storing grant
	grant := &types.Grant{
		ID:                  grantID,
		ClientID:            clientID,
		UserID:              userID,
		Scope:               []string{"read", "write", "admin"},
		Metadata:            map[string]any{"provider": "test", "ip": "127.0.0.1"},
		Props:               map[string]any{"email": "test@example.com", "name": "Test User"},
		CreatedAt:           time.Now().Unix(),
		ExpiresAt:           time.Now().Add(10 * time.Minute).Unix(),
		CodeChallenge:       "test_challenge",
		CodeChallengeMethod: "S256",
	}

	err = db.StoreGrant(grant)
	require.NoError(t, err)

	// Test retrieving grant
	retrievedGrant, err := db.GetGrant(grantID, userID)
	require.NoError(t, err)
	assert.Equal(t, grant.ID, retrievedGrant.ID)
	assert.Equal(t, grant.ClientID, retrievedGrant.ClientID)
	assert.Equal(t, grant.UserID, retrievedGrant.UserID)
	assert.Equal(t, grant.Scope, retrievedGrant.Scope)
	assert.Equal(t, grant.Metadata, retrievedGrant.Metadata)
	assert.Equal(t, grant.Props, retrievedGrant.Props)
	assert.Equal(t, grant.CreatedAt, retrievedGrant.CreatedAt)
	assert.Equal(t, grant.ExpiresAt, retrievedGrant.ExpiresAt)
	assert.Equal(t, grant.CodeChallenge, retrievedGrant.CodeChallenge)
	assert.Equal(t, grant.CodeChallengeMethod, retrievedGrant.CodeChallengeMethod)

	// Test retrieving non-existent grant
	_, err = db.GetGrant("non_existent_grant", "test_user_123")
	assert.Error(t, err)
}

func generateRandomString(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}
func testTokenOperations(t *testing.T, db *Store) {
	grantID, err := generateRandomString(16)
	require.NoError(t, err)

	clientID, err := generateRandomString(16)
	require.NoError(t, err)

	userID, err := generateRandomString(16)
	require.NoError(t, err)

	grant := &types.Grant{
		ID:       grantID,
		ClientID: clientID,
		UserID:   userID,
		Scope:    []string{"read", "write", "admin"},
		Metadata: map[string]any{"provider": "test", "ip": "127.0.0.1"},
	}

	err = db.StoreGrant(grant)
	require.NoError(t, err)

	refreshTokenData, err := generateRandomString(16)
	require.NoError(t, err)

	accessTokenData, err := generateRandomString(16)
	require.NoError(t, err)

	// Test storing token
	tokenData := &types.TokenData{
		AccessToken:           accessTokenData,
		RefreshToken:          refreshTokenData,
		ClientID:              clientID,
		UserID:                userID,
		GrantID:               grantID,
		Scope:                 "read write admin",
		ExpiresAt:             time.Now().Add(time.Hour),
		RefreshTokenExpiresAt: time.Now().Add(30 * 24 * time.Hour), // 30 days
		CreatedAt:             time.Now(),
		Revoked:               false,
	}

	err = db.StoreToken(tokenData)
	require.NoError(t, err)

	// Test retrieving token
	retrievedToken, err := db.GetToken(accessTokenData)
	require.NoError(t, err)
	assert.Equal(t, hashToken(tokenData.AccessToken), retrievedToken.AccessToken)
	assert.Equal(t, hashToken(tokenData.RefreshToken), retrievedToken.RefreshToken)
	assert.Equal(t, tokenData.ClientID, retrievedToken.ClientID)
	assert.Equal(t, tokenData.UserID, retrievedToken.UserID)
	assert.Equal(t, tokenData.GrantID, retrievedToken.GrantID)
	assert.Equal(t, tokenData.Scope, retrievedToken.Scope)
	assert.False(t, retrievedToken.Revoked)
	assert.True(t, retrievedToken.RefreshTokenExpiresAt.After(time.Now().Add(29*24*time.Hour))) // Should be ~30 days

	// Test retrieving token by refresh token
	refreshToken, err := db.GetTokenByRefreshToken(refreshTokenData)
	require.NoError(t, err)
	assert.Equal(t, hashToken(tokenData.AccessToken), refreshToken.AccessToken)
	assert.Equal(t, tokenData.RefreshToken, refreshToken.RefreshToken)
	assert.True(t, refreshToken.RefreshTokenExpiresAt.After(time.Now().Add(29*24*time.Hour))) // Should be ~30 days

	// Test revoking token
	err = db.RevokeToken(accessTokenData)
	require.NoError(t, err)

	revokedToken, err := db.GetToken(accessTokenData)
	require.NoError(t, err)
	assert.True(t, revokedToken.Revoked)
	assert.NotNil(t, revokedToken.RevokedAt)

	// Test retrieving non-existent token
	_, err = db.GetToken("non_existent_token")
	assert.Error(t, err)
}

func testAuthCodeOperations(t *testing.T, db *Store) {
	// Test storing authorization code
	code, err := generateRandomString(16)
	require.NoError(t, err)

	grantID, err := generateRandomString(16)
	require.NoError(t, err)

	userID, err := generateRandomString(16)
	require.NoError(t, err)

	grant := &types.Grant{
		ID:       grantID,
		ClientID: "test_client_db",
		UserID:   userID,
		Scope:    []string{"read", "write", "admin"},
		Metadata: map[string]any{"provider": "test", "ip": "127.0.0.1"},
	}

	err = db.StoreGrant(grant)
	require.NoError(t, err)

	err = db.StoreAuthCode(code, grantID, userID)
	require.NoError(t, err)

	// Test validating authorization code
	retrievedGrantID, retrievedUserID, err := db.ValidateAuthCode(code)
	require.NoError(t, err)
	assert.Equal(t, grantID, retrievedGrantID)
	assert.Equal(t, userID, retrievedUserID)

	// Test deleting authorization code
	err = db.DeleteAuthCode(code)
	require.NoError(t, err)

	// Test validating deleted authorization code
	_, _, err = db.ValidateAuthCode(code)
	assert.Error(t, err)
}

func testCleanupOperations(t *testing.T, db *Store) {
	grantID, err := generateRandomString(16)
	require.NoError(t, err)

	userID, err := generateRandomString(16)
	require.NoError(t, err)

	grant := &types.Grant{
		ID:       grantID,
		ClientID: "test_client_db",
		UserID:   userID,
		Scope:    []string{"read", "write", "admin"},
		Metadata: map[string]any{"provider": "test", "ip": "127.0.0.1"},
	}

	err = db.StoreGrant(grant)
	require.NoError(t, err)

	accessTokenData, err := generateRandomString(16)
	require.NoError(t, err)

	refreshTokenData, err := generateRandomString(16)
	require.NoError(t, err)

	// Create expired tokens
	expiredToken := &types.TokenData{
		AccessToken:           accessTokenData,
		RefreshToken:          refreshTokenData,
		ClientID:              "test_client_db",
		UserID:                userID,
		GrantID:               grantID,
		Scope:                 "read write",
		ExpiresAt:             time.Now().Add(-time.Hour), // Expired
		RefreshTokenExpiresAt: time.Now().Add(-time.Hour), // Also expired
		CreatedAt:             time.Now().Add(-2 * time.Hour),
		Revoked:               false,
	}

	err = db.StoreToken(expiredToken)
	require.NoError(t, err)

	// Test cleanup
	err = db.CleanupExpiredTokens()
	require.NoError(t, err)

	// Verify expired token is cleaned up (both access and refresh tokens expired)
	_, err = db.GetToken(accessTokenData)
	assert.Error(t, err)
}

// TestCodeChallengeGeneration tests PKCE code challenge generation
func TestCodeChallengeGeneration(t *testing.T) {
	codeVerifier := "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	expectedChallenge := "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM"

	challenge := GenerateCodeChallenge(codeVerifier)
	assert.Equal(t, expectedChallenge, challenge)

	// Test with random code verifier
	randomVerifier := "random_code_verifier_123"
	randomChallenge := GenerateCodeChallenge(randomVerifier)
	assert.NotEmpty(t, randomChallenge)
	assert.NotEqual(t, randomVerifier, randomChallenge)
}

// TestTokenHashing tests token hashing functionality
func TestTokenHashing(t *testing.T) {
	token := "test_token_123"
	hash1 := hashToken(token)
	hash2 := hashToken(token)

	// Same token should produce same hash
	assert.Equal(t, hash1, hash2)

	// Different tokens should produce different hashes
	differentToken := "different_token_456"
	hash3 := hashToken(differentToken)
	assert.NotEqual(t, hash1, hash3)

	// Hash should be consistent
	assert.NotEmpty(t, hash1)
	assert.Len(t, hash1, 44) // Base64 encoded SHA256 hash length
}

func testRefreshTokenExpiration(t *testing.T, db *Store) {
	grantID, err := generateRandomString(16)
	require.NoError(t, err)

	clientID, err := generateRandomString(16)
	require.NoError(t, err)

	grant := &types.Grant{
		ID:        grantID,
		ClientID:  clientID,
		UserID:    "test_user_123",
		Scope:     []string{"read", "write", "admin"},
		Metadata:  map[string]any{"provider": "test", "ip": "127.0.0.1"},
		ExpiresAt: time.Now().Add(time.Hour).Unix(),
	}

	err = db.StoreGrant(grant)
	require.NoError(t, err)

	// Test 1: Store token with default refresh token expiration (30 days)
	accessToken1, err := generateRandomString(16)
	require.NoError(t, err)
	refreshToken1, err := generateRandomString(16)
	require.NoError(t, err)

	tokenData1 := &types.TokenData{
		AccessToken:  accessToken1,
		RefreshToken: refreshToken1,
		ClientID:     clientID,
		UserID:       "test_user_123",
		GrantID:      grantID,
		Scope:        "read write admin",
		ExpiresAt:    time.Now().Add(time.Hour),
		// RefreshTokenExpiresAt will be set automatically to 30 days
		CreatedAt: time.Now(),
		Revoked:   false,
	}

	err = db.StoreToken(tokenData1)
	require.NoError(t, err)

	// Verify refresh token expiration was set to 30 days
	retrievedToken1, err := db.GetToken(accessToken1)
	require.NoError(t, err)
	assert.True(t, retrievedToken1.RefreshTokenExpiresAt.After(time.Now().Add(29*24*time.Hour)))
	assert.True(t, retrievedToken1.RefreshTokenExpiresAt.Before(time.Now().Add(31*24*time.Hour)))

	// Test 2: Store token with custom refresh token expiration
	accessToken2, err := generateRandomString(16)
	require.NoError(t, err)
	refreshToken2, err := generateRandomString(16)
	require.NoError(t, err)

	customExpiration := time.Now().Add(7 * 24 * time.Hour) // 7 days
	tokenData2 := &types.TokenData{
		AccessToken:           accessToken2,
		RefreshToken:          refreshToken2,
		ClientID:              clientID,
		UserID:                "test_user_123",
		GrantID:               grantID,
		Scope:                 "read write admin",
		ExpiresAt:             time.Now().Add(time.Hour),
		RefreshTokenExpiresAt: customExpiration,
		CreatedAt:             time.Now(),
		Revoked:               false,
	}

	err = db.StoreToken(tokenData2)
	require.NoError(t, err)

	// Verify custom expiration was preserved
	retrievedToken2, err := db.GetToken(accessToken2)
	require.NoError(t, err)
	// Use tolerance-based comparison instead of exact equality due to database precision differences
	timeDiff := retrievedToken2.RefreshTokenExpiresAt.Sub(customExpiration)
	assert.True(t, timeDiff >= -time.Second && timeDiff <= time.Second,
		"Refresh token expiration should be within 1 second of expected time, got difference of %v", timeDiff)

	// Test 3: Test refresh token expiration check
	expired, err := db.IsRefreshTokenExpired(refreshToken1)
	require.NoError(t, err)
	assert.False(t, expired)

	// Test 4: Test cleanup with mixed expiration scenarios
	// Create token with expired access token but valid refresh token
	accessToken3, err := generateRandomString(16)
	require.NoError(t, err)
	refreshToken3, err := generateRandomString(16)
	require.NoError(t, err)

	tokenData3 := &types.TokenData{
		AccessToken:           accessToken3,
		RefreshToken:          refreshToken3,
		ClientID:              clientID,
		UserID:                "test_user_123",
		GrantID:               grantID,
		Scope:                 "read write admin",
		ExpiresAt:             time.Now().Add(-time.Hour), // Expired access token
		RefreshTokenExpiresAt: time.Now().Add(time.Hour),  // Valid refresh token
		CreatedAt:             time.Now().Add(-2 * time.Hour),
		Revoked:               false,
	}

	err = db.StoreToken(tokenData3)
	require.NoError(t, err)

	// Create token with both expired access and refresh tokens
	accessToken4, err := generateRandomString(16)
	require.NoError(t, err)
	refreshToken4, err := generateRandomString(16)
	require.NoError(t, err)

	tokenData4 := &types.TokenData{
		AccessToken:           accessToken4,
		RefreshToken:          refreshToken4,
		ClientID:              clientID,
		UserID:                "test_user_123",
		GrantID:               grantID,
		Scope:                 "read write admin",
		ExpiresAt:             time.Now().Add(-time.Hour), // Expired access token
		RefreshTokenExpiresAt: time.Now().Add(-time.Hour), // Expired refresh token
		CreatedAt:             time.Now().Add(-2 * time.Hour),
		Revoked:               false,
	}

	err = db.StoreToken(tokenData4)
	require.NoError(t, err)

	// Run cleanup
	err = db.CleanupExpiredTokens()
	require.NoError(t, err)

	// Verify that only the token with both expired access and refresh tokens was cleaned up
	_, err = db.GetToken(accessToken3) // Should still exist (valid refresh token)
	assert.NoError(t, err)

	_, err = db.GetToken(accessToken4) // Should be cleaned up (both expired)
	assert.Error(t, err)

	// Test 5: Test that expired refresh tokens cannot be used
	// Create a token with expired refresh token
	accessToken5, err := generateRandomString(16)
	require.NoError(t, err)
	refreshToken5, err := generateRandomString(16)
	require.NoError(t, err)

	tokenData5 := &types.TokenData{
		AccessToken:           accessToken5,
		RefreshToken:          refreshToken5,
		ClientID:              clientID,
		UserID:                "test_user_123",
		GrantID:               grantID,
		Scope:                 "read write admin",
		ExpiresAt:             time.Now().Add(time.Hour),
		RefreshTokenExpiresAt: time.Now().Add(-time.Hour), // Expired refresh token
		CreatedAt:             time.Now().Add(-2 * time.Hour),
		Revoked:               false,
	}

	err = db.StoreToken(tokenData5)
	require.NoError(t, err)

	// Try to get token by expired refresh token
	_, err = db.GetTokenByRefreshToken(refreshToken5)
	assert.Error(t, err, "Should return error when trying to get token by expired refresh token")

	// Verify the token still exists when accessed by access token
	retrievedToken5, err := db.GetToken(accessToken5)
	require.NoError(t, err)
	assert.Equal(t, hashToken(accessToken5), retrievedToken5.AccessToken)
}
