package main

import (
	"github.com/obot-platform/nanobot/pkg/cli"
	"github.com/obot-platform/nanobot/pkg/cmd"
)

func main() {
	cmd.Main(cli.New())
}
