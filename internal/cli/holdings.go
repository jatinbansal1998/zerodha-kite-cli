package cli

import (
	"github.com/spf13/cobra"
	kiteconnect "github.com/zerodha/gokiteconnect/v4"
)

func newHoldingsCmd(opts *rootOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "holdings",
		Short: "List current holdings",
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

			holdings, err := callWithAuthRetry(ctx, profileName, profile, func(client *kiteconnect.Client) (kiteconnect.Holdings, error) {
				return client.GetHoldings()
			})
			if err != nil {
				return err
			}

			printer := ctx.printer(cmd.OutOrStdout())
			if printer.IsJSON() {
				return printer.JSON(holdings)
			}

			rows := make([][]string, 0, len(holdings))
			for _, holding := range holdings {
				rows = append(rows, []string{
					holding.Tradingsymbol,
					holding.Exchange,
					intToString(holding.Quantity),
					formatFloat(holding.AveragePrice),
					formatFloat(holding.LastPrice),
					formatFloat(holding.PnL),
					formatFloat(holding.DayChangePercentage),
				})
			}
			if len(rows) == 0 {
				rows = append(rows, []string{"-", "-", "0", "0.00", "0.00", "0.00", "0.00"})
			}

			return printer.Table([]string{"SYMBOL", "EXCHANGE", "QTY", "AVG_PRICE", "LTP", "PNL", "DAY_CHANGE_%"},
				rows)
		},
	}
}
