package cli

import (
	"errors"
	"strings"

	"github.com/jatinbansal1998/zerodha-kite-cli/internal/updater"
	"github.com/spf13/cobra"
)

func newSelfUpdateApplyCmd() *cobra.Command {
	var targetPath string
	var sourcePath string
	var cacheDir string
	var version string

	cmd := &cobra.Command{
		Use:                updater.HelperCommandName(),
		Short:              "Internal updater helper command",
		Hidden:             true,
		SilenceUsage:       true,
		SilenceErrors:      true,
		DisableFlagParsing: false,
		RunE: func(_ *cobra.Command, _ []string) error {
			if strings.TrimSpace(targetPath) == "" ||
				strings.TrimSpace(sourcePath) == "" ||
				strings.TrimSpace(cacheDir) == "" {
				return errors.New("target, source, and cache-dir are required")
			}

			return updater.RunApplyHelper(updater.ApplyHelperRequest{
				TargetPath: targetPath,
				SourcePath: sourcePath,
				CacheDir:   cacheDir,
				Version:    version,
			})
		},
	}

	cmd.Flags().StringVar(&targetPath, "target", "", "Target executable path")
	cmd.Flags().StringVar(&sourcePath, "source", "", "Downloaded executable path")
	cmd.Flags().StringVar(&cacheDir, "cache-dir", "", "Updater cache directory")
	cmd.Flags().StringVar(&version, "version", "", "Downloaded version")

	return cmd
}
