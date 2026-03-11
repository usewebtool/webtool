package cmd

import (
	"fmt"
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
			return fmt.Errorf("index must be an integer")
		}
		return client.Switch(cmd.Context(), index)
	},
}

func init() {
	rootCmd.AddCommand(tabCmd)
}
