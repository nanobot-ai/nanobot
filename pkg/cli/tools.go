package cli

import (
	"os"
	"text/tabwriter"

	"github.com/obot-platform/nanobot/pkg/log"
	"github.com/obot-platform/nanobot/pkg/tools"
	"github.com/spf13/cobra"
)

type Tools struct {
	Nanobot *Nanobot
	Server  []string `usage:"Specific MCP server name to query (default: all)" short:"s"`
	Tool    []string `usage:"Specific tool name to query (default: all)" short:"t"`
	Output  string   `usage:"Output format (json, yaml, table)" short:"o" default:"table"`
}

func NewTools(n *Nanobot) *Tools {
	return &Tools{
		Nanobot: n,
	}
}

func (t *Tools) Customize(cmd *cobra.Command) {
	cmd.Short = "List tools the available tools and agents internal to the nanobot"
	cmd.Aliases = []string{"tool", "t"}
	cmd.Args = cobra.ExactArgs(1)
}

func (t *Tools) Run(cmd *cobra.Command, args []string) error {
	log.EnableMessages = false
	r, err := t.Nanobot.GetRuntime(args[0])
	if err != nil {
		return err
	}

	tools, err := r.ListTools(r.WithTempSession(cmd.Context()), tools.ListToolsOptions{
		Servers: t.Server,
		Tools:   t.Tool,
	})
	if err != nil {
		return err
	}

	if display(tools, t.Output) {
		return nil
	}

	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	_, err = tw.Write([]byte("SERVER\tTOOL\tDESCRIPTION\n"))
	if err != nil {
		return err
	}

	for _, tool := range tools {
		for _, t := range tool.Tools {
			_, err := tw.Write([]byte(tool.Server + "\t" + t.Name + "\t" + trim(t.Description) + "\n"))
			if err != nil {
				return err
			}
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
