package cli

import (
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/nanobot-ai/nanobot/pkg/session"
	"github.com/spf13/cobra"
)

type Sessions struct {
	Nanobot *Nanobot
	Output  string `usage:"Output format (json, yaml, table)" short:"o" default:"table"`
	Long    bool   `usage:"Show long output with full session description" short:"l"`
}

func NewSessions(n *Nanobot) *Sessions {
	return &Sessions{
		Nanobot: n,
	}
}

func (t *Sessions) Customize(cmd *cobra.Command) {
	cmd.Use = "sessions [flags]"
	cmd.Short = "List all existing sessions"
	cmd.Aliases = []string{"session", "s"}
	cmd.Args = cobra.NoArgs
}

func (t *Sessions) Run(cmd *cobra.Command, args []string) error {
	store, err := session.NewStoreFromDSN(t.Nanobot.DSN())
	if err != nil {
		return err
	}

	sessions, err := store.List()
	if err != nil {
		return err
	}

	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	_, err = tw.Write([]byte("ID\tDATE\tDESCRIPTION\n"))
	if err != nil {
		return err
	}

	for _, session := range sessions {
		id := session.SessionID
		if !t.Long {
			id = strings.Split(id, "-")[0] // Trim the session ID to 8 characters
		}
		_, _ = tw.Write([]byte(id + "\t" + session.UpdatedAt.Format(time.RFC3339) + "\t" + trim(session.Description) + "\n"))
	}

	return tw.Flush()
}
