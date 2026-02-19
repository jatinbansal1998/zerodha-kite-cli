package cli

import (
	"strings"
	"time"

	"github.com/jatinbansal1998/zerodha-kite-cli/internal/exitcode"
	"github.com/spf13/cobra"
	kiteconnect "github.com/zerodha/gokiteconnect/v4"
)

type mfOrderFlags struct {
	symbol   string
	txnType  string
	quantity float64
	amount   float64
	tag      string
}

type mfSIPPlaceFlags struct {
	symbol        string
	amount        float64
	instalments   int
	frequency     string
	instalmentDay int
	initialAmount float64
	triggerPrice  float64
	stepUp        string
	sipType       string
	tag           string
}

type mfSIPModifyFlags struct {
	sipID         string
	amount        float64
	frequency     string
	instalmentDay int
	instalments   int
	stepUp        string
	status        string
}

func newMFCmd(opts *rootOptions) *cobra.Command {
	mfCmd := &cobra.Command{
		Use:   "mf",
		Short: "Mutual fund orders, SIPs, and holdings",
	}

	ordersCmd := &cobra.Command{
		Use:   "orders",
		Short: "Mutual fund order operations",
	}

	var orderPlaceFlags mfOrderFlags
	orderPlaceCmd := &cobra.Command{
		Use:   "place",
		Short: "Place a mutual fund order",
		RunE: func(cmd *cobra.Command, _ []string) error {
			params, err := mfOrderParamsFromFlags(orderPlaceFlags)
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

			resp, err := callWithAuthRetry(ctx, profileName, profile, func(client *kiteconnect.Client) (kiteconnect.MFOrderResponse, error) {
				return client.PlaceMFOrder(params)
			})
			if err != nil {
				return err
			}

			printer := ctx.printer(cmd.OutOrStdout())
			if printer.IsJSON() {
				return printer.JSON(map[string]any{
					"status":   "ok",
					"order_id": resp.OrderID,
				})
			}
			return printer.KV([][2]string{
				{"status", "ok"},
				{"order_id", resp.OrderID},
			})
		},
	}
	orderPlaceCmd.Flags().StringVar(&orderPlaceFlags.symbol, "symbol", "", "MF trading symbol")
	orderPlaceCmd.Flags().StringVar(&orderPlaceFlags.txnType, "txn", "", "Transaction type (BUY/SELL)")
	orderPlaceCmd.Flags().Float64Var(&orderPlaceFlags.quantity, "qty", 0, "Quantity (optional if --amount is provided)")
	orderPlaceCmd.Flags().Float64Var(&orderPlaceFlags.amount, "amount", 0, "Amount (optional if --qty is provided)")
	orderPlaceCmd.Flags().StringVar(&orderPlaceFlags.tag, "tag", "", "Custom order tag")

	var listFrom string
	var listTo string
	var listLimit int
	orderListCmd := &cobra.Command{
		Use:   "list",
		Short: "List mutual fund orders",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := validateLimit(listLimit); err != nil {
				return err
			}

			from := strings.TrimSpace(listFrom)
			to := strings.TrimSpace(listTo)
			if (from == "") != (to == "") {
				return exitcode.New(exitcode.Validation, "--from and --to must be provided together")
			}
			if from != "" {
				if _, err := parseDateYYYYMMDD(from, "--from"); err != nil {
					return err
				}
				if _, err := parseDateYYYYMMDD(to, "--to"); err != nil {
					return err
				}
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

			var orders kiteconnect.MFOrders
			if from == "" {
				orders, err = callWithAuthRetry(ctx, profileName, profile, func(client *kiteconnect.Client) (kiteconnect.MFOrders, error) {
					return client.GetMFOrders()
				})
			} else {
				orders, err = callWithAuthRetry(ctx, profileName, profile, func(client *kiteconnect.Client) (kiteconnect.MFOrders, error) {
					return client.GetMFOrdersByDate(from, to)
				})
			}
			if err != nil {
				return err
			}
			orders = applyLimit(orders, listLimit)

			printer := ctx.printer(cmd.OutOrStdout())
			if printer.IsJSON() {
				return printer.JSON(orders)
			}

			rows := make([][]string, 0, len(orders))
			for _, order := range orders {
				rows = append(rows, []string{
					order.OrderID,
					order.Tradingsymbol,
					order.TransactionType,
					formatFloat(order.Quantity),
					formatFloat(order.Amount),
					order.Status,
					formatModelTime(order.OrderTimestamp.Time),
				})
			}
			if len(rows) == 0 {
				rows = append(rows, []string{"-", "-", "-", "0.00", "0.00", "-", "-"})
			}
			return printer.Table([]string{"ORDER_ID", "SYMBOL", "TXN", "QTY", "AMOUNT", "STATUS", "ORDERED_AT"}, rows)
		},
	}
	orderListCmd.Flags().StringVar(&listFrom, "from", "", "Start date (YYYY-MM-DD)")
	orderListCmd.Flags().StringVar(&listTo, "to", "", "End date (YYYY-MM-DD)")
	orderListCmd.Flags().IntVar(&listLimit, "limit", 0, "Limit number of rows (0 = no limit)")

	var showOrderID string
	orderShowCmd := &cobra.Command{
		Use:   "show --order-id <id>",
		Short: "Show details of one mutual fund order",
		RunE: func(cmd *cobra.Command, _ []string) error {
			orderID := strings.TrimSpace(showOrderID)
			if orderID == "" {
				return exitcode.New(exitcode.Validation, "--order-id is required")
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

			order, err := callWithAuthRetry(ctx, profileName, profile, func(client *kiteconnect.Client) (kiteconnect.MFOrder, error) {
				return client.GetMFOrderInfo(orderID)
			})
			if err != nil {
				return err
			}

			printer := ctx.printer(cmd.OutOrStdout())
			if printer.IsJSON() {
				return printer.JSON(order)
			}
			return printer.KV([][2]string{
				{"order_id", order.OrderID},
				{"symbol", order.Tradingsymbol},
				{"txn", order.TransactionType},
				{"qty", formatFloat(order.Quantity)},
				{"amount", formatFloat(order.Amount)},
				{"status", order.Status},
				{"status_message", order.StatusMessage},
				{"ordered_at", formatModelTime(order.OrderTimestamp.Time)},
			})
		},
	}
	orderShowCmd.Flags().StringVar(&showOrderID, "order-id", "", "Order ID")

	var cancelOrderID string
	orderCancelCmd := &cobra.Command{
		Use:   "cancel --order-id <id>",
		Short: "Cancel a mutual fund order",
		RunE: func(cmd *cobra.Command, _ []string) error {
			orderID := strings.TrimSpace(cancelOrderID)
			if orderID == "" {
				return exitcode.New(exitcode.Validation, "--order-id is required")
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

			resp, err := callWithAuthRetry(ctx, profileName, profile, func(client *kiteconnect.Client) (kiteconnect.MFOrderResponse, error) {
				return client.CancelMFOrder(orderID)
			})
			if err != nil {
				return err
			}

			printer := ctx.printer(cmd.OutOrStdout())
			if printer.IsJSON() {
				return printer.JSON(map[string]any{
					"status":   "ok",
					"order_id": resp.OrderID,
				})
			}
			return printer.KV([][2]string{
				{"status", "ok"},
				{"order_id", resp.OrderID},
			})
		},
	}
	orderCancelCmd.Flags().StringVar(&cancelOrderID, "order-id", "", "Order ID")

	ordersCmd.AddCommand(orderPlaceCmd, orderListCmd, orderShowCmd, orderCancelCmd)

	sipsCmd := &cobra.Command{
		Use:   "sips",
		Short: "Mutual fund SIP operations",
	}

	var sipPlaceFlags mfSIPPlaceFlags
	sipPlaceCmd := &cobra.Command{
		Use:   "place",
		Short: "Place a mutual fund SIP",
		RunE: func(cmd *cobra.Command, _ []string) error {
			params, err := mfSIPPlaceParamsFromFlags(sipPlaceFlags)
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

			resp, err := callWithAuthRetry(ctx, profileName, profile, func(client *kiteconnect.Client) (kiteconnect.MFSIPResponse, error) {
				return client.PlaceMFSIP(params)
			})
			if err != nil {
				return err
			}

			printer := ctx.printer(cmd.OutOrStdout())
			if printer.IsJSON() {
				return printer.JSON(map[string]any{
					"status":   "ok",
					"sip_id":   resp.SIPID,
					"order_id": resp.OrderID,
				})
			}

			orderID := ""
			if resp.OrderID != nil {
				orderID = *resp.OrderID
			}
			return printer.KV([][2]string{
				{"status", "ok"},
				{"sip_id", resp.SIPID},
				{"order_id", orderID},
			})
		},
	}
	sipPlaceCmd.Flags().StringVar(&sipPlaceFlags.symbol, "symbol", "", "MF trading symbol")
	sipPlaceCmd.Flags().Float64Var(&sipPlaceFlags.amount, "amount", 0, "Instalment amount")
	sipPlaceCmd.Flags().IntVar(&sipPlaceFlags.instalments, "instalments", 0, "Number of instalments")
	sipPlaceCmd.Flags().StringVar(&sipPlaceFlags.frequency, "frequency", "", "SIP frequency")
	sipPlaceCmd.Flags().IntVar(&sipPlaceFlags.instalmentDay, "instalment-day", 0, "Instalment day (1-31)")
	sipPlaceCmd.Flags().Float64Var(&sipPlaceFlags.initialAmount, "initial-amount", 0, "Optional initial amount")
	sipPlaceCmd.Flags().Float64Var(&sipPlaceFlags.triggerPrice, "trigger-price", 0, "Optional trigger price")
	sipPlaceCmd.Flags().StringVar(&sipPlaceFlags.stepUp, "step-up", "", "Optional step-up value")
	sipPlaceCmd.Flags().StringVar(&sipPlaceFlags.sipType, "sip-type", "", "Optional SIP type (e.g. regular/flexi)")
	sipPlaceCmd.Flags().StringVar(&sipPlaceFlags.tag, "tag", "", "Custom SIP tag")

	var sipModifyFlags mfSIPModifyFlags
	sipModifyCmd := &cobra.Command{
		Use:   "modify --sip-id <id>",
		Short: "Modify a mutual fund SIP",
		RunE: func(cmd *cobra.Command, _ []string) error {
			sipID := strings.TrimSpace(sipModifyFlags.sipID)
			if sipID == "" {
				return exitcode.New(exitcode.Validation, "--sip-id is required")
			}

			params, err := mfSIPModifyParamsFromFlags(sipModifyFlags)
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

			resp, err := callWithAuthRetry(ctx, profileName, profile, func(client *kiteconnect.Client) (kiteconnect.MFSIPResponse, error) {
				return client.ModifyMFSIP(sipID, params)
			})
			if err != nil {
				return err
			}

			printer := ctx.printer(cmd.OutOrStdout())
			if printer.IsJSON() {
				return printer.JSON(map[string]any{
					"status":   "ok",
					"sip_id":   resp.SIPID,
					"order_id": resp.OrderID,
				})
			}

			orderID := ""
			if resp.OrderID != nil {
				orderID = *resp.OrderID
			}
			return printer.KV([][2]string{
				{"status", "ok"},
				{"sip_id", resp.SIPID},
				{"order_id", orderID},
			})
		},
	}
	sipModifyCmd.Flags().StringVar(&sipModifyFlags.sipID, "sip-id", "", "SIP ID")
	sipModifyCmd.Flags().Float64Var(&sipModifyFlags.amount, "amount", 0, "Instalment amount")
	sipModifyCmd.Flags().StringVar(&sipModifyFlags.frequency, "frequency", "", "SIP frequency")
	sipModifyCmd.Flags().IntVar(&sipModifyFlags.instalmentDay, "instalment-day", 0, "Instalment day (1-31)")
	sipModifyCmd.Flags().IntVar(&sipModifyFlags.instalments, "instalments", 0, "Number of instalments")
	sipModifyCmd.Flags().StringVar(&sipModifyFlags.stepUp, "step-up", "", "Step-up value")
	sipModifyCmd.Flags().StringVar(&sipModifyFlags.status, "status", "", "SIP status")

	var cancelSipID string
	sipCancelCmd := &cobra.Command{
		Use:   "cancel --sip-id <id>",
		Short: "Cancel a mutual fund SIP",
		RunE: func(cmd *cobra.Command, _ []string) error {
			sipID := strings.TrimSpace(cancelSipID)
			if sipID == "" {
				return exitcode.New(exitcode.Validation, "--sip-id is required")
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

			resp, err := callWithAuthRetry(ctx, profileName, profile, func(client *kiteconnect.Client) (kiteconnect.MFSIPResponse, error) {
				return client.CancelMFSIP(sipID)
			})
			if err != nil {
				return err
			}

			printer := ctx.printer(cmd.OutOrStdout())
			if printer.IsJSON() {
				return printer.JSON(map[string]any{
					"status":   "ok",
					"sip_id":   resp.SIPID,
					"order_id": resp.OrderID,
				})
			}

			orderID := ""
			if resp.OrderID != nil {
				orderID = *resp.OrderID
			}
			return printer.KV([][2]string{
				{"status", "ok"},
				{"sip_id", resp.SIPID},
				{"order_id", orderID},
			})
		},
	}
	sipCancelCmd.Flags().StringVar(&cancelSipID, "sip-id", "", "SIP ID")

	var sipListLimit int
	sipListCmd := &cobra.Command{
		Use:   "list",
		Short: "List mutual fund SIPs",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := validateLimit(sipListLimit); err != nil {
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

			sips, err := callWithAuthRetry(ctx, profileName, profile, func(client *kiteconnect.Client) (kiteconnect.MFSIPs, error) {
				return client.GetMFSIPs()
			})
			if err != nil {
				return err
			}
			sips = applyLimit(sips, sipListLimit)

			printer := ctx.printer(cmd.OutOrStdout())
			if printer.IsJSON() {
				return printer.JSON(sips)
			}

			rows := make([][]string, 0, len(sips))
			for _, sip := range sips {
				rows = append(rows, []string{
					sip.ID,
					sip.Tradingsymbol,
					sip.Frequency,
					formatFloat(sip.InstalmentAmount),
					intToString(sip.Instalments),
					intToString(sip.PendingInstalments),
					sip.Status,
					sip.NextInstalment,
				})
			}
			if len(rows) == 0 {
				rows = append(rows, []string{"-", "-", "-", "0.00", "0", "0", "-", "-"})
			}
			return printer.Table([]string{"SIP_ID", "SYMBOL", "FREQUENCY", "AMOUNT", "INSTALMENTS", "PENDING", "STATUS", "NEXT_INSTALMENT"}, rows)
		},
	}
	sipListCmd.Flags().IntVar(&sipListLimit, "limit", 0, "Limit number of rows (0 = no limit)")

	var showSipID string
	sipShowCmd := &cobra.Command{
		Use:   "show --sip-id <id>",
		Short: "Show details of one mutual fund SIP",
		RunE: func(cmd *cobra.Command, _ []string) error {
			sipID := strings.TrimSpace(showSipID)
			if sipID == "" {
				return exitcode.New(exitcode.Validation, "--sip-id is required")
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

			sip, err := callWithAuthRetry(ctx, profileName, profile, func(client *kiteconnect.Client) (kiteconnect.MFSIP, error) {
				return client.GetMFSIPInfo(sipID)
			})
			if err != nil {
				return err
			}

			printer := ctx.printer(cmd.OutOrStdout())
			if printer.IsJSON() {
				return printer.JSON(sip)
			}
			return printer.KV([][2]string{
				{"sip_id", sip.ID},
				{"symbol", sip.Tradingsymbol},
				{"status", sip.Status},
				{"frequency", sip.Frequency},
				{"instalment_amount", formatFloat(sip.InstalmentAmount)},
				{"instalments", intToString(sip.Instalments)},
				{"pending_instalments", intToString(sip.PendingInstalments)},
				{"next_instalment", sip.NextInstalment},
			})
		},
	}
	sipShowCmd.Flags().StringVar(&showSipID, "sip-id", "", "SIP ID")

	sipsCmd.AddCommand(sipPlaceCmd, sipModifyCmd, sipCancelCmd, sipListCmd, sipShowCmd)

	var holdingsLimit int
	holdingsCmd := &cobra.Command{
		Use:   "holdings",
		Short: "List mutual fund holdings",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := validateLimit(holdingsLimit); err != nil {
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

			holdings, err := callWithAuthRetry(ctx, profileName, profile, func(client *kiteconnect.Client) (kiteconnect.MFHoldings, error) {
				return client.GetMFHoldings()
			})
			if err != nil {
				return err
			}
			holdings = applyLimit(holdings, holdingsLimit)

			printer := ctx.printer(cmd.OutOrStdout())
			if printer.IsJSON() {
				return printer.JSON(holdings)
			}

			rows := make([][]string, 0, len(holdings))
			for _, holding := range holdings {
				rows = append(rows, []string{
					holding.Tradingsymbol,
					holding.Fund,
					holding.Folio,
					formatFloat(holding.Quantity),
					formatFloat(holding.AveragePrice),
					formatFloat(holding.LastPrice),
					formatFloat(holding.Pnl),
					holding.LastPriceDate,
				})
			}
			if len(rows) == 0 {
				rows = append(rows, []string{"-", "-", "-", "0.00", "0.00", "0.00", "0.00", "-"})
			}
			return printer.Table([]string{"SYMBOL", "FUND", "FOLIO", "QTY", "AVG_PRICE", "LAST_PRICE", "PNL", "LAST_PRICE_DATE"}, rows)
		},
	}
	holdingsCmd.Flags().IntVar(&holdingsLimit, "limit", 0, "Limit number of rows (0 = no limit)")

	var holdingISIN string
	var holdingShowLimit int
	holdingShowCmd := &cobra.Command{
		Use:   "show --isin <isin>",
		Short: "Show breakdown for one mutual fund holding ISIN",
		RunE: func(cmd *cobra.Command, _ []string) error {
			isin := strings.TrimSpace(holdingISIN)
			if isin == "" {
				return exitcode.New(exitcode.Validation, "--isin is required")
			}
			if err := validateLimit(holdingShowLimit); err != nil {
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

			breakdown, err := callWithAuthRetry(ctx, profileName, profile, func(client *kiteconnect.Client) (kiteconnect.MFHoldingBreakdown, error) {
				return client.GetMFHoldingInfo(isin)
			})
			if err != nil {
				return err
			}
			breakdown = applyLimit(breakdown, holdingShowLimit)

			printer := ctx.printer(cmd.OutOrStdout())
			if printer.IsJSON() {
				return printer.JSON(breakdown)
			}

			rows := make([][]string, 0, len(breakdown))
			for _, trade := range breakdown {
				rows = append(rows, []string{
					formatModelTime(trade.ExchangeTimestamp.Time),
					trade.Tradingsymbol,
					formatFloat(trade.Quantity),
					formatFloat(trade.Amount),
					formatFloat(trade.AveragePrice),
					trade.Folio,
					trade.Variety,
				})
			}
			if len(rows) == 0 {
				rows = append(rows, []string{"-", "-", "0.00", "0.00", "0.00", "-", "-"})
			}
			return printer.Table([]string{"TIMESTAMP", "SYMBOL", "QTY", "AMOUNT", "AVG_PRICE", "FOLIO", "VARIETY"}, rows)
		},
	}
	holdingShowCmd.Flags().StringVar(&holdingISIN, "isin", "", "ISIN")
	holdingShowCmd.Flags().IntVar(&holdingShowLimit, "limit", 0, "Limit number of rows (0 = no limit)")

	var holdingISINsLimit int
	holdingISINsCmd := &cobra.Command{
		Use:   "isins",
		Short: "List allotted MF ISINs",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := validateLimit(holdingISINsLimit); err != nil {
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

			isins, err := callWithAuthRetry(ctx, profileName, profile, func(client *kiteconnect.Client) (kiteconnect.MFAllottedISINs, error) {
				return client.GetMFAllottedISINs()
			})
			if err != nil {
				return err
			}
			isins = applyLimit(isins, holdingISINsLimit)

			printer := ctx.printer(cmd.OutOrStdout())
			if printer.IsJSON() {
				return printer.JSON(isins)
			}

			rows := make([][]string, 0, len(isins))
			for _, isin := range isins {
				rows = append(rows, []string{isin})
			}
			if len(rows) == 0 {
				rows = append(rows, []string{"-"})
			}
			return printer.Table([]string{"ISIN"}, rows)
		},
	}
	holdingISINsCmd.Flags().IntVar(&holdingISINsLimit, "limit", 0, "Limit number of rows (0 = no limit)")

	holdingsCmd.AddCommand(holdingShowCmd, holdingISINsCmd)

	mfCmd.AddCommand(ordersCmd, sipsCmd, holdingsCmd)
	return mfCmd
}

func mfOrderParamsFromFlags(flags mfOrderFlags) (kiteconnect.MFOrderParams, error) {
	params := kiteconnect.MFOrderParams{
		Tradingsymbol:   strings.TrimSpace(flags.symbol),
		TransactionType: normalizeUpper(flags.txnType),
		Quantity:        flags.quantity,
		Amount:          flags.amount,
		Tag:             strings.TrimSpace(flags.tag),
	}

	if params.Tradingsymbol == "" || params.TransactionType == "" {
		return params, exitcode.New(exitcode.Validation, "--symbol and --txn are required")
	}
	if params.TransactionType != kiteconnect.TransactionTypeBuy && params.TransactionType != kiteconnect.TransactionTypeSell {
		return params, exitcode.New(exitcode.Validation, "invalid --txn; use BUY or SELL")
	}
	if params.Quantity <= 0 && params.Amount <= 0 {
		return params, exitcode.New(exitcode.Validation, "either --qty or --amount must be greater than 0")
	}
	if params.Quantity < 0 || params.Amount < 0 {
		return params, exitcode.New(exitcode.Validation, "--qty and --amount cannot be negative")
	}

	return params, nil
}

func mfSIPPlaceParamsFromFlags(flags mfSIPPlaceFlags) (kiteconnect.MFSIPParams, error) {
	params := kiteconnect.MFSIPParams{
		Tradingsymbol: strings.TrimSpace(flags.symbol),
		Amount:        flags.amount,
		Instalments:   flags.instalments,
		Frequency:     strings.TrimSpace(flags.frequency),
		InstalmentDay: flags.instalmentDay,
		InitialAmount: flags.initialAmount,
		TriggerPrice:  flags.triggerPrice,
		StepUp:        strings.TrimSpace(flags.stepUp),
		SipType:       strings.TrimSpace(flags.sipType),
		Tag:           strings.TrimSpace(flags.tag),
	}

	if params.Tradingsymbol == "" || params.Frequency == "" {
		return params, exitcode.New(exitcode.Validation, "--symbol and --frequency are required")
	}
	if params.Amount <= 0 {
		return params, exitcode.New(exitcode.Validation, "--amount must be greater than 0")
	}
	if params.Instalments <= 0 {
		return params, exitcode.New(exitcode.Validation, "--instalments must be greater than 0")
	}
	if params.InstalmentDay < 0 || params.InstalmentDay > 31 {
		return params, exitcode.New(exitcode.Validation, "--instalment-day must be between 1 and 31 when provided")
	}
	if params.InitialAmount < 0 {
		return params, exitcode.New(exitcode.Validation, "--initial-amount cannot be negative")
	}
	if params.TriggerPrice < 0 {
		return params, exitcode.New(exitcode.Validation, "--trigger-price cannot be negative")
	}

	return params, nil
}

func mfSIPModifyParamsFromFlags(flags mfSIPModifyFlags) (kiteconnect.MFSIPModifyParams, error) {
	params := kiteconnect.MFSIPModifyParams{
		Amount:        flags.amount,
		Frequency:     strings.TrimSpace(flags.frequency),
		InstalmentDay: flags.instalmentDay,
		Instalments:   flags.instalments,
		StepUp:        strings.TrimSpace(flags.stepUp),
		Status:        strings.TrimSpace(flags.status),
	}

	if params.Amount < 0 {
		return params, exitcode.New(exitcode.Validation, "--amount cannot be negative")
	}
	if params.InstalmentDay < 0 || params.InstalmentDay > 31 {
		return params, exitcode.New(exitcode.Validation, "--instalment-day must be between 1 and 31 when provided")
	}
	if params.Instalments < 0 {
		return params, exitcode.New(exitcode.Validation, "--instalments cannot be negative")
	}
	if params.Amount == 0 &&
		params.Frequency == "" &&
		params.InstalmentDay == 0 &&
		params.Instalments == 0 &&
		params.StepUp == "" &&
		params.Status == "" {
		return params, exitcode.New(exitcode.Validation, "at least one field must be provided to modify a SIP")
	}

	return params, nil
}

func parseDateYYYYMMDD(input string, flagName string) (string, error) {
	value := strings.TrimSpace(input)
	if value == "" {
		return "", exitcode.New(exitcode.Validation, flagName+" is required")
	}
	if _, err := time.Parse("2006-01-02", value); err != nil {
		return "", exitcode.New(exitcode.Validation, flagName+" has invalid format; use YYYY-MM-DD")
	}
	return value, nil
}
