package config

import (
	"os"
	"path/filepath"
)

const appDirName = ".notebook-cli"

type Config struct {
	BaseDir string
	DBPath  string
}

func Load() (Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Config{}, err
	}

	baseDir := filepath.Join(home, appDirName)
	return Config{
		BaseDir: baseDir,
		DBPath:  filepath.Join(baseDir, "notebook.db"),
	}, nil
}
