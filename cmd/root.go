package cmd

import (
	"fmt"
	"os"

	"github.com/machinae/webtool/browser"
	"github.com/spf13/cobra"
)

var chrome *browser.Browser

var rootCmd = &cobra.Command{
	Use:   "webtool",
	Short: "Fast CLI for your Chrome browser.",
	Long:  "A fast, composable CLI tool that drives a Chrome browser via Chrome DevTools Protocol. Designed for agent-driven workflows.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		chrome, err = newBrowser()
		return err
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		return chrome.Save()
	},
}

// newBrowser creates and configures a Browser from environment and persisted state.
func newBrowser() (*browser.Browser, error) {
	b := browser.New()

	// DataDir override must come before Load so we read from the right place.
	if dir := os.Getenv("WEBTOOL_HOME"); dir != "" {
		b = b.WithDataDir(dir)
	}

	if err := b.Load(); err != nil {
		return nil, err
	}

	// Env vars override persisted state.
	if dir := os.Getenv("WEBTOOL_CHROME_DATA_DIR"); dir != "" {
		b = b.WithChromeDataDir(dir)
	}

	return b, nil
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}
