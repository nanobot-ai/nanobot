package mcp

import (
	"context"
	"fmt"
	"io"
	"maps"
	"net"
	"os"
	"slices"
	"strings"
	"sync"

	"github.com/obot-platform/nanobot/pkg/envvar"
	"github.com/obot-platform/nanobot/pkg/log"
	"github.com/obot-platform/nanobot/pkg/mcp/sandbox"
	"github.com/obot-platform/nanobot/pkg/supervise"
	"github.com/obot-platform/nanobot/pkg/system"
)

type runner struct {
	lock    sync.Mutex
	running map[string]Server
}

var (
	r runner
)

type streamResult struct {
	Stdout io.Reader
	Stdin  io.Writer
	Close  func()
}

func (r *runner) newCommand(ctx context.Context, currentEnv map[string]string, root []Root, config Server) (Server, *sandbox.Cmd, error) {
	var publishPorts []string
	if len(config.Ports) > 0 {
		if currentEnv == nil {
			currentEnv = make(map[string]string)
		} else {
			currentEnv = maps.Clone(currentEnv)
		}
		for _, port := range config.Ports {
			l, err := net.Listen("tcp4", "localhost:0")
			if err != nil {
				return config, nil, fmt.Errorf("failed to allocate port for %s: %w", port, err)
			}
			addrString := l.Addr().String()
			_, portStr, err := net.SplitHostPort(addrString)
			if err != nil {
				_ = l.Close()
				return config, nil, fmt.Errorf("failed to get port for %s, addr %s: %w", port, addrString, err)
			}
			if err := l.Close(); err != nil {
				return config, nil, fmt.Errorf("failed to close listener for %s, addr %s: %w", port, addrString, err)
			}
			publishPorts = append(publishPorts, portStr)
			currentEnv["port:"+port] = portStr
		}
	}

	config.BaseURL = envvar.ReplaceString(currentEnv, config.BaseURL)

	command, args, env := envvar.ReplaceEnv(currentEnv, config.Command, config.Args, config.Env)
	if config.Unsandboxed || command == "nanobot" {
		if command == "nanobot" {
			command = system.Bin()
		}
		cmd := supervise.Cmd(ctx, command, args...)
		cmd.Env = append(os.Environ(), env...)
		return config, &sandbox.Cmd{
			Cmd: cmd,
		}, nil
	}

	var rootPaths []string
	for _, root := range root {
		if strings.HasPrefix(root.URI, "file://") {
			rootPaths = append(rootPaths, root.URI[7:])
		}
	}

	cmd, err := sandbox.NewCmd(ctx, sandbox.Command{
		PublishPorts: publishPorts,
		ReversePorts: config.ReversePorts,
		Roots:        rootPaths,
		Command:      command,
		Args:         args,
		Env:          slices.Collect(maps.Keys(config.Env)),
		BaseImage:    config.BaseImage,
		Dockerfile:   config.Dockerfile,
		Source:       sandbox.Source(config.Source),
	})
	if err != nil {
		return config, nil, fmt.Errorf("failed to create sandbox command: %w", err)
	}

	cmd.Env = append(os.Environ(), env...)
	return config, cmd, nil
}

func (r *runner) doRun(ctx context.Context, serverName string, config Server, cmd *sandbox.Cmd) (Server, error) {
	// hold open stdin for the supervisor
	_, err := cmd.StdinPipe()
	if err != nil {
		return config, fmt.Errorf("failed to get stdin pipe: %w", err)
	}
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return config, fmt.Errorf("failed to get stdout pipe: %w", err)
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return config, fmt.Errorf("failed to get stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return config, fmt.Errorf("failed to start command: %w", err)
	}

	if r.running == nil {
		r.running = make(map[string]Server)
	}
	r.running[serverName] = config

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		sandbox.PipeOut(ctx, stdoutPipe, serverName)
	}()

	go func() {
		defer wg.Done()
		sandbox.PipeOut(ctx, stderrPipe, serverName)
	}()

	go func() {
		wg.Wait()
		err := cmd.Wait()
		if err != nil {
			log.Errorf(ctx, "Command %s exited with error: %v\n", serverName, err)
		}
		r.lock.Lock()
		delete(r.running, serverName)
		r.lock.Unlock()
	}()

	return config, nil
}

func (r *runner) doStream(ctx context.Context, serverName string, cmd *sandbox.Cmd) (*streamResult, error) {
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdout pipe: %w", err)
	}
	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdin pipe: %w", err)
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start command: %w", err)
	}

	go func() {
		sandbox.PipeOut(ctx, stderrPipe, serverName)
		if err := cmd.Wait(); err != nil {
			log.Errorf(ctx, "Command %s exited with error: %v\n", serverName, err)
		}
	}()

	return &streamResult{
		Stdout: stdoutPipe,
		Stdin:  stdinPipe,
	}, nil
}

func (r *runner) Run(ctx context.Context, roots []Root, env map[string]string, serverName string, config Server) (Server, error) {
	r.lock.Lock()
	defer r.lock.Unlock()

	if c, ok := r.running[serverName]; ok {
		return c, nil
	}

	newConfig, cmd, err := r.newCommand(ctx, env, roots, config)
	if err != nil {
		return config, err
	}

	return r.doRun(ctx, serverName, newConfig, cmd)
}

func (r *runner) Stream(ctx context.Context, roots []Root, env map[string]string, serverName string, config Server) (*streamResult, error) {
	ctx, cancel := context.WithCancel(ctx)
	_, cmd, err := r.newCommand(ctx, env, roots, config)
	if err != nil {
		cancel()
		return nil, err
	}
	result, err := r.doStream(ctx, serverName, cmd)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to run stdio command: %w", err)
	}
	result.Close = cancel
	return result, nil
}
