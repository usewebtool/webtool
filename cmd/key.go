package cmd

import (
	"context"

	"github.com/spf13/cobra"
)

var keyCmd = &cobra.Command{
	Use:   "key <name>",
	Short: "Send a key press (e.g. Enter, Escape, Tab, ArrowDown). Case-insensitive.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(cmd.Context(), timeoutFlag)
		defer cancel()

		return client.Key(ctx, args[0])
	},
}

func init() {
	rootCmd.AddCommand(keyCmd)
}
