package mcp

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/nanobot-ai/nanobot/pkg/log"
)

type StdioServer struct {
	MessageHandler MessageHandler
	stdio          *Stdio
	envProvider    func() (map[string]string, error)
}

func NewStdioServer(envProvider func() (map[string]string, error), handler MessageHandler) *StdioServer {
	return &StdioServer{
		envProvider:    envProvider,
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

	s.stdio = NewStdio("proxy", nil, in, out, func() {})

	if err = s.stdio.Start(ctx, func(ctx context.Context, msg Message) {
		if s.envProvider != nil {
			env, err := s.envProvider()
			if err != nil {
				log.Errorf(ctx, "failed to reload environment: %v", err)
			} else {
				session.session.SetEnv(env)
			}
		}

		resp, err := session.Exchange(ctx, msg)
		if errors.Is(err, ErrNoResponse) {
			return
		} else if err != nil {
			log.Errorf(ctx, "failed to exchange message: %v", err)
		}
		if err := s.stdio.Send(ctx, resp); err != nil {
			log.Errorf(ctx, "failed to send message in reply to %v: %v", msg.ID, err)
		}
	}); err != nil {
		return fmt.Errorf("failed to start stdio: %w", err)
	}

	go func() {
		s.stdio.Wait()
		session.Close(false)
	}()

	return nil
}
