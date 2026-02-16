package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/jatinbansal1998/zerodha-kite-cli/internal/exitcode"
	"github.com/spf13/cobra"
)

func newAuthCmd(opts *rootOptions) *cobra.Command {
	authCmd := &cobra.Command{
		Use:   "auth",
		Short: "Authenticate and manage sessions",
	}

	var (
		loginUseCallback bool
		loginPort        int
	)
	loginCmd := &cobra.Command{
		Use:   "login",
		Short: "Login and persist access/refresh tokens for the selected profile",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx, err := newCommandContext(opts)
			if err != nil {
				return err
			}

			profileName, profile, err := ctx.resolveProfile(true)
			if err != nil {
				return err
			}
			if err := ensureProfileCredentials(profile); err != nil {
				return err
			}

			client := newKiteClient(*profile, opts.debug)
			loginURL := client.GetLoginURL()
			if _, err := fmt.Fprintf(cmd.OutOrStdout(), "Login URL: %s\n", loginURL); err != nil {
				return err
			}

			var requestToken string
			if loginUseCallback {
				if _, err := fmt.Fprintf(cmd.OutOrStdout(), "Waiting for callback at http://127.0.0.1:%d/ (timeout: 2m)\n", loginPort); err != nil {
					return err
				}
				token, err := captureRequestToken(loginPort, 2*time.Minute)
				if err != nil {
					return exitcode.Wrap(exitcode.Auth, "failed to receive request_token via callback", err)
				}
				requestToken = token
			} else {
				token, err := promptRequestToken(cmd.InOrStdin(), cmd.OutOrStdout())
				if err != nil {
					return exitcode.Wrap(exitcode.Auth, "failed to read request_token", err)
				}
				requestToken = token
			}

			session, err := client.GenerateSession(requestToken, profile.APISecret)
			if err != nil {
				return wrapKiteError("session generation failed", err)
			}
			if strings.TrimSpace(session.AccessToken) == "" {
				return exitcode.New(exitcode.Auth, "session response did not include an access token")
			}

			profile.AccessToken = session.AccessToken
			profile.RefreshToken = session.RefreshToken
			profile.LastLoginAt = nowUTC()
			ctx.setProfile(profileName, *profile)
			if ctx.cfg.ActiveProfile == "" {
				ctx.cfg.ActiveProfile = profileName
			}
			if err := ctx.save(); err != nil {
				return err
			}

			printer := ctx.printer(cmd.OutOrStdout())
			if printer.IsJSON() {
				return printer.JSON(map[string]any{
					"status":         "ok",
					"profile":        profileName,
					"user_id":        session.UserID,
					"last_login_at":  profile.LastLoginAt,
					"access_token":   profile.AccessToken != "",
					"refresh_token":  profile.RefreshToken != "",
					"active_profile": ctx.cfg.ActiveProfile,
				})
			}
			return printer.KV([][2]string{
				{"status", "ok"},
				{"profile", profileName},
				{"user_id", session.UserID},
				{"active_profile", ctx.cfg.ActiveProfile},
				{"last_login_at", profile.LastLoginAt.Format(time.RFC3339)},
			})
		},
	}
	loginCmd.Flags().BoolVar(&loginUseCallback, "callback", false, "Capture request_token from localhost callback")
	loginCmd.Flags().IntVar(&loginPort, "callback-port", 8787, "Local callback port")

	renewCmd := &cobra.Command{
		Use:   "renew",
		Short: "Renew access token using refresh token",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx, err := newCommandContext(opts)
			if err != nil {
				return err
			}
			profileName, profile, err := ctx.resolveProfile(true)
			if err != nil {
				return err
			}
			if err := ensureProfileCredentials(profile); err != nil {
				return err
			}
			if strings.TrimSpace(profile.RefreshToken) == "" {
				return exitcode.New(exitcode.Auth, "profile has no refresh token; run `zerodha auth login`")
			}

			client := newKiteClient(*profile, opts.debug)
			session, err := client.RenewAccessToken(profile.RefreshToken, profile.APISecret)
			if err != nil {
				return wrapKiteError("token renewal failed", err)
			}
			if strings.TrimSpace(session.AccessToken) == "" {
				return exitcode.New(exitcode.Auth, "token renewal response did not include access token")
			}

			profile.AccessToken = session.AccessToken
			if strings.TrimSpace(session.RefreshToken) != "" {
				profile.RefreshToken = session.RefreshToken
			}
			profile.LastLoginAt = nowUTC()
			ctx.setProfile(profileName, *profile)
			if err := ctx.save(); err != nil {
				return err
			}

			printer := ctx.printer(cmd.OutOrStdout())
			if printer.IsJSON() {
				return printer.JSON(map[string]any{
					"status":        "ok",
					"profile":       profileName,
					"last_login_at": profile.LastLoginAt,
				})
			}
			return printer.KV([][2]string{
				{"status", "ok"},
				{"profile", profileName},
				{"last_login_at", profile.LastLoginAt.Format(time.RFC3339)},
			})
		},
	}

	logoutCmd := &cobra.Command{
		Use:   "logout",
		Short: "Clear local tokens and invalidate current access token if available",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx, err := newCommandContext(opts)
			if err != nil {
				return err
			}
			profileName, profile, err := ctx.resolveProfile(true)
			if err != nil {
				return err
			}
			if err := ensureProfileCredentials(profile); err != nil {
				return err
			}

			if strings.TrimSpace(profile.AccessToken) != "" {
				client := newKiteClient(*profile, opts.debug)
				_, invalidateErr := client.InvalidateAccessToken()
				if invalidateErr != nil && !isTokenError(invalidateErr) {
					return wrapKiteError("failed to invalidate access token", invalidateErr)
				}
			}

			profile.AccessToken = ""
			profile.RefreshToken = ""
			ctx.setProfile(profileName, *profile)
			if err := ctx.save(); err != nil {
				return err
			}

			printer := ctx.printer(cmd.OutOrStdout())
			if printer.IsJSON() {
				return printer.JSON(map[string]any{
					"status":  "ok",
					"profile": profileName,
				})
			}
			return printer.KV([][2]string{
				{"status", "ok"},
				{"profile", profileName},
			})
		},
	}

	authCmd.AddCommand(loginCmd, renewCmd, logoutCmd)
	return authCmd
}
