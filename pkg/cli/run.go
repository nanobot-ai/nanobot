package cli

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/obot-platform/nanobot/pkg/chat"
	"github.com/obot-platform/nanobot/pkg/confirm"
	"github.com/obot-platform/nanobot/pkg/log"
	"github.com/obot-platform/nanobot/pkg/mcp"
	"github.com/obot-platform/nanobot/pkg/runtime"
	"github.com/obot-platform/nanobot/pkg/server"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

type Run struct {
	MCP           bool     `usage:"Run the nanobot as an MCP server" default:"false" short:"m" env:"NANOBOT_MCP"`
	AutoConfirm   bool     `usage:"Automatically confirm all tool calls" default:"false" short:"y"`
	Output        string   `usage:"Output file for the result. Use - for stdout" default:"" short:"o"`
	ListenAddress string   `usage:"Address to listen on (ex: localhost:8099) (implies -m)" default:"stdio" short:"a"`
	Roots         []string `usage:"Roots to expose the MCP server in the form of name:directory" short:"r"`
	Input         string   `usage:"Input file for the prompt" default:"" short:"f"`
	n             *Nanobot
}

func NewRun(n *Nanobot) *Run {
	return &Run{
		n: n,
	}
}

func (r *Run) Customize(cmd *cobra.Command) {
	cmd.Use = "run [flags] NANOBOT [PROMPT]"
	cmd.Short = "Run the nanobot with the specified config file"
	cmd.Example = `
  # Run the nanobot.yaml in the current directory
  nanobot run .

  # Run the nanobot.yaml in the GitHub repo github.com/example/nanobot
  nanobot run example/nanobot

  # Run the nanobot.yaml at the URL
  nanobot run https://....

  # Run a single prompt and exit
  nanobot run . Talk like a pirate

  # Run the nanobot as a MCP Server
  nanobot run --mcp
`
	cmd.Args = cobra.MinimumNArgs(1)
}

func (r *Run) getRoots() ([]mcp.Root, error) {
	var (
		rootDefs = r.Roots
		roots    []mcp.Root
	)

	if len(rootDefs) == 0 {
		rootDefs = []string{"cwd:."}
	}

	for _, root := range rootDefs {
		name, directory, ok := strings.Cut(root, ":")
		if !ok {
			name = filepath.Base(root)
			directory = root
		}
		if !filepath.IsAbs(directory) {
			wd, err := os.Getwd()
			if err != nil {
				return nil, fmt.Errorf("failed to get current working directory: %w", err)
			}
			directory = filepath.Join(wd, directory)
		}
		if _, err := os.Stat(directory); err != nil {
			return nil, fmt.Errorf("failed to stat directory root (%s): %w", name, err)
		}

		roots = append(roots, mcp.Root{
			Name: name,
			URI:  "file://" + directory,
		})
	}

	return roots, nil
}

func (r *Run) Run(cmd *cobra.Command, args []string) error {
	var (
		runtimeOpt runtime.Options
	)

	if r.ListenAddress != "stdio" {
		r.MCP = true
	}

	roots, err := r.getRoots()
	if err != nil {
		return err
	}

	runtimeOpt.Roots = roots

	if r.MCP {
		runtime, err := r.n.GetRuntime(cmd.Context(), args[0], runtimeOpt)
		if err != nil {
			return err
		}

		return r.runMCP(cmd.Context(), runtime, nil)
	}

	runtimeOpt.Confirmations = confirm.NewService()
	runtimeOpt.Confirmations.Start(context.Background())

	runtime, err := r.n.GetRuntime(cmd.Context(), args[0], runtimeOpt)
	if err != nil {
		return err
	}

	if c := runtime.GetConfig(); c.Publish.Entrypoint == "" {
		var (
			agentName string
			example   string
		)
		for name := range c.Agents {
			agentName = name
			break
		}
		if agentName != "" {
			example = ", for example:\n\n```\npublish:\n  entrypoint: " + agentName + "\nagents:\n  " + agentName + ": ...\n```\n"
		}
		return fmt.Errorf("there are no entrypoints defined in the config file, please add one to the publish section%s", example)
	}

	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return fmt.Errorf("failed to pick a local port: %w", err)
	}
	r.ListenAddress = l.Addr().String()

	prompt := strings.Join(args[1:], " ")
	if r.Input != "" {
		input, err := os.ReadFile(r.Input)
		if err != nil {
			return fmt.Errorf("failed to read input file: %w", err)
		}
		prompt = strings.TrimSpace(string(input))
	}

	eg, ctx := errgroup.WithContext(cmd.Context())
	ctx, cancel := context.WithCancel(ctx)
	eg.Go(func() error {
		return r.runMCP(ctx, runtime, l)
	})
	eg.Go(func() error {
		defer cancel()
		return chat.Chat(ctx, r.ListenAddress, runtimeOpt.Confirmations, r.AutoConfirm, prompt, r.Output)
	})
	return eg.Wait()
}

func (r *Run) runMCP(ctx context.Context, runtime *runtime.Runtime, l net.Listener) error {
	env, err := r.n.loadEnv()
	if err != nil {
		return fmt.Errorf("failed to load environment: %w", err)
	}

	address := r.ListenAddress
	if strings.HasPrefix("address", "http://") {
		address = strings.TrimPrefix(address, "http://")
	} else if strings.HasPrefix(address, "https://") {
		return fmt.Errorf("https:// is not supported, use http:// instead")
	}

	mcpServer := server.NewServer(runtime)

	if address == "stdio" {
		stdio := mcp.NewStdioServer(env, mcpServer)
		if err := stdio.Start(ctx, os.Stdin, os.Stdout); err != nil {
			return fmt.Errorf("failed to start stdio server: %w", err)
		}

		stdio.Wait()
		return nil
	}

	httpServer := mcp.NewHTTPServer(env, mcpServer)

	s := &http.Server{
		Addr:    address,
		Handler: httpServer,
	}

	context.AfterFunc(ctx, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = s.Shutdown(ctx)
	})

	if l == nil {
		_, _ = fmt.Fprintf(os.Stderr, "Starting server on %s\n", address)
		err = s.ListenAndServe()
	} else {
		err = s.Serve(l)
	}
	log.Debugf(ctx, "Server stopped: %v", err)
	return err
}
