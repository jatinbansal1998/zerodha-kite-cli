package cli

import (
	"os"

	"github.com/jatinbansal1998/zerodha-kite-cli/internal/buildinfo"
	"github.com/jatinbansal1998/zerodha-kite-cli/internal/paths"
	"github.com/jatinbansal1998/zerodha-kite-cli/internal/updater"
	"github.com/spf13/cobra"
)

type rootOptions struct {
	profile    string
	configPath string
	outputJSON bool
	debug      bool
}

func Execute() error {
	return newRootCmd().Execute()
}

func newRootCmd() *cobra.Command {
	opts := &rootOptions{}
	defaultConfigPath, _ := paths.DefaultConfigPath()

	rootCmd := &cobra.Command{}
	rootCmd.Use = "zerodha"
	rootCmd.Short = "CLI-based tooling for Zerodha account workflows"
	rootCmd.Long = "zerodha provides CLI-driven workflows for Zerodha Kite account operations."
	rootCmd.SilenceUsage = true
	rootCmd.RunE = func(cmd *cobra.Command, _ []string) error {
		return cmd.Help()
	}
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, _ []string) {
		if shouldRunAutoUpdate(cmd) {
			startAutoUpdate()
		}
	}

	rootCmd.PersistentFlags().StringVar(&opts.profile, "profile", "", "Profile name (defaults to active profile)")
	rootCmd.PersistentFlags().StringVar(&opts.configPath, "config", defaultConfigPath, "Path to config file")
	rootCmd.PersistentFlags().BoolVar(&opts.outputJSON, "json", false, "Render output as JSON")
	rootCmd.PersistentFlags().BoolVar(&opts.debug, "debug", false, "Enable SDK HTTP debug logs")

	rootCmd.AddCommand(
		newSelfUpdateApplyCmd(),
		newVersionCmd(),
		newUpdateCmd(opts),
		newAuthCmd(opts),
		newConfigCmd(opts),
		newProfileCmd(opts),
		newQuoteCmd(opts),
		newInstrumentsCmd(opts),
		newGttCmd(opts),
		newMFCmd(opts),
		newOrderCmd(opts),
		newOrdersCmd(opts),
		newPositionsCmd(opts),
		newHoldingsCmd(opts),
		newMarginsCmd(opts),
	)

	return rootCmd
}

func shouldRunAutoUpdate(cmd *cobra.Command) bool {
	if cmd != nil && (cmd.Name() == updater.HelperCommandName() || cmd.Name() == "update") {
		return false
	}

	helperMode := os.Getenv(updater.HelperEnvVar)
	return helperMode == ""
}

func startAutoUpdate() {
	cacheDir, err := paths.DefaultCacheDir()
	if err != nil {
		return
	}

	executablePath, err := os.Executable()
	if err != nil {
		return
	}

	updater.StartBackground(updater.Options{
		CurrentVersion: buildinfo.Version,
		ExecutablePath: executablePath,
		CacheDir:       cacheDir,
		RepoOwner:      updater.DefaultRepoOwner,
		RepoName:       updater.DefaultRepoName,
		Cooldown:       updater.DefaultCooldown,
	})
}
