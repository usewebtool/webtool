package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

var tabCmd = &cobra.Command{
	Use:     "tab <index>",
	Short:   "Switch to a tab by 1-based index from webtool tabs.",
	Long:    `Switch to a tab by its 1-based index. Run "webtool tabs" first to see available indices.`,
	Example: "  webtool tab 2",
	Args:    cobra.ExactArgs(1),
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
