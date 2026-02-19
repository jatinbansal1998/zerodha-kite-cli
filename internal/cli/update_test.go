package cli

import (
	"testing"

	"github.com/jatinbansal1998/zerodha-kite-cli/internal/updater"
)

func TestSummarizeManualUpdateUpdated(t *testing.T) {
	result := summarizeManualUpdate("v0.0.1", updater.State{
		CurrentVersion:    "v0.0.2",
		LatestVersionSeen: "v0.0.2",
	})

	if result.Status != "updated" {
		t.Fatalf("expected status updated, got %q", result.Status)
	}
	if result.UpdatedVersion != "v0.0.2" {
		t.Fatalf("expected updated version v0.0.2, got %q", result.UpdatedVersion)
	}
}

func TestSummarizeManualUpdateApplyPending(t *testing.T) {
	result := summarizeManualUpdate("v0.0.1", updater.State{
		ApplyPending:      true,
		DownloadedVersion: "v0.0.2",
		LatestVersionSeen: "v0.0.2",
	})

	if result.Status != "apply_pending" {
		t.Fatalf("expected status apply_pending, got %q", result.Status)
	}
	if result.PendingVersion != "v0.0.2" {
		t.Fatalf("expected pending version v0.0.2, got %q", result.PendingVersion)
	}
}

func TestSummarizeManualUpdateUpToDate(t *testing.T) {
	result := summarizeManualUpdate("v0.0.2", updater.State{
		CurrentVersion:    "v0.0.2",
		LatestVersionSeen: "v0.0.2",
	})

	if result.Status != "up_to_date" {
		t.Fatalf("expected status up_to_date, got %q", result.Status)
	}
}
