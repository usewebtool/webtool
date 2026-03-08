package cmd

import (
	"github.com/machinae/webtool/agent"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the daemon.",
	Args:  cobra.NoArgs,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return nil // Don't EnsureRunning — we want to stop, not start.
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		c := agent.NewClientWithDataDir(resolveDataDir())
		c.Stop(cmd.Context()) // Ignore error — idempotent if no daemon running.
		return nil
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
