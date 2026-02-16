package cli

import (
	"strings"

	"github.com/spf13/cobra"
	kiteconnect "github.com/zerodha/gokiteconnect/v4"
)

func newMarginsCmd(opts *rootOptions) *cobra.Command {
	var segment string

	marginsCmd := &cobra.Command{
		Use:   "margins",
		Short: "Get available and used margins",
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

			segmentValue := strings.ToLower(strings.TrimSpace(segment))
			if segmentValue == "" || segmentValue == "all" {
				allMargins, err := callWithAuthRetry(ctx, profileName, profile, func(client *kiteconnect.Client) (kiteconnect.AllMargins, error) {
					return client.GetUserMargins()
				})
				if err != nil {
					return err
				}

				printer := ctx.printer(cmd.OutOrStdout())
				if printer.IsJSON() {
					return printer.JSON(allMargins)
				}

				rows := [][]string{
					{"equity", formatFloat(allMargins.Equity.Net), formatFloat(allMargins.Equity.Available.Cash), formatFloat(allMargins.Equity.Used.Debits)},
					{"commodity", formatFloat(allMargins.Commodity.Net), formatFloat(allMargins.Commodity.Available.Cash), formatFloat(allMargins.Commodity.Used.Debits)},
				}
				return printer.Table([]string{"SEGMENT", "NET", "AVAILABLE_CASH", "USED_DEBITS"}, rows)
			}

			margin, err := callWithAuthRetry(ctx, profileName, profile, func(client *kiteconnect.Client) (kiteconnect.Margins, error) {
				return client.GetUserSegmentMargins(segmentValue)
			})
			if err != nil {
				return err
			}

			printer := ctx.printer(cmd.OutOrStdout())
			if printer.IsJSON() {
				return printer.JSON(margin)
			}
			rows := [][]string{
				{segmentValue, formatFloat(margin.Net), formatFloat(margin.Available.Cash), formatFloat(margin.Used.Debits)},
			}
			return printer.Table([]string{"SEGMENT", "NET", "AVAILABLE_CASH", "USED_DEBITS"}, rows)
		},
	}
	marginsCmd.Flags().StringVar(&segment, "segment", "all", "Margin segment: all/equity/commodity")
	return marginsCmd
}
