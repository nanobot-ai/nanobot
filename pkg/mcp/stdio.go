package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	log2 "log"
	"strings"
	"sync"

	"github.com/obot-platform/nanobot/pkg/log"
)

type waiter struct {
	running chan struct{}
	closed  bool
	lock    sync.Mutex
}

func newWaiter() *waiter {
	return &waiter{
		running: make(chan struct{}),
	}
}

func (w *waiter) Wait() {
	<-w.running
}

func (w *waiter) Close() {
	w.lock.Lock()
	if !w.closed {
		w.closed = true
		close(w.running)
	}
	w.lock.Unlock()
}

type Stdio struct {
	stdout         io.Reader
	stdin          io.Writer
	closer         func()
	server         string
	pendingRequest pendingRequest
	waiter         *waiter
	writeLock      sync.Mutex
}

func (s *Stdio) Send(ctx context.Context, req Message) error {
	s.writeLock.Lock()
	defer s.writeLock.Unlock()

	data, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	log.Messages(ctx, s.server, true, data)
	_, err = s.stdin.Write(append(data, '\n'))
	return err
}

func (s *Stdio) Wait() {
	s.waiter.Wait()
}

func (s *Stdio) Close() {
	s.closer()
	s.waiter.Close()
}

func (s *Stdio) Start(ctx context.Context, handler wireHandler) error {
	context.AfterFunc(ctx, func() {
		s.Close()
	})
	go func() {
		defer s.Close()
		err := s.start(ctx, handler)
		if err != nil {
			log2.Fatal(err)
		}
	}()
	return nil
}

func (s *Stdio) start(ctx context.Context, handler wireHandler) error {
	buf := bufio.NewScanner(s.stdout)
	buf.Buffer(make([]byte, 0, 1024), 10*1024*1024)
	for buf.Scan() {
		text := strings.TrimSpace(buf.Text())
		log.Messages(ctx, s.server, false, []byte(text))
		var msg Message
		if err := json.Unmarshal([]byte(text), &msg); err != nil {
			log.Errorf(ctx, "failed to unmarshal message: %v", err)
			continue
		}
		go handler(msg)
	}
	return buf.Err()
}

func newStdioClient(ctx context.Context, roots []Root, env map[string]string, serverName string, config Server) (*Stdio, error) {
	result, err := r.Stream(ctx, roots, env, serverName, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create stream: %w", err)
	}

	s := NewStdio(serverName, result.Stdout, result.Stdin, result.Close)
	return s, nil
}

func NewStdio(server string, in io.Reader, out io.Writer, close func()) *Stdio {
	return &Stdio{
		server: server,
		stdout: in,
		stdin:  out,
		closer: close,
		waiter: newWaiter(),
	}
}
