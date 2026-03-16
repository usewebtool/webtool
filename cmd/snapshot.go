package cmd

import (
	"fmt"

	"github.com/usewebtool/webtool/browser"
	"github.com/spf13/cobra"
)

var (
	snapshotInteractive bool
	snapshotAll         bool
)

var snapshotCmd = &cobra.Command{
	Use:   "snapshot",
	Short: "Print a text snapshot of the current page.",
	Long: `Print a text snapshot of the current page. Each line shows:
  [backendNodeId] role "name" attributes
Use the backendNodeId in action commands (click, type, etc).`,
	Example: `  webtool snapshot
  webtool snapshot -i
  webtool snapshot -a`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		mode := browser.ModeDefault
		if snapshotInteractive {
			mode = browser.ModeInteractive
		} else if snapshotAll {
			mode = browser.ModeAll
		}

		snapshot, err := client.Snapshot(cmd.Context(), mode)
		if err != nil {
			return err
		}
		fmt.Print(snapshot)
		return nil
	},
}

func init() {
	snapshotCmd.Flags().BoolVarP(&snapshotInteractive, "interactive", "i", false, "Show only interactive elements")
	snapshotCmd.Flags().BoolVarP(&snapshotAll, "all", "a", false, "Show all text content including paragraphs and static text")
	snapshotCmd.MarkFlagsMutuallyExclusive("interactive", "all")
	rootCmd.AddCommand(snapshotCmd)
}
