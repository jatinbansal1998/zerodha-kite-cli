package cli

import (
	"github.com/spf13/cobra"
	kiteconnect "github.com/zerodha/gokiteconnect/v4"
)

func newPositionsCmd(opts *rootOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "positions",
		Short: "List current positions",
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

			positions, err := callWithAuthRetry(ctx, profileName, profile, func(client *kiteconnect.Client) (kiteconnect.Positions, error) {
				return client.GetPositions()
			})
			if err != nil {
				return err
			}

			printer := ctx.printer(cmd.OutOrStdout())
			if printer.IsJSON() {
				return printer.JSON(positions)
			}

			rows := make([][]string, 0, len(positions.Net))
			for _, position := range positions.Net {
				rows = append(rows, []string{
					position.Tradingsymbol,
					position.Exchange,
					position.Product,
					intToString(position.Quantity),
					formatFloat(position.AveragePrice),
					formatFloat(position.LastPrice),
					formatFloat(position.PnL),
				})
			}

			if len(rows) == 0 {
				rows = append(rows, []string{"-", "-", "-", "0", "0.00", "0.00", "0.00"})
			}

			return printer.Table([]string{"SYMBOL", "EXCHANGE", "PRODUCT", "QTY", "AVG_PRICE", "LTP", "PNL"}, rows)
		},
	}
}
