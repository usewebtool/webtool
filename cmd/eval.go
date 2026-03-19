package cmd

import (
	"context"

	"github.com/spf13/cobra"
)

var evalCmd = &cobra.Command{
	Use:   "eval <js>",
	Short: "Execute a JavaScript expression and print the result.",
	Long: `Execute a JavaScript expression and print the result.

Only expressions are supported, not statements (const, let, var).
For multi-statement code, wrap in an IIFE:

  webtool eval "(function(){ const a = 1; return a; })()"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(cmd.Context(), timeoutFlag)
		defer cancel()
		result, err := client.Eval(ctx, args[0])
		if err != nil {
			return err
		}
		cmd.Println(result)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(evalCmd)
}
