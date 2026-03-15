package cmd

import (
	"context"

	"github.com/spf13/cobra"
)

var clickCmd = &cobra.Command{
	Use:     "click <selector>",
	Short:   "Click an element by backendNodeId, CSS selector, or XPath.",
	Example: "  webtool click 43821",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(cmd.Context(), timeoutFlag)
		defer cancel()

		return client.Click(ctx, args[0])
	},
}

func init() {
	rootCmd.AddCommand(clickCmd)
}
