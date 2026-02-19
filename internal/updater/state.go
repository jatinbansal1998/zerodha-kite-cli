package updater

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const (
	updaterDirName = "updater"
	downloadsDir   = "downloads"
	stateFileName  = "state.json"
)

var versionPathSanitizer = regexp.MustCompile(`[^a-zA-Z0-9._-]`)

type State struct {
	CurrentVersion    string    `json:"current_version,omitempty"`
	LastCheckedAt     time.Time `json:"last_checked_at,omitempty"`
	LatestVersionSeen string    `json:"latest_version_seen,omitempty"`
	DownloadedVersion string    `json:"downloaded_version,omitempty"`
	DownloadedAsset   string    `json:"downloaded_asset_path,omitempty"`
	ApplyPending      bool      `json:"apply_pending,omitempty"`
	LastError         string    `json:"last_error,omitempty"`
	LastErrorAt       time.Time `json:"last_error_at,omitempty"`
	UpdatedAt         time.Time `json:"updated_at,omitempty"`
}

type StateStore struct {
	baseDir   string
	statePath string
}

func NewStateStore(cacheDir string) *StateStore {
	base := filepath.Join(cacheDir, updaterDirName)
	return &StateStore{
		baseDir:   base,
		statePath: filepath.Join(base, stateFileName),
	}
}

func (s *StateStore) StatePath() string {
	return s.statePath
}

func (s *StateStore) Load() (State, error) {
	data, err := os.ReadFile(s.statePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return State{}, nil
		}
		return State{}, fmt.Errorf("read updater state: %w", err)
	}
	if len(data) == 0 {
		return State{}, nil
	}

	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return State{}, fmt.Errorf("decode updater state: %w", err)
	}
	return state, nil
}

func (s *StateStore) Save(state State) error {
	if err := os.MkdirAll(s.baseDir, 0o700); err != nil {
		return fmt.Errorf("create updater directory: %w", err)
	}

	state.UpdatedAt = time.Now().UTC()

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("encode updater state: %w", err)
	}
	data = append(data, '\n')

	tmp, err := os.CreateTemp(s.baseDir, "state-*.tmp")
	if err != nil {
		return fmt.Errorf("create updater temp state: %w", err)
	}
	tmpPath := tmp.Name()

	cleanup := func() {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
	}

	if _, err := tmp.Write(data); err != nil {
		cleanup()
		return fmt.Errorf("write updater temp state: %w", err)
	}
	if err := tmp.Chmod(0o600); err != nil {
		cleanup()
		return fmt.Errorf("set updater temp permissions: %w", err)
	}
	if err := tmp.Close(); err != nil {
		cleanup()
		return fmt.Errorf("close updater temp state: %w", err)
	}
	if err := os.Rename(tmpPath, s.statePath); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("replace updater state: %w", err)
	}
	if err := os.Chmod(s.statePath, 0o600); err != nil {
		return fmt.Errorf("set updater state permissions: %w", err)
	}
	return nil
}

func (s *StateStore) StagingPath(version, assetName string) (string, error) {
	versionDir := sanitizeVersion(version)
	if versionDir == "" {
		versionDir = "unknown"
	}
	name := strings.TrimSpace(assetName)
	if name == "" {
		return "", errors.New("asset name is required")
	}

	dir := filepath.Join(s.baseDir, downloadsDir, versionDir)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", fmt.Errorf("create updater downloads directory: %w", err)
	}
	return filepath.Join(dir, filepath.Base(name)), nil
}

func sanitizeVersion(version string) string {
	v := strings.TrimSpace(version)
	v = strings.TrimPrefix(v, "v")
	v = strings.TrimPrefix(v, "V")
	return versionPathSanitizer.ReplaceAllString(v, "_")
}
