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

func TestStartBackgroundIsNonBlocking(t *testing.T) {
	t.Setenv(DisableEnvVar, "0")

	done := make(chan struct{}, 1)
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/o/r/releases/latest", func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(250 * time.Millisecond)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"tag_name": "v0.0.1",
			"assets":   []map[string]any{},
		})
		done <- struct{}{}
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	cacheDir := t.TempDir()
	exePath := filepath.Join(t.TempDir(), "zerodha")
	if runtime.GOOS == "windows" {
		exePath += ".exe"
	}
	if err := os.WriteFile(exePath, []byte("current-binary"), 0o755); err != nil {
		t.Fatalf("write fake executable: %v", err)
	}

	start := time.Now()
	StartBackground(Options{
		CurrentVersion: "v0.0.1",
		ExecutablePath: exePath,
		CacheDir:       cacheDir,
		RepoOwner:      "o",
		RepoName:       "r",
		APIBaseURL:     server.URL,
		Cooldown:       time.Minute,
		HTTPClient:     server.Client(),
	})
	if time.Since(start) > 60*time.Millisecond {
		t.Fatalf("expected StartBackground to return quickly")
	}

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatalf("timed out waiting for background updater request")
	}
	time.Sleep(100 * time.Millisecond)
}
