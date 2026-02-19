package updater

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestRunManualBypassesCooldown(t *testing.T) {
	cacheDir := t.TempDir()
	store := NewStateStore(cacheDir)

	initialChecked := time.Now().UTC().Add(-5 * time.Minute)
	if err := store.Save(State{
		CurrentVersion: "v0.0.1",
		LastCheckedAt:  initialChecked,
	}); err != nil {
		t.Fatalf("seed updater state: %v", err)
	}

	calls := 0
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/o/r/releases/latest", func(w http.ResponseWriter, _ *http.Request) {
		calls++
		_ = json.NewEncoder(w).Encode(map[string]any{
			"tag_name": "v0.0.1",
			"assets":   []map[string]any{},
		})
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	executablePath := filepath.Join(t.TempDir(), "zerodha")
	if runtime.GOOS == "windows" {
		executablePath += ".exe"
	}
	if err := os.WriteFile(executablePath, []byte("current"), 0o755); err != nil {
		t.Fatalf("write fake executable: %v", err)
	}

	state, err := RunManual(Options{
		CurrentVersion: "v0.0.1",
		ExecutablePath: executablePath,
		CacheDir:       cacheDir,
		RepoOwner:      "o",
		RepoName:       "r",
		APIBaseURL:     server.URL,
		Cooldown:       24 * time.Hour,
		HTTPClient:     server.Client(),
	})
	if err != nil {
		t.Fatalf("RunManual returned error: %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected exactly one release check call, got %d", calls)
	}
	if !state.LastCheckedAt.After(initialChecked) {
		t.Fatalf("expected last_checked_at to be refreshed")
	}
}
