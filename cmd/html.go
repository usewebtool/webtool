package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var htmlCmd = &cobra.Command{
	Use:   "html [selector]",
	Short: "Extract page content as HTML. Alias for extract --html.",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		selector := ""
		if len(args) > 0 {
			selector = args[0]
		}

		ctx, cancel := context.WithTimeout(cmd.Context(), timeoutFlag)
		defer cancel()

		content, err := client.Extract(ctx, selector, true)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}

		fmt.Print(content)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(htmlCmd)
}
