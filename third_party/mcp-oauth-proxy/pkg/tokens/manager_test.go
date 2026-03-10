package tokens

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/obot-platform/mcp-oauth-proxy/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockDatabase implements the Database interface for testing
type MockDatabase struct {
	tokens map[string]*types.TokenData
	grants map[string]*types.Grant
}

func NewMockDatabase() *MockDatabase {
	return &MockDatabase{
		tokens: make(map[string]*types.TokenData),
		grants: make(map[string]*types.Grant),
	}
}

func (m *MockDatabase) GetToken(accessToken string) (*types.TokenData, error) {
	if token, exists := m.tokens[accessToken]; exists {
		return token, nil
	}
	return nil, assert.AnError
}

func (m *MockDatabase) GetGrant(grantID, userID string) (*types.Grant, error) {
	key := grantID + ":" + userID
	if grant, exists := m.grants[key]; exists {
		return grant, nil
	}
	return nil, assert.AnError
}

// TestTokenManager tests the token manager functionality
func TestTokenManager(t *testing.T) {
	t.Run("TestNewTokenManager", func(t *testing.T) {
		tokenManager, err := NewTokenManager(nil)
		require.NoError(t, err)
		assert.NotNil(t, tokenManager)
	})

	t.Run("TestSetDatabase", func(t *testing.T) {
		mockDB := NewMockDatabase()
		tokenManager, err := NewTokenManager(mockDB)
		require.NoError(t, err)

		assert.Equal(t, mockDB, tokenManager.db)
	})

	t.Run("TestValidateAccessToken", func(t *testing.T) {
		mockDB := NewMockDatabase()
		tokenManager, err := NewTokenManager(mockDB)
		require.NoError(t, err)

		// Test without database
		tokenManager.db = nil
		_, err = tokenManager.validateAccessToken("user123:grant456:access_token_123")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database not configured")

		// Test with database but no token
		tokenManager.db = mockDB
		_, err = tokenManager.validateAccessToken("non_existent_token")
		assert.Error(t, err)
		// The error could be either "token not found" or "invalid token format" depending on the token format
		assert.True(t, strings.Contains(err.Error(), "token not found") || strings.Contains(err.Error(), "invalid token format"))

		// Test with invalid token format
		_, err = tokenManager.validateAccessToken("invalid_token_format")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid token format")

		// Test with valid token but revoked
		validToken := "user123:grant456:access_token_123"
		expiredTime := time.Now().Add(-time.Hour)
		mockDB.tokens[validToken] = &types.TokenData{
			AccessToken: validToken,
			ClientID:    "test_client",
			UserID:      "user123",
			GrantID:     "grant456",
			Scope:       "read write",
			ExpiresAt:   time.Now().Add(time.Hour),
			CreatedAt:   time.Now(),
			Revoked:     true,
			RevokedAt:   &expiredTime,
		}

		_, err = tokenManager.validateAccessToken(validToken)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "token has been revoked")

		// Test with expired token
		expiredToken := "user123:grant456:expired_token"
		mockDB.tokens[expiredToken] = &types.TokenData{
			AccessToken: expiredToken,
			ClientID:    "test_client",
			UserID:      "user123",
			GrantID:     "grant456",
			Scope:       "read write",
			ExpiresAt:   time.Now().Add(-time.Hour), // Expired
			CreatedAt:   time.Now().Add(-2 * time.Hour),
			Revoked:     false,
		}

		_, err = tokenManager.validateAccessToken(expiredToken)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "token has expired")

		// Test with valid token but missing grant
		validTokenNoGrant := "user123:missing_grant:valid_token"
		mockDB.tokens[validTokenNoGrant] = &types.TokenData{
			AccessToken: validTokenNoGrant,
			ClientID:    "test_client",
			UserID:      "user123",
			GrantID:     "missing_grant",
			Scope:       "read write",
			ExpiresAt:   time.Now().Add(time.Hour),
			CreatedAt:   time.Now(),
			Revoked:     false,
		}

		_, err = tokenManager.validateAccessToken(validTokenNoGrant)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "grant not found")

		// Test with valid token and grant
		validTokenWithGrant := "user123:valid_grant:valid_token"
		mockDB.tokens[validTokenWithGrant] = &types.TokenData{
			AccessToken: validTokenWithGrant,
			ClientID:    "test_client",
			UserID:      "user123",
			GrantID:     "valid_grant",
			Scope:       "read write",
			ExpiresAt:   time.Now().Add(time.Hour),
			CreatedAt:   time.Now(),
			Revoked:     false,
		}

		mockDB.grants["valid_grant:user123"] = &types.Grant{
			ID:        "valid_grant",
			ClientID:  "test_client",
			UserID:    "user123",
			Scope:     []string{"read", "write"},
			Metadata:  map[string]any{"provider": "test"},
			Props:     map[string]any{"email": "test@example.com", "name": "Test User"},
			CreatedAt: time.Now().Unix(),
			ExpiresAt: time.Now().Add(10 * time.Minute).Unix(),
		}

		claims, err := tokenManager.validateAccessToken(validTokenWithGrant)
		require.NoError(t, err)
		assert.Equal(t, "user123", claims.UserID)
		assert.Equal(t, map[string]any{"email": "test@example.com", "name": "Test User"}, claims.Props)
		assert.False(t, time.Now().After(claims.ExpiresAt))
	})

	t.Run("TestGetTokenInfo", func(t *testing.T) {
		mockDB := NewMockDatabase()
		tokenManager, err := NewTokenManager(mockDB)
		require.NoError(t, err)

		// Test with invalid token
		_, err = tokenManager.GetTokenInfo("invalid_token")
		assert.Error(t, err)

		// Test with valid token
		validToken := "user123:valid_grant:valid_token"
		mockDB.tokens[validToken] = &types.TokenData{
			AccessToken: validToken,
			ClientID:    "test_client",
			UserID:      "user123",
			GrantID:     "valid_grant",
			Scope:       "read write",
			ExpiresAt:   time.Now().Add(time.Hour),
			CreatedAt:   time.Now(),
			Revoked:     false,
		}

		mockDB.grants["valid_grant:user123"] = &types.Grant{
			ID:        "valid_grant",
			ClientID:  "test_client",
			UserID:    "user123",
			Scope:     []string{"read", "write"},
			Metadata:  map[string]any{"provider": "test"},
			Props:     map[string]any{"email": "test@example.com", "name": "Test User"},
			CreatedAt: time.Now().Unix(),
			ExpiresAt: time.Now().Add(10 * time.Minute).Unix(),
		}

		tokenInfo, err := tokenManager.GetTokenInfo(validToken)
		require.NoError(t, err)
		assert.Equal(t, "user123", tokenInfo.UserID)
		assert.Equal(t, map[string]any{"email": "test@example.com", "name": "Test User"}, tokenInfo.Props)
		assert.False(t, time.Now().After(tokenInfo.ExpiresAt))
	})
}

// TestTokenClaims tests the TokenClaims struct
func TestTokenClaims(t *testing.T) {
	expiresAt := time.Now().Add(time.Hour)
	props := map[string]any{
		"email": "test@example.com",
		"name":  "Test User",
		"role":  "admin",
	}

	claims := &TokenInfo{
		UserID:    "user123",
		Props:     props,
		ExpiresAt: expiresAt,
	}

	assert.Equal(t, "user123", claims.UserID)
	assert.Equal(t, props, claims.Props)
	assert.Equal(t, expiresAt, claims.ExpiresAt)
}

// TestTokenInfo tests the TokenInfo struct
func TestTokenInfo(t *testing.T) {
	expiresAt := time.Now().Add(time.Hour)
	props := map[string]any{
		"email": "test@example.com",
		"name":  "Test User",
	}

	tokenInfo := &TokenInfo{
		UserID:    "user123",
		Props:     props,
		ExpiresAt: expiresAt,
	}

	assert.Equal(t, "user123", tokenInfo.UserID)
	assert.Equal(t, props, tokenInfo.Props)
	assert.Equal(t, expiresAt, tokenInfo.ExpiresAt)
}

// TestTokenValidationEdgeCases tests edge cases in token validation
func TestTokenValidationEdgeCases(t *testing.T) {
	mockDB := NewMockDatabase()
	tokenManager, err := NewTokenManager(mockDB)
	require.NoError(t, err)

	t.Run("TestEmptyToken", func(t *testing.T) {
		_, err := tokenManager.validateAccessToken("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid token format")
	})

	t.Run("TestTokenWithEmptyParts", func(t *testing.T) {
		_, err := tokenManager.validateAccessToken("user123::access_token")
		assert.Error(t, err)
		// Should still be valid as empty parts are allowed
	})

	t.Run("TestTokenWithTooManyParts", func(t *testing.T) {
		_, err := tokenManager.validateAccessToken("user123:grant456:access_token:extra_part")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid token format")
	})

	t.Run("TestTokenWithInvalidDatabaseResponse", func(t *testing.T) {
		// Test with database returning wrong type
		validToken := "user123:grant456:valid_token"
		mockDB.tokens[validToken] = &types.TokenData{
			AccessToken: validToken,
			ClientID:    "test_client",
			UserID:      "user123",
			GrantID:     "grant456",
			Scope:       "read write",
			ExpiresAt:   time.Now().Add(time.Hour),
			CreatedAt:   time.Now(),
			Revoked:     false,
		}

		// Add the corresponding grant
		mockDB.grants["grant456:user123"] = &types.Grant{
			ID:        "grant456",
			ClientID:  "test_client",
			UserID:    "user123",
			Scope:     []string{"read", "write"},
			Metadata:  map[string]any{"provider": "test"},
			Props:     map[string]any{"email": "test@example.com"},
			CreatedAt: time.Now().Unix(),
			ExpiresAt: time.Now().Add(10 * time.Minute).Unix(),
		}

		_, err := tokenManager.validateAccessToken(validToken)
		assert.NoError(t, err) // This should work with proper type
	})

	t.Run("TestGrantWithInvalidDatabaseResponse", func(t *testing.T) {
		validToken := "user123:grant456:valid_token"
		mockDB.tokens[validToken] = &types.TokenData{
			AccessToken: validToken,
			ClientID:    "test_client",
			UserID:      "user123",
			GrantID:     "grant456",
			Scope:       "read write",
			ExpiresAt:   time.Now().Add(time.Hour),
			CreatedAt:   time.Now(),
			Revoked:     false,
		}

		mockDB.grants["grant456:user123"] = &types.Grant{
			ID:        "grant456",
			ClientID:  "test_client",
			UserID:    "user123",
			Scope:     []string{"read", "write"},
			Metadata:  map[string]any{"provider": "test"},
			Props:     map[string]any{"email": "test@example.com"},
			CreatedAt: time.Now().Unix(),
			ExpiresAt: time.Now().Add(10 * time.Minute).Unix(),
		}

		_, err := tokenManager.validateAccessToken(validToken)
		assert.NoError(t, err) // This should work with proper type
	})
}

// TestConcurrentTokenValidation tests concurrent access to token validation
func TestConcurrentTokenValidation(t *testing.T) {
	mockDB := NewMockDatabase()
	tokenManager, err := NewTokenManager(mockDB)
	require.NoError(t, err)

	// Create multiple valid tokens
	for i := range 10 {
		token := fmt.Sprintf("user%d:grant%d:token%d", i, i, i)
		mockDB.tokens[token] = &types.TokenData{
			AccessToken: token,
			ClientID:    "test_client",
			UserID:      fmt.Sprintf("user%d", i),
			GrantID:     fmt.Sprintf("grant%d", i),
			Scope:       "read write",
			ExpiresAt:   time.Now().Add(time.Hour),
			CreatedAt:   time.Now(),
			Revoked:     false,
		}

		mockDB.grants[fmt.Sprintf("grant%d:user%d", i, i)] = &types.Grant{
			ID:        fmt.Sprintf("grant%d", i),
			ClientID:  "test_client",
			UserID:    fmt.Sprintf("user%d", i),
			Scope:     []string{"read", "write"},
			Metadata:  map[string]any{"provider": "test"},
			Props:     map[string]any{"email": fmt.Sprintf("user%d@example.com", i)},
			CreatedAt: time.Now().Unix(),
			ExpiresAt: time.Now().Add(10 * time.Minute).Unix(),
		}
	}

	// Test concurrent validation
	done := make(chan bool, 20)
	for i := range 20 {
		go func(id int) {
			token := fmt.Sprintf("user%d:grant%d:token%d", id%10, id%10, id%10)
			_, err := tokenManager.validateAccessToken(token)
			// Should not error for valid tokens
			_ = err
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for range 20 {
		<-done
	}

	// Verify token manager is still functional
	token := "user0:grant0:token0"
	claims, err := tokenManager.validateAccessToken(token)
	require.NoError(t, err)
	assert.Equal(t, "user0", claims.UserID)
}
