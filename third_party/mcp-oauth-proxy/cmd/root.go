package cmd

import (
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/gptscript-ai/cmd"
	"github.com/obot-platform/mcp-oauth-proxy/pkg/proxy"
	"github.com/obot-platform/mcp-oauth-proxy/pkg/types"
	"github.com/spf13/cobra"
)

var (
	version   = "dev"
	buildTime = "unknown"
)

// RootCmd represents the base command when called without any subcommands
type RootCmd struct {
	// Database configuration
	DatabaseDSN string `name:"database-dsn" env:"DATABASE_DSN" usage:"Database connection string (PostgreSQL or SQLite file path). If empty, uses SQLite at data/oauth_proxy.db"`

	// OAuth Provider configuration
	OAuthClientID     string `name:"oauth-client-id" env:"OAUTH_CLIENT_ID" usage:"OAuth client ID from your OAuth provider" required:"true"`
	OAuthClientSecret string `name:"oauth-client-secret" env:"OAUTH_CLIENT_SECRET" usage:"OAuth client secret from your OAuth provider" required:"true"`
	OAuthAuthorizeURL string `name:"oauth-authorize-url" env:"OAUTH_AUTHORIZE_URL" usage:"Authorization endpoint URL from your OAuth provider (e.g., https://accounts.google.com)" required:"true"`
	OAuthJWKSURL      string `name:"oauth-jwks-url" env:"OAUTH_JWKS_URL" usage:"JWKS endpoint URL from your OAuth provider (e.g., https://accounts.google.com/.well-known/openid-configuration/jwks)"`

	// Scopes and MCP configuration
	ScopesSupported string `name:"scopes-supported" env:"SCOPES_SUPPORTED" usage:"Comma-separated list of supported OAuth scopes (e.g., 'openid,profile,email')" required:"true"`
	MCPServerURL    string `name:"mcp-server-url" env:"MCP_SERVER_URL" usage:"URL of the MCP server to proxy requests to" required:"true"`

	// Security configuration
	EncryptionKey string `name:"encryption-key" env:"ENCRYPTION_KEY" usage:"Base64-encoded 32-byte AES-256 key for encrypting sensitive data (optional)"`

	// Server configuration
	Port        string `name:"port" env:"PORT" usage:"Port to run the server on" default:"8080"`
	Host        string `name:"host" env:"HOST" usage:"Host to bind the server to" default:"localhost"`
	RoutePrefix string `name:"route-prefix" env:"ROUTE_PREFIX" usage:"Optional prefix for all routes (e.g., '/oauth2')"`

	// Logging
	Verbose bool `name:"verbose,v" usage:"Enable verbose logging"`
	Version bool `name:"version" usage:"Show version information"`

	Mode string `name:"mode" env:"MODE" usage:"Mode to run the server in" default:"proxy"`
}

func (c *RootCmd) Run(cobraCmd *cobra.Command, args []string) error {
	if c.Version {
		fmt.Printf("MCP OAuth Proxy\n")
		fmt.Printf("Version: %s\n", version)
		fmt.Printf("Built: %s\n", buildTime)
		return nil
	}

	// Configure logging
	if c.Verbose {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
		log.Println("Verbose logging enabled")
	}

	// Convert CLI config to internal config format
	config := &types.Config{
		DatabaseDSN:       c.DatabaseDSN,
		OAuthClientID:     c.OAuthClientID,
		OAuthClientSecret: c.OAuthClientSecret,
		OAuthAuthorizeURL: c.OAuthAuthorizeURL,
		OAuthJWKSURL:      c.OAuthJWKSURL,
		ScopesSupported:   c.ScopesSupported,
		MCPServerURL:      c.MCPServerURL,
		EncryptionKey:     c.EncryptionKey,
		Mode:              c.Mode,
		RoutePrefix:       c.RoutePrefix,
	}

	// Validate configuration
	if err := c.validateConfig(); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	// Create OAuth proxy
	oauthProxy, err := proxy.NewOAuthProxy(config)
	if err != nil {
		return fmt.Errorf("failed to create OAuth proxy: %w", err)
	}
	defer func() {
		if err := oauthProxy.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}()

	// Get HTTP handler
	handler := oauthProxy.GetHandler()

	// Start server
	address := fmt.Sprintf("%s:%s", c.Host, c.Port)
	log.Printf("Starting OAuth proxy server on %s", address)
	log.Printf("OAuth Provider: %s", c.OAuthAuthorizeURL)
	log.Printf("MCP Server: %s", c.MCPServerURL)
	log.Printf("Database: %s", c.getDatabaseType())

	return http.ListenAndServe(address, handler)
}

func (c *RootCmd) validateConfig() error {
	if c.OAuthClientID == "" {
		return fmt.Errorf("oauth-client-id is required")
	}
	if c.OAuthClientSecret == "" {
		return fmt.Errorf("oauth-client-secret is required")
	}
	if c.OAuthAuthorizeURL == "" {
		return fmt.Errorf("oauth-authorize-url is required")
	}
	if c.ScopesSupported == "" {
		return fmt.Errorf("scopes-supported is required")
	}
	if c.MCPServerURL == "" {
		return fmt.Errorf("mcp-server-url is required")
	}
	if c.Mode == proxy.ModeProxy {
		if u, err := url.Parse(c.MCPServerURL); err != nil || u.Scheme != "http" && u.Scheme != "https" {
			return fmt.Errorf("invalid MCP server URL: %w", err)
		} else if u.Path != "" && u.Path != "/" || u.RawQuery != "" || u.Fragment != "" {
			return fmt.Errorf("MCP server URL must not contain a path, query, or fragment")
		}
	}
	return nil
}

func (c *RootCmd) getDatabaseType() string {
	if c.DatabaseDSN == "" {
		return "SQLite (data/oauth_proxy.db)"
	}
	if len(c.DatabaseDSN) > 10 && (c.DatabaseDSN[:11] == "postgres://" || c.DatabaseDSN[:14] == "postgresql://") {
		return "PostgreSQL"
	}
	return fmt.Sprintf("SQLite (%s)", c.DatabaseDSN)
}

// Customizer interface implementation for additional command customization
func (c *RootCmd) Customize(cobraCmd *cobra.Command) {
	cobraCmd.Use = "mcp-oauth-proxy"
	cobraCmd.Short = "OAuth 2.1 proxy server for MCP (Model Context Protocol)"
	cobraCmd.Long = `MCP OAuth Proxy is a comprehensive OAuth 2.1 proxy server that provides
OAuth authorization server functionality with PostgreSQL/SQLite storage.

This proxy supports multiple OAuth providers (Google, Microsoft, GitHub) and
proxies requests to MCP servers with user context headers.

Examples:
  # Start with environment variables
  export OAUTH_CLIENT_ID="your-google-client-id"
  export OAUTH_CLIENT_SECRET="your-secret"
  export OAUTH_AUTHORIZE_URL="https://accounts.google.com"
  export SCOPES_SUPPORTED="openid,profile,email"
  export MCP_SERVER_URL="http://localhost:3000"
  mcp-oauth-proxy

  # Start with CLI flags
  mcp-oauth-proxy \
    --oauth-client-id="your-google-client-id" \
    --oauth-client-secret="your-secret" \
    --oauth-authorize-url="https://accounts.google.com" \
    --scopes-supported="openid,profile,email" \
    --mcp-server-url="http://localhost:3000"

  # Use PostgreSQL database
  mcp-oauth-proxy \
    --database-dsn="postgres://user:pass@localhost:5432/oauth_db?sslmode=disable" \
    --oauth-client-id="your-client-id" \
    # ... other required flags

Configuration:
  Configuration values are loaded in this order (later values override earlier ones):
  1. Default values
  2. Environment variables
  3. Command line flags

Database Support:
  - PostgreSQL: Full ACID compliance, recommended for production
  - SQLite: Zero configuration, perfect for development and small deployments`

	cobraCmd.Version = version
}

// Execute is the main entry point for the CLI
func Execute() error {
	rootCmd := &RootCmd{}
	cobraCmd := cmd.Command(rootCmd)
	return cobraCmd.Execute()
}
