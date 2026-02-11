package cli

import (
	"strings"
	"testing"

	"github.com/tiaanduplessis/envy/internal/config"
	"github.com/tiaanduplessis/envy/internal/crypto"
)

func TestDecryptCmd(t *testing.T) {
	store := setupEncryptedTestStore(t)
	t.Setenv(crypto.EnvPassphrase, "test-passphrase")

	p, _ := config.NewProject("myapp", []string{"dev"}, "dev")
	p.SetVar("dev", "SECRET", "hunter2")
	if err := store.Save(p); err != nil {
		t.Fatal(err)
	}

	cmd := NewRootCmd(store)
	if _, err := executeCommand(cmd, "encrypt", "myapp"); err != nil {
		t.Fatal(err)
	}

	cmd = NewRootCmd(store)
	out, err := executeCommand(cmd, "decrypt", "myapp")
	if err != nil {
		t.Fatalf("decrypt: %v\n%s", err, out)
	}
	if !strings.Contains(out, "Decrypted") {
		t.Errorf("expected success message, got %q", out)
	}

	raw, err := store.LoadRaw("myapp")
	if err != nil {
		t.Fatal(err)
	}
	if raw.IsEncrypted() {
		t.Fatal("expected project to no longer be encrypted")
	}
	val := raw.Environments["dev"]["SECRET"]
	if val != "hunter2" {
		t.Errorf("value on disk should be plaintext, got %q", val)
	}
}

func TestDecryptCmd_NotEncrypted(t *testing.T) {
	store := setupEncryptedTestStore(t)

	p, _ := config.NewProject("plain", []string{"dev"}, "dev")
	if err := store.Save(p); err != nil {
		t.Fatal(err)
	}

	cmd := NewRootCmd(store)
	_, err := executeCommand(cmd, "decrypt", "plain")
	if err == nil {
		t.Fatal("expected error decrypting non-encrypted project")
	}
	if !strings.Contains(err.Error(), "not encrypted") {
		t.Errorf("expected 'not encrypted' error, got: %v", err)
	}
}

func TestDecryptCmd_WrongPassphrase(t *testing.T) {
	store := setupEncryptedTestStore(t)
	t.Setenv(crypto.EnvPassphrase, "test-passphrase")

	p, _ := config.NewProject("locked", []string{"dev"}, "dev")
	p.SetVar("dev", "KEY", "value")
	if err := store.Save(p); err != nil {
		t.Fatal(err)
	}

	cmd := NewRootCmd(store)
	if _, err := executeCommand(cmd, "encrypt", "locked"); err != nil {
		t.Fatal(err)
	}

	t.Setenv(crypto.EnvPassphrase, "wrong-passphrase")
	cmd = NewRootCmd(store)
	_, err := executeCommand(cmd, "decrypt", "locked")
	if err == nil {
		t.Fatal("expected error with wrong passphrase")
	}
}
