package types

import (
	"context"

	"github.com/obot-platform/mcp-oauth-proxy/pkg/providers"
)

type Context struct {
	User    User
	Config  ConfigFactory
	Profile []string
}

type User providers.UserInfo

type contextKey struct{}

func WithNanobotContext(ctx context.Context, nc Context) context.Context {
	return context.WithValue(ctx, contextKey{}, nc)
}

func NanobotContext(ctx context.Context) Context {
	c, _ := ctx.Value(contextKey{}).(Context)
	return c
}
