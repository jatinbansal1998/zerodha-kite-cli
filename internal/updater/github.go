package updater

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
)

const (
	defaultAPIBaseURL = "https://api.github.com"
	userAgent         = "zerodha-kite-cli-updater"
)

type Release struct {
	TagName string
	Assets  []ReleaseAsset
}

type ReleaseAsset struct {
	Name         string
	DownloadURL  string
	DownloadSize int64
}

func FetchLatestRelease(ctx context.Context, client *http.Client, apiBaseURL, owner, repo string) (Release, error) {
	baseURL := strings.TrimSpace(apiBaseURL)
	if baseURL == "" {
		baseURL = defaultAPIBaseURL
	}

	latestURL, err := url.JoinPath(baseURL, "repos", owner, repo, "releases", "latest")
	if err != nil {
		return Release{}, fmt.Errorf("build latest release url: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, latestURL, nil)
	if err != nil {
		return Release{}, fmt.Errorf("build latest release request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", userAgent)

	resp, err := client.Do(req)
	if err != nil {
		return Release{}, fmt.Errorf("request latest release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Release{}, fmt.Errorf("latest release request returned status %d", resp.StatusCode)
	}

	var body struct {
		TagName string `json:"tag_name"`
		Assets  []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
			Size               int64  `json:"size"`
		} `json:"assets"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return Release{}, fmt.Errorf("decode latest release: %w", err)
	}

	release := Release{
		TagName: strings.TrimSpace(body.TagName),
		Assets:  make([]ReleaseAsset, 0, len(body.Assets)),
	}
	for _, asset := range body.Assets {
		release.Assets = append(release.Assets, ReleaseAsset{
			Name:         strings.TrimSpace(asset.Name),
			DownloadURL:  strings.TrimSpace(asset.BrowserDownloadURL),
			DownloadSize: asset.Size,
		})
	}

	if release.TagName == "" {
		return Release{}, errors.New("latest release did not include tag_name")
	}
	return release, nil
}

func SelectAssetForPlatform(assets []ReleaseAsset, goos, goarch string) (ReleaseAsset, error) {
	targetOS := strings.ToLower(strings.TrimSpace(goos))
	targetArch := strings.ToLower(strings.TrimSpace(goarch))
	if targetOS == "" || targetArch == "" {
		return ReleaseAsset{}, errors.New("goos and goarch are required")
	}

	candidates := make([]ReleaseAsset, 0)
	for _, asset := range assets {
		name := strings.ToLower(strings.TrimSpace(asset.Name))
		if name == "" || strings.TrimSpace(asset.DownloadURL) == "" {
			continue
		}
		if !strings.Contains(name, targetOS) || !strings.Contains(name, targetArch) {
			continue
		}
		if isArchiveOrChecksum(name) {
			continue
		}
		if targetOS == "windows" && !strings.HasSuffix(name, ".exe") {
			continue
		}
		if targetOS != "windows" && strings.HasSuffix(name, ".exe") {
			continue
		}
		candidates = append(candidates, asset)
	}

	if len(candidates) == 0 {
		return ReleaseAsset{}, fmt.Errorf("no release asset matched %s/%s", targetOS, targetArch)
	}
	if len(candidates) == 1 {
		return candidates[0], nil
	}

	// Prefer the shortest name to bias toward the canonical executable.
	best := candidates[0]
	bestLen := len(candidates[0].Name)
	for _, c := range candidates[1:] {
		n := len(c.Name)
		if n < bestLen {
			best = c
			bestLen = n
		}
	}
	return best, nil
}

func isArchiveOrChecksum(name string) bool {
	switch {
	case strings.HasSuffix(name, ".zip"),
		strings.HasSuffix(name, ".tar"),
		strings.HasSuffix(name, ".tar.gz"),
		strings.HasSuffix(name, ".tgz"),
		strings.HasSuffix(name, ".gz"),
		strings.HasSuffix(name, ".bz2"),
		strings.HasSuffix(name, ".xz"),
		strings.HasSuffix(name, ".sha256"),
		strings.HasSuffix(name, ".sha512"),
		strings.HasSuffix(name, ".txt"):
		return true
	}

	base := strings.ToLower(path.Base(name))
	return strings.Contains(base, "checksum") || strings.Contains(base, "sha256")
}
