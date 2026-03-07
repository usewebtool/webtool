package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var openCmd = &cobra.Command{
	Use:   "open URL",
	Short: "Navigate the browser to a URL.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := chrome.Open(cmd.Context(), args[0]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(openCmd)
}
