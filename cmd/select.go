package cmd

import (
	"context"

	"github.com/spf13/cobra"
)

var selectCmd = &cobra.Command{
	Use:   "select <selector> <text>",
	Short: "Select a dropdown option by visible text.",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(cmd.Context(), timeoutFlag)
		defer cancel()
		return client.Select(ctx, args[0], args[1])
	},
}

func init() {
	rootCmd.AddCommand(selectCmd)
}
