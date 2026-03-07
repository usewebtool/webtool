package cmd

import (
	"fmt"
	"os"

	"github.com/machinae/webtool/browser"
	"github.com/spf13/cobra"
)

// chrome is the shared browser instance for subcommands.
// This will be replaced by an agent.Client in Phase 5 (Daemon).
var chrome *browser.Browser

var rootCmd = &cobra.Command{
	Use:   "webtool",
	Short: "Fast CLI for your Chrome browser.",
	Long:  "A fast, composable CLI tool that drives a Chrome browser via Chrome DevTools Protocol. Designed for agent-driven workflows.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		chrome = browser.New()
		if dir := os.Getenv("WEBTOOL_CHROME_DATA_DIR"); dir != "" {
			chrome = chrome.WithChromeDataDir(dir)
		}
		return nil
	},
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}
