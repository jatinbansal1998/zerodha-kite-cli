//go:build windows

package updater

func applyDownloadedBinary(_ string, _ string) error {
	// Windows updates are applied by the helper process.
	return nil
}
