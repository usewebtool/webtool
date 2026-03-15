package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

var (
	extractMain bool
	extractHTML bool
)

var extractCmd = &cobra.Command{
	Use:   "extract [selector]",
	Short: "Extract page content as markdown or HTML.",
	Long: `Extract page content as markdown (default) or raw HTML (with --html).
Use --main to extract only the main content area. Unlike snapshot, this returns
readable page content rather than an interactive element tree.`,
	Example: `  webtool extract
  webtool extract --main
  webtool extract --html 43821`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		selector := ""
		if len(args) > 0 {
			selector = args[0]
		}

		if extractMain && selector != "" {
			return fmt.Errorf("--main and a selector are mutually exclusive")
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
			return err
		}
		fmt.Println(content)
		return nil
	},
}

func init() {
	extractCmd.Flags().BoolVar(&extractMain, "main", false, "extract only the main content area")
	extractCmd.Flags().BoolVar(&extractHTML, "html", false, "return raw HTML instead of markdown")
	rootCmd.AddCommand(extractCmd)
}
