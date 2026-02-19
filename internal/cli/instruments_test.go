package cli

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/jatinbansal1998/zerodha-kite-cli/internal/config"
	kiteconnect "github.com/zerodha/gokiteconnect/v4"
)

func TestSummarizeInstrumentsAggregatesCounts(t *testing.T) {
	instruments := kiteconnect.Instruments{
		{Exchange: "NSE", InstrumentType: "EQ"},
		{Exchange: "BSE", InstrumentType: "EQ"},
		{Exchange: "NSE", InstrumentType: "FUT"},
		{Exchange: "nse", InstrumentType: "eq"},
		{Exchange: "", InstrumentType: ""},
	}

	summary := summarizeInstruments(instruments)
	if summary.Total != 5 {
		t.Fatalf("expected total %d, got %d", 5, summary.Total)
	}

	expectedExchanges := []instrumentExchangeRow{
		{Exchange: "-", Count: 1},
		{Exchange: "BSE", Count: 1},
		{Exchange: "NSE", Count: 3},
	}
	if len(summary.Exchanges) != len(expectedExchanges) {
		t.Fatalf("expected %d exchange groups, got %d", len(expectedExchanges), len(summary.Exchanges))
	}
	for i, expected := range expectedExchanges {
		if summary.Exchanges[i] != expected {
			t.Fatalf("exchange row %d: expected %+v, got %+v", i, expected, summary.Exchanges[i])
		}
	}

	expectedTypes := []instrumentTypeRow{
		{Type: "-", Count: 1},
		{Type: "EQ", Count: 3},
		{Type: "FUT", Count: 1},
	}
	if len(summary.Types) != len(expectedTypes) {
		t.Fatalf("expected %d type groups, got %d", len(expectedTypes), len(summary.Types))
	}
	for i, expected := range expectedTypes {
		if summary.Types[i] != expected {
			t.Fatalf("type row %d: expected %+v, got %+v", i, expected, summary.Types[i])
		}
	}
}

func TestSummaryRowsIncludesTotalAndGroupRows(t *testing.T) {
	summary := instrumentsSummary{
		Total: 3,
		Exchanges: []instrumentExchangeRow{
			{Exchange: "BSE", Count: 1},
			{Exchange: "NSE", Count: 2},
		},
		Types: []instrumentTypeRow{
			{Type: "EQ", Count: 3},
		},
	}

	rows := summaryRows(summary)
	expectedRows := [][]string{
		{"TOTAL", "ALL", "3"},
		{"EXCHANGE", "BSE", "1"},
		{"EXCHANGE", "NSE", "2"},
		{"TYPE", "EQ", "3"},
	}
	if len(rows) != len(expectedRows) {
		t.Fatalf("expected %d rows, got %d", len(expectedRows), len(rows))
	}
	for i := range expectedRows {
		if strings.Join(rows[i], "|") != strings.Join(expectedRows[i], "|") {
			t.Fatalf("row %d: expected %v, got %v", i, expectedRows[i], rows[i])
		}
	}
}

func TestInstrumentsListRejectsExchangeAndAllTogether(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")
	cfg := config.Default()
	cfg.ActiveProfile = "default"
	cfg.Profiles["default"] = config.Profile{
		APIKey:      "test_key",
		APISecret:   "test_secret",
		AccessToken: "test_access_token",
	}
	saveTestConfig(t, configPath, cfg)

	_, _, err := executeCLICommand(
		t,
		configPath,
		"instruments",
		"list",
		"--exchange",
		"NSE",
		"--all",
	)
	if err == nil {
		t.Fatalf("expected validation error, got nil")
	}
	if !strings.Contains(err.Error(), "--exchange and --all cannot be used together") {
		t.Fatalf("expected validation error, got %q", err.Error())
	}
}
