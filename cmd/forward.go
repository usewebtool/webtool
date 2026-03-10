package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var forwardCmd = &cobra.Command{
	Use:   "forward",
	Short: "Navigate forward in browser history.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := client.Forward(cmd.Context()); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(forwardCmd)
}
