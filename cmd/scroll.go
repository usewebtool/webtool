package cmd

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

var scrollUp bool

var scrollCmd = &cobra.Command{
	Use:   "scroll [pixels]",
	Short: "Scroll the page up or down.",
	Long: `Scroll the page by a number of pixels. Defaults to one viewport height.
Scrolls down unless --up is specified. Useful for triggering lazy-loaded content.`,
	Example: `  webtool scroll
  webtool scroll 300
  webtool scroll --up
  webtool scroll 500 --up`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		pixels := 0
		if len(args) > 0 {
			p, err := strconv.Atoi(args[0])
			if err != nil {
				return err
			}
			if p < 0 {
				return fmt.Errorf("pixels must be non-negative (use --up to scroll up)")
			}
			pixels = p
		}

		ctx, cancel := context.WithTimeout(cmd.Context(), timeoutFlag)
		defer cancel()

		return client.Scroll(ctx, pixels, scrollUp)
	},
}

func init() {
	scrollCmd.Flags().BoolVar(&scrollUp, "up", false, "scroll up instead of down")
	rootCmd.AddCommand(scrollCmd)
}
