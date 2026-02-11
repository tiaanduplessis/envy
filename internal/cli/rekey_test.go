package cli

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/tiaanduplessis/envy/internal/config"
	"github.com/tiaanduplessis/envy/internal/crypto"
)

func TestRekeyCmd(t *testing.T) {
	store := setupEncryptedTestStore(t)
	t.Setenv(crypto.EnvPassphrase, "old-passphrase")

	p, _ := config.NewProject("myapp", []string{"dev"}, "dev")
	p.SetVar("dev", "SECRET", "hunter2")
	if err := store.Save(p); err != nil {
		t.Fatal(err)
	}

	cmd := NewRootCmd(store)
	if _, err := executeCommand(cmd, "encrypt", "myapp"); err != nil {
		t.Fatal(err)
	}

	// Use the internal constructor with injectable passphrase funcs
	rekeyCmd := newRekeyCmd(store,
		func(string) (string, error) { return "old-passphrase", nil },
		func() (string, error) { return "new-passphrase", nil },
	)
	out, err := executeCommand(rekeyCmd, "myapp")
	if err != nil {
		t.Fatalf("rekey: %v\n%s", err, out)
	}
	if !strings.Contains(out, "Re-encrypted") {
		t.Errorf("expected success message, got %q", out)
	}

	// Verify the new passphrase works
	raw, err := store.LoadRaw("myapp")
	if err != nil {
		t.Fatal(err)
	}
	salt, err := base64.StdEncoding.DecodeString(raw.Encryption.Salt)
	if err != nil {
		t.Fatal(err)
	}
	newKey := crypto.DeriveKey("new-passphrase", salt, raw.Encryption.Params)
	decrypted, err := crypto.DecryptMap(newKey, raw.Environments["dev"])
	if err != nil {
		t.Fatalf("decrypt with new key: %v", err)
	}
	if decrypted["SECRET"] != "hunter2" {
		t.Errorf("SECRET = %q, want %q", decrypted["SECRET"], "hunter2")
	}

	// Verify the old passphrase no longer works
	oldSalt, _ := base64.StdEncoding.DecodeString(raw.Encryption.Salt)
	oldKey := crypto.DeriveKey("old-passphrase", oldSalt, raw.Encryption.Params)
	_, err = crypto.DecryptMap(oldKey, raw.Environments["dev"])
	if err == nil {
		t.Fatal("expected old passphrase to fail")
	}
}

func TestRekeyCmd_NotEncrypted(t *testing.T) {
	store := setupEncryptedTestStore(t)

	p, _ := config.NewProject("plain", []string{"dev"}, "dev")
	if err := store.Save(p); err != nil {
		t.Fatal(err)
	}

	cmd := NewRootCmd(store)
	_, err := executeCommand(cmd, "rekey", "plain")
	if err == nil {
		t.Fatal("expected error rekeying non-encrypted project")
	}
	if !strings.Contains(err.Error(), "not encrypted") {
		t.Errorf("expected 'not encrypted' error, got: %v", err)
	}
}

func TestRekeyCmd_WrongOldPassphrase(t *testing.T) {
	store := setupEncryptedTestStore(t)
	t.Setenv(crypto.EnvPassphrase, "correct-passphrase")

	p, _ := config.NewProject("locked", []string{"dev"}, "dev")
	p.SetVar("dev", "KEY", "value")
	if err := store.Save(p); err != nil {
		t.Fatal(err)
	}

	cmd := NewRootCmd(store)
	if _, err := executeCommand(cmd, "encrypt", "locked"); err != nil {
		t.Fatal(err)
	}

	rekeyCmd := newRekeyCmd(store,
		func(string) (string, error) { return "wrong-passphrase", nil },
		func() (string, error) { return "new-passphrase", nil },
	)
	_, err := executeCommand(rekeyCmd, "locked")
	if err == nil {
		t.Fatal("expected error with wrong old passphrase")
	}
}
