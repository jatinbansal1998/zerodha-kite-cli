//go:build !windows

package updater

import (
	"fmt"
	"io"
	"os"
)

func applyDownloadedBinary(targetPath, sourcePath string) error {
	backupPath := targetPath + ".old"
	_ = os.Remove(backupPath)

	if err := os.Rename(targetPath, backupPath); err != nil {
		return fmt.Errorf("move current binary to backup: %w", err)
	}
	if err := os.Rename(sourcePath, targetPath); err != nil {
		if copyErr := copyFile(sourcePath, targetPath, 0o755); copyErr != nil {
			_ = os.Rename(backupPath, targetPath)
			return fmt.Errorf("move downloaded binary into place: %w", err)
		}
		_ = os.Remove(sourcePath)
	}
	if err := os.Chmod(targetPath, 0o755); err != nil {
		return fmt.Errorf("set executable permissions on updated binary: %w", err)
	}
	_ = os.Remove(backupPath)
	return nil
}

func copyFile(sourcePath, destinationPath string, mode os.FileMode) error {
	source, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.OpenFile(destinationPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer destination.Close()

	if _, err := io.Copy(destination, source); err != nil {
		return err
	}
	return destination.Sync()
}
