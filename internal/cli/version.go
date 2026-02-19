package cli

import (
	"fmt"
	"strings"

	"github.com/jatinbansal1998/zerodha-kite-cli/internal/buildinfo"
	"github.com/jatinbansal1998/zerodha-kite-cli/internal/paths"
	"github.com/jatinbansal1998/zerodha-kite-cli/internal/updater"
	"github.com/spf13/cobra"
)

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print CLI version",
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, err := fmt.Fprintf(cmd.OutOrStdout(), "zerodha %s\n", buildinfo.Version); err != nil {
				return err
			}

			cacheDir, err := paths.DefaultCacheDir()
			if err != nil {
				return nil
			}
			state, err := updater.LoadState(cacheDir)
			if err != nil {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "update: unavailable (%s)\n", strings.TrimSpace(err.Error()))
				return nil
			}
			_, err = fmt.Fprintln(cmd.OutOrStdout(), updater.SummarizeState(buildinfo.Version, state))
			return err
		},
	}
}
