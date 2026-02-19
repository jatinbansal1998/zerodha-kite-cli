package cli

import (
	"fmt"
	"strings"

	"github.com/jatinbansal1998/zerodha-kite-cli/internal/exitcode"
	"github.com/spf13/cobra"
	kiteconnect "github.com/zerodha/gokiteconnect/v4"
)

type orderFlags struct {
	exchange    string
	symbol      string
	txnType     string
	orderType   string
	product     string
	validity    string
	validityTTL int
	variety     string
	quantity    int
	price       float64
	trigger     float64
	tag         string
}

func newOrderCmd(opts *rootOptions) *cobra.Command {
	orderCmd := &cobra.Command{
		Use:   "order",
		Short: "Place/modify/cancel/exit individual orders",
	}

	var placeFlags orderFlags
	placeCmd := &cobra.Command{
		Use:   "place",
		Short: "Place a new order",
		RunE: func(cmd *cobra.Command, _ []string) error {
			variety, params, err := placeOrderParams(placeFlags)
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

			resp, err := callWithAuthRetry(ctx, profileName, profile, func(client *kiteconnect.Client) (kiteconnect.OrderResponse, error) {
				return client.PlaceOrder(variety, params)
			})
			if err != nil {
				return err
			}

			printer := ctx.printer(cmd.OutOrStdout())
			if printer.IsJSON() {
				return printer.JSON(map[string]any{
					"status":   "ok",
					"order_id": resp.OrderID,
					"variety":  variety,
				})
			}
			return printer.KV([][2]string{
				{"status", "ok"},
				{"order_id", resp.OrderID},
				{"variety", variety},
			})
		},
	}
	bindPlaceFlags(placeCmd, &placeFlags)

	var modifyFlags orderFlags
	var modifyOrderID string
	modifyCmd := &cobra.Command{
		Use:   "modify --order-id <id>",
		Short: "Modify an existing order",
		RunE: func(cmd *cobra.Command, _ []string) error {
			orderID := strings.TrimSpace(modifyOrderID)
			if orderID == "" {
				return exitcode.New(exitcode.Validation, "--order-id is required")
			}
			variety, params, err := modifyOrderParams(modifyFlags)
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

			resp, err := callWithAuthRetry(ctx, profileName, profile, func(client *kiteconnect.Client) (kiteconnect.OrderResponse, error) {
				return client.ModifyOrder(variety, orderID, params)
			})
			if err != nil {
				return err
			}

			printer := ctx.printer(cmd.OutOrStdout())
			if printer.IsJSON() {
				return printer.JSON(map[string]any{
					"status":   "ok",
					"order_id": resp.OrderID,
					"variety":  variety,
				})
			}
			return printer.KV([][2]string{
				{"status", "ok"},
				{"order_id", resp.OrderID},
				{"variety", variety},
			})
		},
	}
	bindModifyFlags(modifyCmd, &modifyFlags, &modifyOrderID)

	var cancelOrderID string
	var cancelVariety string
	var parentOrderID string
	cancelCmd := &cobra.Command{
		Use:   "cancel --order-id <id>",
		Short: "Cancel an order",
		RunE: func(cmd *cobra.Command, _ []string) error {
			orderID := strings.TrimSpace(cancelOrderID)
			if orderID == "" {
				return exitcode.New(exitcode.Validation, "--order-id is required")
			}
			variety := normalizeVariety(cancelVariety)
			var parentID *string
			if strings.TrimSpace(parentOrderID) != "" {
				parentID = new(strings.TrimSpace(parentOrderID))
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

			resp, err := callWithAuthRetry(ctx, profileName, profile, func(client *kiteconnect.Client) (kiteconnect.OrderResponse, error) {
				return client.CancelOrder(variety, orderID, parentID)
			})
			if err != nil {
				return err
			}

			printer := ctx.printer(cmd.OutOrStdout())
			if printer.IsJSON() {
				return printer.JSON(map[string]any{
					"status":   "ok",
					"order_id": resp.OrderID,
					"variety":  variety,
				})
			}
			return printer.KV([][2]string{
				{"status", "ok"},
				{"order_id", resp.OrderID},
				{"variety", variety},
			})
		},
	}
	cancelCmd.Flags().StringVar(&cancelOrderID, "order-id", "", "Order ID")
	cancelCmd.Flags().StringVar(&cancelVariety, "variety", kiteconnect.VarietyRegular, "Order variety")
	cancelCmd.Flags().StringVar(&parentOrderID, "parent-order-id", "", "Parent order ID (for bracket/cover orders)")

	var exitOrderID string
	var exitVariety string
	var exitParentOrderID string
	exitCmd := &cobra.Command{
		Use:   "exit --order-id <id>",
		Short: "Exit an order (alias of cancel in Kite)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			orderID := strings.TrimSpace(exitOrderID)
			if orderID == "" {
				return exitcode.New(exitcode.Validation, "--order-id is required")
			}
			variety := normalizeVariety(exitVariety)
			var parentID *string
			if strings.TrimSpace(exitParentOrderID) != "" {
				parentID = new(strings.TrimSpace(exitParentOrderID))
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

			resp, err := callWithAuthRetry(ctx, profileName, profile, func(client *kiteconnect.Client) (kiteconnect.OrderResponse, error) {
				return client.ExitOrder(variety, orderID, parentID)
			})
			if err != nil {
				return err
			}

			printer := ctx.printer(cmd.OutOrStdout())
			if printer.IsJSON() {
				return printer.JSON(map[string]any{
					"status":   "ok",
					"order_id": resp.OrderID,
					"variety":  variety,
				})
			}
			return printer.KV([][2]string{
				{"status", "ok"},
				{"order_id", resp.OrderID},
				{"variety", variety},
			})
		},
	}
	exitCmd.Flags().StringVar(&exitOrderID, "order-id", "", "Order ID")
	exitCmd.Flags().StringVar(&exitVariety, "variety", kiteconnect.VarietyRegular, "Order variety")
	exitCmd.Flags().StringVar(&exitParentOrderID, "parent-order-id", "", "Parent order ID (for bracket/cover orders)")

	orderCmd.AddCommand(placeCmd, modifyCmd, cancelCmd, exitCmd)
	return orderCmd
}

func bindPlaceFlags(cmd *cobra.Command, flags *orderFlags) {
	cmd.Flags().StringVar(&flags.exchange, "exchange", "", "Exchange (NSE/BSE/NFO/...)")
	cmd.Flags().StringVar(&flags.symbol, "symbol", "", "Trading symbol")
	cmd.Flags().StringVar(&flags.txnType, "txn", "", "Transaction type (BUY/SELL)")
	cmd.Flags().StringVar(&flags.orderType, "type", "", "Order type (MARKET/LIMIT/SL/SL-M)")
	cmd.Flags().StringVar(&flags.product, "product", "", "Product (CNC/MIS/NRML/MTF)")
	cmd.Flags().IntVar(&flags.quantity, "qty", 0, "Quantity")
	cmd.Flags().Float64Var(&flags.price, "price", 0, "Limit price")
	cmd.Flags().Float64Var(&flags.trigger, "trigger-price", 0, "Trigger price")
	cmd.Flags().StringVar(&flags.validity, "validity", kiteconnect.ValidityDay, "Validity (DAY/IOC/TTL)")
	cmd.Flags().IntVar(&flags.validityTTL, "validity-ttl", 0, "Validity TTL in minutes when validity=TTL")
	cmd.Flags().StringVar(&flags.variety, "variety", kiteconnect.VarietyRegular, "Order variety")
	cmd.Flags().StringVar(&flags.tag, "tag", "", "Custom order tag")
}

func bindModifyFlags(cmd *cobra.Command, flags *orderFlags, orderID *string) {
	cmd.Flags().StringVar(orderID, "order-id", "", "Order ID")
	cmd.Flags().StringVar(&flags.exchange, "exchange", "", "Exchange")
	cmd.Flags().StringVar(&flags.symbol, "symbol", "", "Trading symbol")
	cmd.Flags().StringVar(&flags.txnType, "txn", "", "Transaction type (BUY/SELL)")
	cmd.Flags().StringVar(&flags.orderType, "type", "", "Order type (MARKET/LIMIT/SL/SL-M)")
	cmd.Flags().StringVar(&flags.product, "product", "", "Product")
	cmd.Flags().IntVar(&flags.quantity, "qty", 0, "Quantity")
	cmd.Flags().Float64Var(&flags.price, "price", 0, "Price")
	cmd.Flags().Float64Var(&flags.trigger, "trigger-price", 0, "Trigger price")
	cmd.Flags().StringVar(&flags.validity, "validity", "", "Validity")
	cmd.Flags().IntVar(&flags.validityTTL, "validity-ttl", 0, "TTL validity")
	cmd.Flags().StringVar(&flags.variety, "variety", kiteconnect.VarietyRegular, "Order variety")
	cmd.Flags().StringVar(&flags.tag, "tag", "", "Custom order tag")
}

func placeOrderParams(flags orderFlags) (string, kiteconnect.OrderParams, error) {
	var params kiteconnect.OrderParams

	params.Exchange = normalizeUpper(flags.exchange)
	params.Tradingsymbol = strings.TrimSpace(flags.symbol)
	params.TransactionType = normalizeUpper(flags.txnType)
	params.OrderType = normalizeUpper(flags.orderType)
	params.Product = normalizeUpper(flags.product)
	params.Quantity = flags.quantity
	params.Price = flags.price
	params.TriggerPrice = flags.trigger
	params.Validity = normalizeUpper(flags.validity)
	params.ValidityTTL = flags.validityTTL
	params.Tag = strings.TrimSpace(flags.tag)
	variety := normalizeVariety(flags.variety)

	if params.Exchange == "" || params.Tradingsymbol == "" || params.TransactionType == "" || params.OrderType == "" || params.Product == "" {
		return "", params, exitcode.New(exitcode.Validation, "--exchange, --symbol, --txn, --type, --product are required")
	}
	if params.Quantity <= 0 {
		return "", params, exitcode.New(exitcode.Validation, "--qty must be greater than 0")
	}
	if params.Validity == "" {
		params.Validity = kiteconnect.ValidityDay
	}

	switch params.OrderType {
	case kiteconnect.OrderTypeLimit:
		if params.Price <= 0 {
			return "", params, exitcode.New(exitcode.Validation, "--price is required for LIMIT orders")
		}
	case kiteconnect.OrderTypeSL:
		if params.Price <= 0 || params.TriggerPrice <= 0 {
			return "", params, exitcode.New(exitcode.Validation, "--price and --trigger-price are required for SL orders")
		}
	case kiteconnect.OrderTypeSLM:
		if params.TriggerPrice <= 0 {
			return "", params, exitcode.New(exitcode.Validation, "--trigger-price is required for SL-M orders")
		}
	case kiteconnect.OrderTypeMarket:
	default:
		return "", params, exitcode.New(exitcode.Validation, "invalid --type; use MARKET, LIMIT, SL, or SL-M")
	}

	if params.TransactionType != kiteconnect.TransactionTypeBuy && params.TransactionType != kiteconnect.TransactionTypeSell {
		return "", params, exitcode.New(exitcode.Validation, "invalid --txn; use BUY or SELL")
	}
	if params.Validity != kiteconnect.ValidityDay && params.Validity != kiteconnect.ValidityIOC && params.Validity != kiteconnect.ValidityTTL {
		return "", params, exitcode.New(exitcode.Validation, "invalid --validity; use DAY, IOC, or TTL")
	}
	if params.Validity == kiteconnect.ValidityTTL && params.ValidityTTL <= 0 {
		return "", params, exitcode.New(exitcode.Validation, "--validity-ttl must be greater than 0 when validity is TTL")
	}

	return variety, params, nil
}

func modifyOrderParams(flags orderFlags) (string, kiteconnect.OrderParams, error) {
	var params kiteconnect.OrderParams
	params.Exchange = normalizeUpper(flags.exchange)
	params.Tradingsymbol = strings.TrimSpace(flags.symbol)
	params.TransactionType = normalizeUpper(flags.txnType)
	params.OrderType = normalizeUpper(flags.orderType)
	params.Product = normalizeUpper(flags.product)
	params.Quantity = flags.quantity
	params.Price = flags.price
	params.TriggerPrice = flags.trigger
	params.Validity = normalizeUpper(flags.validity)
	params.ValidityTTL = flags.validityTTL
	params.Tag = strings.TrimSpace(flags.tag)

	variety := normalizeVariety(flags.variety)
	if params.Exchange == "" &&
		params.Tradingsymbol == "" &&
		params.TransactionType == "" &&
		params.OrderType == "" &&
		params.Product == "" &&
		params.Quantity == 0 &&
		params.Price == 0 &&
		params.TriggerPrice == 0 &&
		params.Validity == "" &&
		params.ValidityTTL == 0 &&
		params.Tag == "" {
		return "", params, exitcode.New(exitcode.Validation, "at least one field must be provided to modify an order")
	}

	if params.OrderType != "" {
		switch params.OrderType {
		case kiteconnect.OrderTypeMarket, kiteconnect.OrderTypeLimit, kiteconnect.OrderTypeSL, kiteconnect.OrderTypeSLM:
		default:
			return "", params, exitcode.New(exitcode.Validation, "invalid --type; use MARKET, LIMIT, SL, or SL-M")
		}
	}
	if params.TransactionType != "" &&
		params.TransactionType != kiteconnect.TransactionTypeBuy &&
		params.TransactionType != kiteconnect.TransactionTypeSell {
		return "", params, exitcode.New(exitcode.Validation, "invalid --txn; use BUY or SELL")
	}
	if params.Validity != "" &&
		params.Validity != kiteconnect.ValidityDay &&
		params.Validity != kiteconnect.ValidityIOC &&
		params.Validity != kiteconnect.ValidityTTL {
		return "", params, exitcode.New(exitcode.Validation, "invalid --validity; use DAY, IOC, or TTL")
	}

	return variety, params, nil
}

func normalizeUpper(v string) string {
	return strings.ToUpper(strings.TrimSpace(v))
}

func normalizeVariety(v string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	if v == "" {
		return kiteconnect.VarietyRegular
	}
	return v
}

func formatFloat(v float64) string {
	return fmt.Sprintf("%.2f", v)
}
