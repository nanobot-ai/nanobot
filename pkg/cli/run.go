package cli

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/obot-platform/nanobot/pkg/mcp"
	"github.com/obot-platform/nanobot/pkg/server"
	"github.com/spf13/cobra"
)

type Run struct {
	ListenAddress string `usage:"Address to listen on (ex: localhost:8099)" default:"stdio" short:"a"`
	n             *Nanobot
}

func NewRun(n *Nanobot) *Run {
	return &Run{
		n: n,
	}
}

func (r *Run) Customize(cmd *cobra.Command) {
	cmd.Use = "run [flags] NANOBOT"
	cmd.Short = "Run the nanobot with the specified config file"
	cmd.Example = `
  # Run the nanobot.yaml in the current directory
  nanobot run .

  # Run a custom config file in another directory with a custom address using HTTP instead of stdio.
  # Note: The current working directory of nanobot will be changed to ./other/path
  nanobot run ./other/path/custom.yaml --address localhost:8099
`
	cmd.Args = cobra.ExactArgs(1)
}

func (r *Run) Run(cmd *cobra.Command, args []string) error {
	runtime, err := r.n.GetRuntime(args[0])
	if err != nil {
		return err
	}

	address := r.ListenAddress
	if strings.HasPrefix("address", "http://") {
		address = strings.TrimPrefix(address, "http://")
	} else if strings.HasPrefix(address, "https://") {
		address = strings.TrimPrefix(address, "https://")
	}

	mcpServer := server.NewServer(runtime)

	if address == "stdio" {
		stdio := mcp.NewStdioServer(mcpServer)
		if err := stdio.Start(cmd.Context(), os.Stdin, os.Stdout); err != nil {
			return fmt.Errorf("failed to start stdio server: %w", err)
		}

		stdio.Wait()
		return nil
	}

	httpServer := mcp.HTTPServer{
		MessageHandler: mcpServer,
	}

	s := &http.Server{
		Addr:    address,
		Handler: &httpServer,
	}

	_, _ = fmt.Fprintf(os.Stderr, "Starting server on %s\n", address)
	context.AfterFunc(cmd.Context(), func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = s.Shutdown(ctx)
	})

	return s.ListenAndServe()
}
