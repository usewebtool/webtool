package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var tabsCmd = &cobra.Command{
	Use:   "tabs",
	Short: "List open browser tabs.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		tabs, err := client.Tabs(cmd.Context())
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		for _, t := range tabs {
			fmt.Printf("%s %s %s\n", t.ID, t.Title, t.URL)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(tabsCmd)
}
