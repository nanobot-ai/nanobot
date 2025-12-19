package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/nanobot-ai/nanobot/pkg/auth"
	"github.com/nanobot-ai/nanobot/pkg/confirm"
	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/mcp/auditlogs"
	"github.com/nanobot-ai/nanobot/pkg/printer"
	"github.com/nanobot-ai/nanobot/pkg/runtime"
	"github.com/nanobot-ai/nanobot/pkg/types"
	"github.com/spf13/cobra"
)

type Run struct {
	Auth
	ListenAddress                string            `usage:"Address to listen on" default:"localhost:8080" short:"a"`
	DisableUI                    bool              `usage:"Disable the UI"`
	ForceFetchToolList           bool              `usage:"Always fetch tools when listing instead of using session cache"`
	HealthzPath                  string            `usage:"Path to serve healthz on"`
	AuditLogSendURL              string            `usage:"URL to send audit logs to"`
	AuditLogToken                string            `usage:"Token to send audit logs with"`
	AuditLogMetadata             map[string]string `usage:"Metadata to send with audit logs"`
	AuditLogBatchSize            int               `usage:"Batch size for sending audit logs" default:"1000"`
	AuditLogFlushIntervalSeconds int               `usage:"Interval for flushing audit logs" default:"5"`
	Roots                        []string          `usage:"Roots to expose the MCP server in the form of name:directory" short:"r"`
	n                            *Nanobot
}

type Auth auth.Auth

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
		TokenExchangeEndpoint:     r.Auth.OAuthTokenURL,
		TokenExchangeClientID:     r.Auth.OAuthClientID,
		TokenExchangeClientSecret: r.Auth.OAuthClientSecret,
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

	var auditLogCollector *auditlogs.Collector
	if r.AuditLogSendURL != "" {
		auditLogCollector = auditlogs.NewCollector(r.AuditLogSendURL, r.AuditLogToken, r.AuditLogBatchSize, time.Duration(r.AuditLogFlushIntervalSeconds)*time.Second, r.AuditLogMetadata)
		defer auditLogCollector.Close()
	}

	runtime, err := r.n.GetRuntime(runtimeOpt, runtime.Options{
		OAuthRedirectURL:  "http://" + strings.Replace(r.ListenAddress, "127.0.0.1", "localhost", 1) + "/oauth/callback",
		DSN:               r.n.DSN(),
		AuditLogCollector: auditLogCollector,
	})
	if err != nil {
		return err
	}

	return r.n.runMCP(cmd.Context(), cfgFactory, runtime, callbackHandler, auditLogCollector, mcpOpts{
		Auth:               auth.Auth(r.Auth),
		ListenAddress:      r.ListenAddress,
		HealthzPath:        r.HealthzPath,
		ForceFetchToolList: r.ForceFetchToolList,
		StartUI:            !r.DisableUI,
	})
}
