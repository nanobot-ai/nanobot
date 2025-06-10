package main

import (
	"os"

	"github.com/obot-platform/nanobot/pkg/cli"
	"github.com/obot-platform/nanobot/pkg/cmd"
	"github.com/obot-platform/nanobot/pkg/supervise"
)

func main() {
	if len(os.Args) > 2 && os.Args[1] == "_exec" {
		if err := supervise.Daemon(); err != nil {
			os.Exit(1)
		}
		return
	}
	cmd.Main(cli.New())
}
