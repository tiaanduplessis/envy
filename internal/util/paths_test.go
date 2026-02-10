package util

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigDir_EnvOverride(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("ENVY_CONFIG_DIR", dir)

	got, err := ConfigDir()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != dir {
		t.Errorf("ConfigDir() = %q, want %q", got, dir)
	}
}

func TestConfigDir_Default(t *testing.T) {
	t.Setenv("ENVY_CONFIG_DIR", "")

	got, err := ConfigDir()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	base, _ := os.UserConfigDir()
	want := filepath.Join(base, "envy")
	if got != want {
		t.Errorf("ConfigDir() = %q, want %q", got, want)
	}
}

func TestProjectsDir(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("ENVY_CONFIG_DIR", dir)

	got, err := ProjectsDir()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := filepath.Join(dir, "projects")
	if got != want {
		t.Errorf("ProjectsDir() = %q, want %q", got, want)
	}
}
