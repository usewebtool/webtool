package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var keyCmd = &cobra.Command{
	Use:   "key <name>",
	Short: "Send a key press (e.g. Enter, Escape, Tab, ArrowDown). Case-insensitive.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(cmd.Context(), timeoutFlag)
		defer cancel()

		if err := client.Key(ctx, args[0]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(keyCmd)
}
