package cli

import (
	"os"

	"github.com/obot-platform/nanobot/pkg/chat"
	"github.com/spf13/cobra"
)

type Call struct {
	File   string `usage:"File to read input from" default:"" short:"f"`
	Output string `usage:"Output format (json, pretty)" default:"pretty" short:"o"`
	n      *Nanobot
}

func NewCall(n *Nanobot) *Call {
	return &Call{
		n: n,
	}
}

func (e *Call) Customize(cmd *cobra.Command) {
	cmd.Use = "call [flags] NANOBOT_CONFIG TARGET_NAME [INPUT...]"
	cmd.Short = "Call a single tool, agent, or flow in the nanobot. Use \"nanobot targets\" to list available targets."
	cmd.Example = `
  # Run a tool, passing in a JSON object as input. Tools expect a JSON object as input.
  nanobot call . server1/tool1 '{"arg1": "value1", "arg2": "value2"}'

  # Run a tool, passing in the same input as above, but using a friendly format.
  nanobot call . server1/tool1 --arg1=value1 --arg2 value2

  # Run an agent, passing in a string as input. If the input is JSON it will be based as is.
  nanobot call . agent1 "What is the weather like today?"
`
	cmd.Args = cobra.MinimumNArgs(2)
	cmd.Flags().SetInterspersed(false)
}

func (e *Call) Run(cmd *cobra.Command, args []string) error {
	runtime, err := e.n.GetRuntime(cmd.Context(), args[0])
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

	return chat.PrintResult(os.Stdout, result)
}
