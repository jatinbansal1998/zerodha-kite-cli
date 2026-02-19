package cli

import (
	"fmt"
	"strings"

	"github.com/jatinbansal1998/zerodha-kite-cli/internal/config"
	"github.com/jatinbansal1998/zerodha-kite-cli/internal/exitcode"
	"github.com/spf13/cobra"
)

func newConfigCmd(opts *rootOptions) *cobra.Command {
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Manage local CLI configuration",
	}
	configCmd.AddCommand(newConfigProfileCmd(opts))
	return configCmd
}

func newConfigProfileCmd(opts *rootOptions) *cobra.Command {
	profileCmd := &cobra.Command{
		Use:   "profile",
		Short: "Manage named profiles",
	}

	var (
		addAPIKey    string
		addAPISecret string
		addSetActive bool
	)
	addCmd := &cobra.Command{
		Use:   "add <name>",
		Short: "Add or update a profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := strings.TrimSpace(args[0])
			if name == "" {
				return exitcode.New(exitcode.Validation, "profile name cannot be empty")
			}
			if strings.TrimSpace(addAPIKey) == "" || strings.TrimSpace(addAPISecret) == "" {
				return exitcode.New(exitcode.Validation, "--api-key and --api-secret are required")
			}

			ctx, err := newCommandContext(opts)
			if err != nil {
				return err
			}

			profile := upsertProfileCredentials(ctx.cfg.Profiles[name], addAPIKey, addAPISecret)
			ctx.setProfile(name, profile)
			if ctx.cfg.ActiveProfile == "" || addSetActive {
				ctx.cfg.ActiveProfile = name
			}
			if err := ctx.save(); err != nil {
				return err
			}

			printer := ctx.printer(cmd.OutOrStdout())
			if printer.IsJSON() {
				return printer.JSON(map[string]string{
					"status":  "ok",
					"profile": name,
				})
			}

			return printer.KV([][2]string{
				{"status", "ok"},
				{"profile", name},
				{"active_profile", ctx.cfg.ActiveProfile},
				{"config_path", ctx.store.Path()},
			})
		},
	}
	addCmd.Flags().StringVar(&addAPIKey, "api-key", "", "Kite API key")
	addCmd.Flags().StringVar(&addAPISecret, "api-secret", "", "Kite API secret")
	addCmd.Flags().BoolVar(&addSetActive, "set-active", false, "Set this profile as active")

	var setAPIKeyValue string
	setAPIKeyCmd := &cobra.Command{
		Use:   "set-api-key <name>",
		Short: "Set only the API key for a profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := strings.TrimSpace(args[0])
			if name == "" {
				return exitcode.New(exitcode.Validation, "profile name cannot be empty")
			}
			if strings.TrimSpace(setAPIKeyValue) == "" {
				return exitcode.New(exitcode.Validation, "--api-key is required")
			}

			ctx, err := updateExistingProfile(opts, name, func(profile *config.Profile) {
				profile.APIKey = strings.TrimSpace(setAPIKeyValue)
			})
			if err != nil {
				return err
			}

			printer := ctx.printer(cmd.OutOrStdout())
			if printer.IsJSON() {
				return printer.JSON(map[string]string{
					"status":  "ok",
					"profile": name,
				})
			}
			return printer.KV([][2]string{
				{"status", "ok"},
				{"profile", name},
			})
		},
	}
	setAPIKeyCmd.Flags().StringVar(&setAPIKeyValue, "api-key", "", "Kite API key")

	var setAPISecretValue string
	setAPISecretCmd := &cobra.Command{
		Use:   "set-api-secret <name>",
		Short: "Set only the API secret for a profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := strings.TrimSpace(args[0])
			if name == "" {
				return exitcode.New(exitcode.Validation, "profile name cannot be empty")
			}
			if strings.TrimSpace(setAPISecretValue) == "" {
				return exitcode.New(exitcode.Validation, "--api-secret is required")
			}

			ctx, err := updateExistingProfile(opts, name, func(profile *config.Profile) {
				profile.APISecret = strings.TrimSpace(setAPISecretValue)
			})
			if err != nil {
				return err
			}

			printer := ctx.printer(cmd.OutOrStdout())
			if printer.IsJSON() {
				return printer.JSON(map[string]string{
					"status":  "ok",
					"profile": name,
				})
			}
			return printer.KV([][2]string{
				{"status", "ok"},
				{"profile", name},
			})
		},
	}
	setAPISecretCmd.Flags().StringVar(&setAPISecretValue, "api-secret", "", "Kite API secret")

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List configured profiles",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx, err := newCommandContext(opts)
			if err != nil {
				return err
			}

			names := ctx.profileNames()
			printer := ctx.printer(cmd.OutOrStdout())
			if printer.IsJSON() {
				type item struct {
					Name   string `json:"name"`
					Active bool   `json:"active"`
				}
				items := make([]item, 0, len(names))
				for _, name := range names {
					items = append(items, item{Name: name, Active: name == ctx.cfg.ActiveProfile})
				}
				return printer.JSON(map[string]any{
					"active_profile": ctx.cfg.ActiveProfile,
					"profiles":       items,
				})
			}

			if len(names) == 0 {
				_, err := fmt.Fprintln(cmd.OutOrStdout(), "No profiles found. Add one with `zerodha config profile add <name> --api-key ... --api-secret ...`.")
				return err
			}

			rows := make([][]string, 0, len(names))
			for _, name := range names {
				active := ""
				if name == ctx.cfg.ActiveProfile {
					active = "yes"
				}
				rows = append(rows, []string{name, active})
			}
			return printer.Table([]string{"PROFILE", "ACTIVE"}, rows)
		},
	}

	useCmd := &cobra.Command{
		Use:   "use <name>",
		Short: "Set active profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := strings.TrimSpace(args[0])
			if name == "" {
				return exitcode.New(exitcode.Validation, "profile name cannot be empty")
			}

			ctx, err := newCommandContext(opts)
			if err != nil {
				return err
			}
			if _, ok := ctx.cfg.Profiles[name]; !ok {
				return exitcode.New(exitcode.Config, fmt.Sprintf("profile %q not found", name))
			}
			ctx.cfg.ActiveProfile = name
			if err := ctx.save(); err != nil {
				return err
			}

			printer := ctx.printer(cmd.OutOrStdout())
			if printer.IsJSON() {
				return printer.JSON(map[string]string{
					"status":         "ok",
					"active_profile": name,
				})
			}
			return printer.KV([][2]string{
				{"status", "ok"},
				{"active_profile", name},
			})
		},
	}

	removeCmd := &cobra.Command{
		Use:   "remove <name>",
		Short: "Remove a profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := strings.TrimSpace(args[0])
			if name == "" {
				return exitcode.New(exitcode.Validation, "profile name cannot be empty")
			}

			ctx, err := newCommandContext(opts)
			if err != nil {
				return err
			}
			if _, ok := ctx.cfg.Profiles[name]; !ok {
				return exitcode.New(exitcode.Config, fmt.Sprintf("profile %q not found", name))
			}

			ctx.deleteProfile(name)
			if ctx.cfg.ActiveProfile == name {
				ctx.cfg.ActiveProfile = firstRemainingProfile(ctx.profileNames(), "")
			}
			if err := ctx.save(); err != nil {
				return err
			}

			printer := ctx.printer(cmd.OutOrStdout())
			if printer.IsJSON() {
				return printer.JSON(map[string]string{
					"status":         "ok",
					"removed":        name,
					"active_profile": ctx.cfg.ActiveProfile,
				})
			}
			return printer.KV([][2]string{
				{"status", "ok"},
				{"removed", name},
				{"active_profile", ctx.cfg.ActiveProfile},
			})
		},
	}

	profileCmd.AddCommand(addCmd, setAPIKeyCmd, setAPISecretCmd, listCmd, useCmd, removeCmd)
	return profileCmd
}

func upsertProfileCredentials(profile config.Profile, apiKey, apiSecret string) config.Profile {
	profile.APIKey = strings.TrimSpace(apiKey)
	profile.APISecret = strings.TrimSpace(apiSecret)
	return profile
}

func updateExistingProfile(
	opts *rootOptions,
	name string,
	update func(*config.Profile),
) (*commandContext, error) {
	ctx, err := newCommandContext(opts)
	if err != nil {
		return nil, err
	}

	profile, ok := ctx.cfg.Profiles[name]
	if !ok {
		return nil, exitcode.New(exitcode.Config, fmt.Sprintf("profile %q not found", name))
	}

	update(&profile)
	ctx.setProfile(name, profile)
	if err := ctx.save(); err != nil {
		return nil, err
	}

	return ctx, nil
}
