package cache

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
)

type FSStore struct {
	baseDir string
}

func NewFSStore(baseDir string) *FSStore {
	return &FSStore{baseDir: baseDir}
}

func (s *FSStore) BaseDir() string {
	return s.baseDir
}

func (s *FSStore) Put(key string, data []byte) error {
	if err := s.ensureDir(); err != nil {
		return err
	}
	path := s.keyPath(key)
	return os.WriteFile(path, data, 0o600)
}

func (s *FSStore) Get(key string) ([]byte, error) {
	path := s.keyPath(key)
	return os.ReadFile(path)
}

func (s *FSStore) Delete(key string) error {
	path := s.keyPath(key)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (s *FSStore) ensureDir() error {
	if err := os.MkdirAll(s.baseDir, 0o700); err != nil {
		return fmt.Errorf("create cache directory: %w", err)
	}
	return nil
}

func (s *FSStore) keyPath(key string) string {
	sum := sha1.Sum([]byte(key))
	return filepath.Join(s.baseDir, hex.EncodeToString(sum[:])+".cache")
}
