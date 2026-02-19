package cli

import (
	"strings"

	"github.com/jatinbansal1998/zerodha-kite-cli/internal/exitcode"
	"github.com/spf13/cobra"
	kiteconnect "github.com/zerodha/gokiteconnect/v4"
)

func newOrdersCmd(opts *rootOptions) *cobra.Command {
	ordersCmd := &cobra.Command{
		Use:   "orders",
		Short: "Orderbook operations",
	}

	var listLimit int
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List orders",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := validateLimit(listLimit); err != nil {
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

			orders, err := callWithAuthRetry(ctx, profileName, profile, func(client *kiteconnect.Client) (kiteconnect.Orders, error) {
				return client.GetOrders()
			})
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
					order.TradingSymbol,
					order.Exchange,
					order.TransactionType,
					formatFloat(order.Quantity),
					formatFloat(order.Price),
					order.Status,
					order.OrderType,
					order.Product,
					order.OrderTimestamp.Time.Format("2006-01-02 15:04:05"),
				})
			}

			return printer.Table([]string{
				"ORDER_ID",
				"SYMBOL",
				"EXCHANGE",
				"TXN",
				"QTY",
				"PRICE",
				"STATUS",
				"TYPE",
				"PRODUCT",
				"TIMESTAMP",
			}, rows)
		},
	}
	listCmd.Flags().IntVar(&listLimit, "limit", 0, "Limit number of rows (0 = no limit)")

	var showOrderID string
	var showLimit int
	showCmd := &cobra.Command{
		Use:   "show --order-id <id>",
		Short: "Show order history for an order ID",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if showOrderID == "" {
				return exitcode.New(exitcode.Validation, "--order-id is required")
			}
			if err := validateLimit(showLimit); err != nil {
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

			history, err := callWithAuthRetry(ctx, profileName, profile, func(client *kiteconnect.Client) ([]kiteconnect.Order, error) {
				return client.GetOrderHistory(showOrderID)
			})
			if err != nil {
				return err
			}
			history = applyLimit(history, showLimit)

			printer := ctx.printer(cmd.OutOrStdout())
			if printer.IsJSON() {
				return printer.JSON(history)
			}

			rows := make([][]string, 0, len(history))
			for _, event := range history {
				rows = append(rows, []string{
					event.Status,
					formatFloat(event.FilledQuantity),
					formatFloat(event.PendingQuantity),
					formatFloat(event.AveragePrice),
					event.StatusMessage,
					event.OrderTimestamp.Time.Format("2006-01-02 15:04:05"),
				})
			}
			return printer.Table([]string{"STATUS", "FILLED_QTY", "PENDING_QTY", "AVG_PRICE", "MESSAGE", "TIMESTAMP"}, rows)
		},
	}
	showCmd.Flags().StringVar(&showOrderID, "order-id", "", "Order ID")
	showCmd.Flags().IntVar(&showLimit, "limit", 0, "Limit number of rows (0 = no limit)")

	var tradesOrderID string
	var tradesLimit int
	tradesCmd := &cobra.Command{
		Use:   "trades",
		Short: "List trades (optionally filtered by order ID)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := validateLimit(tradesLimit); err != nil {
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

			orderID := strings.TrimSpace(tradesOrderID)
			var trades []kiteconnect.Trade
			if orderID == "" {
				result, err := callWithAuthRetry(ctx, profileName, profile, func(client *kiteconnect.Client) (kiteconnect.Trades, error) {
					return client.GetTrades()
				})
				if err != nil {
					return err
				}
				trades = []kiteconnect.Trade(result)
			} else {
				result, err := callWithAuthRetry(ctx, profileName, profile, func(client *kiteconnect.Client) ([]kiteconnect.Trade, error) {
					return client.GetOrderTrades(orderID)
				})
				if err != nil {
					return err
				}
				trades = result
			}
			trades = applyLimit(trades, tradesLimit)

			printer := ctx.printer(cmd.OutOrStdout())
			if printer.IsJSON() {
				return printer.JSON(trades)
			}

			rows := make([][]string, 0, len(trades))
			for _, trade := range trades {
				rows = append(rows, []string{
					trade.TradeID,
					trade.OrderID,
					trade.TradingSymbol,
					trade.Exchange,
					trade.TransactionType,
					formatFloat(trade.Quantity),
					formatFloat(trade.AveragePrice),
					trade.FillTimestamp.Time.Format("2006-01-02 15:04:05"),
				})
			}
			if len(rows) == 0 {
				rows = append(rows, []string{"-", "-", "-", "-", "-", "0.00", "0.00", "-"})
			}
			return printer.Table([]string{"TRADE_ID", "ORDER_ID", "SYMBOL", "EXCHANGE", "TXN", "QTY", "AVG_PRICE", "FILL_TIMESTAMP"}, rows)
		},
	}
	tradesCmd.Flags().StringVar(&tradesOrderID, "order-id", "", "Filter trades for a specific order ID")
	tradesCmd.Flags().IntVar(&tradesLimit, "limit", 0, "Limit number of rows (0 = no limit)")

	ordersCmd.AddCommand(listCmd, showCmd, tradesCmd)
	return ordersCmd
}
