package cmd

import (
	"github.com/spf13/cobra"
	"github.com/usewebtool/webtool/agent"
	"github.com/usewebtool/webtool/browser"
	"github.com/usewebtool/webtool/policy"
)

var servePolicyFlag string

var serveCmd = &cobra.Command{
	Use:    "_serve",
	Short:  "Run the daemon in foreground (internal use).",
	Hidden: true,
	Args:   cobra.NoArgs,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		agent.HomeDir = resolveHome()
		return nil // Skip root PersistentPreRunE — we are the daemon.
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		b := browser.New().WithChromeDataDir(resolveDataDir())

		if servePolicyFlag != "" {
			p, err := policy.Load(servePolicyFlag)
			if err != nil {
				return err
			}
			b.WithPolicy(p)
		}

		return agent.NewServer(b).Start()
	},
}

func init() {
	serveCmd.Flags().StringVar(&servePolicyFlag, "policy", "", "path to security policy YAML file")
	rootCmd.AddCommand(serveCmd)
}
