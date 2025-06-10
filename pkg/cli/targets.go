package cli

import (
	"os"
	"text/tabwriter"

	"github.com/obot-platform/nanobot/pkg/log"
	"github.com/obot-platform/nanobot/pkg/tools"
	"github.com/spf13/cobra"
)

type Targets struct {
	Nanobot   *Nanobot
	MCPServer []string `usage:"Specific MCP server name to query (default: all)" short:"s" name:"mcp-server"`
	Output    string   `usage:"Output format (json, yaml, table)" short:"o" default:"table"`
}

func NewTargets(n *Nanobot) *Targets {
	return &Targets{
		Nanobot: n,
	}
}

func (t *Targets) Customize(cmd *cobra.Command) {
	cmd.Use = "targets [flags] NANOBOT"
	cmd.Short = "List the available tools, agents, flows that can be called using \"nanobot call\"."
	cmd.Aliases = []string{"target", "t"}
	cmd.Args = cobra.ExactArgs(1)
	cmd.Example = `
  # List the tools from nanobot.yaml in the current directory
  nanobot run .
`
}

func (t *Targets) Run(cmd *cobra.Command, args []string) error {
	log.EnableMessages = false
	r, err := t.Nanobot.GetRuntime(cmd.Context(), args[0])
	if err != nil {
		return err
	}

	tools, err := r.ListTools(r.WithTempSession(cmd.Context()), tools.ListToolsOptions{
		Servers: t.MCPServer,
	})
	if err != nil {
		return err
	}

	if display(tools, t.Output) {
		return nil
	}

	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	_, err = tw.Write([]byte("TARGET\tTYPE\tDESCRIPTION\n"))
	if err != nil {
		return err
	}

	c := r.GetConfig()
	for _, tool := range tools {
		for _, t := range tool.Tools {
			target := tool.Server
			targetType := "agent"
			if _, ok := c.MCPServers[t.Name]; ok {
				targetType = "tool"
				target = target + "/" + t.Name
			} else if _, ok := c.Flows[t.Name]; ok {
				targetType = "flow"
			}

			_, _ = tw.Write([]byte(target + "\t" + targetType + "\t" + trim(t.Description) + "\n"))
		}
	}

	return tw.Flush()
}

func trim(s string) string {
	if len(s) > 70 {
		return s[:70] + "..."
	}
	return s
}
