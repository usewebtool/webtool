package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var snapshotCmd = &cobra.Command{
	Use:   "snapshot",
	Short: "Print a text snapshot of interactive elements on the current page.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		snapshot, err := client.Snapshot(cmd.Context())
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		fmt.Print(snapshot)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(snapshotCmd)
}
