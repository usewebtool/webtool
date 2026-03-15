package cmd

import (
	"context"

	"github.com/spf13/cobra"
)

var waitCmd = &cobra.Command{
	Use:   "wait <duration|selector>",
	Short: "Wait for a duration (e.g. 2s) or until an element exists.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(cmd.Context(), timeoutFlag)
		defer cancel()

		return client.Wait(ctx, args[0])
	},
}

func init() {
	rootCmd.AddCommand(waitCmd)
}
