package types

import (
	"context"

	apitypes "github.com/obot-platform/obot/apiclient/types"
)

type Context struct {
	User    apitypes.User
	Config  ConfigFactory
	Profile []string
}
type contextKey struct{}

func WithNanobotContext(ctx context.Context, nc Context) context.Context {
	return context.WithValue(ctx, contextKey{}, nc)
}

func NanobotContext(ctx context.Context) Context {
	c, _ := ctx.Value(contextKey{}).(Context)
	return c
}
