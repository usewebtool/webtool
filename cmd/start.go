package cmd

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

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
			src := startPolicyFlag
			if !strings.HasPrefix(src, "http://") && !strings.HasPrefix(src, "https://") {
				abs, err := filepath.Abs(src)
				if err != nil {
					return err
				}
				src = abs
			}
			// Validate early so the user sees errors immediately.
			ctx, cancel := context.WithTimeout(cmd.Context(), timeoutFlag)
			defer cancel()
			if _, err := policy.Load(ctx, src); err != nil {
				return fmt.Errorf("error in policy file: %w", err)
			}
			extraArgs = append(extraArgs, "--policy", src)
		}
		return client.Start(cmd.Context(), extraArgs...)
	},
}

func init() {
	startCmd.Flags().StringVarP(&startPolicyFlag, "policy", "p", "", "path or URL to security policy YAML file")
	rootCmd.AddCommand(startCmd)
}
