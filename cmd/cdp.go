package cmd

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

var cdpCmd = &cobra.Command{
	Use:   "cdp <method> [params-json]",
	Short: "Send a raw Chrome DevTools Protocol command.",
	Long: `Send a raw CDP command to the active page. Use this as a fallback for
CDP methods not covered by dedicated commands (e.g. Input.insertText
for canvas-based apps like Google Docs).

Examples:
  webtool cdp Input.insertText '{"text": "hello"}'
  webtool cdp Page.reload
  webtool cdp DOM.getDocument`,
	Args: cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(cmd.Context(), timeoutFlag)
		defer cancel()

		method := args[0]

		var params json.RawMessage
		if len(args) > 1 {
			if !json.Valid([]byte(args[1])) {
				return fmt.Errorf("invalid JSON params: %s", args[1])
			}
			params = json.RawMessage(args[1])
		}

		result, err := client.CDP(ctx, method, params)
		if err != nil {
			return err
		}
		if len(result) > 0 && string(result) != "null" {
			cmd.Println(string(result))
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(cdpCmd)
}
