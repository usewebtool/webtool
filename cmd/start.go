package cmd

import (
	"github.com/machinae/webtool/agent"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the daemon in the background.",
	Args:  cobra.NoArgs,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		agent.HomeDir = resolveHome()
		client = agent.NewClientWithDataDir(resolveDataDir())
		return nil // Skip root PersistentPreRunE — we are starting the daemon.
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return client.Start(cmd.Context())
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
