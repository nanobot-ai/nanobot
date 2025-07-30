package cli

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/nanobot-ai/nanobot/pkg/chat"
	"github.com/nanobot-ai/nanobot/pkg/confirm"
	"github.com/nanobot-ai/nanobot/pkg/log"
	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/runtime"
	"github.com/nanobot-ai/nanobot/pkg/server"
	"github.com/nanobot-ai/nanobot/pkg/session"
	"github.com/nanobot-ai/nanobot/pkg/types"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

type Run struct {
	MCP           bool     `usage:"Run the nanobot as an MCP server" default:"false" short:"m" env:"NANOBOT_MCP"`
	UI            bool     `usage:"Run the nanobot with a UI" default:"false" short:"u"`
	AutoConfirm   bool     `usage:"Automatically confirm all tool calls" default:"false" short:"y"`
	Output        string   `usage:"Output file for the result. Use - for stdout" default:"" short:"o"`
	ListenAddress string   `usage:"Address to listen on (ex: localhost:8099) (implies -m)" default:"stdio" short:"a"`
	Port          string   `usage:"Port to listen on for stdio" default:"8099"`
	HealthzPath   string   `usage:"Path to serve healthz on"`
	Roots         []string `usage:"Roots to expose the MCP server in the form of name:directory" short:"r"`
	Input         string   `usage:"Input file for the prompt" default:"" short:"f"`
	Session       string   `usage:"Session ID to resume" default:"" short:"s"`
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

func (r *Run) reload(ctx context.Context, client *mcp.Client, cfgPath string, runtimeOpt runtime.Options) error {
	_, err := r.n.ReadConfig(ctx, cfgPath, runtimeOpt)
	if err != nil {
		return fmt.Errorf("failed to reload config: %w", err)
	}

	return client.Session.Exchange(ctx, "initialize", mcp.InitializeRequest{}, &mcp.InitializeResult{})
}

func (r *Run) Run(cmd *cobra.Command, args []string) error {
	var (
		runtimeOpt runtime.Options
	)

	if r.ListenAddress != "stdio" {
		r.MCP = true
	}

	if r.UI {
		runtimeOpt.Profiles = append(runtimeOpt.Profiles, "nanobot.ui")
	}

	roots, err := r.getRoots()
	if err != nil {
		return err
	}

	cfgPath := "nanobot.default"
	if len(args) > 0 {
		cfgPath = args[0]
	}

	config, err := r.n.ReadConfig(cmd.Context(), cfgPath, runtimeOpt)
	if err != nil {
		return fmt.Errorf("failed to read config file %q: %w", args[0], err)
	}

	oauthcallbackHandler := mcp.NewCallbackServer(confirm.New())

	runtimeOpt.Roots = roots
	runtimeOpt.MaxConcurrency = r.n.MaxConcurrency
	runtimeOpt.CallbackHandler = oauthcallbackHandler

	if r.MCP {
		runtime, err := r.n.GetRuntime(runtimeOpt, runtime.Options{
			OAuthRedirectURL: "http://" + strings.Replace(r.ListenAddress, "127.0.0.1", "localhost", 1) + "/oauth/callback",
			DSN:              r.n.DSN(),
		})
		if err != nil {
			return err
		}

		return r.runMCP(cmd.Context(), *config, runtime, oauthcallbackHandler, nil, r.HealthzPath)
	}
	if r.Port == "" {
		r.Port = "0"
	}

	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return fmt.Errorf("failed to pick a local port: %w", err)
	}
	r.ListenAddress = l.Addr().String()

	runtime, err := r.n.GetRuntime(runtimeOpt, runtime.Options{
		OAuthRedirectURL: "http://" + strings.Replace(r.ListenAddress, "127.0.0.1", "localhost", 1) + "/oauth/callback",
		DSN:              r.n.DSN(),
	})
	if err != nil {
		return err
	}

	if len(config.Publish.Entrypoint) == 0 {
		if _, ok := config.Agents["main"]; !ok {
			var (
				agentName string
				example   string
			)
			for name := range config.Agents {
				agentName = name
				break
			}
			if agentName != "" {
				example = ", for example:\n\n```\npublish:\n  entrypoint: " + agentName + "\nagents:\n  " + agentName + ": ...\n```\n"
			}
			return fmt.Errorf("there are no entrypoints defined in the config file, please add one to the publish section%s", example)
		}
	}

	var prompt string
	if len(args) > 0 {
		prompt = strings.Join(args[1:], " ")
	}
	if r.Input != "" {
		input, err := os.ReadFile(r.Input)
		if err != nil {
			return fmt.Errorf("failed to read input file: %w", err)
		}
		prompt = strings.TrimSpace(string(input))
	}

	var clientOpt mcp.ClientOption

	if r.Session != "" {
		store, err := session.NewStoreFromDSN(r.n.DSN())
		if err != nil {
			return fmt.Errorf("failed to open session store: %w", err)
		}
		sessions, err := store.FindByPrefix(cmd.Context(), r.Session)
		if err != nil {
			return fmt.Errorf("failed to find session: %w", err)
		} else if len(sessions) > 1 {
			return fmt.Errorf("multiple sessions found with prefix %q, please specify a full session ID", r.Session)
		} else if len(sessions) == 0 {
			return fmt.Errorf("no sessions found with prefix %q", r.Session)
		}
		clientOpt.SessionState = (*mcp.SessionState)(&sessions[0].State)
		clientOpt.SessionState.ID = sessions[0].SessionID
		if len(sessions[0].Config.Agents) > 0 && len(args) == 0 {
			config = (*types.Config)(&sessions[0].Config)
		}
		if clientOpt.SessionState.Attributes == nil {
			clientOpt.SessionState.Attributes = make(map[string]any)
		}
	}

	eg, ctx := errgroup.WithContext(cmd.Context())
	ctx, cancel := context.WithCancel(ctx)
	eg.Go(func() error {
		return r.runMCP(ctx, *config, runtime, oauthcallbackHandler, l, r.HealthzPath)
	})
	eg.Go(func() error {
		defer cancel()
		return chat.Chat(ctx, r.ListenAddress, r.AutoConfirm, prompt, r.Output,
			func(client *mcp.Client) error {
				return r.reload(ctx, client, args[0], runtimeOpt)
			}, clientOpt)
	})
	return eg.Wait()
}

func (r *Run) runMCP(ctx context.Context, config types.Config, runtime *runtime.Runtime, oauthCallbackHandler mcp.CallbackServer, l net.Listener, healthzPath string) error {
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

	var mcpServer mcp.MessageHandler = server.NewServer(runtime)

	if address == "stdio" {
		stdio := mcp.NewStdioServer(env, mcpServer)
		if err := stdio.Start(ctx, os.Stdin, os.Stdout); err != nil {
			return fmt.Errorf("failed to start stdio server: %w", err)
		}

		stdio.Wait()
		return nil
	}

	sessionManager, err := session.NewManager(mcpServer, r.n.DSN(), config)
	if err != nil {
		return err
	}

	httpServer := mcp.NewHTTPServer(env, mcpServer, mcp.HTTPServerOptions{
		SessionStore: sessionManager,
		HealthzPath:  healthzPath,
	})

	mux := http.NewServeMux()
	mux.Handle("/", httpServer)
	if oauthCallbackHandler != nil {
		mux.Handle("/oauth/callback", oauthCallbackHandler)
	}

	s := &http.Server{
		Addr:    address,
		Handler: mux,
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
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	log.Debugf(ctx, "Server stopped: %v", err)
	return err
}
