package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var typeCmd = &cobra.Command{
	Use:   "type <selector> <text>",
	Short: "Type text into an element by backendNodeId, CSS selector, or XPath.",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(cmd.Context(), timeoutFlag)
		defer cancel()

		if err := client.Type(ctx, args[0], args[1]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(typeCmd)
}
