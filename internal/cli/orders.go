package cli

import (
	"github.com/jatinbansal1998/zerodha-kite-cli/internal/exitcode"
	"github.com/spf13/cobra"
	kiteconnect "github.com/zerodha/gokiteconnect/v4"
)

func newOrdersCmd(opts *rootOptions) *cobra.Command {
	ordersCmd := &cobra.Command{
		Use:   "orders",
		Short: "Orderbook operations",
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List orders",
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

			orders, err := callWithAuthRetry(ctx, profileName, profile, func(client *kiteconnect.Client) (kiteconnect.Orders, error) {
				return client.GetOrders()
			})
			if err != nil {
				return err
			}

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

	var showOrderID string
	showCmd := &cobra.Command{
		Use:   "show --order-id <id>",
		Short: "Show order history for an order ID",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if showOrderID == "" {
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

			history, err := callWithAuthRetry(ctx, profileName, profile, func(client *kiteconnect.Client) ([]kiteconnect.Order, error) {
				return client.GetOrderHistory(showOrderID)
			})
			if err != nil {
				return err
			}

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

	ordersCmd.AddCommand(listCmd, showCmd)
	return ordersCmd
}
