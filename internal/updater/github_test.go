package updater

import "testing"

func TestSelectAssetForPlatform(t *testing.T) {
	assets := []ReleaseAsset{
		{Name: "checksums.txt", DownloadURL: "https://example.com/checksums.txt"},
		{Name: "zerodha_darwin_arm64", DownloadURL: "https://example.com/darwin-arm64"},
		{Name: "zerodha_linux_amd64.tar.gz", DownloadURL: "https://example.com/linux-amd64.tar.gz"},
		{Name: "zerodha_windows_amd64.exe", DownloadURL: "https://example.com/windows-amd64.exe"},
	}

	asset, err := SelectAssetForPlatform(assets, "darwin", "arm64")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if asset.Name != "zerodha_darwin_arm64" {
		t.Fatalf("expected darwin arm64 raw binary, got %q", asset.Name)
	}

	asset, err = SelectAssetForPlatform(assets, "windows", "amd64")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if asset.Name != "zerodha_windows_amd64.exe" {
		t.Fatalf("expected windows exe asset, got %q", asset.Name)
	}
}

func TestSelectAssetForPlatformNoMatch(t *testing.T) {
	assets := []ReleaseAsset{
		{Name: "zerodha_linux_amd64.tar.gz", DownloadURL: "https://example.com/linux-amd64.tar.gz"},
	}

	if _, err := SelectAssetForPlatform(assets, "linux", "amd64"); err == nil {
		t.Fatalf("expected no-match error")
	}
}
