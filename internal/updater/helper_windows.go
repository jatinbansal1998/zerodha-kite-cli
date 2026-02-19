//go:build windows

package updater

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

func SpawnApplyHelper(executablePath string, req ApplyHelperRequest) error {
	exe := strings.TrimSpace(executablePath)
	if exe == "" {
		return errors.New("executable path is required")
	}
	if strings.TrimSpace(req.TargetPath) == "" || strings.TrimSpace(req.SourcePath) == "" || strings.TrimSpace(req.CacheDir) == "" {
		return errors.New("helper apply request is incomplete")
	}

	cmd := exec.Command(
		exe,
		helperCommandName,
		"--target",
		req.TargetPath,
		"--source",
		req.SourcePath,
		"--cache-dir",
		req.CacheDir,
		"--version",
		req.Version,
	)
	cmd.Env = append(os.Environ(), HelperEnvVar+"=1")
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start windows apply helper: %w", err)
	}
	return nil
}

func RunApplyHelper(req ApplyHelperRequest) error {
	if strings.TrimSpace(req.TargetPath) == "" || strings.TrimSpace(req.SourcePath) == "" || strings.TrimSpace(req.CacheDir) == "" {
		return errors.New("helper apply request is incomplete")
	}

	backupPath := req.TargetPath + ".old"
	var lastErr error

	for i := 0; i < 300; i++ {
		_ = os.Remove(backupPath)
		if err := os.Rename(req.TargetPath, backupPath); err != nil {
			lastErr = err
			time.Sleep(100 * time.Millisecond)
			continue
		}

		if err := os.Rename(req.SourcePath, req.TargetPath); err != nil {
			lastErr = err
			_ = os.Rename(backupPath, req.TargetPath)
			time.Sleep(100 * time.Millisecond)
			continue
		}

		_ = os.Chmod(req.TargetPath, 0o755)
		_ = os.Remove(backupPath)
		markHelperApplySuccess(req.CacheDir, req.Version)
		return nil
	}

	if lastErr == nil {
		lastErr = errors.New("failed to replace executable")
	}
	markHelperApplyFailure(req.CacheDir, lastErr)
	return fmt.Errorf("apply staged update on windows: %w", lastErr)
}

func markHelperApplySuccess(cacheDir, version string) {
	store := NewStateStore(cacheDir)
	state, err := store.Load()
	if err != nil {
		return
	}
	state.CurrentVersion = version
	state.ApplyPending = false
	state.DownloadedVersion = ""
	state.DownloadedAsset = ""
	clearStateError(&state)
	_ = store.Save(state)
}

func markHelperApplyFailure(cacheDir string, err error) {
	store := NewStateStore(cacheDir)
	state, loadErr := store.Load()
	if loadErr != nil {
		return
	}
	setStateError(&state, err)
	_ = store.Save(state)
}
