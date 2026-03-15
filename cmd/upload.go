package cmd

import (
	"context"
	"path/filepath"

	"github.com/spf13/cobra"
)

var uploadCmd = &cobra.Command{
	Use:     "upload <selector> <file> [file...]",
	Short:   "Set one or more files on a file input element.",
	Example: "  webtool upload 43821 photo.jpg",
	Args:    cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(cmd.Context(), timeoutFlag)
		defer cancel()

		var paths []string
		for _, f := range args[1:] {
			abs, err := filepath.Abs(f)
			if err != nil {
				return err
			}
			paths = append(paths, abs)
		}

		return client.Upload(ctx, args[0], paths)
	},
}

func init() {
	rootCmd.AddCommand(uploadCmd)
}
