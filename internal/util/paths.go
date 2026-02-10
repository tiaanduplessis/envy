package util

import (
	"os"
	"path/filepath"
)

// ConfigDir returns the base envy configuration directory.
// Uses ENVY_CONFIG_DIR if set, otherwise falls back to os.UserConfigDir()/envy.
func ConfigDir() (string, error) {
	if dir := os.Getenv("ENVY_CONFIG_DIR"); dir != "" {
		return dir, nil
	}
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "envy"), nil
}

// ProjectsDir returns the directory where project YAML files are stored.
func ProjectsDir() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "projects"), nil
}
