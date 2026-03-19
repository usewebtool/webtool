package cmd

import (
	"github.com/spf13/cobra"
)

var tabsCmd = &cobra.Command{
	Use:   "tabs",
	Short: "List open browser tabs.",
	Long: `List open browser tabs. Tab order is arbitrary and may not match the
order you see in Chrome's tab bar. Use the index shown here with "webtool tab".`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		tabs, err := client.Tabs(cmd.Context())
		if err != nil {
			return err
		}
		for _, t := range tabs {
			if t.Active {
				cmd.Printf("%d %s %s [active]\n", t.Index, t.Title, t.URL)
			} else {
				cmd.Printf("%d %s %s\n", t.Index, t.Title, t.URL)
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(tabsCmd)
}
