package cmd

import (
	"context"

	"github.com/spf13/cobra"
)

var reloadCmd = &cobra.Command{
	Use:   "reload",
	Short: "Reload the current page.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(cmd.Context(), timeoutFlag)
		defer cancel()
		return client.Reload(ctx)
	},
}

func init() {
	rootCmd.AddCommand(reloadCmd)
}
