package cmd

import (
	"context"

	"github.com/spf13/cobra"
)

var hoverCmd = &cobra.Command{
	Use:     "hover <selector>",
	Short:   "Hover over an element to reveal hidden buttons or menus.",
	Long:    `Hover over an element by backendNodeId, CSS selector, or XPath. Useful for revealing hidden UI like delete buttons, edit icons, or dropdown menus that only appear on hover.`,
	Example: "  webtool hover 43821",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(cmd.Context(), timeoutFlag)
		defer cancel()

		return client.Hover(ctx, args[0])
	},
}

func init() {
	rootCmd.AddCommand(hoverCmd)
}
