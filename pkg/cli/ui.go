package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/runtime"
	"github.com/nanobot-ai/nanobot/pkg/types"
	"github.com/spf13/cobra"
)

type UI struct {
	ListenAddress string   `usage:"Address to listen on (ex: localhost:8099)" default:"localhost:8080" short:"a"`
	Port          string   `usage:"Port to listen on for stdio" default:"8099"`
	HealthzPath   string   `usage:"Path to serve healthz on"`
	Roots         []string `usage:"Roots to expose the MCP server in the form of name:directory" short:"r"`
	Input         string   `usage:"Input file for the prompt" default:"" short:"f"`
	Session       string   `usage:"Session ID to resume" default:"" short:"s"`
	n             *Nanobot
}

func NewUI(n *Nanobot) *UI {
	return &UI{
		n: n,
	}
}

func (r *UI) Customize(cmd *cobra.Command) {
	cmd.Use = "ui [flags] [NANOBOT]"
	cmd.Short = "Run the UI for the nanobot"
	cmd.Example = `
  # UI the nanobot.yaml in the current directory
  nanobot ui .

  # UI the nanobot.yaml in the GitHub repo github.com/example/nanobot
  nanobot ui example/nanobot

  # UI the nanobot.yaml at the URL
  nanobot run https://....
`
}

func (r *UI) getRoots() ([]mcp.Root, error) {
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

func (r *UI) reload(ctx context.Context, client *mcp.Client, cfgPath string, runtimeOpt runtime.Options) error {
	_, err := r.n.ReadConfig(ctx, cfgPath, runtimeOpt)
	if err != nil {
		return fmt.Errorf("failed to reload config: %w", err)
	}

	return client.Session.Exchange(ctx, "initialize", mcp.InitializeRequest{}, &mcp.InitializeResult{})
}

func (r *UI) Run(cmd *cobra.Command, args []string) (err error) {
	var (
		runtimeOpt = runtime.Options{
			Profiles: []string{"nanobot.ui"},
		}
		config *types.Config
	)

	roots, err := r.getRoots()
	if err != nil {
		return err
	}

	cfgPath := "nanobot.default"
	if len(args) > 0 {
		cfgPath = args[0]
	}

	if strings.HasPrefix(cfgPath, "http://") || strings.HasPrefix(cfgPath, "https://") {
		r.n.Env = append(r.n.Env, types.AgentPassthroughEnv+"=true")
		config = &types.Config{
			Publish: types.Publish{
				Entrypoint: []string{"agent/chat"},
			},
			MCPServers: map[string]mcp.Server{
				"agent": {
					BaseURL: cfgPath,
				},
			},
		}
	}

	if config == nil {
		config, err = r.n.ReadConfig(cmd.Context(), cfgPath, runtimeOpt)
		if err != nil {
			return fmt.Errorf("failed to read config file %q: %w", args[0], err)
		}
	} else {
		config, err = r.n.ReadConfigType(cmd.Context(), config, runtimeOpt)
		if err != nil {
			return fmt.Errorf("failed to read config for URL %q: %w", args[0], err)
		}
	}

	runtimeOpt.Roots = roots
	runtimeOpt.MaxConcurrency = r.n.MaxConcurrency

	runtime, err := r.n.GetRuntime(runtimeOpt, runtime.Options{
		DSN: r.n.DSN(),
	})
	if err != nil {
		return err
	}

	return r.n.runMCP(cmd.Context(), *config, runtime, nil, nil, r.ListenAddress, r.HealthzPath, true)
}
