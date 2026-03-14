package cmd

import (
	"context"

	"github.com/spf13/cobra"
)

var hoverCmd = &cobra.Command{
	Use:   "hover <selector>",
	Short: "Hover over an element by backendNodeId, CSS selector, or XPath.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(cmd.Context(), timeoutFlag)
		defer cancel()

		return client.Hover(ctx, args[0])
	},
}

func init() {
	rootCmd.AddCommand(hoverCmd)
}
