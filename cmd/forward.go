package cmd

import (
	"context"

	"github.com/spf13/cobra"
)

var forwardCmd = &cobra.Command{
	Use:   "forward",
	Short: "Navigate forward in browser history.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(cmd.Context(), timeoutFlag)
		defer cancel()
		return client.Forward(ctx)
	},
}

func init() {
	rootCmd.AddCommand(forwardCmd)
}
