package mcp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"maps"

	"github.com/nanobot-ai/nanobot/pkg/log"
)

type StdioServer struct {
	MessageHandler MessageHandler
	stdio          *Stdio
	env            map[string]string
}

func NewStdioServer(env map[string]string, handler MessageHandler) *StdioServer {
	return &StdioServer{
		env:            env,
		MessageHandler: handler,
	}
}

func (s *StdioServer) Wait() {
	if s.stdio != nil {
		s.stdio.Wait()
	}
}

func (s *StdioServer) Start(ctx context.Context, in io.ReadCloser, out io.WriteCloser) error {
	session, err := NewServerSession(ctx, s.MessageHandler)
	if err != nil {
		return fmt.Errorf("failed to create stdio session: %w", err)
	}

	maps.Copy(session.session.EnvMap(), s.env)

	s.stdio = NewStdio("proxy", nil, in, out, func() {})

	return s.stdio.Start(ctx, func(ctx context.Context, msg Message) {
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
