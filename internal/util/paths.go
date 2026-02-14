package util

import (
	"os"
	"path/filepath"
)

// ConfigDir returns the base envy configuration directory.
// Uses ENVY_CONFIG_DIR if set, otherwise falls back to ~/.config/envy.
func ConfigDir() (string, error) {
	if dir := os.Getenv("ENVY_CONFIG_DIR"); dir != "" {
		return dir, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "envy"), nil
}

func ProjectsDir() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "projects"), nil
}
