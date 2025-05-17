package mcp

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/obot-platform/nanobot/pkg/log"
)

type StdioServer struct {
	MessageHandler MessageHandler
	stdio          *Stdio
}

func NewStdioServer(handler MessageHandler) *StdioServer {
	return &StdioServer{
		MessageHandler: handler,
	}
}

func (s *StdioServer) Wait() {
	if s.stdio != nil {
		s.stdio.Wait()
	}
}

func (s *StdioServer) Start(ctx context.Context, in io.ReadCloser, out io.WriteCloser) error {
	session, err := newServerSession(ctx, s.MessageHandler)
	if err != nil {
		return fmt.Errorf("failed to create stdio session: %w", err)
	}

	s.stdio = NewStdio("proxy", in, out)

	return s.stdio.Start(ctx, func(msg Message) {
		resp, err := session.Exchange(ctx, msg)
		if errors.Is(err, ErrNoResponse) {
			return
		} else if err != nil {
			log.Errorf(ctx, "failed to exchange message: %v", err)
		}
		if err := s.stdio.Send(ctx, resp); err != nil {
			log.Errorf(ctx, "failed to send message in reply to %v: %v", msg.ID, err)
		}
	})
}
