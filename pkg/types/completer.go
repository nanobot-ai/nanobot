package types

import (
	"context"
	"encoding/json"
)

type Completer interface {
	Complete(ctx context.Context, req Request, opts ...CompletionOptions) (*Response, error)
}

type CompletionOptions struct {
	Progress chan<- json.RawMessage
	Continue bool
}
