package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nanobot-ai/nanobot/pkg/confirm"
	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/printer"
	"github.com/nanobot-ai/nanobot/pkg/runtime"
	"github.com/nanobot-ai/nanobot/pkg/types"
	"github.com/spf13/cobra"
)

type Run struct {
	ListenAddress             string   `usage:"Address to listen on" default:"localhost:8080" short:"a"`
	DisableUI                 bool     `usage:"Disable the UI"`
	HealthzPath               string   `usage:"Path to serve healthz on"`
	TrustedIssuer             string   `usage:"Trusted issuer for JWT tokens"`
	JWKS                      string   `usage:"Base64 encoded JWKS blob for validating JWT tokens"`
	TrustedAudiences          []string `usage:"Trusted audiences for JWT tokens"`
	TokenExchangeEndpoint     string   `usage:"Endpoint for token exchange"`
	TokenExchangeClientID     string   `usage:"Client ID for token exchange"`
	TokenExchangeClientSecret string   `usage:"Client secret for token exchange"`
	Roots                     []string `usage:"Roots to expose the MCP server in the form of name:directory" short:"r"`
	n                         *Nanobot
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

func (r *Run) Run(cmd *cobra.Command, args []string) (err error) {
	if (r.TrustedIssuer != "") != (len(r.TrustedAudiences) != 0) {
		return fmt.Errorf("trusted issuer and audience must be set together")
	}

	roots, err := r.getRoots()
	if err != nil {
		return err
	}

	callbackHandler := mcp.NewCallbackServer(confirm.New())
	runtimeOpt := runtime.Options{
		Roots:                     roots,
		MaxConcurrency:            r.n.MaxConcurrency,
		CallbackHandler:           callbackHandler,
		TokenExchangeEndpoint:     r.TokenExchangeEndpoint,
		TokenExchangeClientID:     r.TokenExchangeClientID,
		TokenExchangeClientSecret: r.TokenExchangeClientSecret,
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

	once, err := cfgFactory(cmd.Context(), "")
	if err != nil {
		return fmt.Errorf("failed to read config file %q: %w", args[0], err)
	}

	cfg, _ := json.MarshalIndent(once, "", "  ")
	printer.Prefix("config", string(cfg))

	runtime, err := r.n.GetRuntime(runtimeOpt, runtime.Options{
		OAuthRedirectURL: "http://" + strings.Replace(r.ListenAddress, "127.0.0.1", "localhost", 1) + "/oauth/callback",
		DSN:              r.n.DSN(),
	})
	if err != nil {
		return err
	}

	return r.n.runMCP(cmd.Context(), cfgFactory, runtime, callbackHandler, mcpOpts{
		TrustedIssuer:    r.TrustedIssuer,
		JWKS:             r.JWKS,
		TrustedAudiences: r.TrustedAudiences,
		ListenAddress:    r.ListenAddress,
		HealthzPath:      r.HealthzPath,
		StartUI:          !r.DisableUI,
	})
}
