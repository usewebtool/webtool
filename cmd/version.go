package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Set at build time via ldflags. See .goreleaser.yaml.
var (
	Version = "dev"
	Commit  = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version and build info",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Fprintf(cmd.OutOrStdout(), "webtool %s (%s)\n", Version, Commit)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
