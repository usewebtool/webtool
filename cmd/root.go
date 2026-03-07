package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "webtool",
	Short: "Fast CLI for your Chrome browser.",
	Long:  "A fast, composable CLI tool that drives a Chrome browser via Chrome DevTools Protocol. Designed for agent-driven workflows.",
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}
