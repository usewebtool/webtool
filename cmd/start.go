package cmd

import (
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the daemon in the background.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil // PersistentPreRunE already called EnsureRunning.
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
