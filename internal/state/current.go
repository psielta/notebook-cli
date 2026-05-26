package state

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type CurrentStore struct {
	path string
}

func New(dir string) *CurrentStore {
	return &CurrentStore{path: filepath.Join(dir, ".current")}
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
	if err := os.MkdirAll(filepath.Dir(c.path), 0o755); err != nil {
		return err
	}

	tmp := fmt.Sprintf("%s.%d.tmp", c.path, os.Getpid())
	if err := os.WriteFile(tmp, []byte(name), 0o644); err != nil {
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
