package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var clickCmd = &cobra.Command{
	Use:   "click <selector>",
	Short: "Click an element by backendNodeId, CSS selector, or XPath.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(cmd.Context(), timeoutFlag)
		defer cancel()

		if err := client.Click(ctx, args[0]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(clickCmd)
}
