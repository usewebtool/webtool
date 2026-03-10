package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var selectCmd = &cobra.Command{
	Use:   "select <selector> <text>",
	Short: "Select a dropdown option by visible text.",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := client.Select(cmd.Context(), args[0], args[1]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(selectCmd)
}
