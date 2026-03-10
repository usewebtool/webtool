package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var backCmd = &cobra.Command{
	Use:   "back",
	Short: "Navigate back in browser history.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := client.Back(cmd.Context()); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(backCmd)
}
