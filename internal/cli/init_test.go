package cli

import (
	"strings"
	"testing"

	"github.com/tiaanduplessis/envy/internal/crypto"
)

func TestInitCmd_Basic(t *testing.T) {
	store := setupTestStore(t)
	cmd := NewRootCmd(store)

	out, err := executeCommand(cmd, "init", "myapp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Created project") {
		t.Errorf("output = %q, want it to contain 'Created project'", out)
	}

	p, err := store.Load("myapp")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if p.DefaultEnv != "dev" {
		t.Errorf("DefaultEnv = %q, want %q", p.DefaultEnv, "dev")
	}
	if _, ok := p.Environments["dev"]; !ok {
		t.Error("expected 'dev' environment")
	}
}

func TestInitCmd_WithEnvsAndPaths(t *testing.T) {
	store := setupTestStore(t)
	cmd := NewRootCmd(store)

	_, err := executeCommand(cmd, "init", "mono",
		"--env", "dev", "--env", "staging",
		"--path", "services/api",
		"--default-env", "staging")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	p, err := store.Load("mono")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if p.DefaultEnv != "staging" {
		t.Errorf("DefaultEnv = %q, want %q", p.DefaultEnv, "staging")
	}
	if _, ok := p.Environments["dev"]; !ok {
		t.Error("expected 'dev' environment")
	}
	if _, ok := p.Environments["staging"]; !ok {
		t.Error("expected 'staging' environment")
	}
	if _, ok := p.Paths["services/api"]; !ok {
		t.Error("expected 'services/api' path")
	}
}

func TestInitCmd_DuplicateProject(t *testing.T) {
	store := setupTestStore(t)
	cmd := NewRootCmd(store)

	_, err := executeCommand(cmd, "init", "dup")
	if err != nil {
		t.Fatalf("first init: %v", err)
	}

	cmd = NewRootCmd(store)
	_, err = executeCommand(cmd, "init", "dup")
	if err == nil {
		t.Error("expected error for duplicate project")
	}
}

func TestInitCmd_InvalidName(t *testing.T) {
	store := setupTestStore(t)
	cmd := NewRootCmd(store)

	_, err := executeCommand(cmd, "init", "../bad")
	if err == nil {
		t.Error("expected error for invalid name")
	}
}

func TestInitCmd_WithEncrypt(t *testing.T) {
	store := setupEncryptedTestStore(t)
	t.Setenv(crypto.EnvPassphrase, "test-passphrase")

	cmd := NewRootCmd(store)
	out, err := executeCommand(cmd, "init", "secure", "--encrypt", "--env", "dev")
	if err != nil {
		t.Fatalf("init --encrypt: %v\n%s", err, out)
	}
	if !strings.Contains(out, "Created project") {
		t.Errorf("expected success message, got %q", out)
	}

	raw, err := store.LoadRaw("secure")
	if err != nil {
		t.Fatal(err)
	}
	if !raw.IsEncrypted() {
		t.Fatal("expected project to be encrypted from creation")
	}

	loaded, err := store.Load("secure")
	if err != nil {
		t.Fatal(err)
	}
	if !loaded.IsEncrypted() {
		t.Fatal("loaded project should report as encrypted")
	}
}
