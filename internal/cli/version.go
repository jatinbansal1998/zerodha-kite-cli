package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

const version = "dev"

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print CLI version",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := fmt.Fprintf(cmd.OutOrStdout(), "zerodha %s\n", version)
			return err
		},
	}
}
