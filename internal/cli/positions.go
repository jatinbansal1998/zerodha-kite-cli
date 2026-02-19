package cli

import (
	"strings"

	"github.com/jatinbansal1998/zerodha-kite-cli/internal/exitcode"
	"github.com/spf13/cobra"
	kiteconnect "github.com/zerodha/gokiteconnect/v4"
)

func newPositionsCmd(opts *rootOptions) *cobra.Command {
	positionsCmd := &cobra.Command{
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

	var (
		exchange     string
		symbol       string
		oldProduct   string
		newProduct   string
		positionType string
		txnType      string
		quantity     int
	)
	convertCmd := &cobra.Command{
		Use:   "convert",
		Short: "Convert a position product type",
		RunE: func(cmd *cobra.Command, _ []string) error {
			params := kiteconnect.ConvertPositionParams{
				Exchange:        normalizeUpper(exchange),
				TradingSymbol:   strings.TrimSpace(symbol),
				OldProduct:      normalizeUpper(oldProduct),
				NewProduct:      normalizeUpper(newProduct),
				PositionType:    strings.ToLower(strings.TrimSpace(positionType)),
				TransactionType: normalizeUpper(txnType),
				Quantity:        quantity,
			}

			if params.Exchange == "" || params.TradingSymbol == "" || params.OldProduct == "" || params.NewProduct == "" || params.PositionType == "" || params.TransactionType == "" {
				return exitcode.New(exitcode.Validation, "--exchange, --symbol, --old-product, --new-product, --position-type, and --txn are required")
			}
			if params.Quantity <= 0 {
				return exitcode.New(exitcode.Validation, "--qty must be greater than 0")
			}
			if params.TransactionType != kiteconnect.TransactionTypeBuy && params.TransactionType != kiteconnect.TransactionTypeSell {
				return exitcode.New(exitcode.Validation, "invalid --txn; use BUY or SELL")
			}
			if params.PositionType != kiteconnect.PositionTypeDay && params.PositionType != kiteconnect.PositionTypeOvernight {
				return exitcode.New(exitcode.Validation, "invalid --position-type; use day or overnight")
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

			converted, err := callWithAuthRetry(ctx, profileName, profile, func(client *kiteconnect.Client) (bool, error) {
				return client.ConvertPosition(params)
			})
			if err != nil {
				return err
			}
			if !converted {
				return exitcode.New(exitcode.API, "position conversion was not accepted")
			}

			printer := ctx.printer(cmd.OutOrStdout())
			if printer.IsJSON() {
				return printer.JSON(map[string]any{
					"status":        "ok",
					"exchange":      params.Exchange,
					"symbol":        params.TradingSymbol,
					"old_product":   params.OldProduct,
					"new_product":   params.NewProduct,
					"position_type": params.PositionType,
					"txn":           params.TransactionType,
					"qty":           params.Quantity,
				})
			}
			return printer.KV([][2]string{
				{"status", "ok"},
				{"exchange", params.Exchange},
				{"symbol", params.TradingSymbol},
				{"old_product", params.OldProduct},
				{"new_product", params.NewProduct},
				{"position_type", params.PositionType},
				{"txn", params.TransactionType},
				{"qty", intToString(params.Quantity)},
			})
		},
	}
	convertCmd.Flags().StringVar(&exchange, "exchange", "", "Exchange (NSE/BSE/NFO/...)")
	convertCmd.Flags().StringVar(&symbol, "symbol", "", "Trading symbol")
	convertCmd.Flags().StringVar(&oldProduct, "old-product", "", "Existing product type (CNC/MIS/NRML/MTF)")
	convertCmd.Flags().StringVar(&newProduct, "new-product", "", "Target product type (CNC/MIS/NRML/MTF)")
	convertCmd.Flags().StringVar(&positionType, "position-type", kiteconnect.PositionTypeDay, "Position type (day/overnight)")
	convertCmd.Flags().StringVar(&txnType, "txn", "", "Transaction type (BUY/SELL)")
	convertCmd.Flags().IntVar(&quantity, "qty", 0, "Quantity")

	positionsCmd.AddCommand(convertCmd)
	return positionsCmd
}
