package cli

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"
	kiteconnect "github.com/zerodha/gokiteconnect/v4"
)

func newQuoteCmd(opts *rootOptions) *cobra.Command {
	quoteCmd := &cobra.Command{
		Use:   "quote",
		Short: "Market quote utilities",
	}

	getCmd := &cobra.Command{
		Use:   "get <EXCHANGE:SYMBOL> [EXCHANGE:SYMBOL...]",
		Short: "Fetch snapshot quotes",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
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

			quotes, err := callWithAuthRetry(ctx, profileName, profile, func(client *kiteconnect.Client) (kiteconnect.Quote, error) {
				return client.GetQuote(args...)
			})
			if err != nil {
				return err
			}

			printer := ctx.printer(cmd.OutOrStdout())
			if printer.IsJSON() {
				return printer.JSON(quotes)
			}

			keys := make([]string, 0, len(quotes))
			for k := range quotes {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			rows := make([][]string, 0, len(keys))
			for _, key := range keys {
				q := quotes[key]
				rows = append(rows, []string{
					key,
					fmt.Sprintf("%.2f", q.LastPrice),
					fmt.Sprintf("%.2f", q.NetChange),
					fmt.Sprintf("%d", q.Volume),
					q.Timestamp.Time.Format("2006-01-02 15:04:05"),
				})
			}
			return printer.Table([]string{"INSTRUMENT", "LTP", "NET_CHANGE", "VOLUME", "TIMESTAMP"}, rows)
		},
	}

	quoteCmd.AddCommand(getCmd)
	return quoteCmd
}
