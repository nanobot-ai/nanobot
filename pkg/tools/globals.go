package tools

import (
	"context"
	"fmt"
	"maps"
	"strings"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/types"
)

func (s *Service) callPrompt(ctx context.Context, prompt string, args map[string]string) (string, error) {
	server, prompt, ok := strings.Cut(prompt, "/")
	if !ok {
		prompt = server
	}

	promptResult, err := s.GetPrompt(ctx, server, prompt, args)
	if err != nil {
		return "", fmt.Errorf("failed to get prompt %s from server %s: %w", prompt, server, err)
	}

	for _, msg := range promptResult.Messages {
		if msg.Content.Text != "" {
			return msg.Content.Text, nil
		}
	}
	return "", nil
}

func (s *Service) newGlobals(ctx context.Context, vars map[string]any) map[string]any {
	session := mcp.SessionFromContext(ctx)
	attr := session.Attributes()
	//data["nanobot"] = attr
	servers := map[string]any{}

	for k, v := range attr {
		cf, ok := v.(*clientFactory)
		if !ok {
			continue
		}
		serverName, ok := strings.CutPrefix(k, "clients/")
		if !ok {
			continue
		}
		var instructions string
		if cf.client != nil && cf.client.Session != nil {
			instructions = cf.client.Session.InitializeResult.Instructions
		}
		servers[serverName] = map[string]any{
			"instructions": instructions,
		}
	}

	c := types.ConfigFromContext(ctx)
	for serverName := range c.MCPServers {
		if _, ok := servers[serverName]; !ok {
			servers[serverName] = map[string]any{
				"instructions": "",
			}
		}
	}

	data := map[string]any{
		"nanobot": map[string]any{
			"prompt": func(target string, args map[string]string) (string, error) {
				return s.callPrompt(ctx, target, args)
			},
			"servers": servers,
		},
	}
	maps.Copy(data, vars)
	return data
}
