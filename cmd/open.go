package cmd

import (
	"context"

	"github.com/spf13/cobra"
)

var openCmd = &cobra.Command{
	Use:   "open URL",
	Short: "Navigate the browser to a URL.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(cmd.Context(), timeoutFlag)
		defer cancel()

		return client.Open(ctx, args[0])
	},
}

func init() {
	rootCmd.AddCommand(openCmd)
}
