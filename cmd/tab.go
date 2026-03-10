package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
)

var tabCmd = &cobra.Command{
	Use:   "tab <index>",
	Short: "Switch to a tab by its index (from webtool tabs).",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		index, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Fprintln(os.Stderr, "index must be an integer")
			os.Exit(2)
		}

		if err := client.Switch(cmd.Context(), index); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(tabCmd)
}
