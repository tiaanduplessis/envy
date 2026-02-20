package cli

import (
	"strings"
	"testing"

	"github.com/tiaanduplessis/envy/internal/config"
	"github.com/tiaanduplessis/envy/internal/crypto"
)

func TestGetCmd_ExistingKey(t *testing.T) {
	store := setupTestStore(t)
	p, _ := config.NewProject("foo", []string{"dev"}, "dev")
	p.SetVar("dev", "DB", "localhost")
	store.Save(p)

	root := NewRootCmd(store)
	out, err := executeCommand(root, "get", "foo", "DB")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := strings.TrimSpace(out); got != "localhost" {
		t.Errorf("output = %q, want %q", got, "localhost")
	}
}

func TestGetCmd_MissingKey(t *testing.T) {
	store := setupTestStore(t)
	p, _ := config.NewProject("foo", []string{"dev"}, "dev")
	store.Save(p)

	root := NewRootCmd(store)
	_, err := executeCommand(root, "get", "foo", "MISSING")
	if err == nil {
		t.Error("expected error for missing key")
	}
}

func TestGetCmd_WithPathInheritance(t *testing.T) {
	store := setupTestStore(t)
	p, _ := config.NewProject("mono", []string{"dev"}, "dev")
	p.SetVar("dev", "DB", "localhost")
	p.SetVar("dev", "SHARED", "yes")
	p.SetPathVar("services/api", "dev", "DB", "api-db")
	p.SetPathVar("services/api", "dev", "PORT", "3000")
	store.Save(p)

	root := NewRootCmd(store)
	out, err := executeCommand(root, "get", "mono", "DB", "--path", "services/api")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := strings.TrimSpace(out); got != "api-db" {
		t.Errorf("DB = %q, want %q (should be overridden)", got, "api-db")
	}

	root = NewRootCmd(store)
	out, err = executeCommand(root, "get", "mono", "SHARED", "--path", "services/api")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := strings.TrimSpace(out); got != "yes" {
		t.Errorf("SHARED = %q, want %q (should be inherited)", got, "yes")
	}
}

func TestGetCmd_NonexistentProject(t *testing.T) {
	store := setupTestStore(t)
	root := NewRootCmd(store)
	_, err := executeCommand(root, "get", "nope", "KEY")
	if err == nil {
		t.Error("expected error for nonexistent project")
	}
}

func TestGetCmd_Encrypted(t *testing.T) {
	store := setupEncryptedTestStore(t)
	t.Setenv(crypto.EnvPassphrase, "test-passphrase")

	p, _ := config.NewProject("foo", []string{"dev"}, "dev")
	p.SetVar("dev", "SECRET", "hunter2")
	store.Save(p)

	cmd := NewRootCmd(store)
	if _, err := executeCommand(cmd, "encrypt", "foo"); err != nil {
		t.Fatal(err)
	}

	cmd = NewRootCmd(store)
	out, err := executeCommand(cmd, "get", "foo", "SECRET")
	if err != nil {
		t.Fatalf("get on encrypted project: %v", err)
	}
	if got := strings.TrimSpace(out); got != "hunter2" {
		t.Errorf("output = %q, want %q", got, "hunter2")
	}
}
