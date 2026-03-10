package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var reloadCmd = &cobra.Command{
	Use:   "reload",
	Short: "Reload the current page.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := client.Reload(cmd.Context()); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(reloadCmd)
}
