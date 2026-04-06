package types

import (
	"context"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
)

type Context struct {
	User    mcp.User
	Config  ConfigFactory
	Profile []string
}

type contextKey struct{}
type internalLLMRequestTypeKey struct{}

const (
	InternalLLMRequestTypeHeader = "X-Obot-Internal-Request-Type"
	ThreadTitleRequestType       = "nanobot.summary.thread_title"
)

func WithNanobotContext(ctx context.Context, nc Context) context.Context {
	return context.WithValue(ctx, contextKey{}, nc)
}

func NanobotContext(ctx context.Context) Context {
	c, _ := ctx.Value(contextKey{}).(Context)
	return c
}

func WithInternalLLMRequestType(ctx context.Context, requestType string) context.Context {
	if requestType == "" {
		return ctx
	}
	return context.WithValue(ctx, internalLLMRequestTypeKey{}, requestType)
}

func InternalLLMRequestType(ctx context.Context) string {
	requestType, _ := ctx.Value(internalLLMRequestTypeKey{}).(string)
	return requestType
}

func WithThreadTitleRequest(ctx context.Context) context.Context {
	return WithInternalLLMRequestType(ctx, ThreadTitleRequestType)
}
