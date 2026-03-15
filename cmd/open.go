package cmd

import (
	"context"

	"github.com/spf13/cobra"
)

var openNewTab bool

var openCmd = &cobra.Command{
	Use:   "open URL",
	Short: "Navigate the browser to a URL.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(cmd.Context(), timeoutFlag)
		defer cancel()

		return client.Open(ctx, args[0], openNewTab)
	},
}

func init() {
	openCmd.Flags().BoolVar(&openNewTab, "new", false, "open in a new tab")
	rootCmd.AddCommand(openCmd)
}
