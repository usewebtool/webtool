package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var tabsCmd = &cobra.Command{
	Use:   "tabs",
	Short: "List open browser tabs.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		tabs, err := client.Tabs(cmd.Context())
		if err != nil {
			return err
		}
		for _, t := range tabs {
			fmt.Printf("%d %s %s\n", t.Index, t.Title, t.URL)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(tabsCmd)
}
