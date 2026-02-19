package cli

import (
	"fmt"
	"strings"

	"github.com/jatinbansal1998/zerodha-kite-cli/internal/exitcode"
	"github.com/spf13/cobra"
	kiteconnect "github.com/zerodha/gokiteconnect/v4"
)

func newMarginsCmd(opts *rootOptions) *cobra.Command {
	var segment string

	marginsCmd := &cobra.Command{
		Use:   "margins",
		Short: "Margin summaries and calculators",
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

	var (
		orderFlags   marginOrderFlags
		orderCompact bool
	)
	orderCmd := &cobra.Command{
		Use:   "order",
		Short: "Estimate margin for an order",
		RunE: func(cmd *cobra.Command, _ []string) error {
			param, err := marginOrderParamFromFlags(orderFlags)
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

			margins, err := callWithAuthRetry(ctx, profileName, profile, func(client *kiteconnect.Client) ([]kiteconnect.OrderMargins, error) {
				return client.GetOrderMargins(kiteconnect.GetMarginParams{
					OrderParams: []kiteconnect.OrderMarginParam{param},
					Compact:     orderCompact,
				})
			})
			if err != nil {
				return err
			}

			printer := ctx.printer(cmd.OutOrStdout())
			if printer.IsJSON() {
				return printer.JSON(margins)
			}

			rows := make([][]string, 0, len(margins))
			for _, margin := range margins {
				rows = append(rows, []string{
					margin.TradingSymbol,
					margin.Exchange,
					formatFloat(margin.SPAN),
					formatFloat(margin.Exposure),
					formatFloat(margin.Total),
					formatFloat(margin.Charges.Total),
				})
			}
			if len(rows) == 0 {
				rows = append(rows, []string{"-", "-", "0.00", "0.00", "0.00", "0.00"})
			}
			return printer.Table([]string{"SYMBOL", "EXCHANGE", "SPAN", "EXPOSURE", "TOTAL_MARGIN", "TOTAL_CHARGES"}, rows)
		},
	}
	bindMarginOrderFlags(orderCmd, &orderFlags)
	orderCmd.Flags().BoolVar(&orderCompact, "compact", false, "Request compact response")

	var (
		basketFlags             marginOrderFlags
		basketCompact           bool
		basketConsiderPositions bool
	)
	basketCmd := &cobra.Command{
		Use:   "basket",
		Short: "Estimate basket margin (single leg in current CLI)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			param, err := marginOrderParamFromFlags(basketFlags)
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

			margins, err := callWithAuthRetry(ctx, profileName, profile, func(client *kiteconnect.Client) (kiteconnect.BasketMargins, error) {
				return client.GetBasketMargins(kiteconnect.GetBasketParams{
					OrderParams:       []kiteconnect.OrderMarginParam{param},
					Compact:           basketCompact,
					ConsiderPositions: basketConsiderPositions,
				})
			})
			if err != nil {
				return err
			}

			printer := ctx.printer(cmd.OutOrStdout())
			if printer.IsJSON() {
				return printer.JSON(margins)
			}

			if err := printer.KV([][2]string{
				{"initial_total_margin", formatFloat(margins.Initial.Total)},
				{"initial_total_charges", formatFloat(margins.Initial.Charges.Total)},
				{"final_total_margin", formatFloat(margins.Final.Total)},
				{"final_total_charges", formatFloat(margins.Final.Charges.Total)},
			}); err != nil {
				return err
			}
			if _, err := fmt.Fprintln(cmd.OutOrStdout()); err != nil {
				return err
			}

			rows := make([][]string, 0, len(margins.Orders))
			for _, margin := range margins.Orders {
				rows = append(rows, []string{
					margin.TradingSymbol,
					margin.Exchange,
					formatFloat(margin.SPAN),
					formatFloat(margin.Exposure),
					formatFloat(margin.Total),
					formatFloat(margin.Charges.Total),
				})
			}
			if len(rows) == 0 {
				rows = append(rows, []string{"-", "-", "0.00", "0.00", "0.00", "0.00"})
			}
			return printer.Table([]string{"SYMBOL", "EXCHANGE", "SPAN", "EXPOSURE", "TOTAL_MARGIN", "TOTAL_CHARGES"}, rows)
		},
	}
	bindMarginOrderFlags(basketCmd, &basketFlags)
	basketCmd.Flags().BoolVar(&basketCompact, "compact", false, "Request compact response")
	basketCmd.Flags().BoolVar(&basketConsiderPositions, "consider-positions", false, "Factor current positions in basket margin")

	var chargesFlags marginChargesFlags
	chargesCmd := &cobra.Command{
		Use:   "charges",
		Short: "Estimate order charges",
		RunE: func(cmd *cobra.Command, _ []string) error {
			param, err := marginChargesParamFromFlags(chargesFlags)
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

			charges, err := callWithAuthRetry(ctx, profileName, profile, func(client *kiteconnect.Client) ([]kiteconnect.OrderCharges, error) {
				return client.GetOrderCharges(kiteconnect.GetChargesParams{
					OrderParams: []kiteconnect.OrderChargesParam{param},
				})
			})
			if err != nil {
				return err
			}

			printer := ctx.printer(cmd.OutOrStdout())
			if printer.IsJSON() {
				return printer.JSON(charges)
			}

			rows := make([][]string, 0, len(charges))
			for _, charge := range charges {
				rows = append(rows, []string{
					charge.Tradingsymbol,
					charge.Exchange,
					charge.TransactionType,
					charge.Product,
					charge.OrderType,
					formatFloat(charge.Quantity),
					formatFloat(charge.Price),
					formatFloat(charge.Charges.Total),
				})
			}
			if len(rows) == 0 {
				rows = append(rows, []string{"-", "-", "-", "-", "-", "0.00", "0.00", "0.00"})
			}
			return printer.Table([]string{"SYMBOL", "EXCHANGE", "TXN", "PRODUCT", "TYPE", "QTY", "PRICE", "TOTAL_CHARGES"}, rows)
		},
	}
	bindMarginChargesFlags(chargesCmd, &chargesFlags)

	marginsCmd.AddCommand(orderCmd, basketCmd, chargesCmd)
	return marginsCmd
}

type marginOrderFlags struct {
	exchange  string
	symbol    string
	txnType   string
	orderType string
	product   string
	variety   string
	quantity  int
	price     float64
	trigger   float64
}

type marginChargesFlags struct {
	marginOrderFlags
	orderID      string
	averagePrice float64
}

func bindMarginOrderFlags(cmd *cobra.Command, flags *marginOrderFlags) {
	cmd.Flags().StringVar(&flags.exchange, "exchange", "", "Exchange (NSE/BSE/NFO/...)")
	cmd.Flags().StringVar(&flags.symbol, "symbol", "", "Trading symbol")
	cmd.Flags().StringVar(&flags.txnType, "txn", "", "Transaction type (BUY/SELL)")
	cmd.Flags().StringVar(&flags.orderType, "type", "", "Order type (MARKET/LIMIT/SL/SL-M)")
	cmd.Flags().StringVar(&flags.product, "product", "", "Product (CNC/MIS/NRML/MTF)")
	cmd.Flags().StringVar(&flags.variety, "variety", kiteconnect.VarietyRegular, "Order variety")
	cmd.Flags().IntVar(&flags.quantity, "qty", 0, "Quantity")
	cmd.Flags().Float64Var(&flags.price, "price", 0, "Price")
	cmd.Flags().Float64Var(&flags.trigger, "trigger-price", 0, "Trigger price")
}

func bindMarginChargesFlags(cmd *cobra.Command, flags *marginChargesFlags) {
	bindMarginOrderFlags(cmd, &flags.marginOrderFlags)
	cmd.Flags().StringVar(&flags.orderID, "order-id", "", "Order ID (optional)")
	cmd.Flags().Float64Var(&flags.averagePrice, "avg-price", 0, "Average execution price")
}

func marginOrderParamFromFlags(flags marginOrderFlags) (kiteconnect.OrderMarginParam, error) {
	param := kiteconnect.OrderMarginParam{
		Exchange:        normalizeUpper(flags.exchange),
		Tradingsymbol:   strings.TrimSpace(flags.symbol),
		TransactionType: normalizeUpper(flags.txnType),
		Variety:         normalizeVariety(flags.variety),
		Product:         normalizeUpper(flags.product),
		OrderType:       normalizeUpper(flags.orderType),
		Quantity:        float64(flags.quantity),
		Price:           flags.price,
		TriggerPrice:    flags.trigger,
	}

	if param.Exchange == "" || param.Tradingsymbol == "" || param.TransactionType == "" || param.Product == "" || param.OrderType == "" {
		return param, exitcode.New(exitcode.Validation, "--exchange, --symbol, --txn, --product, and --type are required")
	}
	if flags.quantity <= 0 {
		return param, exitcode.New(exitcode.Validation, "--qty must be greater than 0")
	}
	if param.TransactionType != kiteconnect.TransactionTypeBuy && param.TransactionType != kiteconnect.TransactionTypeSell {
		return param, exitcode.New(exitcode.Validation, "invalid --txn; use BUY or SELL")
	}

	switch param.OrderType {
	case kiteconnect.OrderTypeMarket:
	case kiteconnect.OrderTypeLimit:
		if param.Price <= 0 {
			return param, exitcode.New(exitcode.Validation, "--price is required for LIMIT orders")
		}
	case kiteconnect.OrderTypeSL:
		if param.Price <= 0 || param.TriggerPrice <= 0 {
			return param, exitcode.New(exitcode.Validation, "--price and --trigger-price are required for SL orders")
		}
	case kiteconnect.OrderTypeSLM:
		if param.TriggerPrice <= 0 {
			return param, exitcode.New(exitcode.Validation, "--trigger-price is required for SL-M orders")
		}
	default:
		return param, exitcode.New(exitcode.Validation, "invalid --type; use MARKET, LIMIT, SL, or SL-M")
	}

	return param, nil
}

func marginChargesParamFromFlags(flags marginChargesFlags) (kiteconnect.OrderChargesParam, error) {
	marginParam, err := marginOrderParamFromFlags(flags.marginOrderFlags)
	if err != nil {
		return kiteconnect.OrderChargesParam{}, err
	}

	if flags.averagePrice <= 0 {
		return kiteconnect.OrderChargesParam{}, exitcode.New(exitcode.Validation, "--avg-price must be greater than 0")
	}

	return kiteconnect.OrderChargesParam{
		OrderID:         strings.TrimSpace(flags.orderID),
		Exchange:        marginParam.Exchange,
		Tradingsymbol:   marginParam.Tradingsymbol,
		TransactionType: marginParam.TransactionType,
		Variety:         marginParam.Variety,
		Product:         marginParam.Product,
		OrderType:       marginParam.OrderType,
		Quantity:        marginParam.Quantity,
		AveragePrice:    flags.averagePrice,
	}, nil
}
