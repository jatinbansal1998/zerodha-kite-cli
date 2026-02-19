package cli

import (
	"fmt"
	"strings"

	"github.com/jatinbansal1998/zerodha-kite-cli/internal/exitcode"
	"github.com/spf13/cobra"
	kiteconnect "github.com/zerodha/gokiteconnect/v4"
)

type gttFlags struct {
	exchange    string
	symbol      string
	lastPrice   float64
	txnType     string
	product     string
	triggerType string

	trigger    float64
	limitPrice float64
	quantity   float64

	lowerTrigger    float64
	lowerLimitPrice float64
	lowerQuantity   float64
	upperTrigger    float64
	upperLimitPrice float64
	upperQuantity   float64
}

func newGttCmd(opts *rootOptions) *cobra.Command {
	gttCmd := &cobra.Command{
		Use:   "gtt",
		Short: "Good Till Triggered (GTT) operations",
	}

	var placeFlags gttFlags
	placeCmd := &cobra.Command{
		Use:   "place",
		Short: "Place a new GTT trigger",
		RunE: func(cmd *cobra.Command, _ []string) error {
			params, err := gttParamsFromFlags(placeFlags)
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

			resp, err := callWithAuthRetry(ctx, profileName, profile, func(client *kiteconnect.Client) (kiteconnect.GTTResponse, error) {
				return client.PlaceGTT(params)
			})
			if err != nil {
				return err
			}

			printer := ctx.printer(cmd.OutOrStdout())
			if printer.IsJSON() {
				return printer.JSON(map[string]any{
					"status":     "ok",
					"trigger_id": resp.TriggerID,
				})
			}
			return printer.KV([][2]string{
				{"status", "ok"},
				{"trigger_id", intToString(resp.TriggerID)},
			})
		},
	}
	bindGTTFlags(placeCmd, &placeFlags)

	var modifyFlags gttFlags
	var modifyTriggerID int
	modifyCmd := &cobra.Command{
		Use:   "modify --trigger-id <id>",
		Short: "Modify an existing GTT trigger",
		RunE: func(cmd *cobra.Command, _ []string) error {
			triggerID, err := parseTriggerID(modifyTriggerID)
			if err != nil {
				return err
			}

			params, err := gttParamsFromFlags(modifyFlags)
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

			resp, err := callWithAuthRetry(ctx, profileName, profile, func(client *kiteconnect.Client) (kiteconnect.GTTResponse, error) {
				return client.ModifyGTT(triggerID, params)
			})
			if err != nil {
				return err
			}

			printer := ctx.printer(cmd.OutOrStdout())
			if printer.IsJSON() {
				return printer.JSON(map[string]any{
					"status":     "ok",
					"trigger_id": resp.TriggerID,
				})
			}
			return printer.KV([][2]string{
				{"status", "ok"},
				{"trigger_id", intToString(resp.TriggerID)},
			})
		},
	}
	bindGTTFlags(modifyCmd, &modifyFlags)
	modifyCmd.Flags().IntVar(&modifyTriggerID, "trigger-id", 0, "Trigger ID")

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all GTT triggers",
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

			gtts, err := callWithAuthRetry(ctx, profileName, profile, func(client *kiteconnect.Client) (kiteconnect.GTTs, error) {
				return client.GetGTTs()
			})
			if err != nil {
				return err
			}

			printer := ctx.printer(cmd.OutOrStdout())
			if printer.IsJSON() {
				return printer.JSON(gtts)
			}

			rows := make([][]string, 0, len(gtts))
			for _, gtt := range gtts {
				rows = append(rows, []string{
					intToString(gtt.ID),
					string(gtt.Type),
					gtt.Status,
					gtt.Condition.Tradingsymbol,
					gtt.Condition.Exchange,
					formatFloat(gtt.Condition.LastPrice),
					formatFloatSlice(gtt.Condition.TriggerValues),
					formatModelTime(gtt.CreatedAt.Time),
					formatModelTime(gtt.UpdatedAt.Time),
					formatModelTime(gtt.ExpiresAt.Time),
				})
			}
			if len(rows) == 0 {
				rows = append(rows, []string{"0", "-", "-", "-", "-", "0.00", "-", "-", "-", "-"})
			}
			return printer.Table(
				[]string{"TRIGGER_ID", "TYPE", "STATUS", "SYMBOL", "EXCHANGE", "LAST_PRICE", "TRIGGER_VALUES", "CREATED_AT", "UPDATED_AT", "EXPIRES_AT"},
				rows,
			)
		},
	}

	var showTriggerID int
	showCmd := &cobra.Command{
		Use:   "show --trigger-id <id>",
		Short: "Show details of one GTT trigger",
		RunE: func(cmd *cobra.Command, _ []string) error {
			triggerID, err := parseTriggerID(showTriggerID)
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

			gtt, err := callWithAuthRetry(ctx, profileName, profile, func(client *kiteconnect.Client) (kiteconnect.GTT, error) {
				return client.GetGTT(triggerID)
			})
			if err != nil {
				return err
			}

			printer := ctx.printer(cmd.OutOrStdout())
			if printer.IsJSON() {
				return printer.JSON(gtt)
			}

			if err := printer.KV([][2]string{
				{"trigger_id", intToString(gtt.ID)},
				{"type", string(gtt.Type)},
				{"status", gtt.Status},
				{"symbol", gtt.Condition.Tradingsymbol},
				{"exchange", gtt.Condition.Exchange},
				{"last_price", formatFloat(gtt.Condition.LastPrice)},
				{"trigger_values", formatFloatSlice(gtt.Condition.TriggerValues)},
				{"created_at", formatModelTime(gtt.CreatedAt.Time)},
				{"updated_at", formatModelTime(gtt.UpdatedAt.Time)},
				{"expires_at", formatModelTime(gtt.ExpiresAt.Time)},
				{"rejection_reason", strings.TrimSpace(gtt.Meta.RejectionReason)},
			}); err != nil {
				return err
			}
			if _, err := fmt.Fprintln(cmd.OutOrStdout()); err != nil {
				return err
			}

			rows := make([][]string, 0, len(gtt.Orders))
			for _, order := range gtt.Orders {
				rows = append(rows, []string{
					order.TradingSymbol,
					order.Exchange,
					order.TransactionType,
					formatFloat(order.Quantity),
					formatFloat(order.Price),
					order.OrderType,
					order.Product,
					order.Status,
				})
			}
			if len(rows) == 0 {
				rows = append(rows, []string{"-", "-", "-", "0.00", "0.00", "-", "-", "-"})
			}
			return printer.Table([]string{"SYMBOL", "EXCHANGE", "TXN", "QTY", "PRICE", "TYPE", "PRODUCT", "STATUS"}, rows)
		},
	}
	showCmd.Flags().IntVar(&showTriggerID, "trigger-id", 0, "Trigger ID")

	var deleteTriggerID int
	deleteCmd := &cobra.Command{
		Use:   "delete --trigger-id <id>",
		Short: "Delete a GTT trigger",
		RunE: func(cmd *cobra.Command, _ []string) error {
			triggerID, err := parseTriggerID(deleteTriggerID)
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

			resp, err := callWithAuthRetry(ctx, profileName, profile, func(client *kiteconnect.Client) (kiteconnect.GTTResponse, error) {
				return client.DeleteGTT(triggerID)
			})
			if err != nil {
				return err
			}

			printer := ctx.printer(cmd.OutOrStdout())
			if printer.IsJSON() {
				return printer.JSON(map[string]any{
					"status":     "ok",
					"trigger_id": resp.TriggerID,
				})
			}
			return printer.KV([][2]string{
				{"status", "ok"},
				{"trigger_id", intToString(resp.TriggerID)},
			})
		},
	}
	deleteCmd.Flags().IntVar(&deleteTriggerID, "trigger-id", 0, "Trigger ID")

	gttCmd.AddCommand(placeCmd, modifyCmd, listCmd, showCmd, deleteCmd)
	return gttCmd
}

func bindGTTFlags(cmd *cobra.Command, flags *gttFlags) {
	cmd.Flags().StringVar(&flags.exchange, "exchange", "", "Exchange (NSE/BSE/NFO/...)")
	cmd.Flags().StringVar(&flags.symbol, "symbol", "", "Trading symbol")
	cmd.Flags().Float64Var(&flags.lastPrice, "last-price", 0, "Last traded price")
	cmd.Flags().StringVar(&flags.txnType, "txn", "", "Transaction type (BUY/SELL)")
	cmd.Flags().StringVar(&flags.product, "product", kiteconnect.ProductCNC, "Product (CNC/MIS/NRML/MTF)")
	cmd.Flags().StringVar(&flags.triggerType, "type", string(kiteconnect.GTTTypeSingle), "GTT type (single/two-leg)")

	cmd.Flags().Float64Var(&flags.trigger, "trigger", 0, "Trigger price for single GTT")
	cmd.Flags().Float64Var(&flags.limitPrice, "limit-price", 0, "Limit price for single GTT")
	cmd.Flags().Float64Var(&flags.quantity, "qty", 0, "Quantity for single GTT")

	cmd.Flags().Float64Var(&flags.lowerTrigger, "lower-trigger", 0, "Lower trigger for two-leg GTT")
	cmd.Flags().Float64Var(&flags.lowerLimitPrice, "lower-limit-price", 0, "Lower limit price for two-leg GTT")
	cmd.Flags().Float64Var(&flags.lowerQuantity, "lower-qty", 0, "Lower quantity for two-leg GTT")
	cmd.Flags().Float64Var(&flags.upperTrigger, "upper-trigger", 0, "Upper trigger for two-leg GTT")
	cmd.Flags().Float64Var(&flags.upperLimitPrice, "upper-limit-price", 0, "Upper limit price for two-leg GTT")
	cmd.Flags().Float64Var(&flags.upperQuantity, "upper-qty", 0, "Upper quantity for two-leg GTT")
}

func gttParamsFromFlags(flags gttFlags) (kiteconnect.GTTParams, error) {
	params := kiteconnect.GTTParams{
		Tradingsymbol:   strings.TrimSpace(flags.symbol),
		Exchange:        normalizeUpper(flags.exchange),
		LastPrice:       flags.lastPrice,
		TransactionType: normalizeUpper(flags.txnType),
		Product:         normalizeUpper(flags.product),
	}
	if params.Product == "" {
		params.Product = kiteconnect.ProductCNC
	}

	if params.Exchange == "" || params.Tradingsymbol == "" || params.TransactionType == "" {
		return params, exitcode.New(exitcode.Validation, "--exchange, --symbol, and --txn are required")
	}
	if params.LastPrice <= 0 {
		return params, exitcode.New(exitcode.Validation, "--last-price must be greater than 0")
	}
	if params.TransactionType != kiteconnect.TransactionTypeBuy && params.TransactionType != kiteconnect.TransactionTypeSell {
		return params, exitcode.New(exitcode.Validation, "invalid --txn; use BUY or SELL")
	}
	switch params.Product {
	case kiteconnect.ProductCNC, kiteconnect.ProductMIS, kiteconnect.ProductNRML, kiteconnect.ProductMTF:
	default:
		return params, exitcode.New(exitcode.Validation, "invalid --product; use CNC, MIS, NRML, or MTF")
	}

	switch strings.ToLower(strings.TrimSpace(flags.triggerType)) {
	case string(kiteconnect.GTTTypeSingle):
		if flags.trigger <= 0 || flags.limitPrice <= 0 || flags.quantity <= 0 {
			return params, exitcode.New(exitcode.Validation, "--trigger, --limit-price, and --qty are required for type=single")
		}
		params.Trigger = &kiteconnect.GTTSingleLegTrigger{
			TriggerParams: kiteconnect.TriggerParams{
				TriggerValue: flags.trigger,
				LimitPrice:   flags.limitPrice,
				Quantity:     flags.quantity,
			},
		}
	case string(kiteconnect.GTTTypeOCO):
		if flags.lowerTrigger <= 0 || flags.lowerLimitPrice <= 0 || flags.lowerQuantity <= 0 {
			return params, exitcode.New(exitcode.Validation, "--lower-trigger, --lower-limit-price, and --lower-qty are required for type=two-leg")
		}
		if flags.upperTrigger <= 0 || flags.upperLimitPrice <= 0 || flags.upperQuantity <= 0 {
			return params, exitcode.New(exitcode.Validation, "--upper-trigger, --upper-limit-price, and --upper-qty are required for type=two-leg")
		}
		params.Trigger = &kiteconnect.GTTOneCancelsOtherTrigger{
			Lower: kiteconnect.TriggerParams{
				TriggerValue: flags.lowerTrigger,
				LimitPrice:   flags.lowerLimitPrice,
				Quantity:     flags.lowerQuantity,
			},
			Upper: kiteconnect.TriggerParams{
				TriggerValue: flags.upperTrigger,
				LimitPrice:   flags.upperLimitPrice,
				Quantity:     flags.upperQuantity,
			},
		}
	default:
		return params, exitcode.New(exitcode.Validation, "invalid --type; use single or two-leg")
	}

	return params, nil
}

func parseTriggerID(triggerID int) (int, error) {
	if triggerID <= 0 {
		return 0, exitcode.New(exitcode.Validation, "--trigger-id must be greater than 0")
	}
	return triggerID, nil
}

func formatFloatSlice(values []float64) string {
	if len(values) == 0 {
		return "-"
	}
	parts := make([]string, 0, len(values))
	for _, value := range values {
		parts = append(parts, formatFloat(value))
	}
	return strings.Join(parts, ",")
}
