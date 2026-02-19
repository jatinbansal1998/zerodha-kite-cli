package cli

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/jatinbansal1998/zerodha-kite-cli/internal/exitcode"
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

	ltpCmd := &cobra.Command{
		Use:   "ltp <EXCHANGE:SYMBOL> [EXCHANGE:SYMBOL...]",
		Short: "Fetch last traded prices",
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

			quotes, err := callWithAuthRetry(ctx, profileName, profile, func(client *kiteconnect.Client) (kiteconnect.QuoteLTP, error) {
				return client.GetLTP(args...)
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
				})
			}
			return printer.Table([]string{"INSTRUMENT", "LTP"}, rows)
		},
	}

	ohlcCmd := &cobra.Command{
		Use:   "ohlc <EXCHANGE:SYMBOL> [EXCHANGE:SYMBOL...]",
		Short: "Fetch OHLC snapshots",
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

			quotes, err := callWithAuthRetry(ctx, profileName, profile, func(client *kiteconnect.Client) (kiteconnect.QuoteOHLC, error) {
				return client.GetOHLC(args...)
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
					fmt.Sprintf("%.2f", q.OHLC.Open),
					fmt.Sprintf("%.2f", q.OHLC.High),
					fmt.Sprintf("%.2f", q.OHLC.Low),
					fmt.Sprintf("%.2f", q.OHLC.Close),
					fmt.Sprintf("%.2f", q.LastPrice),
				})
			}
			return printer.Table([]string{"INSTRUMENT", "OPEN", "HIGH", "LOW", "CLOSE", "LTP"}, rows)
		},
	}

	var (
		hInstrumentToken int
		hInterval        string
		hFrom            string
		hTo              string
		hContinuous      bool
		hOI              bool
	)
	historicalCmd := &cobra.Command{
		Use:   "historical",
		Short: "Fetch historical candles by instrument token",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if hInstrumentToken <= 0 {
				return exitcode.New(exitcode.Validation, "--instrument-token must be greater than 0")
			}
			interval := strings.TrimSpace(hInterval)
			if interval == "" {
				return exitcode.New(exitcode.Validation, "--interval is required")
			}

			from, err := parseHistoricalTime(hFrom, "--from")
			if err != nil {
				return err
			}
			to, err := parseHistoricalTime(hTo, "--to")
			if err != nil {
				return err
			}
			if from.After(to) {
				return exitcode.New(exitcode.Validation, "--from must be before or equal to --to")
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

			candles, err := callWithAuthRetry(ctx, profileName, profile, func(client *kiteconnect.Client) ([]kiteconnect.HistoricalData, error) {
				return client.GetHistoricalData(hInstrumentToken, interval, from, to, hContinuous, hOI)
			})
			if err != nil {
				return err
			}

			printer := ctx.printer(cmd.OutOrStdout())
			if printer.IsJSON() {
				return printer.JSON(candles)
			}

			rows := make([][]string, 0, len(candles))
			for _, candle := range candles {
				rows = append(rows, []string{
					candle.Date.Time.Format("2006-01-02 15:04:05"),
					formatFloat(candle.Open),
					formatFloat(candle.High),
					formatFloat(candle.Low),
					formatFloat(candle.Close),
					intToString(candle.Volume),
					intToString(candle.OI),
				})
			}
			if len(rows) == 0 {
				rows = append(rows, []string{"-", "0.00", "0.00", "0.00", "0.00", "0", "0"})
			}
			return printer.Table([]string{"DATE", "OPEN", "HIGH", "LOW", "CLOSE", "VOLUME", "OI"}, rows)
		},
	}
	historicalCmd.Flags().IntVar(&hInstrumentToken, "instrument-token", 0, "Instrument token")
	historicalCmd.Flags().StringVar(&hInterval, "interval", "", "Candle interval (minute, 3minute, 5minute, 10minute, 15minute, 30minute, 60minute, day)")
	historicalCmd.Flags().StringVar(&hFrom, "from", "", "Start timestamp (YYYY-MM-DD, YYYY-MM-DD HH:MM:SS, or RFC3339)")
	historicalCmd.Flags().StringVar(&hTo, "to", "", "End timestamp (YYYY-MM-DD, YYYY-MM-DD HH:MM:SS, or RFC3339)")
	historicalCmd.Flags().BoolVar(&hContinuous, "continuous", false, "Set continuous=true for continuous futures data")
	historicalCmd.Flags().BoolVar(&hOI, "oi", false, "Include open interest")

	quoteCmd.AddCommand(getCmd, ltpCmd, ohlcCmd, historicalCmd)
	return quoteCmd
}

func parseHistoricalTime(raw string, flag string) (time.Time, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return time.Time{}, exitcode.New(exitcode.Validation, flag+" is required")
	}

	layouts := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, value); err == nil {
			return t, nil
		}
	}

	return time.Time{}, exitcode.New(exitcode.Validation, flag+" has invalid format; use YYYY-MM-DD, YYYY-MM-DD HH:MM:SS, or RFC3339")
}
