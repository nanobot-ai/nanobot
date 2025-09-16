package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/nanobot-ai/nanobot/pkg/api"
	"github.com/nanobot-ai/nanobot/pkg/auth"
	"github.com/nanobot-ai/nanobot/pkg/cmd"
	"github.com/nanobot-ai/nanobot/pkg/complete"
	"github.com/nanobot-ai/nanobot/pkg/config"
	"github.com/nanobot-ai/nanobot/pkg/llm"
	"github.com/nanobot-ai/nanobot/pkg/llm/anthropic"
	"github.com/nanobot-ai/nanobot/pkg/llm/responses"
	"github.com/nanobot-ai/nanobot/pkg/log"
	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/runtime"
	"github.com/nanobot-ai/nanobot/pkg/server"
	"github.com/nanobot-ai/nanobot/pkg/session"
	"github.com/nanobot-ai/nanobot/pkg/types"
	"github.com/nanobot-ai/nanobot/pkg/version"
	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

func New() *cobra.Command {
	n := &Nanobot{}

	root := cmd.Command(n,
		NewCall(n),
		NewTargets(n),
		NewSessions(n),
		NewRun(n))
	return root
}

type Nanobot struct {
	Debug            bool              `usage:"Enable debug logging"`
	Trace            bool              `usage:"Enable trace logging"`
	Env              []string          `usage:"Environment variables to set in the form of KEY=VALUE, or KEY to load from current environ" short:"e"`
	EnvFile          string            `usage:"Path to the environment file (default: ./nanobot.env)"`
	EmptyEnv         bool              `usage:"Do not load environment variables from the environment by default"`
	DefaultModel     string            `usage:"Default model to use for completions" default:"gpt-4.1" env:"NANOBOT_DEFAULT_MODEL" name:"default-model"`
	OpenAIAPIKey     string            `usage:"OpenAI API key" env:"OPENAI_API_KEY" name:"openai-api-key"`
	OpenAIBaseURL    string            `usage:"OpenAI API URL" env:"OPENAI_BASE_URL" name:"openai-base-url"`
	OpenAIHeaders    map[string]string `usage:"OpenAI API headers" env:"OPENAI_HEADERS" name:"openai-headers"`
	AnthropicAPIKey  string            `usage:"Anthropic API key" env:"ANTHROPIC_API_KEY" name:"anthropic-api-key"`
	AnthropicBaseURL string            `usage:"Anthropic API URL" env:"ANTHROPIC_BASE_URL" name:"anthropic-base-url"`
	AnthropicHeaders map[string]string `usage:"Anthropic API headers" env:"ANTHROPIC_HEADERS" name:"anthropic-headers"`
	MaxConcurrency   int               `usage:"The maximum number of concurrent tasks in a parallel loop" default:"10"`
	Chdir            string            `usage:"Change directory to this path before running the nanobot" default:"." short:"C"`
	State            string            `usage:"Path to the state file" default:"${XDG_CONFIG_HOME}/nanobot/state.db"`

	env map[string]string
}

func ensureDirectoryForDSN(dsn string) error {
	dsnFile, _, _ := strings.Cut(dsn, "?")
	dsnFile = strings.TrimPrefix(dsnFile, "file:")
	if !strings.HasSuffix(dsnFile, ".db") {
		return nil
	}

	dir := filepath.Dir(dsnFile)
	if dir == "." {
		return nil
	}

	_, err := os.Stat(dir)
	if !errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	return os.MkdirAll(dir, 0o700)
}

func (n *Nanobot) DSN() string {
	dsn := os.Expand(n.State, func(s string) string {
		if s == "XDG_CONFIG_HOME" {
			config, err := os.UserConfigDir()
			if err != nil {
				log.Fatalf(context.Background(), "Failed to get user config directory: %v", err)
			}
			return config
		}
		return os.Getenv(s)
	})

	if err := ensureDirectoryForDSN(dsn); err != nil {
		log.Fatalf(context.Background(), "Failed to ensure directory for state file %s: %v", dsn, err)
	}

	return dsn
}

func (n *Nanobot) Customize(cmd *cobra.Command) {
	cmd.Short = "Nanobot: Build, Run, Share AI Agents"
	cmd.Example = `
	# Run the Welcome bot
	nanobot run nanobot-ai/welcome
`
	cmd.CompletionOptions.HiddenDefaultCmd = true
	cmd.Version = version.Get().String()
}

func (n *Nanobot) PersistentPre(cmd *cobra.Command, _ []string) error {
	if n.Chdir != "." {
		if err := os.Chdir(n.Chdir); err != nil {
			return fmt.Errorf("failed to change directory to %s: %w", n.Chdir, err)
		}
	}

	if n.Debug {
		log.EnableMessages = true
		log.DebugLog = true
	}

	if n.Trace {
		log.EnableMessages = true
		log.EnableProgress = true
		log.DebugLog = true
	}

	for _, sub := range cmd.Commands() {
		if sub.Name() == "help" {
			sub.Hidden = true
			sub.Use = " help"
		}
	}
	// Don't need to do anything here, this is just to ensure the env vars get parsed and set always.
	// To be honest don't know why this is needed, but it is.
	return nil
}

func display(obj any, format string) bool {
	if format == "json" {
		data, _ := json.MarshalIndent(obj, "", "  ")
		fmt.Println(string(data))
		return true
	} else if format == "yaml" {
		data, _ := yaml.Marshal(obj)
		fmt.Println(string(data))
		return true
	}
	return false
}

func (n *Nanobot) llmConfig() llm.Config {
	return llm.Config{
		DefaultModel: n.DefaultModel,
		Responses: responses.Config{
			APIKey:  n.OpenAIAPIKey,
			BaseURL: n.OpenAIBaseURL,
			Headers: n.OpenAIHeaders,
		},
		Anthropic: anthropic.Config{
			APIKey:  n.AnthropicAPIKey,
			BaseURL: n.AnthropicBaseURL,
			Headers: n.AnthropicHeaders,
		},
	}
}

func (n *Nanobot) loadEnv() (map[string]string, error) {
	if n.env != nil {
		return n.env, nil
	}

	env := map[string]string{}
	cwd, err := os.Getwd()
	if err == nil {
		env["PWD"] = cwd
		env["CWD"] = cwd
	}

	if !n.EmptyEnv {
		for _, kv := range os.Environ() {
			k, v, _ := strings.Cut(kv, "=")
			env[k] = v
		}
	}

	defaultFile := n.EnvFile == ""
	if defaultFile {
		n.EnvFile = "./nanobot.env"
	}

	data, err := os.ReadFile(n.EnvFile)
	if errors.Is(err, fs.ErrNotExist) && defaultFile {
	} else if err != nil {
		return nil, err
	} else {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			k, v, _ := strings.Cut(line, "=")
			env[k] = v
		}
	}

	if _, ok := env["NANOBOT_MCP"]; !ok {
		env["NANOBOT_MCP"] = "true"
	}

	for _, kv := range n.Env {
		k, v, ok := strings.Cut(kv, "=")
		if !ok {
			v = os.Getenv(k)
		}
		env[k] = v
	}

	n.env = env
	return env, nil
}

func (n *Nanobot) ReadConfig(ctx context.Context, cfgPath string, opts ...runtime.Options) (*types.Config, error) {
	cfg, _, err := config.Load(ctx, cfgPath, complete.Complete(opts...).Profiles...)
	return cfg, err
}

func (n *Nanobot) ReadConfigType(ctx context.Context, cfg *types.Config, opts ...runtime.Options) (*types.Config, error) {
	cfg, _, err := config.LoadFromConfig(ctx, *cfg, complete.Complete(opts...).Profiles...)
	return cfg, err
}

func (n *Nanobot) GetRuntime(opts ...runtime.Options) (*runtime.Runtime, error) {
	return runtime.NewRuntime(n.llmConfig(), opts...)
}

func (n *Nanobot) Run(cmd *cobra.Command, _ []string) error {
	return cmd.Help()
}

func (n *Nanobot) runMCP(ctx context.Context, config types.ConfigFactory, runt *runtime.Runtime, oauthCallbackHandler mcp.CallbackServer, l net.Listener, listenAddress string, healthzPath string, startUI bool) error {
	env, err := n.loadEnv()
	if err != nil {
		return fmt.Errorf("failed to load environment: %w", err)
	}

	address := listenAddress
	if strings.HasPrefix("address", "http://") {
		address = strings.TrimPrefix(address, "http://")
	} else if strings.HasPrefix(address, "https://") {
		return fmt.Errorf("https:// is not supported, use http:// instead")
	}

	sessionManager, err := session.NewManager(n.DSN())
	if err != nil {
		return err
	}

	var mcpServer mcp.MessageHandler = server.NewServer(runt, config, sessionManager)

	if address == "stdio" {
		stdio := mcp.NewStdioServer(env, mcpServer)
		if err := stdio.Start(ctx, os.Stdin, os.Stdout); err != nil {
			return fmt.Errorf("failed to start stdio server: %w", err)
		}

		stdio.Wait()
		return nil
	}

	httpServer := mcp.NewHTTPServer(env, mcpServer, mcp.HTTPServerOptions{
		SessionStore: sessionManager,
		HealthzPath:  healthzPath,
	})

	mux := http.NewServeMux()
	if oauthCallbackHandler != nil {
		mux.Handle("/oauth/callback", oauthCallbackHandler)
	}
	if startUI {
		mux.Handle("/api/", api.Handler(sessionManager))
		mux.Handle("/", session.UISession(httpServer, sessionManager))
	} else {
		mux.Handle("/", httpServer)
	}

	authCfg, err := config(ctx, "")
	if err != nil {
		return err
	}

	handler, err := auth.Wrap(env, authCfg, n.DSN(), mux)
	if err != nil {
		return fmt.Errorf("failed to setup auth: %w", err)
	}

	s := &http.Server{
		Addr:    address,
		Handler: handler,
	}

	context.AfterFunc(ctx, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = s.Shutdown(ctx)
	})

	if l == nil {
		_, _ = fmt.Fprintf(os.Stderr, "Starting server on http://%s\n", address)
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
