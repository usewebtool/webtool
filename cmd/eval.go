package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

var evalCmd = &cobra.Command{
	Use:   "eval <js>",
	Short: "Execute a JavaScript expression and print the result.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(cmd.Context(), timeoutFlag)
		defer cancel()
		result, err := client.Eval(ctx, args[0])
		if err != nil {
			return err
		}
		fmt.Println(result)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(evalCmd)
}
