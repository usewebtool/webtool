package cmd

import (
	"context"

	"github.com/spf13/cobra"
)

var typeCmd = &cobra.Command{
	Use:   "type <selector> <text>",
	Short: "Type text into an element by backendNodeId, CSS selector, or XPath.",
	Long: `Type text into an element. Makes a best effort to clear existing content
before typing (not guaranteed to succeed on all inputs). No need to clear manually.`,
	Example: "  webtool type 43821 \"hello world\"",
	Args:    cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(cmd.Context(), timeoutFlag)
		defer cancel()

		return client.Type(ctx, args[0], args[1])
	},
}

func init() {
	rootCmd.AddCommand(typeCmd)
}
