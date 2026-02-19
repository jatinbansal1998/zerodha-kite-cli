package updater

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	DefaultRepoOwner = "jatinbansal1998"
	DefaultRepoName  = "zerodha-kite-cli"
	DefaultCooldown  = 12 * time.Hour

	DisableEnvVar = "ZERODHA_AUTO_UPDATE_DISABLED"
	HelperEnvVar  = "ZERODHA_AUTO_UPDATE_HELPER"

	helperCommandName = "__self-update-apply"
)

type Options struct {
	CurrentVersion string
	ExecutablePath string
	CacheDir       string
	RepoOwner      string
	RepoName       string
	APIBaseURL     string
	Cooldown       time.Duration
	HTTPClient     *http.Client
}

type normalizedOptions struct {
	CurrentVersion string
	ExecutablePath string
	CacheDir       string
	RepoOwner      string
	RepoName       string
	APIBaseURL     string
	Cooldown       time.Duration
	HTTPClient     *http.Client
}

func StartBackground(opts Options) {
	if updateDisabled() {
		return
	}

	normalized, ok := normalizeOptions(opts)
	if !ok {
		return
	}

	go func() {
		_, _ = run(normalized, false)
	}()
}

func LoadState(cacheDir string) (State, error) {
	return NewStateStore(cacheDir).Load()
}

func RunManual(opts Options) (State, error) {
	normalized, ok := normalizeOptions(opts)
	if !ok {
		return State{}, errors.New("manual update options are incomplete")
	}
	return run(normalized, true)
}

func normalizeOptions(opts Options) (normalizedOptions, bool) {
	version := strings.TrimSpace(opts.CurrentVersion)
	if version == "" {
		return normalizedOptions{}, false
	}

	executable := strings.TrimSpace(opts.ExecutablePath)
	if executable == "" {
		return normalizedOptions{}, false
	}

	cacheDir := strings.TrimSpace(opts.CacheDir)
	if cacheDir == "" {
		return normalizedOptions{}, false
	}

	repoOwner := strings.TrimSpace(opts.RepoOwner)
	if repoOwner == "" {
		repoOwner = DefaultRepoOwner
	}
	repoName := strings.TrimSpace(opts.RepoName)
	if repoName == "" {
		repoName = DefaultRepoName
	}

	cooldown := opts.Cooldown
	if cooldown <= 0 {
		cooldown = DefaultCooldown
	}

	client := opts.HTTPClient
	if client == nil {
		client = &http.Client{
			Timeout: 15 * time.Second,
		}
	}

	return normalizedOptions{
		CurrentVersion: version,
		ExecutablePath: executable,
		CacheDir:       cacheDir,
		RepoOwner:      repoOwner,
		RepoName:       repoName,
		APIBaseURL:     strings.TrimSpace(opts.APIBaseURL),
		Cooldown:       cooldown,
		HTTPClient:     client,
	}, true
}

func run(opts normalizedOptions, forceCheck bool) (State, error) {
	store := NewStateStore(opts.CacheDir)
	state, err := store.Load()
	if err != nil {
		state = State{}
		setStateError(&state, err)
		_ = store.Save(state)
		return state, err
	}
	state.CurrentVersion = opts.CurrentVersion

	if err := tryApplyStagedUpdate(&state, store, opts); err != nil {
		setStateError(&state, err)
		_ = store.Save(state)
		return state, err
	}

	now := time.Now().UTC()
	if !forceCheck && !state.LastCheckedAt.IsZero() && now.Sub(state.LastCheckedAt) < opts.Cooldown {
		_ = store.Save(state)
		return state, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()

	release, err := FetchLatestRelease(ctx, opts.HTTPClient, opts.APIBaseURL, opts.RepoOwner, opts.RepoName)
	state.LastCheckedAt = now
	if err != nil {
		setStateError(&state, err)
		_ = store.Save(state)
		return state, err
	}
	state.LatestVersionSeen = release.TagName

	if !IsNewerVersion(release.TagName, opts.CurrentVersion) {
		clearStateError(&state)
		_ = store.Save(state)
		return state, nil
	}

	if state.ApplyPending &&
		state.DownloadedVersion == release.TagName &&
		strings.TrimSpace(state.DownloadedAsset) != "" {
		clearStateError(&state)
		_ = store.Save(state)
		return state, nil
	}

	asset, err := SelectAssetForPlatform(release.Assets, runtime.GOOS, runtime.GOARCH)
	if err != nil {
		setStateError(&state, err)
		_ = store.Save(state)
		return state, err
	}

	stagePath, err := store.StagingPath(release.TagName, asset.Name)
	if err != nil {
		setStateError(&state, err)
		_ = store.Save(state)
		return state, err
	}

	if err := downloadAssetToPath(ctx, opts.HTTPClient, asset.DownloadURL, stagePath); err != nil {
		setStateError(&state, err)
		_ = store.Save(state)
		return state, err
	}
	if err := os.Chmod(stagePath, 0o755); err != nil {
		cause := fmt.Errorf("mark downloaded binary executable: %w", err)
		setStateError(&state, cause)
		_ = store.Save(state)
		return state, cause
	}

	state.DownloadedVersion = release.TagName
	state.DownloadedAsset = stagePath
	state.ApplyPending = true
	clearStateError(&state)
	_ = store.Save(state)

	if err := tryApplyStagedUpdate(&state, store, opts); err != nil {
		setStateError(&state, err)
		_ = store.Save(state)
		return state, err
	}
	clearStateError(&state)
	_ = store.Save(state)
	return state, nil
}

func tryApplyStagedUpdate(state *State, store *StateStore, opts normalizedOptions) error {
	if !state.ApplyPending || strings.TrimSpace(state.DownloadedAsset) == "" {
		return nil
	}

	if runtime.GOOS == "windows" {
		return SpawnApplyHelper(opts.ExecutablePath, ApplyHelperRequest{
			TargetPath: opts.ExecutablePath,
			SourcePath: state.DownloadedAsset,
			CacheDir:   opts.CacheDir,
			Version:    state.DownloadedVersion,
		})
	}

	if err := applyDownloadedBinary(opts.ExecutablePath, state.DownloadedAsset); err != nil {
		return err
	}

	state.CurrentVersion = state.DownloadedVersion
	state.DownloadedVersion = ""
	state.DownloadedAsset = ""
	state.ApplyPending = false
	_ = store.Save(*state)
	return nil
}

func downloadAssetToPath(ctx context.Context, client *http.Client, downloadURL, destinationPath string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, nil)
	if err != nil {
		return fmt.Errorf("build download request: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("download release asset: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download release asset returned status %d", resp.StatusCode)
	}

	if err := os.MkdirAll(filepath.Dir(destinationPath), 0o700); err != nil {
		return fmt.Errorf("create download directory: %w", err)
	}

	tmpPath := destinationPath + ".tmp"
	tmpFile, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o700)
	if err != nil {
		return fmt.Errorf("create temporary download file: %w", err)
	}

	cleanup := func() {
		_ = tmpFile.Close()
		_ = os.Remove(tmpPath)
	}

	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		cleanup()
		return fmt.Errorf("write downloaded asset: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		cleanup()
		return fmt.Errorf("close downloaded asset: %w", err)
	}
	if err := os.Rename(tmpPath, destinationPath); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("finalize downloaded asset: %w", err)
	}
	return nil
}

func updateDisabled() bool {
	value := strings.ToLower(strings.TrimSpace(os.Getenv(DisableEnvVar)))
	return value == "1" || value == "true" || value == "yes" || value == "on"
}

func setStateError(state *State, err error) {
	if err == nil {
		return
	}
	state.LastError = err.Error()
	state.LastErrorAt = time.Now().UTC()
}

func clearStateError(state *State) {
	state.LastError = ""
	state.LastErrorAt = time.Time{}
}

func HelperCommandName() string {
	return helperCommandName
}
