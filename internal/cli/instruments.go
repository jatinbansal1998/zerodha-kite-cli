package cli

import (
	"sort"
	"strings"

	"github.com/jatinbansal1998/zerodha-kite-cli/internal/exitcode"
	"github.com/spf13/cobra"
	kiteconnect "github.com/zerodha/gokiteconnect/v4"
)

func newInstrumentsCmd(opts *rootOptions) *cobra.Command {
	instrumentsCmd := &cobra.Command{
		Use:   "instruments",
		Short: "Instrument master data",
	}

	var exchange string
	var listAll bool
	var listLimit int
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "Summarize instruments by exchange/type (default)",
		Long: strings.Join([]string{
			"By default, this command prints instrument counts grouped by exchange and type.",
			"Use --exchange to print row-level instruments for a single exchange.",
			"Use --all to print row-level instruments across all exchanges.",
		}, " "),
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

			exchangeValue := normalizeUpper(exchange)
			if exchangeValue != "" && listAll {
				return exitcode.New(exitcode.Validation, "--exchange and --all cannot be used together")
			}
			outputRows := exchangeValue != "" || listAll
			if listLimit > 0 && !outputRows {
				return exitcode.New(exitcode.Validation, "--limit can only be used with --exchange or --all")
			}

			var instruments kiteconnect.Instruments
			if exchangeValue == "" {
				instruments, err = callWithAuthRetry(ctx, profileName, profile, func(client *kiteconnect.Client) (kiteconnect.Instruments, error) {
					return client.GetInstruments()
				})
			} else {
				instruments, err = callWithAuthRetry(ctx, profileName, profile, func(client *kiteconnect.Client) (kiteconnect.Instruments, error) {
					return client.GetInstrumentsByExchange(exchangeValue)
				})
			}
			if err != nil {
				return err
			}
			if outputRows {
				instruments = applyLimit(instruments, listLimit)
			}

			printer := ctx.printer(cmd.OutOrStdout())
			if printer.IsJSON() {
				if !outputRows {
					return printer.JSON(summarizeInstruments(instruments))
				}
				return printer.JSON(instruments)
			}
			if !outputRows {
				summary := summarizeInstruments(instruments)
				return printer.Table(
					[]string{"GROUP", "KEY", "COUNT"},
					summaryRows(summary),
				)
			}

			rows := make([][]string, 0, len(instruments))
			for _, instrument := range instruments {
				expiry := "-"
				if !instrument.Expiry.Time.IsZero() {
					expiry = instrument.Expiry.Time.Format("2006-01-02")
				}

				rows = append(rows, []string{
					intToString(instrument.InstrumentToken),
					instrument.Tradingsymbol,
					instrument.Name,
					instrument.Exchange,
					instrument.InstrumentType,
					expiry,
					formatFloat(instrument.StrikePrice),
					formatFloat(instrument.LastPrice),
					formatFloat(instrument.TickSize),
					formatFloat(instrument.LotSize),
				})
			}
			if len(rows) == 0 {
				rows = append(rows, []string{"0", "-", "-", "-", "-", "-", "0.00", "0.00", "0.00", "0.00"})
			}

			return printer.Table(
				[]string{"TOKEN", "SYMBOL", "NAME", "EXCHANGE", "TYPE", "EXPIRY", "STRIKE", "LAST_PRICE", "TICK_SIZE", "LOT_SIZE"},
				rows,
			)
		},
	}
	listCmd.Flags().StringVar(&exchange, "exchange", "", "Print row-level instruments for one exchange (NSE/BSE/NFO/MCX/...)")
	listCmd.Flags().BoolVar(&listAll, "all", false, "Print row-level instruments across all exchanges (large output)")
	listCmd.Flags().IntVar(&listLimit, "limit", 0, "Limit number of rows (0 = no limit, only with --exchange or --all)")

	var mfLimit int
	mfCmd := &cobra.Command{
		Use:   "mf",
		Short: "List mutual fund instruments",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := validateLimit(mfLimit); err != nil {
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

			instruments, err := callWithAuthRetry(ctx, profileName, profile, func(client *kiteconnect.Client) (kiteconnect.MFInstruments, error) {
				return client.GetMFInstruments()
			})
			if err != nil {
				return err
			}
			instruments = applyLimit(instruments, mfLimit)

			printer := ctx.printer(cmd.OutOrStdout())
			if printer.IsJSON() {
				return printer.JSON(instruments)
			}

			rows := make([][]string, 0, len(instruments))
			for _, instrument := range instruments {
				lastPriceDate := "-"
				if !instrument.LastPriceDate.Time.IsZero() {
					lastPriceDate = instrument.LastPriceDate.Time.Format("2006-01-02")
				}
				rows = append(rows, []string{
					instrument.Tradingsymbol,
					instrument.Name,
					instrument.AMC,
					formatFloat(instrument.LastPrice),
					boolToYesNo(instrument.PurchaseAllowed),
					boolToYesNo(instrument.RedemtpionAllowed),
					instrument.SchemeType,
					instrument.Plan,
					lastPriceDate,
				})
			}
			if len(rows) == 0 {
				rows = append(rows, []string{"-", "-", "-", "0.00", "no", "no", "-", "-", "-"})
			}

			return printer.Table(
				[]string{"SYMBOL", "NAME", "AMC", "LAST_PRICE", "PURCHASE_ALLOWED", "REDEMPTION_ALLOWED", "SCHEME_TYPE", "PLAN", "LAST_PRICE_DATE"},
				rows,
			)
		},
	}
	mfCmd.Flags().IntVar(&mfLimit, "limit", 0, "Limit number of rows (0 = no limit)")

	instrumentsCmd.AddCommand(listCmd, mfCmd)
	return instrumentsCmd
}

func boolToYesNo(v bool) string {
	if v {
		return "yes"
	}
	return "no"
}

type instrumentsSummary struct {
	Total     int                     `json:"total"`
	Exchanges []instrumentExchangeRow `json:"exchanges"`
	Types     []instrumentTypeRow     `json:"types"`
}

type instrumentExchangeRow struct {
	Exchange string `json:"exchange"`
	Count    int    `json:"count"`
}

type instrumentTypeRow struct {
	Type  string `json:"type"`
	Count int    `json:"count"`
}

func summarizeInstruments(instruments kiteconnect.Instruments) instrumentsSummary {
	exchangeCounts := make(map[string]int)
	typeCounts := make(map[string]int)

	for _, instrument := range instruments {
		exchange := normalizeUpper(instrument.Exchange)
		if exchange == "" {
			exchange = "-"
		}
		exchangeCounts[exchange]++

		instrumentType := normalizeUpper(instrument.InstrumentType)
		if instrumentType == "" {
			instrumentType = "-"
		}
		typeCounts[instrumentType]++
	}

	exchangeKeys := make([]string, 0, len(exchangeCounts))
	for key := range exchangeCounts {
		exchangeKeys = append(exchangeKeys, key)
	}
	sort.Strings(exchangeKeys)

	typeKeys := make([]string, 0, len(typeCounts))
	for key := range typeCounts {
		typeKeys = append(typeKeys, key)
	}
	sort.Strings(typeKeys)

	summary := instrumentsSummary{
		Total:     len(instruments),
		Exchanges: make([]instrumentExchangeRow, 0, len(exchangeKeys)),
		Types:     make([]instrumentTypeRow, 0, len(typeKeys)),
	}

	for _, key := range exchangeKeys {
		summary.Exchanges = append(summary.Exchanges, instrumentExchangeRow{
			Exchange: key,
			Count:    exchangeCounts[key],
		})
	}
	for _, key := range typeKeys {
		summary.Types = append(summary.Types, instrumentTypeRow{
			Type:  key,
			Count: typeCounts[key],
		})
	}

	return summary
}

func summaryRows(summary instrumentsSummary) [][]string {
	rows := make([][]string, 0, 1+len(summary.Exchanges)+len(summary.Types))
	rows = append(rows, []string{"TOTAL", "ALL", intToString(summary.Total)})

	for _, row := range summary.Exchanges {
		rows = append(rows, []string{"EXCHANGE", row.Exchange, intToString(row.Count)})
	}
	for _, row := range summary.Types {
		rows = append(rows, []string{"TYPE", row.Type, intToString(row.Count)})
	}

	return rows
}
