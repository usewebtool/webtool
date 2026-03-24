package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/usewebtool/webtool/agent"
	"github.com/usewebtool/webtool/policy"
)

var startPolicyFlag string

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the daemon in the background.",
	Args:  cobra.NoArgs,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		agent.HomeDir = resolveHome()
		client = agent.NewClientWithDataDir(resolveDataDir())
		return nil // Skip root PersistentPreRunE — we are starting the daemon.
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		var extraArgs []string
		if startPolicyFlag != "" {
			abs, err := filepath.Abs(startPolicyFlag)
			if err != nil {
				return err
			}
			// Validate early so the user sees errors immediately.
			if _, err := policy.Load(abs); err != nil {
				return fmt.Errorf("error in policy file: %w", err)
			}
			extraArgs = append(extraArgs, "--policy", abs)
		}
		return client.Start(cmd.Context(), extraArgs...)
	},
}

func init() {
	startCmd.Flags().StringVarP(&startPolicyFlag, "policy", "p", "", "path to security policy YAML file")
	rootCmd.AddCommand(startCmd)
}
