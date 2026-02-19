package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/jatinbansal1998/zerodha-kite-cli/internal/exitcode"
	"github.com/spf13/cobra"
	kiteconnect "github.com/zerodha/gokiteconnect/v4"
)

func newAuthCmd(opts *rootOptions) *cobra.Command {
	authCmd := &cobra.Command{
		Use:   "auth",
		Short: "Authenticate and manage sessions",
	}

	var (
		loginUseCallback bool
		loginPort        int
		loginRequest     string
	)
	loginCmd := &cobra.Command{
		Use:   "login",
		Short: "Login and persist access/refresh tokens for the selected profile",
		Long: "Login and persist access/refresh tokens for the selected profile.\n\n" +
			"Exactly one token acquisition mode is required:\n" +
			"  - --request-token <token_or_redirect_url>\n" +
			"  - --callback [--callback-port <1-65535>]",
		RunE: func(cmd *cobra.Command, _ []string) error {
			requestToken, err := validateAuthLoginFlags(
				loginUseCallback,
				loginRequest,
				loginPort,
				cmd.Flags().Changed("callback-port"),
			)
			if err != nil {
				return err
			}

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

			if loginUseCallback {
				if _, err := fmt.Fprintf(cmd.OutOrStdout(), "Waiting for callback at http://127.0.0.1:%d/ (timeout: 2m)\n", loginPort); err != nil {
					return err
				}
				token, err := captureRequestToken(loginPort, 2*time.Minute)
				if err != nil {
					return exitcode.Wrap(exitcode.Auth, "failed to receive request_token via callback", err)
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
	loginCmd.Flags().BoolVar(&loginUseCallback, "callback", false, "Capture request_token from localhost callback (cannot be used with --request-token)")
	loginCmd.Flags().IntVar(&loginPort, "callback-port", 8787, "Local callback port (1-65535, only with --callback)")
	loginCmd.Flags().StringVar(&loginRequest, "request-token", "", "Request token or full redirect URL (required unless --callback is used)")

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

	var revokeRefreshToken string
	revokeRefreshCmd := &cobra.Command{
		Use:   "revoke-refresh",
		Short: "Invalidate a refresh token",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx, err := newCommandContext(opts)
			if err != nil {
				return err
			}
			profileName, profile, err := ctx.resolveProfile(true)
			if err != nil {
				return err
			}
			if strings.TrimSpace(profile.APIKey) == "" {
				return exitcode.New(exitcode.Config, "profile missing api_key; set via `zerodha config profile add` or `zerodha config profile set-api-key`")
			}
			if err := ensureAccessToken(profile); err != nil {
				return err
			}

			token := strings.TrimSpace(revokeRefreshToken)
			storedToken := strings.TrimSpace(profile.RefreshToken)
			if token == "" {
				token = storedToken
			}
			if token == "" {
				return exitcode.New(exitcode.Validation, "no refresh token to revoke; pass --refresh-token or login again to store one")
			}

			revoked, err := callWithAuthRetry(ctx, profileName, profile, func(client *kiteconnect.Client) (bool, error) {
				return client.InvalidateRefreshToken(token)
			})
			if err != nil {
				return err
			}
			if !revoked {
				return exitcode.New(exitcode.API, "refresh token invalidation was not accepted")
			}

			refreshCleared := false
			if storedToken != "" && token == storedToken {
				profile.RefreshToken = ""
				ctx.setProfile(profileName, *profile)
				if err := ctx.save(); err != nil {
					return err
				}
				refreshCleared = true
			}

			printer := ctx.printer(cmd.OutOrStdout())
			if printer.IsJSON() {
				return printer.JSON(map[string]any{
					"status":                "ok",
					"profile":               profileName,
					"refresh_token_revoked": true,
					"refresh_token_cleared": refreshCleared,
				})
			}
			return printer.KV([][2]string{
				{"status", "ok"},
				{"profile", profileName},
				{"refresh_token_revoked", "yes"},
				{"refresh_token_cleared", boolToYesNo(refreshCleared)},
			})
		},
	}
	revokeRefreshCmd.Flags().StringVar(&revokeRefreshToken, "refresh-token", "", "Refresh token (defaults to stored profile refresh token)")

	authCmd.AddCommand(loginCmd, renewCmd, logoutCmd, revokeRefreshCmd)
	return authCmd
}
