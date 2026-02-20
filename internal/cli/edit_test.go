package cli

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/tiaanduplessis/envy/internal/config"
	"github.com/tiaanduplessis/envy/internal/crypto"
)

func TestEditCmd_ProjectNotFound(t *testing.T) {
	store := setupTestStore(t)
	cmd := newEditCmd(store, func(path string) error { return nil })
	_, err := executeCommand(cmd, "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent project")
	}
	if want := `project "nonexistent" not found`; err.Error() != want {
		t.Errorf("got %q, want %q", err.Error(), want)
	}
}

func TestEditCmd_EditorCalledWithCorrectPath(t *testing.T) {
	store := setupTestStore(t)
	p := &config.Project{Name: "my-app", DefaultEnv: "dev"}
	if err := store.Save(p); err != nil {
		t.Fatal(err)
	}

	var calledWith string
	cmd := newEditCmd(store, func(path string) error {
		calledWith = path
		return nil
	})

	_, err := executeCommand(cmd, "my-app")
	if err != nil {
		t.Fatal(err)
	}

	want := store.ProjectPath("my-app")
	if calledWith != want {
		t.Errorf("editor called with %q, want %q", calledWith, want)
	}
}

func TestEditCmd_EditorError(t *testing.T) {
	store := setupTestStore(t)
	p := &config.Project{Name: "my-app", DefaultEnv: "dev"}
	if err := store.Save(p); err != nil {
		t.Fatal(err)
	}

	editorErr := fmt.Errorf("editor crashed")
	cmd := newEditCmd(store, func(path string) error {
		return editorErr
	})

	_, err := executeCommand(cmd, "my-app")
	if err != editorErr {
		t.Errorf("got %v, want %v", err, editorErr)
	}
}

func TestEditCmd_EncryptedWarning(t *testing.T) {
	store := setupEncryptedTestStore(t)
	t.Setenv(crypto.EnvPassphrase, "test-passphrase")

	p, _ := config.NewProject("secure", []string{"dev"}, "dev")
	p.SetVar("dev", "KEY", "value")
	store.Save(p)

	root := NewRootCmd(store)
	if _, err := executeCommand(root, "encrypt", "secure"); err != nil {
		t.Fatal(err)
	}

	var stderr bytes.Buffer
	cmd := newEditCmd(store, func(path string) error { return nil })
	cmd.SetErr(&stderr)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetArgs([]string{"secure"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(stderr.String(), "Warning") || !strings.Contains(stderr.String(), "encrypted") {
		t.Errorf("expected encryption warning on stderr, got: %q", stderr.String())
	}
}

func TestEditCmd_NoWarningForPlaintext(t *testing.T) {
	store := setupTestStore(t)
	p, _ := config.NewProject("plain", []string{"dev"}, "dev")
	store.Save(p)

	var stderr bytes.Buffer
	cmd := newEditCmd(store, func(path string) error { return nil })
	cmd.SetErr(&stderr)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetArgs([]string{"plain"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	if strings.Contains(stderr.String(), "Warning") {
		t.Errorf("should not warn for plaintext project, got: %q", stderr.String())
	}
}
