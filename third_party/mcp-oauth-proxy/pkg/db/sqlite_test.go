package db

import (
	"os"
	"testing"
	"time"

	"github.com/obot-platform/mcp-oauth-proxy/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSQLiteDatabase(t *testing.T) {
	// Clean up any existing test database
	if err := os.RemoveAll("data"); err != nil {
		t.Logf("Error removing data directory: %v", err)
	}
	defer func() {
		if err := os.RemoveAll("data"); err != nil {
			t.Logf("Error removing data directory: %v", err)
		}
	}()

	// Test SQLite database without DSN
	db, err := New("")
	require.NoError(t, err)
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("Error closing database: %v", err)
		}
	}()

	// Verify it's using SQLite
	assert.Equal(t, "sqlite", db.dbType)

	t.Run("TestClientOperations", func(t *testing.T) {
		// Test client operations
		client := &types.ClientInfo{
			ClientID:     "test_client_sqlite",
			ClientSecret: "test_secret",
			RedirectUris: []string{"http://localhost:8080/callback"},
			ClientName:   "Test SQLite Client",
		}

		// Store client
		err := db.StoreClient(client)
		require.NoError(t, err)

		// Retrieve client
		retrievedClient, err := db.GetClient("test_client_sqlite")
		require.NoError(t, err)
		require.NotNil(t, retrievedClient)

		assert.Equal(t, client.ClientID, retrievedClient.ClientID)
		assert.Equal(t, client.ClientName, retrievedClient.ClientName)
		assert.Equal(t, client.RedirectUris, retrievedClient.RedirectUris)
	})

	t.Run("TestGrantOperations", func(t *testing.T) {
		// Test grant operations
		grant := &types.Grant{
			ID:        "test_grant_sqlite",
			ClientID:  "test_client_sqlite",
			UserID:    "test_user",
			Scope:     []string{"openid", "profile"},
			Metadata:  map[string]any{"test": "value"},
			Props:     map[string]any{"prop": "value"},
			CreatedAt: time.Now().Unix(),
			ExpiresAt: time.Now().Add(time.Hour).Unix(),
		}

		// Store grant
		err := db.StoreGrant(grant)
		require.NoError(t, err)

		// Retrieve grant
		retrievedGrant, err := db.GetGrant("test_grant_sqlite", "test_user")
		require.NoError(t, err)
		require.NotNil(t, retrievedGrant)

		assert.Equal(t, grant.ID, retrievedGrant.ID)
		assert.Equal(t, grant.ClientID, retrievedGrant.ClientID)
		assert.Equal(t, grant.UserID, retrievedGrant.UserID)
		assert.Equal(t, grant.Scope, retrievedGrant.Scope)
	})

	t.Run("TestTokenOperations", func(t *testing.T) {
		// Test token operations
		tokenData := &types.TokenData{
			AccessToken:  "test_access_token_sqlite",
			RefreshToken: "test_refresh_token_sqlite",
			ClientID:     "test_client_sqlite",
			UserID:       "test_user",
			GrantID:      "test_grant_sqlite",
			Scope:        "openid profile",
			ExpiresAt:    time.Now().Add(time.Hour),
			// RefreshTokenExpiresAt will be set automatically to 30 days
		}

		// Store token
		err := db.StoreToken(tokenData)
		require.NoError(t, err)

		// Retrieve token
		retrievedToken, err := db.GetToken("test_access_token_sqlite")
		require.NoError(t, err)
		require.NotNil(t, retrievedToken)

		assert.Equal(t, tokenData.ClientID, retrievedToken.ClientID)
		assert.Equal(t, tokenData.UserID, retrievedToken.UserID)
		assert.Equal(t, tokenData.GrantID, retrievedToken.GrantID)
		assert.Equal(t, tokenData.Scope, retrievedToken.Scope)
		assert.True(t, retrievedToken.RefreshTokenExpiresAt.After(time.Now().Add(29*24*time.Hour))) // Should be ~30 days

		// Test retrieving by refresh token
		refreshToken, err := db.GetTokenByRefreshToken("test_refresh_token_sqlite")
		require.NoError(t, err)
		assert.Equal(t, tokenData.ClientID, refreshToken.ClientID)
		assert.True(t, refreshToken.RefreshTokenExpiresAt.After(time.Now().Add(29*24*time.Hour)))

		// Test refresh token expiration check
		expired, err := db.IsRefreshTokenExpired("test_refresh_token_sqlite")
		require.NoError(t, err)
		assert.False(t, expired)
	})

	t.Run("TestAuthCodeOperations", func(t *testing.T) {
		// Test authorization code operations
		code := "test_auth_code_sqlite"
		grantID := "test_grant_sqlite"
		userID := "test_user"

		// Store auth code
		err := db.StoreAuthCode(code, grantID, userID)
		require.NoError(t, err)

		// Validate auth code
		retrievedGrantID, retrievedUserID, err := db.ValidateAuthCode(code)
		require.NoError(t, err)
		assert.Equal(t, grantID, retrievedGrantID)
		assert.Equal(t, userID, retrievedUserID)

		// Delete auth code
		err = db.DeleteAuthCode(code)
		require.NoError(t, err)

		// Verify it's deleted
		_, _, err = db.ValidateAuthCode(code)
		assert.Error(t, err) // Should fail because code is deleted
	})
}
