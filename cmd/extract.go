package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	extractMain bool
	extractHTML bool
)

var extractCmd = &cobra.Command{
	Use:   "extract [selector]",
	Short: "Extract page content as markdown (or HTML with --html).",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		selector := ""
		if len(args) > 0 {
			selector = args[0]
		}

		if extractMain && selector != "" {
			fmt.Fprintln(os.Stderr, "--main and a selector are mutually exclusive")
			os.Exit(2)
		}

		if extractMain {
			selector = "main, [role='main']"
		}

		ctx, cancel := context.WithTimeout(cmd.Context(), timeoutFlag)
		defer cancel()

		content, err := client.Extract(ctx, selector, extractHTML)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}

		fmt.Print(content)
		return nil
	},
}

func init() {
	extractCmd.Flags().BoolVar(&extractMain, "main", false, "extract only the main content area")
	extractCmd.Flags().BoolVar(&extractHTML, "html", false, "return raw HTML instead of markdown")
	rootCmd.AddCommand(extractCmd)
}
