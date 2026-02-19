package updater

import (
	"testing"
	"time"
)

func TestStateStoreSaveLoad(t *testing.T) {
	store := NewStateStore(t.TempDir())

	want := State{
		CurrentVersion:    "v0.0.1",
		LastCheckedAt:     time.Date(2026, 2, 19, 10, 0, 0, 0, time.UTC),
		LatestVersionSeen: "v0.0.2",
		DownloadedVersion: "v0.0.2",
		DownloadedAsset:   "/tmp/zerodha-v0.0.2",
		ApplyPending:      true,
	}
	if err := store.Save(want); err != nil {
		t.Fatalf("save state: %v", err)
	}

	got, err := store.Load()
	if err != nil {
		t.Fatalf("load state: %v", err)
	}

	if got.CurrentVersion != want.CurrentVersion {
		t.Fatalf("expected current version %q, got %q", want.CurrentVersion, got.CurrentVersion)
	}
	if !got.LastCheckedAt.Equal(want.LastCheckedAt) {
		t.Fatalf("expected last checked at %v, got %v", want.LastCheckedAt, got.LastCheckedAt)
	}
	if got.DownloadedAsset != want.DownloadedAsset {
		t.Fatalf("expected downloaded asset %q, got %q", want.DownloadedAsset, got.DownloadedAsset)
	}
	if !got.ApplyPending {
		t.Fatalf("expected apply pending true")
	}
}

func TestStateStoreStagingPath(t *testing.T) {
	store := NewStateStore(t.TempDir())
	path, err := store.StagingPath("v1.2.3", "zerodha_darwin_arm64")
	if err != nil {
		t.Fatalf("staging path: %v", err)
	}
	if path == "" {
		t.Fatalf("expected non-empty staging path")
	}
}
