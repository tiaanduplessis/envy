package cli

import (
	"strings"
	"testing"

	"github.com/tiaanduplessis/envy/internal/config"
	"github.com/tiaanduplessis/envy/internal/crypto"
)

func TestEncryptCmd(t *testing.T) {
	store := setupEncryptedTestStore(t)

	p, _ := config.NewProject("myapp", []string{"dev"}, "dev")
	p.SetVar("dev", "SECRET", "hunter2")
	if err := store.Save(p); err != nil {
		t.Fatal(err)
	}

	cmd := NewRootCmd(store)
	t.Setenv(crypto.EnvPassphrase, "test-passphrase")
	out, err := executeCommand(cmd, "encrypt", "myapp")
	if err != nil {
		t.Fatalf("encrypt: %v\n%s", err, out)
	}
	if !strings.Contains(out, "Encrypted") {
		t.Errorf("expected success message, got %q", out)
	}

	raw, err := store.LoadRaw("myapp")
	if err != nil {
		t.Fatal(err)
	}
	if !raw.IsEncrypted() {
		t.Fatal("expected project to be encrypted")
	}
	val := raw.Environments["dev"]["SECRET"]
	if !crypto.IsEncrypted(val) {
		t.Errorf("value on disk should be encrypted, got %q", val)
	}

	loaded, err := store.Load("myapp")
	if err != nil {
		t.Fatal(err)
	}
	if got := loaded.Environments["dev"]["SECRET"]; got != "hunter2" {
		t.Errorf("decrypted value = %q, want %q", got, "hunter2")
	}
}

func TestEncryptCmd_AlreadyEncrypted(t *testing.T) {
	store := setupEncryptedTestStore(t)

	p, _ := config.NewProject("myapp", []string{"dev"}, "dev")
	p.SetVar("dev", "KEY", "value")
	if err := store.Save(p); err != nil {
		t.Fatal(err)
	}

	cmd := NewRootCmd(store)
	t.Setenv(crypto.EnvPassphrase, "test-passphrase")
	_, err := executeCommand(cmd, "encrypt", "myapp")
	if err != nil {
		t.Fatal(err)
	}

	cmd = NewRootCmd(store)
	_, err = executeCommand(cmd, "encrypt", "myapp")
	if err == nil {
		t.Fatal("expected error encrypting already-encrypted project")
	}
	if !strings.Contains(err.Error(), "already encrypted") {
		t.Errorf("expected 'already encrypted' error, got: %v", err)
	}
}

func TestEncryptCmd_NonexistentProject(t *testing.T) {
	store := setupEncryptedTestStore(t)
	cmd := NewRootCmd(store)
	t.Setenv(crypto.EnvPassphrase, "test-passphrase")
	_, err := executeCommand(cmd, "encrypt", "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent project")
	}
}

func TestEncryptCmd_WithPaths(t *testing.T) {
	store := setupEncryptedTestStore(t)

	p, _ := config.NewProject("mono", []string{"dev"}, "dev")
	p.SetVar("dev", "SHARED", "root-value")
	p.SetPathVar("services/api", "dev", "PORT", "3000")
	if err := store.Save(p); err != nil {
		t.Fatal(err)
	}

	cmd := NewRootCmd(store)
	t.Setenv(crypto.EnvPassphrase, "test-passphrase")
	_, err := executeCommand(cmd, "encrypt", "mono")
	if err != nil {
		t.Fatal(err)
	}

	raw, err := store.LoadRaw("mono")
	if err != nil {
		t.Fatal(err)
	}
	if !crypto.IsEncrypted(raw.Environments["dev"]["SHARED"]) {
		t.Error("root var should be encrypted")
	}
	if !crypto.IsEncrypted(raw.Paths["services/api"]["dev"]["PORT"]) {
		t.Error("path var should be encrypted")
	}

	loaded, err := store.Load("mono")
	if err != nil {
		t.Fatal(err)
	}
	if got := loaded.Paths["services/api"]["dev"]["PORT"]; got != "3000" {
		t.Errorf("PORT = %q, want %q", got, "3000")
	}
}
