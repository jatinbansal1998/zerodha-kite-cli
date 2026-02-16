package cli

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jatinbansal1998/zerodha-kite-cli/internal/exitcode"
	"github.com/jatinbansal1998/zerodha-kite-cli/internal/paths"
	"github.com/jatinbansal1998/zerodha-kite-cli/internal/tui"
	"github.com/spf13/cobra"
)

type rootOptions struct {
	interactive bool
	profile     string
	configPath  string
	outputJSON  bool
	debug       bool
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
	rootCmd.Long = "zerodha provides interactive and non-interactive workflows for Zerodha Kite account operations."
	rootCmd.SilenceUsage = true
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, _ []string) error {
		if opts.interactive && cmd != rootCmd {
			return exitcode.New(exitcode.Validation, "interactive mode (-i) cannot be combined with subcommands")
		}
		return nil
	}
	rootCmd.RunE = func(cmd *cobra.Command, _ []string) error {
		if opts.interactive {
			return runInteractive(opts.profile, cmd)
		}
		return cmd.Help()
	}

	rootCmd.PersistentFlags().BoolVarP(&opts.interactive, "interactive", "i", false, "Run in interactive mode")
	rootCmd.PersistentFlags().StringVar(&opts.profile, "profile", "", "Profile name (defaults to active profile)")
	rootCmd.PersistentFlags().StringVar(&opts.configPath, "config", defaultConfigPath, "Path to config file")
	rootCmd.PersistentFlags().BoolVar(&opts.outputJSON, "json", false, "Render output as JSON")
	rootCmd.PersistentFlags().BoolVar(&opts.debug, "debug", false, "Enable SDK HTTP debug logs")

	rootCmd.AddCommand(
		newVersionCmd(),
		newAuthCmd(opts),
		newConfigCmd(opts),
		newProfileCmd(opts),
		newQuoteCmd(opts),
		newOrderCmd(opts),
		newOrdersCmd(opts),
		newPositionsCmd(opts),
		newHoldingsCmd(opts),
		newMarginsCmd(opts),
	)

	return rootCmd
}

func runInteractive(profile string, cmd *cobra.Command) error {
	if profile == "" {
		profile = "active"
	}

	model := tui.NewModel(profile)
	program := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := program.Run(); err != nil {
		return fmt.Errorf("start interactive mode: %w", err)
	}

	return nil
}
