package cli

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/obot-platform/nanobot/pkg/mcp"
	"github.com/spf13/cobra"
)

type Exec struct {
	File   string `usage:"File to read input from" default:"" short:"f"`
	Output string `usage:"Output format (json, pretty)" default:"pretty" short:"o"`
	n      *Nanobot
}

func NewExec(n *Nanobot) *Exec {
	return &Exec{
		n: n,
	}
}

func (e *Exec) Customize(cmd *cobra.Command) {
	cmd.Use = "exec [flags] NANOBOT_CONFIG TOOL_NAME|AGENT_NAME [AGENT PROMPT]"
	cmd.Short = "Execute a single tool or agent in the nanobot"
	cmd.Example = `
  # Run a tool, passing in a JSON object as input. Tools expect a JSON object as input.
  nanobot exec . server1/tool1 '{"arg1": "value1", "arg2": "value2"}'

  # Run a tool, passing in the same input as above, but using a friendly format.
  nanobot exec . server1/tool1 --arg1=value1 --arg2 value2

  # Run an agent, passing in a string as input. If the input is JSON it will be based as is.
  # Note: The current working directory of nanobot will be changed to ./other/path
  nanobot exec . agent1 "What is the weather like today?"
`
	cmd.Args = cobra.MinimumNArgs(2)
}

func writeData(result mcp.Content) error {
	data, err := base64.StdEncoding.DecodeString(result.Data)
	if err != nil {
		return err
	}
	i := 1
	for {
		filename := fmt.Sprintf("output%d.data", i)
		_, err := os.Stat(filename)
		if !errors.Is(err, fs.ErrNotExist) {
			i++
			continue
		}

		if err := os.WriteFile(filename, data, 0644); err != nil {
			return err
		}
		name := result.Type
		if result.MIMEType != "" {
			name += "(" + result.MIMEType + ")"
		}
		fmt.Printf("%s written to %s\n", name, filename)
		return nil
	}
}

func (e *Exec) Run(cmd *cobra.Command, args []string) error {
	runtime, err := e.n.GetRuntime(args[0])
	if err != nil {
		return err
	}

	ctx := runtime.WithTempSession(cmd.Context())

	result, err := runtime.CallFromCLI(ctx, args[1], args[2:]...)
	if err != nil {
		return err
	}

	if display(result, e.Output) {
		return nil
	}

	for _, out := range result.Content {
		if out.Text != "" {
			fmt.Println(out.Text)
		} else if out.Data != "" {
			if err := writeData(out); err != nil {
				return err
			}
		}
	}

	return nil
}
