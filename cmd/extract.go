package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var (
	extractMain bool
	extractHTML bool
)

var extractCmd = &cobra.Command{
	Use:   "extract [selector]",
	Short: "Extract page content as markdown (or HTML with --html). Default timeout 1s.",
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

		// Default to 1s for extract instead of the global 30s. CSS/XPath
		// selectors retry until the context expires, and 30s is too long
		// to wait on a typo. The user can still override with --timeout.
		timeout := timeoutFlag
		if !cmd.Flags().Changed("timeout") {
			timeout = 1 * time.Second
		}

		ctx, cancel := context.WithTimeout(cmd.Context(), timeout)
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
