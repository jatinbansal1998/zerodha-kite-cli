package cli

import (
	"strings"
	"time"

	"github.com/jatinbansal1998/zerodha-kite-cli/internal/exitcode"
	"github.com/spf13/cobra"
	kiteconnect "github.com/zerodha/gokiteconnect/v4"
)

func newHoldingsCmd(opts *rootOptions) *cobra.Command {
	holdingsCmd := &cobra.Command{
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

	auctionsCmd := &cobra.Command{
		Use:   "auctions",
		Short: "List auction-eligible holdings",
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

			instruments, err := callWithAuthRetry(ctx, profileName, profile, func(client *kiteconnect.Client) ([]kiteconnect.AuctionInstrument, error) {
				return client.GetAuctionInstruments()
			})
			if err != nil {
				return err
			}

			printer := ctx.printer(cmd.OutOrStdout())
			if printer.IsJSON() {
				return printer.JSON(instruments)
			}

			rows := make([][]string, 0, len(instruments))
			for _, instrument := range instruments {
				rows = append(rows, []string{
					instrument.TradingSymbol,
					instrument.Exchange,
					instrument.AuctionNumber,
					intToString(instrument.Quantity),
					intToString(instrument.AuthorisedQuantity),
					formatFloat(instrument.LastPrice),
					formatFloat(instrument.Pnl),
					formatFloat(instrument.DayChangePercentage),
				})
			}
			if len(rows) == 0 {
				rows = append(rows, []string{"-", "-", "-", "0", "0", "0.00", "0.00", "0.00"})
			}
			return printer.Table(
				[]string{"SYMBOL", "EXCHANGE", "AUCTION_NO", "QTY", "AUTH_QTY", "LAST_PRICE", "PNL", "DAY_CHANGE_%"},
				rows,
			)
		},
	}

	var (
		authType         string
		authTransferType string
		authExecDate     string
		authISINs        []string
		authQuantities   []float64
	)
	authInitiateCmd := &cobra.Command{
		Use:   "auth-initiate",
		Short: "Initiate holdings authorization flow",
		RunE: func(cmd *cobra.Command, _ []string) error {
			params, err := holdingAuthParamsFromFlags(authType, authTransferType, authExecDate, authISINs, authQuantities)
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
			if err := ensureAccessToken(profile); err != nil {
				return err
			}

			resp, err := callWithAuthRetry(ctx, profileName, profile, func(client *kiteconnect.Client) (kiteconnect.HoldingsAuthResp, error) {
				return client.InitiateHoldingsAuth(params)
			})
			if err != nil {
				return err
			}

			printer := ctx.printer(cmd.OutOrStdout())
			if printer.IsJSON() {
				return printer.JSON(resp)
			}
			return printer.KV([][2]string{
				{"status", "ok"},
				{"request_id", resp.RequestID},
				{"redirect_url", resp.RedirectURL},
			})
		},
	}
	authInitiateCmd.Flags().StringVar(&authType, "type", "", "Authorization type (equity/mf)")
	authInitiateCmd.Flags().StringVar(&authTransferType, "transfer-type", "", "Transfer type (pre/post/off/gift)")
	authInitiateCmd.Flags().StringVar(&authExecDate, "exec-date", "", "Execution date (YYYY-MM-DD)")
	authInitiateCmd.Flags().StringSliceVar(&authISINs, "isin", nil, "ISIN to authorize (repeatable)")
	authInitiateCmd.Flags().Float64SliceVar(&authQuantities, "qty", nil, "Quantity for each --isin in the same order")

	holdingsCmd.AddCommand(auctionsCmd, authInitiateCmd)
	return holdingsCmd
}

func holdingAuthParamsFromFlags(authType, transferType, execDate string, isins []string, quantities []float64) (kiteconnect.HoldingAuthParams, error) {
	params := kiteconnect.HoldingAuthParams{
		Type:         strings.ToLower(strings.TrimSpace(authType)),
		TransferType: strings.ToLower(strings.TrimSpace(transferType)),
		ExecDate:     strings.TrimSpace(execDate),
	}

	if params.Type != "" && params.Type != kiteconnect.HolAuthTypeEquity && params.Type != kiteconnect.HolAuthTypeMF {
		return params, exitcode.New(exitcode.Validation, "invalid --type; use equity or mf")
	}
	if params.TransferType != "" &&
		params.TransferType != kiteconnect.HolAuthTransferTypePreTrade &&
		params.TransferType != kiteconnect.HolAuthTransferTypePostTrade &&
		params.TransferType != kiteconnect.HolAuthTransferTypeOffMarket &&
		params.TransferType != kiteconnect.HolAuthTransferTypeGift {
		return params, exitcode.New(exitcode.Validation, "invalid --transfer-type; use pre, post, off, or gift")
	}
	if params.ExecDate != "" {
		if _, err := time.Parse("2006-01-02", params.ExecDate); err != nil {
			return params, exitcode.New(exitcode.Validation, "--exec-date has invalid format; use YYYY-MM-DD")
		}
	}

	if len(isins) == 0 && len(quantities) == 0 {
		return params, nil
	}
	if len(isins) != len(quantities) {
		return params, exitcode.New(exitcode.Validation, "--isin and --qty must be provided in matching counts")
	}

	params.Instruments = make([]kiteconnect.HoldingsAuthInstruments, 0, len(isins))
	for i := range isins {
		isin := strings.TrimSpace(isins[i])
		if isin == "" {
			return params, exitcode.New(exitcode.Validation, "--isin values cannot be empty")
		}
		if quantities[i] <= 0 {
			return params, exitcode.New(exitcode.Validation, "--qty values must be greater than 0")
		}
		params.Instruments = append(params.Instruments, kiteconnect.HoldingsAuthInstruments{
			ISIN:     isin,
			Quantity: quantities[i],
		})
	}

	return params, nil
}
