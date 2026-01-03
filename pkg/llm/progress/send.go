package progress

import (
	"context"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/types"
)

// ProgressInterceptor is a function that can intercept progress messages
type ProgressInterceptor func(ctx context.Context, progress *types.CompletionProgress)

type interceptorKeyType struct{}

var interceptorKey = interceptorKeyType{}

// WithInterceptor adds a progress interceptor to the context
func WithInterceptor(ctx context.Context, interceptor ProgressInterceptor) context.Context {
	return context.WithValue(ctx, interceptorKey, interceptor)
}

func Send(ctx context.Context, progress *types.CompletionProgress, progressToken any) {
	if progressToken == "" || progressToken == nil {
		return
	}

	// Call interceptor if present
	if interceptor, ok := ctx.Value(interceptorKey).(ProgressInterceptor); ok && interceptor != nil {
		interceptor(ctx, progress)
	}

	session := mcp.SessionFromContext(ctx)
	if session == nil {
		return
	}

	_ = session.SendPayload(ctx, "notifications/progress", mcp.NotificationProgressRequest{
		ProgressToken: progressToken,
		Meta: map[string]any{
			types.CompletionProgressMetaKey: progress,
		},
	})
}
