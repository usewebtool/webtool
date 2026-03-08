package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/machinae/webtool/agent"
	"github.com/machinae/webtool/browser"
	"github.com/spf13/cobra"
)

// client is the shared daemon client for subcommands.
var client *agent.Client

var timeoutFlag time.Duration

var rootCmd = &cobra.Command{
	Use:           "webtool",
	Short:         "Fast CLI for your Chrome browser.",
	Long:          "A fast, composable CLI tool that drives a Chrome browser via Chrome DevTools Protocol. Designed for agent-driven workflows.",
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		agent.HomeDir = resolveHome()
		client = agent.NewClientWithDataDir(resolveDataDir())
		return client.EnsureRunning(cmd.Context())
	},
}

// resolveHome returns the webtool home directory from WEBTOOL_HOME or ~/.webtool.
func resolveHome() string {
	if dir := os.Getenv("WEBTOOL_HOME"); dir != "" {
		return dir
	}
	home, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Sprintf("cannot determine home directory: %v. export $WEBTOOL_HOME to set", err))
	}
	return filepath.Join(home, ".webtool")
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

func init() {
	rootCmd.PersistentFlags().DurationVar(&timeoutFlag, "timeout", 30*time.Second, "timeout for the command (e.g. 5s, 1m)")
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}
