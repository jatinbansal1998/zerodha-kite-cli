package cli

import (
	"strings"

	"github.com/spf13/cobra"
	kiteconnect "github.com/zerodha/gokiteconnect/v4"
)

func newProfileCmd(opts *rootOptions) *cobra.Command {
	profileCmd := &cobra.Command{
		Use:   "profile",
		Short: "Fetch account profile data",
	}

	showCmd := &cobra.Command{
		Use:   "show",
		Short: "Show Zerodha user profile",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx, err := newCommandContext(opts)
			if err != nil {
				return err
			}
			profileName, profile, err := ctx.resolveProfile(true)
			if err != nil {
				return err
			}
			if err := ensureAccessToken(profile); err != nil {
				return err
			}

			userProfile, err := callWithAuthRetry(ctx, profileName, profile, func(client *kiteconnect.Client) (kiteconnect.UserProfile, error) {
				return client.GetUserProfile()
			})
			if err != nil {
				return err
			}

			printer := ctx.printer(cmd.OutOrStdout())
			if printer.IsJSON() {
				return printer.JSON(userProfile)
			}
			return printer.KV([][2]string{
				{"user_id", userProfile.UserID},
				{"user_name", userProfile.UserName},
				{"user_shortname", userProfile.UserShortName},
				{"email", userProfile.Email},
				{"broker", userProfile.Broker},
				{"products", strings.Join(userProfile.Products, ",")},
				{"order_types", strings.Join(userProfile.OrderTypes, ",")},
				{"exchanges", strings.Join(userProfile.Exchanges, ",")},
			})
		},
	}

	fullCmd := &cobra.Command{
		Use:   "full",
		Short: "Show full Zerodha user profile",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx, err := newCommandContext(opts)
			if err != nil {
				return err
			}
			profileName, profile, err := ctx.resolveProfile(true)
			if err != nil {
				return err
			}
			if err := ensureAccessToken(profile); err != nil {
				return err
			}

			userProfile, err := callWithAuthRetry(ctx, profileName, profile, func(client *kiteconnect.Client) (kiteconnect.FullUserProfile, error) {
				return client.GetFullUserProfile()
			})
			if err != nil {
				return err
			}

			printer := ctx.printer(cmd.OutOrStdout())
			if printer.IsJSON() {
				return printer.JSON(userProfile)
			}
			return printer.KV([][2]string{
				{"user_id", userProfile.UserID},
				{"user_name", userProfile.UserName},
				{"user_shortname", userProfile.UserShortName},
				{"email", userProfile.Email},
				{"phone", userProfile.Phone},
				{"user_type", userProfile.UserType},
				{"broker", userProfile.Broker},
				{"twofa_type", userProfile.TwoFAType},
				{"pan", userProfile.PAN},
				{"products", strings.Join(userProfile.Products, ",")},
				{"order_types", strings.Join(userProfile.OrderTypes, ",")},
				{"exchanges", strings.Join(userProfile.Exchanges, ",")},
				{"dp_ids", strings.Join(userProfile.DPIDs, ",")},
				{"tags", strings.Join(userProfile.Tags, ",")},
				{"account_blocks", strings.Join(userProfile.Meta.AccountBlocks, ",")},
			})
		},
	}

	profileCmd.AddCommand(showCmd, fullCmd)
	return profileCmd
}
