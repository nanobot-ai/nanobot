package obot

import (
	"context"
	"fmt"
	"net/http"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
)

type Config struct {
	BaseURL    string
	AuthHeader string
}

func GetConfig(ctx context.Context) (Config, error) {
	session := mcp.SessionFromContext(ctx)
	if session == nil {
		return Config{}, fmt.Errorf("no session found")
	}

	envMap := session.GetEnvMap()
	baseURL := envMap["OBOT_URL"]
	if baseURL == "" {
		return Config{}, fmt.Errorf("OBOT_URL is not configured")
	}

	var authHeader string
	if apiKey := envMap["MCP_API_KEY"]; apiKey != "" {
		authHeader = "Bearer " + apiKey
	}

	return Config{
		BaseURL:    baseURL,
		AuthHeader: authHeader,
	}, nil
}

func ApplyAuth(req *http.Request, cfg Config) {
	if cfg.AuthHeader != "" {
		req.Header.Set("Authorization", cfg.AuthHeader)
	}
}
