package state

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

type CurrentStore struct {
	path string
}

func New(baseDir, sessionID string) *CurrentStore {
	return &CurrentStore{
		path: filepath.Join(baseDir, "sessions", sessionID+".current"),
	}
}

func (c *CurrentStore) Get() (string, error) {
	data, err := os.ReadFile(c.path)
	if errors.Is(err, os.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func (c *CurrentStore) Set(name string) error {
	dir := filepath.Dir(c.path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	file, err := os.CreateTemp(dir, ".current.*.tmp")
	if err != nil {
		return err
	}
	tmp := file.Name()
	defer os.Remove(tmp)

	if _, err := file.WriteString(name); err != nil {
		_ = file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return os.Rename(tmp, c.path)
}

func (c *CurrentStore) Clear() error {
	err := os.Remove(c.path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}
