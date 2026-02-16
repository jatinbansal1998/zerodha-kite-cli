package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

type FileStore struct {
	path string
}

func NewFileStore(path string) *FileStore {
	return &FileStore{path: path}
}

func (s *FileStore) Path() string {
	return s.path
}

func (s *FileStore) Load() (Config, error) {
	if err := s.ensureFile(); err != nil {
		return Config{}, err
	}

	content, err := os.ReadFile(s.path)
	if err != nil {
		return Config{}, fmt.Errorf("read config file: %w", err)
	}

	if len(content) == 0 {
		cfg := Default()
		if err := s.Save(cfg); err != nil {
			return Config{}, err
		}
		return cfg, nil
	}

	cfg := Default()
	if err := json.Unmarshal(content, &cfg); err != nil {
		return Config{}, fmt.Errorf("decode config file: %w", err)
	}

	if cfg.Version == 0 {
		cfg.Version = CurrentVersion
	}
	if cfg.Profiles == nil {
		cfg.Profiles = make(map[string]Profile)
	}

	return cfg, nil
}

func (s *FileStore) Save(cfg Config) error {
	if err := s.ensureDir(); err != nil {
		return err
	}

	cfg.Version = CurrentVersion
	if cfg.Profiles == nil {
		cfg.Profiles = make(map[string]Profile)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("encode config file: %w", err)
	}
	data = append(data, '\n')

	dir := filepath.Dir(s.path)
	tmpFile, err := os.CreateTemp(dir, "config-*.tmp")
	if err != nil {
		return fmt.Errorf("create temp config file: %w", err)
	}
	tmpPath := tmpFile.Name()

	cleanup := func() {
		_ = tmpFile.Close()
		_ = os.Remove(tmpPath)
	}

	if _, err := tmpFile.Write(data); err != nil {
		cleanup()
		return fmt.Errorf("write temp config file: %w", err)
	}
	if err := tmpFile.Chmod(0o600); err != nil {
		cleanup()
		if !isPermissionErr(err) {
			return fmt.Errorf("set temp config permissions: %w", err)
		}
	}
	if err := tmpFile.Close(); err != nil {
		cleanup()
		return fmt.Errorf("close temp config file: %w", err)
	}
	if err := os.Rename(tmpPath, s.path); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("replace config file: %w", err)
	}
	if err := os.Chmod(s.path, 0o600); err != nil {
		if !isPermissionErr(err) {
			return fmt.Errorf("set config permissions: %w", err)
		}
	}

	return nil
}

func (s *FileStore) ensureFile() error {
	if err := s.ensureDir(); err != nil {
		return err
	}

	_, err := os.Stat(s.path)
	if err == nil {
		return nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("stat config file: %w", err)
	}

	if err := os.WriteFile(s.path, []byte{}, 0o600); err != nil {
		return fmt.Errorf("create config file: %w", err)
	}
	if err := os.Chmod(s.path, 0o600); err != nil {
		if !isPermissionErr(err) {
			return fmt.Errorf("set config permissions: %w", err)
		}
	}

	return nil
}

func (s *FileStore) ensureDir() error {
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}
	if err := os.Chmod(dir, 0o700); err != nil {
		if !isPermissionErr(err) {
			return fmt.Errorf("set config directory permissions: %w", err)
		}
	}
	return nil
}

func isPermissionErr(err error) bool {
	if errors.Is(err, os.ErrPermission) || errors.Is(err, syscall.EPERM) {
		return true
	}
	return false
}
