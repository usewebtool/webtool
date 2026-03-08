package cmd

import (
	"github.com/machinae/webtool/agent"
	"github.com/machinae/webtool/browser"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:    "_serve",
	Short:  "Run the daemon in foreground (internal use).",
	Hidden: true,
	Args:   cobra.NoArgs,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		agent.HomeDir = resolveHome()
		return nil // Skip root PersistentPreRunE — we are the daemon.
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		b := browser.New().WithChromeDataDir(resolveDataDir())
		if err := b.Connect(); err != nil {
			return err
		}
		return agent.NewServer(b).Start()
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
