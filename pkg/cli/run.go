package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nanobot-ai/nanobot/pkg/confirm"
	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/runtime"
	"github.com/nanobot-ai/nanobot/pkg/types"
	"github.com/spf13/cobra"
)

type Serve struct {
	ListenAddress string   `usage:"Address to listen on" default:"localhost:8080" short:"a"`
	DisableUI     bool     `usage:"Disable the UI"`
	HealthzPath   string   `usage:"Path to serve healthz on"`
	Roots         []string `usage:"Roots to expose the MCP server in the form of name:directory" short:"r"`
	n             *Nanobot
}

func NewRun(n *Nanobot) *Serve {
	return &Serve{
		n: n,
	}
}

func (r *Serve) Customize(cmd *cobra.Command) {
	cmd.Use = "run [flags] NANOBOT [PROMPT]"
	cmd.Short = "Serve the nanobot with the specified config file"
	cmd.Example = `
  # Serve the nanobot.yaml in the current directory
  nanobot serve .

  # Serve the nanobot.yaml in the GitHub repo github.com/example/nanobot
  nanobot serve example/nanobot

  # Serve the nanobot.yaml at the URL
  nanobot serve https://....
`
}

func (r *Serve) getRoots() ([]mcp.Root, error) {
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

func (r *Serve) reload(ctx context.Context, client *mcp.Client, cfgPath string, runtimeOpt runtime.Options) error {
	_, err := r.n.ReadConfig(ctx, cfgPath, runtimeOpt)
	if err != nil {
		return fmt.Errorf("failed to reload config: %w", err)
	}

	return client.Session.Exchange(ctx, "initialize", mcp.InitializeRequest{}, &mcp.InitializeResult{})
}

func (r *Serve) Run(cmd *cobra.Command, args []string) (err error) {
	roots, err := r.getRoots()
	if err != nil {
		return err
	}

	callbackHandler := mcp.NewCallbackServer(confirm.New())
	runtimeOpt := runtime.Options{
		Roots:           roots,
		MaxConcurrency:  r.n.MaxConcurrency,
		CallbackHandler: callbackHandler,
	}

	cfgPath := "nanobot.default"
	if len(args) > 0 {
		cfgPath = args[0]
	}

	cfgFactory := types.ConfigFactory(func(ctx context.Context, profiles string) (types.Config, error) {
		optCopy := runtimeOpt
		if profiles != "" {
			optCopy.Profiles = append(optCopy.Profiles, strings.Split(profiles, ",")...)
		}
		cfg, err := r.n.ReadConfig(cmd.Context(), cfgPath, optCopy)
		if err != nil {
			return types.Config{}, err
		}
		return *cfg, nil
	})

	_, err = cfgFactory(cmd.Context(), "")
	if err != nil {
		return fmt.Errorf("failed to read config file %q: %w", args[0], err)
	}

	runtime, err := r.n.GetRuntime(runtimeOpt, runtime.Options{
		OAuthRedirectURL: "http://" + strings.Replace(r.ListenAddress, "127.0.0.1", "localhost", 1) + "/oauth/callback",
		DSN:              r.n.DSN(),
	})
	if err != nil {
		return err
	}

	return r.n.runMCP(cmd.Context(), cfgFactory, runtime, callbackHandler, r.ListenAddress, r.HealthzPath, !r.DisableUI)
}
