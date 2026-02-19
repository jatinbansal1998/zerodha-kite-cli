//go:build !windows

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

func TestRunManualAppliesUpdateOnUnix(t *testing.T) {
	cacheDir := t.TempDir()
	executablePath := filepath.Join(t.TempDir(), "zerodha")
	if err := os.WriteFile(executablePath, []byte("old-binary"), 0o755); err != nil {
		t.Fatalf("write current executable: %v", err)
	}

	assetName := "zerodha_" + runtime.GOOS + "_" + runtime.GOARCH
	assetPath := "/download/" + assetName
	downloadURL := ""

	mux := http.NewServeMux()
	mux.HandleFunc("/repos/o/r/releases/latest", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"tag_name": "v0.0.2",
			"assets": []map[string]any{
				{
					"name":                 assetName,
					"browser_download_url": downloadURL,
				},
			},
		})
	})
	mux.HandleFunc(assetPath, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("new-binary"))
	})
	server := httptest.NewServer(mux)
	defer server.Close()
	downloadURL = server.URL + assetPath

	state, err := RunManual(Options{
		CurrentVersion: "v0.0.1",
		ExecutablePath: executablePath,
		CacheDir:       cacheDir,
		RepoOwner:      "o",
		RepoName:       "r",
		APIBaseURL:     server.URL,
		Cooldown:       time.Hour,
		HTTPClient:     server.Client(),
	})
	if err != nil {
		t.Fatalf("RunManual returned error: %v", err)
	}

	if state.CurrentVersion != "v0.0.2" {
		t.Fatalf("expected current version v0.0.2, got %q", state.CurrentVersion)
	}
	if state.ApplyPending {
		t.Fatalf("expected apply pending to be false")
	}
	if state.DownloadedVersion != "" || state.DownloadedAsset != "" {
		t.Fatalf("expected downloaded state to be cleared after apply")
	}

	updatedBytes, err := os.ReadFile(executablePath)
	if err != nil {
		t.Fatalf("read updated executable: %v", err)
	}
	if string(updatedBytes) != "new-binary" {
		t.Fatalf("expected updated executable content, got %q", string(updatedBytes))
	}
}
