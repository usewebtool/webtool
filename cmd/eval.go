package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var evalCmd = &cobra.Command{
	Use:   "eval <js>",
	Short: "Execute a JavaScript expression and print the result.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		result, err := client.Eval(cmd.Context(), args[0])
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}

		fmt.Println(result)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(evalCmd)
}
