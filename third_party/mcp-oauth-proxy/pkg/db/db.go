package db

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/obot-platform/mcp-oauth-proxy/pkg/types"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Store represents the database connection and operations
type Store struct {
	db     *gorm.DB
	dbType string // "postgres" or "sqlite"
}

// New creates a new database connection and sets up the schema
func New(dsn string) (*Store, error) {
	var gormDB *gorm.DB
	var dbType string
	var err error

	// Configure GORM logger
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // Change to logger.Info for debugging
	}

	// If DSN is empty, use SQLite with local file
	if dsn == "" {
		// Create data directory if it doesn't exist
		dataDir := "data"
		if err := os.MkdirAll(dataDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create data directory: %w", err)
		}

		// Use local SQLite database
		sqlitePath := filepath.Join(dataDir, "oauth_proxy.db")
		gormDB, err = gorm.Open(sqlite.Open(sqlitePath), gormConfig)
		dbType = "sqlite"
	} else {
		// Check if it's a PostgreSQL DSN
		if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
			gormDB, err = gorm.Open(postgres.Open(dsn), gormConfig)
			dbType = "postgres"
		} else {
			// Assume SQLite file path
			gormDB, err = gorm.Open(sqlite.Open(dsn), gormConfig)
			dbType = "sqlite"
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	database := &Store{db: gormDB, dbType: dbType}

	// Setup schema using GORM AutoMigrate
	if err := database.setupSchema(); err != nil {
		return nil, fmt.Errorf("failed to setup schema: %w", err)
	}

	return database, nil
}

// setupSchema creates the necessary tables and handles migrations
func (d *Store) setupSchema() error {
	// Use GORM's AutoMigrate to create/update tables
	err := d.db.AutoMigrate(
		&types.ClientInfo{},
		&types.Grant{},
		&types.AuthorizationCode{},
		&types.TokenData{},
		&types.StoredAuthRequest{},
	)
	if err != nil {
		return fmt.Errorf("failed to auto-migrate database schema: %w", err)
	}

	return nil
}

// GetClient retrieves a client by ID
func (d *Store) GetClient(clientID string) (*types.ClientInfo, error) {
	var client types.ClientInfo
	err := d.db.First(&client, "client_id = ?", clientID).Error
	if err != nil {
		return nil, err
	}
	return &client, nil
}

// StoreClient stores a new client
func (d *Store) StoreClient(client *types.ClientInfo) error {
	return d.db.Create(client).Error
}

// StoreGrant stores a new grant
func (d *Store) StoreGrant(grant *types.Grant) error {
	// Convert []string to StringSlice and map[string]any to JSON for GORM
	gormGrant := &types.Grant{
		ID:                  grant.ID,
		ClientID:            grant.ClientID,
		UserID:              grant.UserID,
		Scope:               types.StringSlice(grant.Scope),
		Metadata:            types.JSON(grant.Metadata),
		Props:               types.JSON(grant.Props),
		CreatedAt:           grant.CreatedAt,
		ExpiresAt:           grant.ExpiresAt,
		CodeChallenge:       grant.CodeChallenge,
		CodeChallengeMethod: grant.CodeChallengeMethod,
	}
	return d.db.Create(gormGrant).Error
}

// UpdateGrant updates an existing grant's properties
func (d *Store) UpdateGrant(grant *types.Grant) error {
	// Convert types for GORM
	updateGrant := &types.Grant{
		Scope:     types.StringSlice(grant.Scope),
		Metadata:  types.JSON(grant.Metadata),
		Props:     types.JSON(grant.Props),
		ExpiresAt: grant.ExpiresAt,
	}

	result := d.db.Model(&types.Grant{}).Where("id = ? AND user_id = ?", grant.ID, grant.UserID).Updates(updateGrant)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("grant not found: id=%s, user_id=%s", grant.ID, grant.UserID)
	}

	return nil
}

// GetGrant retrieves a grant by ID and user ID
func (d *Store) GetGrant(grantID, userID string) (*types.Grant, error) {
	var grant types.Grant
	err := d.db.First(&grant, "id = ? AND user_id = ?", grantID, userID).Error
	if err != nil {
		return nil, err
	}

	// Convert GORM types back to original types
	result := &types.Grant{
		ID:                  grant.ID,
		ClientID:            grant.ClientID,
		UserID:              grant.UserID,
		Scope:               []string(grant.Scope),
		Metadata:            map[string]any(grant.Metadata),
		Props:               map[string]any(grant.Props),
		CreatedAt:           grant.CreatedAt,
		ExpiresAt:           grant.ExpiresAt,
		CodeChallenge:       grant.CodeChallenge,
		CodeChallengeMethod: grant.CodeChallengeMethod,
	}

	return result, nil
}

// StoreAuthCode stores an authorization code
func (d *Store) StoreAuthCode(code, grantID, userID string) error {
	authCode := &types.AuthorizationCode{
		Code:      code,
		GrantID:   grantID,
		UserID:    userID,
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}
	return d.db.Create(authCode).Error
}

// ValidateAuthCode validates an authorization code and returns grant info
func (d *Store) ValidateAuthCode(code string) (string, string, error) {
	var authCode types.AuthorizationCode
	err := d.db.First(&authCode, "code = ? AND expires_at > ?", code, time.Now()).Error
	if err != nil {
		return "", "", err
	}
	return authCode.GrantID, authCode.UserID, nil
}

// DeleteAuthCode deletes an authorization code (single-use)
func (d *Store) DeleteAuthCode(code string) error {
	return d.db.Delete(&types.AuthorizationCode{}, "code = ?", code).Error
}

// hashRefreshToken creates a SHA-256 hash of the refresh token for secure storage
func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return base64.StdEncoding.EncodeToString(hash[:])
}

// StoreToken stores an access token and refresh token
func (d *Store) StoreToken(data *types.TokenData) error {
	// Hash the tokens for secure storage
	hashedAccessToken := hashToken(data.AccessToken)
	hashedRefreshToken := hashToken(data.RefreshToken)

	// Set refresh token expiration to 30 days from now if not already set
	if data.RefreshTokenExpiresAt.IsZero() {
		data.RefreshTokenExpiresAt = time.Now().Add(30 * 24 * time.Hour)
	}

	tokenData := &types.TokenData{
		AccessToken:           hashedAccessToken,
		RefreshToken:          hashedRefreshToken,
		ClientID:              data.ClientID,
		UserID:                data.UserID,
		GrantID:               data.GrantID,
		Scope:                 data.Scope,
		ExpiresAt:             data.ExpiresAt,
		RefreshTokenExpiresAt: data.RefreshTokenExpiresAt,
		Revoked:               false,
	}

	return d.db.Create(tokenData).Error
}

// GetToken retrieves a token by access token
func (d *Store) GetToken(accessToken string) (*types.TokenData, error) {
	hashedAccessToken := hashToken(accessToken)

	var data types.TokenData
	err := d.db.First(&data, "access_token = ?", hashedAccessToken).Error
	if err != nil {
		return nil, err
	}

	return &data, nil
}

// GetTokenByRefreshToken retrieves a token by refresh token
func (d *Store) GetTokenByRefreshToken(refreshToken string) (*types.TokenData, error) {
	// Hash the refresh token for lookup
	hashedRefreshToken := hashToken(refreshToken)

	var data types.TokenData
	err := d.db.First(&data, "refresh_token = ? AND refresh_token_expires_at > ?", hashedRefreshToken, time.Now()).Error
	if err != nil {
		return nil, err
	}

	// Store the original refresh token in the struct for the caller
	data.RefreshToken = refreshToken

	return &data, nil
}

// IsRefreshTokenExpired checks if a refresh token is expired
func (d *Store) IsRefreshTokenExpired(refreshToken string) (bool, error) {
	hashedRefreshToken := hashToken(refreshToken)

	var data types.TokenData
	err := d.db.Select("refresh_token_expires_at").First(&data, "refresh_token = ?", hashedRefreshToken).Error
	if err != nil {
		return false, err
	}

	return time.Now().After(data.RefreshTokenExpiresAt), nil
}

// RevokeToken revokes an access token
func (d *Store) RevokeToken(token string) error {
	hashedToken := hashToken(token)
	now := time.Now()

	// First try to revoke as access token
	result := d.db.Model(&types.TokenData{}).Where("access_token = ?", hashedToken).Updates(map[string]any{
		"revoked":    true,
		"revoked_at": &now,
	})
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected > 0 {
		return nil
	}

	// If not found as access token, try as refresh token
	result = d.db.Model(&types.TokenData{}).Where("refresh_token = ?", hashedToken).Updates(map[string]any{
		"revoked":    true,
		"revoked_at": &now,
	})
	return result.Error
}

// CleanupExpiredTokens removes expired tokens and authorization codes
func (d *Store) CleanupExpiredTokens() error {
	now := time.Now()
	nowUnix := now.Unix()

	// Delete expired access tokens (both access and refresh tokens expired) or revoked tokens
	result := d.db.Where("(expires_at < ? AND refresh_token_expires_at < ?) OR revoked = ?", now, now, true).Delete(&types.TokenData{})
	if result.Error != nil {
		return fmt.Errorf("failed to cleanup expired tokens: %w", result.Error)
	}
	if result.RowsAffected > 0 {
		fmt.Printf("Deleted %d expired/revoked tokens\n", result.RowsAffected)
	}

	// Delete expired authorization codes
	result = d.db.Where("expires_at < ?", now).Delete(&types.AuthorizationCode{})
	if result.Error != nil {
		return fmt.Errorf("failed to cleanup expired authorization codes: %w", result.Error)
	}
	if result.RowsAffected > 0 {
		fmt.Printf("Deleted %d expired authorization codes\n", result.RowsAffected)
	}

	// Delete expired grants
	result = d.db.Where("expires_at < ?", nowUnix).Delete(&types.Grant{})
	if result.Error != nil {
		return fmt.Errorf("failed to cleanup expired grants: %w", result.Error)
	}
	if result.RowsAffected > 0 {
		fmt.Printf("Deleted %d expired grants\n", result.RowsAffected)
	}

	// Delete expired auth requests
	if err := d.CleanupExpiredAuthRequests(); err != nil {
		return fmt.Errorf("failed to cleanup expired auth requests: %w", err)
	}

	return nil
}

// StoreAuthRequest stores an authorization request with a 15-minute TTL
func (d *Store) StoreAuthRequest(key string, data map[string]any) error {
	authRequest := &types.StoredAuthRequest{
		Key:       key,
		Data:      types.JSON(data),
		ExpiresAt: time.Now().Add(15 * time.Minute), // 15-minute TTL
	}
	return d.db.Create(authRequest).Error
}

// GetAuthRequest retrieves an authorization request by key and checks TTL
func (d *Store) GetAuthRequest(key string) (map[string]any, error) {
	var authRequest types.StoredAuthRequest
	err := d.db.First(&authRequest, "key = ? AND expires_at > ?", key, time.Now()).Error
	if err != nil {
		return nil, err
	}

	// Convert JSON back to map
	return map[string]any(authRequest.Data), nil
}

// DeleteAuthRequest deletes an authorization request by key
func (d *Store) DeleteAuthRequest(key string) error {
	return d.db.Delete(&types.StoredAuthRequest{}, "key = ?", key).Error
}

// CleanupExpiredAuthRequests removes expired authorization requests
func (d *Store) CleanupExpiredAuthRequests() error {
	result := d.db.Where("expires_at < ?", time.Now()).Delete(&types.StoredAuthRequest{})
	if result.Error != nil {
		return fmt.Errorf("failed to cleanup expired auth requests: %w", result.Error)
	}
	if result.RowsAffected > 0 {
		fmt.Printf("Deleted %d expired auth requests\n", result.RowsAffected)
	}
	return nil
}

// Close closes the database connection
func (d *Store) Close() error {
	sqlDB, err := d.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// GenerateCodeChallenge generates a PKCE code challenge from a code verifier
func GenerateCodeChallenge(codeVerifier string) string {
	hash := sha256.Sum256([]byte(codeVerifier))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}
