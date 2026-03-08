package cmd

import (
	"fmt"
	"os"

	"github.com/machinae/webtool/agent"
	"github.com/machinae/webtool/browser"
	"github.com/spf13/cobra"
)

// client is the shared daemon client for subcommands.
var client *agent.Client

var rootCmd = &cobra.Command{
	Use:   "webtool",
	Short: "Fast CLI for your Chrome browser.",
	Long:  "A fast, composable CLI tool that drives a Chrome browser via Chrome DevTools Protocol. Designed for agent-driven workflows.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		client = agent.NewClientWithDataDir(resolveDataDir())
		return client.EnsureRunning(cmd.Context())
	},
}

// resolveDataDir returns the Chrome data directory from the environment or OS default.
// Panics if the OS is unsupported.
func resolveDataDir() string {
	if dir := os.Getenv("WEBTOOL_CHROME_DATA_DIR"); dir != "" {
		return dir
	}
	dir, err := browser.DefaultChromeUserDataDir()
	if err != nil {
		panic(fmt.Sprintf("unsupported OS: %v", err))
	}
	return dir
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}
